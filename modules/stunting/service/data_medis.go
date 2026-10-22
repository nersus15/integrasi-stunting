package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/satusehat"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/repository"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"
)

// alasan hanya terisi kalau sengaja tidak disimpan; kegagalan sebenarnya lewat error
func (s *StreamService) simpanDataMedis(anak *types.Anak, ihsAnak *string, kunjungan *types.Kunjungan,
	refFaskes *string, observasi []satusehat.HasilObservasi, diagnosa []types.Diagnosa,
	layanan []types.Layanan, rujukan []types.Rujukan, episode []types.Episode, anakBaru *entity.Anak, asalKunjungan asal) (*string, bool, string, error) {
	idAnak := ""
	if anak != nil {
		idAnak = anak.Id
	}
	// bundle tanpa Patient: anaknya dicari lewat IHS id dari envelope
	if !utils.IsStrFilled(idAnak) && utils.IsFilled(ihsAnak) {
		if a, err := s.repository.FindAnakBySatusehatId(*ihsAnak, false); err == nil && a != nil {
			idAnak = a.ID
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

	data := &repository.DataMedis{AnakBaru: anakBaru}

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
			asalKunjungan.kunjungan(e)
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

	if sudahAda && len(observasi) > 0 {
		ukur := &types.Kunjungan{}

		angkatKeKunjungan(ukur, observasi)
		data.IdKunjunganUkur = idKunjungan
		data.KolomUkur = satusehat.KolomUkurTerisi(ukur)
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
	if len(data.Rujukan) > 0 {
		s.repository.LupakanRujukanTerakhir(idAnak)
	}

	return idKunjungan, true, "", nil
}

func (s *StreamService) entityKunjungan(idAnak string, k *types.Kunjungan, refFaskes *string) *entity.Kunjungan {
	k.Id = uuid.NewString()
	k.IdAnak = idAnak

	if utils.IsFilled(refFaskes) {
		if f, err := s.repository.FindFaskesByOrgid(*refFaskes); err == nil && f != nil {
			k.IdFaskes = &f.ID
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
		ID:             k.Id,
		IDAnak:         k.IdAnak,
		CaraUkur:       k.CaraUkur,
		BeratBadan:     k.BeratBadan,
		TinggiBadan:    k.TinggiBadan,
		LingkarLengan:  k.LingkarLengan,
		LingkarKepala:  k.LingkarKepala,
		LingkarDada:    k.LingkarDada,
		ASIBulan0:      k.ASIBulan0,
		ASIBulan1:      k.ASIBulan1,
		ASIBulan2:      k.ASIBulan2,
		ASIBulan3:      k.ASIBulan3,
		ASIBulan4:      k.ASIBulan4,
		ASIBulan5:      k.ASIBulan5,
		ASIBulan6:      k.ASIBulan6,
		VitBiru:        k.VitBiru,
		VitMerah:       k.VitMerah,
		PittingEdema:   k.PittingEdema,
		KelasIbuBalita: k.KelasIbuBalita,
		StatusBBU:      k.StatusBbu,
		StatusTBU:      k.StatusTbu,
		StatusBBTB:     k.StatusBbtb,
		ZScoreBBU:      k.ZscoreBbu,
		ZScoreTBU:      k.ZscoreTbu,
		ZScoreBBTB:     k.ZscoreBbtb,
		IDFaskes:       k.IdFaskes,
		SatusehatId:    k.IdSatusehat,
		IDEpisode:      k.IdEpisode,
		RefEpisode:     k.RefEpisode,
		IDRujukan:      k.IdRujukan,
		RefRujukan:     k.RefRujukan,
		Stunting:       k.Stunting,
	}

	if t := satusehat.WaktuDariTanggal(k.TanggalPengukuran); t != nil {
		e.TanggalPengukuran = *t
	}
	if k.TanggalSelesai != nil {
		e.TanggalSelesai = satusehat.WaktuDariTanggal(*k.TanggalSelesai)
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
	if e, err := s.repository.FindEpisodeBySatusehatId(*refEpisode, false); err == nil && e != nil {
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
		return (*types.Kunjungan)(nil).FromEntity(k)
	}
	return nil
}

// observasinya tetap tersimpan utuh di tabel observasi
func angkatKeKunjungan(k *types.Kunjungan, observasi []satusehat.HasilObservasi) {
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

		// status ikut z-score pasangannya, kosong berarti NULL
		case "zscore_bbu":
			k.ZscoreBbu, k.StatusBbu = h.Angka, isiStatus(h.Status)
		case "zscore_tbu":
			k.ZscoreTbu, k.StatusTbu = h.Angka, isiStatus(h.Status)
		case "zscore_bbtb":
			k.ZscoreBbtb, k.StatusBbtb = h.Angka, isiStatus(h.Status)
		}
	}
}

func isiStatus(s string) *string {
	if !utils.IsStrFilled(s) {
		return nil
	}
	return &s
}

func adaTandaStunting(diagnosa []types.Diagnosa, observasi []satusehat.HasilObservasi) bool {
	for _, d := range diagnosa {
		if satusehat.TandaStunting(d.System, d.Kode) {
			return true
		}
	}

	for _, h := range observasi {
		if h.Induk.Interpretasi != nil && satusehat.TandaStunting("", *h.Induk.Interpretasi) {
			return true
		}
		if h.Induk.NilaiKode != nil && satusehat.TandaStunting(satusehat.StrKosong(h.Induk.NilaiSystem), *h.Induk.NilaiKode) {
			return true
		}
		for _, c := range h.Component {
			if c.Interpretasi != nil && satusehat.TandaStunting("", *c.Interpretasi) {
				return true
			}
			if c.NilaiKode != nil && satusehat.TandaStunting(satusehat.StrKosong(c.NilaiSystem), *c.NilaiKode) {
				return true
			}
		}
	}

	return false
}

func refEncounterPertama(observasi []satusehat.HasilObservasi, diagnosa []types.Diagnosa, layanan []types.Layanan, rujukan []types.Rujukan) *string {
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
	return &f.ID, strings.ToLower(strings.TrimSpace(f.Jenis))
}

func (s *StreamService) arahRujukanTersimpan(r *types.Rujukan) error {
	if r.Jenis != "" {
		return nil
	}

	lama, err := s.repository.FindRujukanBySatusehatId(utils.StrPtr(r.IdSatusehat))
	if err != nil {
		return err
	}

	refAsal, refTujuan := r.RefFaskesAsal, r.RefFaskesTujuan
	if !utils.IsFilled(refAsal) {
		refAsal = lama.RefFaskesAsal
	}
	if !utils.IsFilled(refTujuan) {
		refTujuan = lama.RefFaskesTujuan
	}

	_, jenisAsal := s.faskesDariRef(refAsal)
	_, jenisTujuan := s.faskesDariRef(refTujuan)
	r.Jenis = arahRujukan("", refAsal, refTujuan, jenisAsal, jenisTujuan)

	return nil
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
