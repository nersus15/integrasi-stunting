package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/entity"
	exceptions "github.com/nersus15/integrasi/mod-stunting/helper/Exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/repository"
	backgroundworker "github.com/nersus15/lib-background-worker"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
	"github.com/semanggilab/lib-go-fhir/processor"
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

func (s *StreamService) RetryKafkaTransaction(key string) error {
	tr, err := s.repository.FindKafkaTransaction(key)
	if err != nil {
		return err
	}

	err = s.prosesUlang(tr)
	if err != nil {
		// simpan error terbaru dan naikkan attempt
		if e := s.SaveKafkaTransaction(key, []byte(tr.Message), err); e != nil {
			logger.Error("RetryKafkaTransaction: gagal memperbarui catatan "+key, "error", e)
		}
		return err
	}

	return s.repository.TandaiKafkaTransactionSelesai(key)
}

func (s *StreamService) RetryAllKafkaTransactions() (berhasil int, gagal int, err error) {
	daftar, err := s.repository.FindKafkaTransactions(nil)
	if err != nil {
		return 0, 0, err
	}

	for _, tr := range daftar {
		if tr == nil {
			continue
		}
		if e := s.RetryKafkaTransaction(tr.TransactionId); e != nil {
			logger.Error("RetryAllKafkaTransactions: "+tr.TransactionId, "error", e)
			gagal++
			continue
		}
		berhasil++
	}

	return berhasil, gagal, nil
}

func (s *StreamService) prosesUlang(tr *types.KafkaTransaction) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic saat memproses ulang transaksi FHIR",
				"err", r, "stack", string(debug.Stack()))
			err = exceptions.Internal.Messagef("panic saat memproses ulang: %v", r)
		}
	}()

	if tr == nil || len(tr.Message) == 0 {
		return exceptions.TidakDitemukan.Messagef("pesan kafka kosong, tidak ada yang bisa diproses ulang")
	}

	transaction := types2.PackMediator{}
	if e := json.Unmarshal([]byte(tr.Message), &transaction); e != nil {
		return exceptions.BodyRusak.WithMessage("Gagal Unmarshall Data. Json Invalid", e)
	}

	bundle, _, e := s.SusunUlangBundle(transaction)
	if e != nil {
		return exceptions.Internal.WithMessage("Gagal menerjemahkan data transaksi FHIR: "+e.Error(), e)
	}
	if bundle == nil {
		return exceptions.Internal.Messagef("Bundle transaksi FHIR kosong")
	}

	return s.ProsesTransaksiFHIR(bundle, transaction.Patient)
}

// DaftarKafkaTransactionGagal dipakai endpoint/UI untuk menampilkan antrean.
// patientId opsional untuk menyaring.
func (s *StreamService) DaftarKafkaTransactionGagal(patientId *string) ([]*types.KafkaTransaction, error) {
	return s.repository.FindKafkaTransactions(patientId)
}

func (s *StreamService) SaveKafkaTransaction(key string, message []byte, err error) error {
	pesan := "tanpa keterangan"
	if err != nil {
		pesan = err.Error()
	}

	eBytes, _ := json.Marshal(pesan)
	e := json.RawMessage(eBytes)

	patientId := ekstrakPatientId(message)

	data := types.KafkaTransactionPayload{
		Id:            uuid.NewString(),
		TransactionId: key,
		GroupId:       s.Config.Kafka.GroupID,
		PatientId:     patientId,
		Message:       string(message),
		Error:         &e,
		Attempt:       1,
		Created:       time.Now(),
	}
	return s.repository.SaveKafkaTransaction(data)
}

func ekstrakPatientId(message []byte) *string {
	tx := types2.PackMediator{}
	if err := json.Unmarshal(message, &tx); err != nil {
		return nil
	}

	if !utils.IsFilled(tx.Patient) {
		return nil
	}

	return tx.Patient
}

func (s *StreamService) ProsesTransaksiFHIR(bundle *types2.Bundle, ihsAnak *string) error {
	if utils.DumpAktif() {
		logger.Debug("Bundle: " + helper.ToLogJSON(bundle))
	}
	var anak *types.Anak
	var orangtua *types.Orangtua
	// var kunjungan *types.Kunjungan

	// dari envelope PackMediator.patient, cadangannya reference di resource
	idAnakSatusehat := ihsAnak
	orangtuaPerluUpdate := false

	var kunjungan *types.Kunjungan
	var refFaskes *string
	var observasi []utils.HasilObservasi
	var diagnosa []types.Diagnosa
	var layanan []types.Layanan
	var rujukan []types.Rujukan
	var episode []types.Episode

	var anak_ke *int

	// Urutkan bundle
	utils.UrutkanEntry(bundle.Entry)
	payload := types.KunjunganNestedPayload{}

	for _, entry := range bundle.Entry {
		if entry.Base == nil || !utils.IsFilled(entry.Base.ResourceType) {
			continue
		}
		if utils.DumpAktif() {
			logger.Debug("Entry: " + helper.ToLogJSON(entry.Resource))
		}

		// Untuk Update tidak mungkin entry bundle berisi lebih dari 1, karena tidak ada method PUT untuk resouceType Bundle di satusehat,
		// jadi PUT harus per resource yang kemudian dibungkus dalam BundlePack ketika dikirim ke Kafka
		update := entry.Request != nil && entry.Request.Method == fhir.HTTPVerbPUT

		switch strings.ToLower(*entry.Base.ResourceType) {
		case "patient":
			tmp, err := utils.PatientToAnak(entry)

			if err != nil {
				logger.Error("PatientToAnak", err)
				continue
			}

			// cari data anak di database
			anak, _ = s.repository.FindAnak(nil, tmp.Nik, nil, nil)

			if anak != nil {
				berubah := timpaAnakDariFHIR(anak, tmp)

				if berubah && !utils.IsFilled(anak.UpdatedBy) && anak.IdSatusehat != nil {
					anak.UpdatedBy = new("Orgid")

					if _, err := s.repository.UpdateAnakById(anak.ToPayload().ToEntity(), nil, true); err != nil {
						logger.Error("ProsesTransaksiFHIR:UpdateDataAnak => " + err.Error())
					}
				}
			} else {
				if update {
					if utils.DumpAktif() {
						logger.Debug("ProsesTransaksiFHIR:Patient:Update:Anak Tidak Ditemukan => continue")
					}
					continue
				}
				anak = &types.Anak{
					Nik:          tmp.Nik,
					IdSatusehat:  tmp.IdSatusehat,
					Nama:         tmp.Nama,
					TanggalLahir: tmp.TanggalLahir,
					JenisKelamin: tmp.JenisKelamin,
				}
			}

		case "relatedperson":
			// Orangtua
			tmp, ihs, err := utils.RelatedPersonToOrangtua(entry)

			if err != nil {
				logger.Error("RelatedPersonToOrangtua", err)
				continue
			}

			// nik satu-satunya kunci pencocokan
			if !utils.IsStrFilled(tmp.Nik) {
				logger.Error("ProsesTransaksiFHIR:RelatedPerson => nik kosong, entry dilewati")
				continue
			}

			if utils.IsFilled(ihs) {
				idAnakSatusehat = ihs
			}

			orangtua, _ = s.repository.FindOrangTuaByNik(tmp.Nik)

			if orangtua != nil {
				orangtuaPerluUpdate = timpaDariFHIR(orangtua, tmp)
			} else {
				if update {
					if utils.DumpAktif() {
						logger.Debug("ProsesTransaksiFHIR:Patient:Update:Orangtua Tidak Ditemukan => continue")
					}
					continue
				}
				orangtua = tmp
			}

		case "encounter":
			tmp, ihs, ihsFaskes, err := utils.EncounterToKunjungan(entry)
			if err != nil {
				logger.Error("EncounterToKunjungan", err)
				continue
			}
			if utils.IsFilled(ihs) {
				idAnakSatusehat = ihs
			}
			kunjungan = tmp
			refFaskes = ihsFaskes

			if update {
				_, err := s.repository.UpdateKunjunganBySatusehatId(tmp.ToPayload().ToEntity())
				return abaikanJikaBelumTersimpan(err)
			}

		case "observation":
			tmp, resource, ihs, refEnc, err := utils.ObservationToObservasi(entry)
			if err != nil {
				logger.Error("ObservationToObservasi", err)
				continue
			}
			if utils.IsFilled(ihs) {
				idAnakSatusehat = ihs
			}

			if resource != nil {
				u := utils.FindAnakKe(resource)

				if u != nil {
					anak_ke = u
				}
			}
			tmp.Induk.RefEncounter = refEnc
			observasi = append(observasi, *tmp)

			if update {
				_, err := s.repository.UpdateObservasiBySatusehatId(tmp.Induk.ToEntity())
				return abaikanJikaBelumTersimpan(err)
			}

		case "condition":
			tmp, ihs, refEnc, err := utils.ConditionToDiagnosa(entry)
			if err != nil {
				logger.Error("ConditionToDiagnosa", err)
				continue
			}
			if utils.IsFilled(ihs) {
				idAnakSatusehat = ihs
			}
			tmp.RefEncounter = refEnc
			diagnosa = append(diagnosa, *tmp)

			if update {
				return abaikanJikaBelumTersimpan(s.repository.UpdateDiagnosaBySatusehatId(tmp.ToEntity()))
			}

		case "procedure", "medicationdispense", "nutritionorder", "immunization":
			tmp, ihs, refEnc, err := petakanLayanan(entry)
			if err != nil {
				logger.Error("petakanLayanan", err)
				continue
			}
			if utils.IsFilled(ihs) {
				idAnakSatusehat = ihs
			}
			tmp.RefEncounter = refEnc
			layanan = append(layanan, *tmp)

			if update {
				return abaikanJikaBelumTersimpan(s.repository.UpdateLayananBySatusehatId(tmp.ToEntity()))
			}

		case "servicerequest":
			tmp, ihs, refEnc, err := utils.ServiceRequestToRujukan(entry)
			if err != nil {
				logger.Error("ServiceRequestToRujukan", err)
				continue
			}
			if utils.IsFilled(ihs) {
				idAnakSatusehat = ihs
			}
			tmp.RefEncounter = refEnc
			rujukan = append(rujukan, *tmp)

			if update {
				return abaikanJikaBelumTersimpan(s.repository.UpdateRujukanBySatusehatId(tmp.ToEntity()))
			}

		case "episodeofcare":
			tmp, ihs, ihsFaskes, err := utils.EpisodeOfCareToEpisode(entry)
			if err != nil {
				logger.Error("EpisodeOfCareToEpisode", err)
				continue
			}
			if utils.IsFilled(ihs) {
				idAnakSatusehat = ihs
			}
			tmp.RefFaskes = ihsFaskes
			episode = append(episode, *tmp)

			if update {
				return abaikanJikaBelumTersimpan(s.repository.UpdateEpisodeBySatusehatId(tmp.ToEntity()))
			}

		case "allergyintolerance":
			tmp, ihs, refEnc, err := utils.AllergyIntoleranceToDiagnosa(entry)
			if err != nil {
				logger.Error("AllergyIntoleranceToDiagnosa", err)
				continue
			}
			if utils.IsFilled(ihs) {
				idAnakSatusehat = ihs
			}
			tmp.RefEncounter = refEnc
			diagnosa = append(diagnosa, *tmp)

			if update {
				return abaikanJikaBelumTersimpan(s.repository.UpdateDiagnosaBySatusehatId(tmp.ToEntity()))
			}

		case "questionnaireresponse":
			// ditunda: linkId string bebas, bukan kode terminologi
		case "composition":
			// Composition ditunda dulu
		}
	}

	if anak != nil && !utils.IsStrFilled(anak.Id) {
		a, err := s.PastikanAnak(anak, anak_ke)
		if err != nil {
			return err
		}
		if a == nil {
			return nil
		}
		anak = a
		if payload.Kunjungan != nil {
			payload.Kunjungan.IDAnak = a.Id
		}
	}

	if anak_ke != nil && utils.IsFilled(idAnakSatusehat) {
		// Update data anak
		anak_db, err := s.repository.FindAnakBySatusehatId(*idAnakSatusehat)
		if err != nil {
			logger.Error("ProsesTransaksiFHIR:UpdateDataAnak => " + err.Error())
		}

		if anak_db != nil {
			anak_db.AnakKe = int16(*anak_ke)
			if _, err := s.repository.UpdateAnakById(anak_db.ToPayload().ToEntity(), nil, true); err != nil {
				logger.Error("ProsesTransaksiFHIR:UpdateDataAnak => " + err.Error())
			}
		}
	}

	if kunjungan != nil || len(observasi) > 0 || len(diagnosa) > 0 || len(layanan) > 0 || len(rujukan) > 0 || len(episode) > 0 {
		_, _, alasan, err := s.simpanDataMedis(anak, idAnakSatusehat, kunjungan, refFaskes, observasi, diagnosa, layanan, rujukan, episode)

		if err != nil {
			return err
		}
		if alasan != "" {
			logger.Info("ProsesTransaksiFHIR: data medis tidak disimpan => " + alasan)
		}
	}

	if orangtua != nil {
		s.simpanOrangtua(orangtua, orangtuaPerluUpdate, idAnakSatusehat)
	}

	body, err := json.Marshal(payload.Kunjungan)
	if err != nil {
		logger.Error("ProsesTransaksiFHIR: Gagal Marshall Payload", err)
		return exceptions.Internal.WithMessage("ProsesTransaksiFHIR: Gagal Marshall Payload", err)
	}

	if payload.Kunjungan != nil {
		// untuk membuat kunjungan, variasi payload yang akan digunakan adalah flat (payload kunjungan saja)
		_, err = s.stunting.CreateKunjungan(body)
		if err != nil {
			logger.Error("ProsesTransaksiFHIR: CreateKunjungan", err)
			return exceptions.Internal.WithMessage("ProsesTransaksiFHIR: "+err.Error(), err)
		}
	}

	return nil
}

// penghubungan butuh id hasil create, jadi berurutan
func (s *StreamService) simpanOrangtua(orangtua *types.Orangtua, perluUpdate bool, idAnakSatusehat *string) {
	idOrangtua := orangtua.Id

	switch {
	case !utils.IsStrFilled(idOrangtua):
		// belum terdaftar di database, buat baru
		orangtua.Id = uuid.NewString()

		p := orangtua.ToPayload()
		if err := utils.ValidateOrangtuaStream(*p); err != nil {
			logger.Error("ProsesTransaksiFHIR:CreateOrangtua validasi => " + err.Error())
			return
		}

		e := p.ToEntity()
		if orangtua.IdSatusehat != nil {
			e.SatusehatId = *orangtua.IdSatusehat
		}

		o, err := s.repository.CreateOrangTua(e, nil)
		if err != nil {
			logger.Error("ProsesTransaksiFHIR:CreateOrangtua => " + err.Error())
			return
		}
		idOrangtua = o.Id

	case perluUpdate:
		// sudah ada tapi satusehat_id-nya masih kosong
		e := orangtua.ToPayload().ToEntity()
		if orangtua.IdSatusehat != nil {
			e.SatusehatId = *orangtua.IdSatusehat
		}

		if _, err := s.repository.UpdateOrangTuaById(e, nil, true); err != nil {
			logger.Error("ProsesTransaksiFHIR:UpdateOrangtua => " + err.Error())
		}
	}

	if !utils.IsFilled(idAnakSatusehat) {
		return
	}

	n, err := s.repository.HubungkanAnakKeOrangTua(*idAnakSatusehat, idOrangtua, nil)
	if err != nil {
		logger.Error("ProsesTransaksiFHIR:HubungkanAnakKeOrangTua => " + err.Error())
		return
	}

	if n == 0 {
		logger.Info(fmt.Sprintf("ProsesTransaksiFHIR: anak %s tidak dihubungkan ke orangtua %s (belum ada atau sudah punya orangtua)", *idAnakSatusehat, idOrangtua))
	}
}

// FHIR menimpa jakantro, hanya untuk field yang terisi di resource
func timpaAnakDariFHIR(db *types.Anak, fhirData *types.Anak) bool {
	berubah := false

	timpaStr := func(tujuan *string, nilai string) {
		if utils.IsStrFilled(nilai) && *tujuan != nilai {
			*tujuan = nilai
			berubah = true
		}
	}

	timpaStr(&db.Nama, fhirData.Nama)
	timpaStr(&db.TanggalLahir, fhirData.TanggalLahir)
	timpaStr(&db.JenisKelamin, fhirData.JenisKelamin)

	if utils.IsFilled(fhirData.Nik) && (!utils.IsFilled(db.Nik) || *db.Nik != *fhirData.Nik) {
		db.Nik = fhirData.Nik
		berubah = true
	}

	if utils.IsFilled(fhirData.IdSatusehat) &&
		(!utils.IsFilled(db.IdSatusehat) || *db.IdSatusehat != *fhirData.IdSatusehat) {
		db.IdSatusehat = fhirData.IdSatusehat
		berubah = true
	}

	return berubah
}

// FHIR menimpa jakantro, hanya untuk field yang terisi di resource
func timpaDariFHIR(db *types.Orangtua, fhirData *types.Orangtua) bool {
	berubah := false

	timpaStr := func(tujuan *string, nilai string) {
		if utils.IsStrFilled(nilai) && *tujuan != nilai {
			*tujuan = nilai
			berubah = true
		}
	}

	timpaStr(&db.NamaAyah, fhirData.NamaAyah)
	timpaStr(&db.NamaIbu, fhirData.NamaIbu)
	timpaStr(&db.Telepon, fhirData.Telepon)
	timpaStr(&db.Alamat, fhirData.Alamat)

	if utils.IsFilled(fhirData.IdSatusehat) &&
		(!utils.IsFilled(db.IdSatusehat) || *db.IdSatusehat != *fhirData.IdSatusehat) {
		db.IdSatusehat = fhirData.IdSatusehat
		berubah = true
	}

	return berubah
}

func (s *StreamService) extractRequestAndResponse(tx types2.PackMediator) (*fhir.Bundle, []types2.BundleEntryResponse, error) {
	// Contoh untuk request body (diasumsikan sebagai Bundle)
	var request *fhir.Bundle
	var response []types2.BundleEntryResponse

	if tx.Input != nil {
		request = tx.Input
	} else {
		return nil, nil, fmt.Errorf("request body kosong")
	}

	// Response body (bisa Bundle atau OperationOutcome), hanya proses yang Bundle
	if tx.Response != nil {
		response = tx.Response
		if len(request.Entry) != len(response) {
			return nil, nil, fmt.Errorf("jumlah entry Bundle request dan response tidak sama")
		}
	} else {
		return nil, nil, fmt.Errorf("response body kosong")
	}

	return request, response, nil
}

func (s *StreamService) SusunUlangBundle(tx types2.PackMediator) (*types2.Bundle, *types2.SetReference, error) {
	// Ekstrak request dan response dari transaksi
	requestBundle, response, err := s.extractRequestAndResponse(tx)

	if err != nil {
		return nil, nil, err
	}
	processor.NormalizeTemporaryReferences(requestBundle)

	if utils.DumpAktif() {
		logger.Debug("Request Bundle Setelah Normalisasi: " + helper.ToLogJSON(requestBundle))
	}

	var register types2.SetReference = make(types2.SetReference)
	newBundle := types2.CastToBundle(*requestBundle)
	idBaru := []types2.NewPost{} // Mapping ID Response ke request Resource yang baru dipost
	for i := range requestBundle.Entry {
		if response[i].Id == nil || *response[i].Id == "" {
			continue
		}

		responseId := response[i].Id
		entry, err := processor.NewBundleEntry(requestBundle.Entry[i], responseId)
		if entry == nil || err != nil {
			continue
		}

		idBaru = s.mapAllReferences(&register, idBaru, entry, tx.Response[i], i, 0)

		newBundle.Entry[i] = *entry

	}
	if len(idBaru) > 0 {
		for i := range newBundle.Entry {
			processor.UpdateTemporaryReference(&newBundle.Entry[i], &register, idBaru)
		}
	}
	return &newBundle, &register, err
}

func (h *StreamService) mapAllReferences(register *types2.SetReference, baru []types2.NewPost, entry *types2.BundleEntry, response types2.BundleEntryResponse, i int, priority int) []types2.NewPost {
	// jika response.Id null skip saja
	if response.Id == nil {
		return baru
	}

	// ID Resource sebagai key untuk register
	id := response.Id

	// Handle Id setiap resource di dalam bundle entry
	urlUrn := entry.GetTemporaryURN()
	if urlUrn != nil {
		if _, ok := (*register)[*id]; !ok {
			urn := (*urlUrn)[9:]
			sre := &types2.SetReferenceEntry{
				Index:              i,
				Priority:           priority,
				Temporary:          true,
				ReferredByTemp:     []string{},
				ResourceType:       *response.ResourceType,
				URN:                &urn,
				Id:                 id,
				Entry:              entry,
				IsSatusehatChecked: true,
				InSatusehat:        types2.DONE,
			}

			// ID Resource sebagai key untuk register
			(*register)[*id] = sre

			// Resource baru yang di-POST
			baru = append(baru, types2.NewPost{
				Index:        i,
				ResourceType: *entry.Base.ResourceType,
				NewID:        *id,
				TempID:       urn,
			})
		}
	} else {
		sre, ok := (*register)[*id]
		if !ok {
			sre = &types2.SetReferenceEntry{
				Index:              i,
				Priority:           priority,
				ResourceType:       *response.ResourceType,
				Id:                 id,
				Entry:              entry,
				IsSatusehatChecked: true,
				InSatusehat:        types2.DONE,
			}

			(*register)[*id] = sre
		} else {
			if priority > sre.Priority {
				sre.Priority = priority
			}

			if !sre.IsSatusehatChecked {
				sre.IsSatusehatChecked = true
				sre.InSatusehat = types2.DONE
			}
		}
	}

	// Ambil semua references di resource ini
	references := processor.ExtractReferences(*entry.Base)

	// daftarkan ke register untuk reference temuan
	for _, r := range references {
		if r.Reference == nil || *r.Reference == "" {
			continue
		}

		// Temporary Id biarkan dikerjakan di level resource, disini hanya mencari id reference yang bukan non-temporary
		// id := types2.GetReferenceID(&r)
		rp := strings.Split(*r.Reference, "/")
		if len(rp) == 2 {
			restype := rp[0]
			id := rp[1]
			reg, ok := (*register)[id]
			newPriority := priority + 1

			if !ok {
				sre := &types2.SetReferenceEntry{
					Index:              -1, // bukan reference yang menuju resource dalam di dalam bundle entry (resource ekternal yang mungkin belum tersedia)
					Priority:           newPriority,
					ResourceType:       restype,
					Id:                 &id,
					IsSatusehatChecked: true,
					InSatusehat:        types2.DONE,
				}
				(*register)[id] = sre
			} else {
				if reg.Priority < newPriority {
					(*register)[id].Priority = newPriority
				}
			}
		}
	}

	return baru
}

func petakanLayanan(entry types2.BundleEntry) (*types.Layanan, *string, *string, error) {
	switch strings.ToLower(*entry.Base.ResourceType) {
	case "procedure":
		return utils.ProcedureToLayanan(entry)
	case "medicationdispense":
		return utils.MedicationDispenseToLayanan(entry)
	case "nutritionorder":
		return utils.NutritionOrderToLayanan(entry)
	case "immunization":
		return utils.ImmunizationToLayanan(entry)
	}
	return nil, nil, nil, fmt.Errorf("resource %s bukan layanan", *entry.Base.ResourceType)
}

// alasan hanya terisi kalau sengaja tidak disimpan; kegagalan sebenarnya lewat error
func (s *StreamService) simpanDataMedis(anak *types.Anak, ihsAnak *string, kunjungan *types.Kunjungan,
	refFaskes *string, observasi []utils.HasilObservasi, diagnosa []types.Diagnosa,
	layanan []types.Layanan, rujukan []types.Rujukan, episode []types.Episode) (*string, bool, string, error) {
	idAnak := ""
	if anak != nil {
		idAnak = anak.Id
	}
	// bundle tanpa Patient: anaknya dicari lewat IHS id dari envelope
	if !utils.IsStrFilled(idAnak) && utils.IsFilled(ihsAnak) {
		if a, err := s.repository.FindAnakBySatusehatId(*ihsAnak); err == nil && a != nil {
			idAnak = a.Id
		}
	}
	if !utils.IsStrFilled(idAnak) {
		logger.Info("ProsesTransaksiFHIR:simpanDataMedis => anak tidak ditemukan, data medis dilewati")
		return nil, false, "anak tidak ditemukan", nil
	}

	// kunjungan tanpa data klinis punya stunting NULL, tidak jadi acuan
	terakhirStunting := false
	if k, err := s.repository.FindStatusStuntingTerakhir(idAnak); err == nil && k != nil {
		terakhirStunting = utils.SmallintToBool(k.Stunting)
	}

	adaDataKlinis := len(observasi) > 0 || len(diagnosa) > 0
	stunting := adaTandaStunting(diagnosa, observasi)

	if !stunting && !terakhirStunting {
		logger.Info(fmt.Sprintf("ProsesTransaksiFHIR:simpanDataMedis => data medis dilewati (%v, %v)", stunting, terakhirStunting))
		return nil, false, "tidak ada tanda stunting dan riwayat terakhir bukan stunting", nil
	}

	logger.Info(fmt.Sprintf("ProsesTransaksiFHIR:simpanDataMedis => data medis memenuhi syarat (%v, %v)", stunting, terakhirStunting))

	data := &repository.DataMedis{}

	// episode lebih dulu: kunjungan menunjuk ke sana lewat FK
	for i := range episode {
		ep := &episode[i]
		ep.IdAnak = idAnak

		if lama := s.cariEpisode(ep.IdSatusehat); lama != nil {
			ep.Id = *lama
			continue
		}

		ep.Id = uuid.NewString()
		if id, _ := s.faskesDariRef(ep.RefFaskes); id != nil {
			ep.IdFaskes = id
		}
		data.Episode = append(data.Episode, ep.ToEntity())
	}

	var idKunjungan *string
	var statusLama *int
	var faskesKunjungan *string
	sudahAda := false

	if kunjungan != nil {
		angkatKeKunjungan(kunjungan, observasi)
		if adaDataKlinis {
			kunjungan.Stunting = utils.BoolToSmallintPtr(stunting)
		}

		// Encounter yang sama bisa tiba dua kali, kafka at-least-once
		if lama := s.cariKunjungan(kunjungan.IdSatusehat); lama != nil {
			idKunjungan, statusLama, faskesKunjungan, sudahAda = &lama.Id, lama.Stunting, lama.IdFaskes, true
		} else {
			e := s.entityKunjungan(idAnak, kunjungan, refFaskes)
			data.Kunjungan = e
			data.RefEncounter = kunjungan.IdSatusehat
			idKunjungan, faskesKunjungan = &e.ID, e.IDFaskes
		}
	} else {
		// tanpa Encounter di bundle ini: dicari lewat reference, jangan dibuat baru
		if lama := s.cariKunjungan(refEncounterPertama(observasi, diagnosa, layanan, rujukan)); lama != nil {
			idKunjungan, statusLama, faskesKunjungan, sudahAda = &lama.Id, lama.Stunting, lama.IdFaskes, true
		}
		if !sudahAda {
			logger.Info("ProsesTransaksiFHIR:simpanDataMedis => kunjungan belum ada, turunan disimpan menggantung")
		}
	}

	// status menyusul saat data klinis tiba; yang sudah 1 tidak diturunkan
	if sudahAda && adaDataKlinis && (statusLama == nil || (stunting && *statusLama != 1)) {
		data.IdKunjunganStatus = idKunjungan
		data.Stunting = utils.BoolToSmallintPtr(stunting)
	}

	for _, h := range observasi {
		induk := h.Induk
		induk.Id = uuid.NewString()
		induk.IdAnak = idAnak
		induk.IdKunjungan = idKunjungan
		data.Observasi = append(data.Observasi, induk.ToEntity())

		for _, c := range h.Component {
			c.Id = uuid.NewString()
			c.IdAnak = idAnak
			c.IdKunjungan = idKunjungan
			c.IdInduk = &induk.Id
			c.RefEncounter = induk.RefEncounter
			data.Komponen = append(data.Komponen, c.ToEntity())
		}
	}

	for _, d := range diagnosa {
		d.IdAnak = idAnak
		d.Id = uuid.NewString()
		d.IdKunjungan = idKunjungan
		data.Diagnosa = append(data.Diagnosa, d.ToEntity())
	}

	for _, l := range layanan {
		l.IdAnak = idAnak
		l.Id = uuid.NewString()
		l.IdKunjungan = idKunjungan
		data.Layanan = append(data.Layanan, l.ToEntity())
	}

	for _, r := range rujukan {
		r.IdAnak = idAnak
		r.Id = uuid.NewString()
		r.IdKunjungan = idKunjungan
		// requester bisa Practitioner; asalnya jatuh ke faskes Encounter-nya
		refAsal := r.RefFaskesAsal
		if !utils.IsFilled(refAsal) {
			refAsal = refFaskes
		}
		idAsal, jenisAsal := s.faskesDariRef(refAsal)
		if idAsal == nil {
			idAsal = faskesKunjungan
		}
		idTujuan, jenisTujuan := s.faskesDariRef(r.RefFaskesTujuan)
		r.RefFaskesAsal = refAsal
		r.IdFaskesAsal, r.IdFaskesTujuan = idAsal, idTujuan
		r.Jenis = arahRujukan(r.Jenis, refAsal, r.RefFaskesTujuan, jenisAsal, jenisTujuan)
		data.Rujukan = append(data.Rujukan, r.ToEntity())
	}

	if utils.DumpAktif() {
		logger.Debug("DataMedis: " + helper.ToLogJSON(data))
	}

	if err := s.repository.SimpanDataMedis(data, nil); err != nil {
		logger.Error("ProsesTransaksiFHIR:SimpanDataMedis => " + err.Error())
		return nil, false, "", err
	}

	return idKunjungan, true, "", nil
}

func (s *StreamService) entityKunjungan(idAnak string, k *types.Kunjungan, refFaskes *string) *entity.Kunjungan {
	k.Id = uuid.NewString()
	k.IdAnak = idAnak

	if utils.IsFilled(refFaskes) {
		if f, err := s.repository.FindFaskesByOrgid(*refFaskes); err == nil && f != nil {
			k.IdFaskes = &f.Id
		}
	}

	// episode dan rujukan yang dipenuhi mungkin sudah tiba di bundle lain
	if utils.IsFilled(k.RefEpisode) {
		k.IdEpisode = s.cariEpisode(k.RefEpisode)
	}
	if utils.IsFilled(k.RefRujukan) {
		k.IdRujukan = s.cariRujukan(k.RefRujukan)
	}

	e := &entity.Kunjungan{
		ID:            k.Id,
		IDAnak:        k.IdAnak,
		CaraUkur:      k.CaraUkur,
		BeratBadan:    k.BeratBadan,
		TinggiBadan:   k.TinggiBadan,
		LingkarLengan: k.LingkarLengan,
		LingkarKepala: k.LingkarKepala,
		LingkarDada:   k.LingkarDada,
		IDFaskes:      k.IdFaskes,
		SatusehatId:   k.IdSatusehat,
		IDEpisode:     k.IdEpisode,
		RefEpisode:    k.RefEpisode,
		IDRujukan:     k.IdRujukan,
		RefRujukan:    k.RefRujukan,
		Stunting:      k.Stunting,
	}
	if t := utils.WaktuDariTanggal(k.TanggalPengukuran); t != nil {
		e.TanggalPengukuran = *t
	}
	if k.TanggalSelesai != nil {
		e.TanggalSelesai = utils.WaktuDariTanggal(*k.TanggalSelesai)
	}

	return e
}

func (s *StreamService) cariRujukan(refRujukan *string) *string {
	if !utils.IsFilled(refRujukan) {
		return nil
	}
	if r, err := s.repository.FindRujukanBySatusehatId(*refRujukan); err == nil && r != nil {
		return &r.ID
	}
	return nil
}

func (s *StreamService) cariEpisode(refEpisode *string) *string {
	if !utils.IsFilled(refEpisode) {
		return nil
	}
	if e, err := s.repository.FindEpisodeBySatusehatId(*refEpisode); err == nil && e != nil {
		return &e.ID
	}
	return nil
}

// dipanggil sekali per bundle, bukan per turunan
func (s *StreamService) cariKunjungan(refEncounter *string) *types.Kunjungan {
	if !utils.IsFilled(refEncounter) {
		return nil
	}
	if k, err := s.repository.FindKunjunganBySatusehatId(*refEncounter); err == nil {
		return k
	}
	return nil
}

// observasinya tetap tersimpan utuh di tabel observasi
func angkatKeKunjungan(k *types.Kunjungan, observasi []utils.HasilObservasi) {
	if k == nil {
		return
	}

	for _, h := range observasi {
		if h.Kolom == "" || h.Angka == nil {
			continue
		}
		switch h.Kolom {
		case "berat_badan":
			k.BeratBadan = h.Angka
		case "tinggi_badan":
			k.TinggiBadan = h.Angka
			if h.CaraUkur != "" {
				cara := h.CaraUkur
				k.CaraUkur = &cara
			}
		case "lingkar_kepala":
			k.LingkarKepala = h.Angka
		case "lingkar_lengan":
			k.LingkarLengan = h.Angka
		}
	}
}

func adaTandaStunting(diagnosa []types.Diagnosa, observasi []utils.HasilObservasi) bool {
	for _, d := range diagnosa {
		if utils.TandaStunting(d.System, d.Kode) {
			return true
		}
	}

	for _, h := range observasi {
		if h.Induk.Interpretasi != nil && utils.TandaStunting("", *h.Induk.Interpretasi) {
			return true
		}
		if h.Induk.NilaiKode != nil && utils.TandaStunting(utils.StrKosong(h.Induk.NilaiSystem), *h.Induk.NilaiKode) {
			return true
		}
		for _, c := range h.Component {
			if c.Interpretasi != nil && utils.TandaStunting("", *c.Interpretasi) {
				return true
			}
			if c.NilaiKode != nil && utils.TandaStunting(utils.StrKosong(c.NilaiSystem), *c.NilaiKode) {
				return true
			}
		}
	}

	return false
}

func refEncounterPertama(observasi []utils.HasilObservasi, diagnosa []types.Diagnosa, layanan []types.Layanan, rujukan []types.Rujukan) *string {
	for _, h := range observasi {
		if utils.IsFilled(h.Induk.RefEncounter) {
			return h.Induk.RefEncounter
		}
	}
	for _, d := range diagnosa {
		if utils.IsFilled(d.RefEncounter) {
			return d.RefEncounter
		}
	}
	for _, l := range layanan {
		if utils.IsFilled(l.RefEncounter) {
			return l.RefEncounter
		}
	}
	for _, r := range rujukan {
		if utils.IsFilled(r.RefEncounter) {
			return r.RefEncounter
		}
	}
	return nil
}

// dari IHS Organization id ke id internal + jenisnya, sekali cari
func (s *StreamService) faskesDariRef(ref *string) (*string, string) {
	if !utils.IsFilled(ref) {
		return nil, ""
	}
	f, err := s.repository.FindFaskesByOrgid(*ref)
	if err != nil || f == nil {
		return nil, ""
	}
	return &f.Id, strings.ToLower(strings.TrimSpace(f.Jenis))
}

func arahRujukan(dariKategori string, refAsal, refTujuan *string, jenisAsal, jenisTujuan string) string {
	if dariKategori != "" {
		return dariKategori
	}
	if !utils.IsFilled(refTujuan) {
		return entity.RujukanInternal
	}
	if utils.IsFilled(refAsal) && *refAsal == *refTujuan {
		return entity.RujukanInternal
	}
	if strings.Contains(jenisAsal, "rs") && strings.Contains(jenisTujuan, "puskesmas") {
		return entity.RujukBalik
	}
	return entity.RujukanKeluar
}

// nil tanpa error berarti bukan sasaran pemantauan, bukan kegagalan
func (s *StreamService) PastikanAnak(anak *types.Anak, anakKe *int) (*types.Anak, error) {
	if anak == nil {
		return nil, exceptions.Validasi.Messagef("data anak kosong")
	}
	if utils.IsStrFilled(anak.Id) {
		return anak, nil
	}

	bd, err := time.Parse("2006-01-02", anak.TanggalLahir)
	if err != nil {
		return nil, exceptions.Validasi.Messagef("anak.tanggal_lahir tidak valid: %q", anak.TanggalLahir)
	}
	if bd.After(time.Now()) {
		return nil, exceptions.Validasi.Messagef("anak.tanggal_lahir belum terjadi")
	}
	if !time.Now().Before(bd.AddDate(5, 0, 0)) {
		logger.Info(fmt.Sprintf("PastikanAnak: %s sudah lewat 5 tahun, dilewati", utils.Nilai(anak.IdSatusehat)))
		return nil, nil
	}

	anak.Id = uuid.NewString()
	p := anak.ToPayload()
	if anakKe != nil {
		p.AnakKe = int16(*anakKe)
	}
	if err := utils.ValidateAnakStream(*p); err != nil {
		return nil, err
	}
	return s.repository.CreateAnak(p.ToEntity(), nil)
}

func (s *StreamService) CariAnak(a *types.Anak) *types.Anak {
	if a == nil {
		return nil
	}
	if utils.IsFilled(a.IdSatusehat) {
		if k, err := s.repository.FindAnakBySatusehatId(*a.IdSatusehat); err == nil && k != nil {
			return k
		}
	}
	if utils.IsFilled(a.Nik) {
		if k, err := s.repository.FindAnak(nil, a.Nik, nil, nil); err == nil && k != nil {
			return k
		}
	}
	return nil
}

// orgid dari API key, bukan dari body
func (s *StreamService) SimpanPemeriksaanFaskes(p *types.PemeriksaanFaskes, orgid string) (*types.HasilPemeriksaan, error) {
	if p == nil {
		return nil, exceptions.BodyRusak.New(nil)
	}
	if !utils.IsStrFilled(orgid) {
		return nil, exceptions.Forbidden.WithMessage("orgid faskes tidak diketahui", nil)
	}

	faskes, err := s.repository.FindFaskesByOrgid(orgid)
	if err != nil || faskes == nil {
		return nil, exceptions.Forbidden.WithMessage("faskes dengan orgid "+orgid+" belum terdaftar", err)
	}

	observasi := make([]utils.HasilObservasi, 0, len(p.Observasi))
	for _, o := range p.Observasi {
		observasi = append(observasi, utils.HasilDariObservasi(o, o.Component))
	}

	anak := s.CariAnak(&p.Anak)
	if anak == nil {
		if !adaTandaStunting(p.Diagnosa, observasi) {
			return nil, exceptions.TidakDisimpan.Messagef(
				"tidak ada tanda stunting dan anak belum punya riwayat")
		}

		anak, err = s.PastikanAnak(&p.Anak, nil)
		if err != nil {
			return nil, err
		}
		if anak == nil {
			return nil, exceptions.TidakDisimpan.Messagef(
				"anak bukan sasaran pemantauan, umurnya sudah lewat 5 tahun")
		}
	}

	kunjungan := p.Kunjungan
	kunjungan.IdFaskes = &faskes.Id

	idKunjungan, disimpan, alasan, err := s.simpanDataMedis(anak, anak.IdSatusehat, &kunjungan, &orgid,
		observasi, p.Diagnosa, p.Layanan, p.Rujukan, p.Episode)

	if err != nil {
		return nil, err
	}

	if !disimpan {
		return nil, exceptions.TidakDisimpan.Messagef("%s", alasan)
	}

	return &types.HasilPemeriksaan{
		IdAnak:      anak.Id,
		IdKunjungan: idKunjungan,
	}, nil
}

func abaikanJikaBelumTersimpan(err error) error {
	if err == nil {
		return nil
	}

	var siap *exceptions.Error
	if errors.As(err, &siap) && siap.ErrorCode == exceptions.TidakDitemukan.ErrorCode {
		logger.Info("ProsesTransaksiFHIR:Update => resource tidak ada di database, dilewati")
		return nil
	}

	return err
}
