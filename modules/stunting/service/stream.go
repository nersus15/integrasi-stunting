package service

import (
	"fmt"
	"strings"

	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/satusehat"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/repository"
	backgroundworker "github.com/nersus15/lib-background-worker"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"
)

type StreamService struct {
	Context    *core.AppContext
	Config     *config.ModuleConfig
	repository *repository.StuntingRepository
	stunting   *StuntingService
	background *backgroundworker.BackgroundWorker
}

func NewStreamService(ctx *core.AppContext, stunting *StuntingService, cfg *config.ModuleConfig, repo *repository.StuntingRepository, bworker *backgroundworker.BackgroundWorker) *StreamService {
	return &StreamService{
		Context:    ctx,
		Config:     cfg,
		repository: repo,
		stunting:   stunting,
		background: bworker,
	}
}

func (s *StreamService) ProsesTransaksiFHIR(bundle *types2.Bundle, ihsAnak *string) error {
	if utils.DumpAktif() {
		logger.Debug("Bundle: " + helper.ToLogJSON(bundle))
	}

	// ihs pasien dari envelope PackMediator.patient, cadangannya reference di resource
	isi := &isiBundle{idAnakSatusehat: ihsAnak}
	satusehat.UrutkanEntry(bundle.Entry)

	for _, entry := range bundle.Entry {
		if entry.Base == nil || !utils.IsFilled(entry.Base.ResourceType) {
			continue
		}
		if utils.DumpAktif() {
			logger.Debug("Entry: " + helper.ToLogJSON(entry.Resource))
		}

		// PUT selalu satu resource per pesan, satusehat tidak punya PUT untuk Bundle
		update := entry.Request != nil && entry.Request.Method == fhir.HTTPVerbPUT

		if selesai, err := s.bacaEntry(entry, update, isi); selesai {
			return err
		}
	}

	return s.simpanIsiBundle(isi)
}

// isi satu bundle yang dikumpulkan sebelum disimpan
type isiBundle struct {
	anak                *types.Anak
	orangtua            *types.Orangtua
	orangtuaPerluUpdate bool
	idAnakSatusehat     *string
	anakKe              *int

	kunjungan *types.Kunjungan
	refFaskes *string
	observasi []satusehat.HasilObservasi
	diagnosa  []types.Diagnosa
	layanan   []types.Layanan
	rujukan   []types.Rujukan
	episode   []types.Episode
}

func (isi *isiBundle) catatPasien(ihs *string) {
	if utils.IsFilled(ihs) {
		isi.idAnakSatusehat = ihs
	}
}

func (isi *isiBundle) adaDataMedis() bool {
	return isi.kunjungan != nil || len(isi.observasi) > 0 || len(isi.diagnosa) > 0 || len(isi.layanan) > 0 || len(isi.rujukan) > 0 || len(isi.episode) > 0
}

// selesai: entry update, sisa bundle tidak diproses
func (s *StreamService) bacaEntry(entry types2.BundleEntry, update bool, isi *isiBundle) (selesai bool, err error) {
	switch strings.ToLower(*entry.Base.ResourceType) {
	case "patient":
		s.bacaPatient(entry, update, isi)

	case "relatedperson":
		s.bacaRelatedPerson(entry, update, isi)

	case "encounter":
		return bacaKlinis(isi, entry, update, "EncounterToKunjungan", satusehat.EncounterToKunjungan,
			func(k *types.Kunjungan, ihsFaskes *string) { isi.kunjungan, isi.refFaskes = k, ihsFaskes },
			func(k *types.Kunjungan) error {
				_, err := s.repository.UpdateKunjunganBySatusehatId(k.ToPayload().ToEntity(), true)
				return err
			})

	case "observation":
		return s.bacaObservation(entry, update, isi)

	case "condition":
		return bacaKlinis(isi, entry, update, "ConditionToDiagnosa", satusehat.ConditionToDiagnosa, isi.tampungDiagnosa, s.perbaruiDiagnosa)

	case "allergyintolerance":
		return bacaKlinis(isi, entry, update, "AllergyIntoleranceToDiagnosa", satusehat.AllergyIntoleranceToDiagnosa, isi.tampungDiagnosa, s.perbaruiDiagnosa)

	case "procedure", "medicationdispense", "nutritionorder", "immunization":
		return bacaKlinis(isi, entry, update, "petakanLayanan", petakanLayanan,
			func(l *types.Layanan, refEnc *string) { l.RefEncounter = refEnc; isi.layanan = append(isi.layanan, *l) },
			func(l *types.Layanan) error {
				_, err := s.repository.UpdateLayananBySatusehatId(l.ToEntity(), true)
				return err
			})

	case "servicerequest":
		return bacaKlinis(isi, entry, update, "ServiceRequestToRujukan", satusehat.ServiceRequestToRujukan,
			func(r *types.Rujukan, refEnc *string) { r.RefEncounter = refEnc; isi.rujukan = append(isi.rujukan, *r) },
			func(r *types.Rujukan) error {
				if err := s.arahRujukanTersimpan(r); err != nil {
					return err
				}
				_, err := s.repository.UpdateRujukanBySatusehatId(r.ToEntity(), true)
				return err
			})

	case "episodeofcare":
		return bacaKlinis(isi, entry, update, "EpisodeOfCareToEpisode", satusehat.EpisodeOfCareToEpisode,
			func(e *types.Episode, ihsFaskes *string) {
				e.RefFaskes = ihsFaskes
				isi.episode = append(isi.episode, *e)
			},
			func(e *types.Episode) error {
				_, err := s.repository.UpdateEpisodeBySatusehatId(e.ToEntity(), true)
				return err
			})

	case "questionnaireresponse":
		// ditunda: linkId string bebas, bukan kode terminologi
	case "composition":
		// Composition ditunda dulu
	}
	return false, nil
}

// ref: reference encounter, atau faskes untuk Encounter/EpisodeOfCare
func bacaKlinis[T any](isi *isiBundle, entry types2.BundleEntry, update bool, label string,
	terjemah func(types2.BundleEntry) (*T, *string, *string, error),
	tampung func(t *T, ref *string),
	perbarui func(t *T) error) (bool, error) {
	tmp, ihs, ref, err := terjemah(entry)
	if err != nil {
		logger.Error(label, err)
		return false, nil
	}
	isi.catatPasien(ihs)
	tampung(tmp, ref)

	if !update {
		return false, nil
	}
	return true, exceptions.AbaikanJikaBelumTersimpan(perbarui(tmp))
}

func (isi *isiBundle) tampungDiagnosa(d *types.Diagnosa, refEnc *string) {
	d.RefEncounter = refEnc
	isi.diagnosa = append(isi.diagnosa, *d)
}

func (s *StreamService) perbaruiDiagnosa(d *types.Diagnosa) error {
	_, err := s.repository.UpdateDiagnosaBySatusehatId(d.ToEntity(), true)
	return err
}

func (s *StreamService) bacaPatient(entry types2.BundleEntry, update bool, isi *isiBundle) {
	tmp, err := satusehat.PatientToAnak(entry)
	if err != nil {
		logger.Error("PatientToAnak", err)
		return
	}

	lama, _ := s.repository.FindAnak(nil, tmp.Nik, nil, nil, false)
	isi.anak = anakDTO(lama, nil)
	if isi.anak != nil {
		if timpaAnakDariFHIR(isi.anak, tmp) && isi.anak.IdSatusehat != nil {
			lama.Override(*isi.anak.ToPayload().ToEntity(), false)
			lama.UpdatedBy = new(pembuatSatusehat)
			if _, err := s.repository.UpdateAnakById(lama, nil, false); err != nil {
				logger.Error("ProsesTransaksiFHIR:UpdateDataAnak => " + err.Error())
			}
		}
		return
	}

	if update {
		if utils.DumpAktif() {
			logger.Debug("ProsesTransaksiFHIR:Patient:Update:Anak Tidak Ditemukan => continue")
		}
		return
	}
	isi.anak = &types.Anak{
		Nik:          tmp.Nik,
		IdSatusehat:  tmp.IdSatusehat,
		Nama:         tmp.Nama,
		TanggalLahir: tmp.TanggalLahir,
		JenisKelamin: tmp.JenisKelamin,
	}
}

func (s *StreamService) bacaRelatedPerson(entry types2.BundleEntry, update bool, isi *isiBundle) {
	tmp, ihs, err := satusehat.RelatedPersonToOrangtua(entry)
	if err != nil {
		logger.Error("RelatedPersonToOrangtua", err)
		return
	}

	// nik satu-satunya kunci pencocokan
	if !utils.IsStrFilled(tmp.Nik) {
		logger.Error("ProsesTransaksiFHIR:RelatedPerson => nik kosong, entry dilewati")
		return
	}
	isi.catatPasien(ihs)

	isi.orangtua = orangtuaDTO(s.repository.FindOrangTuaByNik(tmp.Nik, false))
	if isi.orangtua != nil {
		isi.orangtuaPerluUpdate = timpaDariFHIR(isi.orangtua, tmp)
		return
	}

	if update {
		if utils.DumpAktif() {
			logger.Debug("ProsesTransaksiFHIR:Patient:Update:Orangtua Tidak Ditemukan => continue")
		}
		return
	}
	isi.orangtua = tmp
}

func (s *StreamService) bacaObservation(entry types2.BundleEntry, update bool, isi *isiBundle) (bool, error) {
	tmp, resource, ihs, refEnc, err := satusehat.ObservationToObservasi(entry)
	if err != nil {
		logger.Error("ObservationToObservasi", err)
		return false, nil
	}
	isi.catatPasien(ihs)

	if resource != nil {
		if u := satusehat.FindAnakKe(resource); u != nil {
			isi.anakKe = u
		}
	}
	tmp.Induk.RefEncounter = refEnc
	isi.observasi = append(isi.observasi, *tmp)

	if !update {
		return false, nil
	}
	_, err = s.repository.UpdateObservasiBySatusehatId(tmp.Induk.ToEntity(), true)
	if err == nil {
		err = s.stunting.perbaruiUkurKunjungan(utils.Nilai(tmp.Induk.IdSatusehat))
	}
	return true, exceptions.AbaikanJikaBelumTersimpan(err)
}

func (s *StreamService) simpanIsiBundle(isi *isiBundle) error {
	if isi.anak != nil && !utils.IsStrFilled(isi.anak.Id) {
		a, err := s.PastikanAnak(isi.anak, isi.anakKe)
		if err != nil {
			return err
		}
		if a == nil {
			return nil
		}
		isi.anak = a
	}

	if isi.anakKe != nil && utils.IsFilled(isi.idAnakSatusehat) {
		s.perbaruiAnakKe(*isi.idAnakSatusehat, *isi.anakKe)
	}

	if isi.adaDataMedis() {
		_, _, alasan, err := s.simpanDataMedis(isi.anak, isi.idAnakSatusehat, isi.kunjungan, isi.refFaskes, isi.observasi, isi.diagnosa, isi.layanan, isi.rujukan, isi.episode, nil, asalStream(utils.StrPtr(isi.refFaskes)))
		if err != nil {
			return err
		}
		if alasan != "" {
			logger.Info("ProsesTransaksiFHIR: data medis tidak disimpan => " + alasan)
		}
	}

	if isi.orangtua != nil {
		s.simpanOrangtua(isi.orangtua, isi.orangtuaPerluUpdate, isi.idAnakSatusehat)
	}
	return nil
}

func (s *StreamService) perbaruiAnakKe(idAnakSatusehat string, anakKe int) {
	anak, err := s.repository.FindAnakBySatusehatId(idAnakSatusehat, false)
	if err != nil {
		logger.Error("ProsesTransaksiFHIR:UpdateDataAnak => " + err.Error())
	}
	if anak == nil {
		return
	}

	anak.AnakKe = int16(anakKe)
	if _, err := s.repository.UpdateAnakById(anak, nil, true); err != nil {
		logger.Error("ProsesTransaksiFHIR:UpdateDataAnak => " + err.Error())
	}
}

func petakanLayanan(entry types2.BundleEntry) (*types.Layanan, *string, *string, error) {
	switch strings.ToLower(*entry.Base.ResourceType) {
	case "procedure":
		return satusehat.ProcedureToLayanan(entry)
	case "medicationdispense":
		return satusehat.MedicationDispenseToLayanan(entry)
	case "nutritionorder":
		return satusehat.NutritionOrderToLayanan(entry)
	case "immunization":
		return satusehat.ImmunizationToLayanan(entry)
	}
	return nil, nil, nil, fmt.Errorf("resource %s bukan layanan", *entry.Base.ResourceType)
}
