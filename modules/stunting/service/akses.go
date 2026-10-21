package service

import (
	"slices"
	"strings"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
)

type Aksi string

const (
	AksiBaca  Aksi = "baca"
	AksiUbah  Aksi = "ubah"
	AksiHapus Aksi = "hapus"
)

// data yang diakses; cukup isi yang relevan
type Sasaran struct {
	Jenis      string
	Id         string
	Pembuat    *string // created_by; data medis memakai pembuat kunjungannya
	IdAnak     string
	IdOrangtua string
	IdPosyandu *string
}

func sasaranAnak(e *entity.Anak) Sasaran {
	return Sasaran{Jenis: "anak", Id: e.ID, Pembuat: e.CreatedBy, IdAnak: e.ID, IdOrangtua: e.IDOrangtua}
}

func sasaranOrangtua(e *entity.Orangtua) Sasaran {
	return Sasaran{Jenis: "orangtua", Id: e.ID, Pembuat: e.CreatedBy, IdOrangtua: e.ID, IdPosyandu: e.IDPosyandu}
}

func sasaranKunjungan(e *entity.Kunjungan) Sasaran {
	return Sasaran{Jenis: "kunjungan", Id: e.ID, Pembuat: e.CreatedBy, IdAnak: e.IDAnak}
}

func sasaranKesehatan(e *entity.Kesehatan) Sasaran {
	return Sasaran{Jenis: "kesehatan", Id: e.ID, Pembuat: e.CreatedBy, IdAnak: e.IDAnak}
}

// satu pintu aturan akses semua jalur, tabelnya di docs/index.md#aturan-akses
func (s *StuntingService) Akses(orgid string, aksi Aksi, t Sasaran) error {
	if !utils.IsStrFilled(orgid) {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses", nil)
	}

	switch {
	case aksi == AksiHapus || (aksi == AksiUbah && milikPembuat(t.Jenis)):
		if !dibuatOleh(t.Pembuat, orgid) {
			return exceptions.Forbidden.Messagef("%s %s bukan milik Anda, hanya pembuatnya yang boleh mengubah atau menghapus", t.Jenis, t.Id)
		}
		return nil
	case orgid == pembuatJakantro:
		return nil
	case aksi == AksiBaca:
		return s.aturanBaca(orgid, t)
	}
	return s.hakAksesIdentitas(orgid, t)
}

// kunjungan beserta data medisnya, dan kesehatan, hanya diubah pembuatnya
func milikPembuat(jenis string) bool {
	return slices.Contains([]string{"kunjungan", "kesehatan", "observasi", "diagnosa", "layanan", "rujukan", "episode"}, jenis)
}

func dibuatOleh(createdBy *string, orgid string) bool {
	return utils.StrPtr(createdBy) == orgid
}

// aturan baca faskes; sementara sama dengan hak akses identitas
func (s *StuntingService) aturanBaca(orgid string, t Sasaran) error {
	return s.hakAksesIdentitas(orgid, t)
}

// faskes atas anak/orangtua: wilayah posyandu, rujukan aktif, atau kunjungan yang ia catat untuk anak tanpa orangtua
func (s *StuntingService) hakAksesIdentitas(orgid string, t Sasaran) error {
	idAnak, idOrangtua, idPosyandu := t.IdAnak, t.IdOrangtua, t.IdPosyandu

	if t.Jenis != "orangtua" {
		if idOrangtua == "" && idAnak != "" {
			a, err := s.Repository.FindAnak(&idAnak, nil, nil, nil, true)
			if err != nil {
				return exceptions.Classify(err)
			}
			idOrangtua = a.IDOrangtua
		}
		if utils.IsStrFilled(idOrangtua) {
			o, err := s.Repository.FindOrangTuaById(idOrangtua, true)
			if err != nil {
				return exceptions.Classify(err)
			}
			idPosyandu = o.IDPosyandu
		}
		idOrangtua = ""
	} else {
		idAnak = ""
	}

	errWilayah := s.aksesWilayah(orgid, idPosyandu)
	if errWilayah == nil {
		return nil
	}

	ada, err := s.Repository.AksesKlinis(orgid, idAnak, idOrangtua)
	if err != nil {
		return err
	}
	if ada {
		return nil
	}
	return errWilayah
}

// puskesmas/pustu atas posyandu di wilayah kerjanya
func (s *StuntingService) aksesWilayah(orgid string, idPosyandu *string) error {
	// null berarti belum terikat posyandu, beda dari posyandu tidak dikenal
	if !utils.IsFilled(idPosyandu) {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, orangtua belum terikat posyandu", nil)
	}

	posyandu, err := s.Repository.FindPosyanduDetail(*idPosyandu)
	if err != nil {
		if tidakDitemukan(err) {
			return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Posyandu Tidak ditemukan", nil)
		}
		return exceptions.Forbidden.WithMessage(err.Error(), err)
	}
	if posyandu == nil {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Posyandu Tidak ditemukan", nil)
	}
	if posyandu.Puskesmas == nil {
		return exceptions.Forbidden.WithMessage("Tidak bisa verifikasi akses, posyandu tidak terhubung dengan puskesmas", nil)
	}

	puskesmas := posyandu.Puskesmas
	if utils.StrPtr(puskesmas.SatusehatID) == orgid {
		return nil
	}
	if puskesmas.SatusehatID == nil && puskesmas.Induk != nil && utils.StrPtr(puskesmas.Induk.SatusehatID) == orgid {
		return nil
	}

	if !utils.IsFilled(puskesmas.Wilayah) {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Wilayah Faskes posyandu kosong", nil)
	}

	faskes, err := s.Repository.FindFaskesByOrgid(orgid)
	if err != nil {
		if tidakDitemukan(err) {
			return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Faskes user Tidak ditemukan", nil)
		}
		return exceptions.Forbidden.WithMessage(err.Error(), err)
	}
	if faskes == nil {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Faskes user Tidak ditemukan", nil)
	}

	// RS tidak memakai wilayah kerja, aksesnya lewat rujukan
	if slices.Contains([]string{"PUSTU", "PUSKESMAS"}, strings.ToUpper(faskes.Jenis)) {
		if !utils.IsFilled(faskes.Wilayah) {
			return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Wilayah Faskes user kosong", nil)
		}
		kecamatan := utils.Substr(*faskes.Wilayah, 0, 9) + "0000"
		switch entity.LevelWilayah(puskesmas.Wilayah) {
		case 3:
			if kecamatan == *puskesmas.Wilayah {
				return nil
			}
		case 4:
			if kecamatan == utils.Substr(*puskesmas.Wilayah, 0, 9)+"0000" {
				return nil
			}
		}
	}

	return exceptions.Forbidden.WithMessage("Tidak bisa akses data", nil)
}

// jakantro tidak dibatasi baca/ubah, jadi barisnya tidak perlu dimuat
func (s *StuntingService) AksesAnak(orgid string, aksi Aksi, idAnak string) error {
	if orgid == pembuatJakantro && aksi != AksiHapus {
		return nil
	}
	a, err := s.Repository.FindAnak(&idAnak, nil, nil, nil, true)
	if err != nil {
		return exceptions.Classify(err)
	}
	return s.Akses(orgid, aksi, sasaranAnak(a))
}

func (s *StuntingService) AksesOrangtua(orgid string, aksi Aksi, idOrangtua string) error {
	if orgid == pembuatJakantro && aksi != AksiHapus {
		return nil
	}
	o, err := s.Repository.FindOrangTuaById(idOrangtua, true)
	if err != nil {
		return exceptions.Classify(err)
	}
	return s.Akses(orgid, aksi, sasaranOrangtua(o))
}

// encounter yang dirujuk pemeriksaan harus kunjungan buatan pemanggil
func (s *StuntingService) AksesEncounter(orgid, refEncounter string, bolehBelumAda bool) error {
	k, err := s.Repository.FindKunjunganBySatusehatId(refEncounter)
	if err != nil {
		if !tidakDitemukan(err) {
			return err
		}
		if bolehBelumAda {
			return nil
		}
		return exceptions.Forbidden.Messagef("encounter %s belum tersimpan, kepemilikannya tidak bisa diverifikasi", refEncounter)
	}
	t := sasaranKunjungan(k)
	t.Id = refEncounter
	return s.Akses(orgid, AksiUbah, t)
}

// anak dicari lewat id_satusehat bila dikirim, selain itu lewat id
func (s *StuntingService) TemukanAnak(id, idSatusehat *string) (*entity.Anak, error) {
	if utils.IsFilled(idSatusehat) {
		a, err := s.Repository.FindAnakBySatusehatId(*idSatusehat, false)
		if err == nil {
			return a, nil
		}
		if !utils.IsFilled(id) || !tidakDitemukan(err) {
			return nil, exceptions.Classify(err)
		}
	}
	a, err := s.Repository.FindAnak(id, nil, nil, nil, false)
	if err != nil {
		return nil, exceptions.Classify(err)
	}
	return a, nil
}

func (s *StuntingService) TemukanOrangtua(id, idSatusehat *string) (*entity.Orangtua, error) {
	if utils.IsFilled(idSatusehat) {
		o, err := s.Repository.FindOrangtuaBySatusehatId(*idSatusehat, true)
		if err == nil {
			return o, nil
		}
		if !utils.IsFilled(id) || !tidakDitemukan(err) {
			return nil, exceptions.Classify(err)
		}
	}
	o, err := s.Repository.FindOrangTuaById(utils.StrPtr(id), true)
	if err != nil {
		return nil, exceptions.Classify(err)
	}
	return o, nil
}

func tidakDitemukan(err error) bool {
	return exceptions.Classify(err).ErrorCode == exceptions.TidakDitemukan.ErrorCode
}

// pembuat data medis adalah pembuat kunjungannya; episode milik faskes episode itu
func (s *StuntingService) pembuatDataMedis(jenis, satusehatId string) (*string, error) {
	memkey := "PEMBUAT::" + jenis + "::" + satusehatId
	if p := s.Repository.GetStringByKey(memkey); p != "" {
		return &p, nil
	}

	var pembuat *string
	var err error
	switch jenis {
	case "kunjungan":
		var k *entity.Kunjungan
		if k, err = s.Repository.FindKunjunganBySatusehatId(satusehatId); err == nil {
			pembuat = k.CreatedBy
		}
	case "observasi":
		var o *entity.Observasi
		if o, err = s.Repository.FindObservasiBySatusehatId(satusehatId); err == nil {
			pembuat, err = s.pembuatKunjungan(o.RefEncounter, o.IDKunjungan)
		}
	case "diagnosa":
		var d *entity.Diagnosa
		if d, err = s.Repository.FindDiagnosaBySatusehatId(satusehatId); err == nil {
			pembuat, err = s.pembuatKunjungan(d.RefEncounter, d.IDKunjungan)
		}
	case "layanan":
		var l *entity.Layanan
		if l, err = s.Repository.FindLayananBySatusehatId(satusehatId); err == nil {
			pembuat, err = s.pembuatKunjungan(l.RefEncounter, l.IDKunjungan)
		}
	case "rujukan":
		var r *entity.Rujukan
		if r, err = s.Repository.FindRujukanBySatusehatId(satusehatId); err == nil {
			pembuat, err = s.pembuatKunjungan(r.RefEncounter, r.IDKunjungan)
		}
	case "episode":
		var e *entity.Episode
		if e, err = s.Repository.FindEpisodeBySatusehatId(satusehatId, true); err == nil && e.Faskes != nil {
			pembuat = e.Faskes.SatusehatID
		}
	}
	if err != nil {
		return nil, exceptions.Classify(err)
	}

	if utils.IsFilled(pembuat) {
		s.Repository.SetStringByKey(memkey, *pembuat, 30*24*time.Hour)
	}
	return pembuat, nil
}

func (s *StuntingService) pembuatKunjungan(refEncounter, idKunjungan *string) (*string, error) {
	if utils.IsFilled(refEncounter) {
		if k, err := s.Repository.FindKunjunganBySatusehatId(*refEncounter); err == nil {
			return k.CreatedBy, nil
		}
	}
	if utils.IsFilled(idKunjungan) {
		if k, err := s.Repository.FindKunjunganById(*idKunjungan); err == nil {
			return k.CreatedBy, nil
		}
	}
	return nil, exceptions.Forbidden.Messagef("Tidak Bisa Verifikasi Akses: pemilik data tidak diketahui")
}
