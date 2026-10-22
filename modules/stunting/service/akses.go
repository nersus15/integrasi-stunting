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

// faskes atas anak/orangtua: pembuatnya, atau pemegang akses anak itu
func (s *StuntingService) hakAksesIdentitas(orgid string, t Sasaran) error {
	if dibuatOleh(t.Pembuat, orgid) {
		return nil
	}

	var boleh bool
	var err error
	if t.Jenis == "orangtua" {
		boleh, err = s.pemegangAksesOrangtua(orgid, t.IdOrangtua, t.IdPosyandu)
	} else {
		boleh, err = s.pemegangAksesAnak(orgid, t.IdAnak, t.IdOrangtua)
	}
	if err != nil {
		return err
	}
	if !boleh {
		return exceptions.Forbidden.Messagef("%s %s bukan buatan Anda dan tidak sedang dirujuk ke faskes Anda", t.Jenis, t.Id)
	}
	return nil
}

// tujuan rujukan terakhir; anak yang belum pernah dirujuk dipegang faskes posyandunya
func (s *StuntingService) pemegangAksesAnak(orgid, idAnak, idOrangtua string) (bool, error) {
	tujuan, dirujuk, err := s.Repository.TujuanRujukanTerakhir(idAnak)
	if err != nil {
		return false, err
	}
	if dirujuk {
		return tujuan == orgid, nil
	}

	if idOrangtua == "" {
		a, err := s.Repository.FindAnak(&idAnak, nil, nil, nil, true)
		if err != nil {
			return false, exceptions.Classify(err)
		}
		idOrangtua = a.IDOrangtua
	}
	if !utils.IsStrFilled(idOrangtua) {
		return false, nil
	}
	o, err := s.Repository.FindOrangTuaById(idOrangtua, true)
	if err != nil {
		return false, exceptions.Classify(err)
	}
	return s.faskesPosyandu(orgid, o.IDPosyandu)
}

// orangtua ikut anaknya; yang belum punya anak dipegang faskes posyandunya
func (s *StuntingService) pemegangAksesOrangtua(orgid, idOrangtua string, idPosyandu *string) (bool, error) {
	semuaAnak, err := s.Repository.IdAnakOrangtua(idOrangtua)
	if err != nil {
		return false, err
	}

	adaBelumDirujuk := len(semuaAnak) == 0
	for _, idAnak := range semuaAnak {
		tujuan, dirujuk, err := s.Repository.TujuanRujukanTerakhir(idAnak)
		if err != nil {
			return false, err
		}
		if dirujuk && tujuan == orgid {
			return true, nil
		}
		adaBelumDirujuk = adaBelumDirujuk || !dirujuk
	}
	if !adaBelumDirujuk {
		return false, nil
	}
	return s.faskesPosyandu(orgid, idPosyandu)
}

// puskesmas yang terikat posyandu, atau sekecamatan bila posyandu tidak terikat; RS sekab/kota
func (s *StuntingService) faskesPosyandu(orgid string, idPosyandu *string) (bool, error) {
	if !utils.IsFilled(idPosyandu) {
		return false, nil
	}
	posyandu, err := s.Repository.FindPosyanduDetail(*idPosyandu)
	if err != nil {
		if tidakDitemukan(err) {
			return false, nil
		}
		return false, err
	}
	faskes, err := s.Repository.FindFaskesByOrgid(orgid)
	if err != nil {
		if tidakDitemukan(err) {
			return false, nil
		}
		return false, err
	}
	if posyandu == nil || faskes == nil {
		return false, nil
	}

	switch jenisFaskes(faskes.Jenis) {
	case "rs":
		return samaWilayah(faskes.Wilayah, posyandu.IDKelurahan, digitKabKota), nil
	case "puskesmas":
		if p := posyandu.Puskesmas; p != nil {
			if p.SatusehatID == nil && p.Induk != nil {
				p = p.Induk
			}
			return utils.StrPtr(p.SatusehatID) == orgid, nil
		}
		return samaWilayah(faskes.Wilayah, posyandu.IDKelurahan, digitKecamatan), nil
	}
	return false, nil
}

const (
	digitKabKota   = 4
	digitKecamatan = 6
)

func jenisFaskes(jenis string) string {
	jenis = strings.ToLower(strings.TrimSpace(jenis))
	switch {
	case strings.Contains(jenis, "puskesmas"), jenis == "pustu":
		return "puskesmas"
	case strings.Contains(jenis, "rs"):
		return "rs"
	}
	return jenis
}

// kode wilayah dibandingkan sebagai digit, titik diabaikan
func samaWilayah(a, b *string, digit int) bool {
	x, y := digitWilayah(a), digitWilayah(b)
	return len(x) >= digit && len(y) >= digit && x[:digit] == y[:digit]
}

func digitWilayah(kode *string) string {
	return strings.ReplaceAll(strings.TrimSpace(utils.StrPtr(kode)), ".", "")
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
