package utils

import (
	"fmt"
	"strings"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
)

func ValidateCariOrangtua(id, nik, nokk, namaAyah, namaIbu string) error {
	if !IsStrFilled(id) && !IsStrFilled(nik) && !IsStrFilled(nokk) &&
		!IsStrFilled(namaAyah) && !IsStrFilled(namaIbu) {
		return fmt.Errorf("sebutkan minimal satu kata kunci (id | nik | nokk | nama_ayah | nama_ibu)")
	}

	return nil
}

func ValidateCariAnak(id string) error {
	if IsStrFilled(id) {
		return nil
	}

	return fmt.Errorf("Id Anak Tidak Valid")
}
func ValidateOrangtua(p types.OrangtuaPayload) error {
	// id datang dari jakantro, bukan digenerate server
	if !IsStrFilled(p.Id) {
		return fmt.Errorf("id tidak boleh kosong")
	}

	if !IsStrFilled(p.IdPosyandu) {
		return fmt.Errorf("id_posyandu tidak boleh kosong")
	}

	if !IsStrFilled(p.NoKk) {
		return fmt.Errorf("no_kk tidak boleh kosong")
	} else if !IsDigitsLen(p.NoKk, 16) {
		return fmt.Errorf("no_kk harus 16 digit angka")
	}

	if !IsStrFilled(p.NamaAyah) {
		return fmt.Errorf("nama_ayah tidak boleh kosong")
	}
	if !IsStrFilled(p.NamaIbu) {
		return fmt.Errorf("nama_ibu tidak boleh kosong")
	}

	if !IsStrFilled(p.Nik) {
		return fmt.Errorf("nik tidak boleh kosong")
	} else if !IsDigitsLen(p.Nik, 16) {
		return fmt.Errorf("nik harus 16 digit angka")
	}

	if !IsStrFilled(p.Telepon) {
		return fmt.Errorf("telepon tidak boleh kosong")
	}
	if !IsStrFilled(p.Rt) {
		return fmt.Errorf("rt tidak boleh kosong")
	}
	if !IsStrFilled(p.Rw) {
		return fmt.Errorf("rw tidak boleh kosong")
	}
	if !IsStrFilled(p.Alamat) {
		return fmt.Errorf("alamat tidak boleh kosong")
	}

	if p.Kia != 0 && p.Kia != 1 {
		return fmt.Errorf("kia harus bernilai 0 atau 1")
	}

	if p.UsiaHamil != nil && *p.UsiaHamil < 0 {
		return fmt.Errorf("usia_hamil tidak boleh bernilai negatif")
	}
	if p.KiaBayiKecil != nil && *p.KiaBayiKecil != 0 && *p.KiaBayiKecil != 1 {
		return fmt.Errorf("kia_bayi_kecil harus bernilai 0 atau 1")
	}

	return nil
}

// versi longgar untuk data dari Kafka
func ValidateOrangtuaStream(p types.OrangtuaPayload) error {
	if !IsStrFilled(p.Id) {
		return fmt.Errorf("id tidak boleh kosong")
	}

	if !IsStrFilled(p.Nik) {
		return fmt.Errorf("nik tidak boleh kosong")
	} else if !IsDigitsLen(p.Nik, 16) {
		return fmt.Errorf("nik harus 16 digit angka")
	}

	if !IsStrFilled(p.NamaAyah) && !IsStrFilled(p.NamaIbu) {
		return fmt.Errorf("nama_ayah atau nama_ibu harus terisi salah satu")
	}

	return nil
}

// identitas anak, berlaku untuk semua sumber data
func validateAnakDasar(p types.AnakPayload) error {
	if !IsStrFilled(p.Id) {
		return fmt.Errorf("id tidak boleh kosong")
	}
	if !IsStrFilled(p.Nama) {
		return fmt.Errorf("nama tidak boleh kosong")
	}

	if IsFilled(p.NIK) && !IsDigitsLen(*p.NIK, 16) {
		return fmt.Errorf("nik harus 16 digit angka")
	}

	if !IsStrFilled(p.TanggalLahir) {
		return fmt.Errorf("tanggal_lahir tidak boleh kosong")
	} else if !IsValidDate(p.TanggalLahir) {
		return fmt.Errorf("tanggal_lahir harus berformat YYYY-MM-DD")
	}

	jk := strings.ToUpper(strings.TrimSpace(p.JenisKelamin))
	if jk != "L" && jk != "P" {
		return fmt.Errorf("jenis_kelamin harus 'L' atau 'P'")
	}

	return nil
}

// versi ketat untuk data dari jakantro
func ValidateAnak(p types.AnakPayload) error {
	if err := validateAnakDasar(p); err != nil {
		return err
	}

	if !IsStrFilled(p.IDOrangtua) {
		return fmt.Errorf("id_orangtua tidak boleh kosong")
	}

	if p.AnakKe <= 0 {
		return fmt.Errorf("anak_ke harus lebih dari 0")
	}
	if p.IMD != 0 && p.IMD != 1 {
		return fmt.Errorf("imd harus bernilai 0 atau 1")
	}

	if p.BBLahir <= 0 {
		return fmt.Errorf("bb_lahir harus lebih dari 0")
	}
	if p.TBLahir <= 0 {
		return fmt.Errorf("tb_lahir harus lebih dari 0")
	}
	if p.LKLahir <= 0 {
		return fmt.Errorf("lk_lahir harus lebih dari 0")
	}

	if !IsStrFilled(p.SourceData) {
		return fmt.Errorf("source_data tidak boleh kosong")
	}

	return nil
}

// versi longgar untuk data dari Kafka
func ValidateAnakStream(p types.AnakPayload) error {
	return validateAnakDasar(p)
}

func ValidateKesehatan(p types.KesehatanPayload) error {
	if !IsStrFilled(p.Id) {
		return fmt.Errorf("id tidak boleh kosong")
	}
	if !IsStrFilled(p.IDAnak) {
		return fmt.Errorf("id_anak tidak boleh kosong")
	}

	if !IsStrFilled(p.TanggalPemantauan) {
		return fmt.Errorf("tanggal_pemantauan tidak boleh kosong")
	}
	if !IsValidDate(p.TanggalPemantauan) {
		return fmt.Errorf("tanggal_pemantauan harus berformat YYYY-MM-DD")
	}

	// Field layanan & tbc semuanya nullable di DB, validasi hanya jika pointer diisi (!= nil)
	flags := map[string]*int16{
		"tbc_batuk":           p.TBCBatuk,
		"tbc_demam":           p.TBCDemam,
		"tbc_bb":              p.TBCBB,
		"tbc_kontak":          p.TBCKontak,
		"layanan_asi_eks":     p.LayananASIEks,
		"layanan_mpasi":       p.LayananMPASI,
		"layanan_imunisasi":   p.LayananImunisasi,
		"layanan_vit_a":       p.LayananVitA,
		"layanan_obat_cacing": p.LayananObatCacing,
		"layanan_mt_pangan":   p.LayananMTPangan,
		"penyuluhan_edukasi":  p.PenyuluhanEdukasi,
		"penyuluhan_rujukan":  p.PenyuluhanRujukan,
	}

	for fieldName, val := range flags {
		if val != nil && *val != 0 && *val != 1 {
			return fmt.Errorf("%s harus bernilai 0 atau 1", fieldName)
		}
	}

	return nil
}

func ValidateKunjungan(p types.KunjunganPayload) error {
	if !IsStrFilled(p.Id) {
		return fmt.Errorf("id tidak boleh kosong")
	}
	if !IsStrFilled(p.IDAnak) {
		return fmt.Errorf("id_anak tidak boleh kosong")
	}

	if !IsStrFilled(p.TanggalPengukuran) {
		return fmt.Errorf("tanggal_pengukuran tidak boleh kosong")
	}
	if !IsValidDate(p.TanggalPengukuran) {
		return fmt.Errorf("tanggal_pengukuran harus berformat YYYY-MM-DD")
	}

	// Semua indikator fisik bernilai nullable di DB, validasi angka > 0 dilakukan jika data dikirim (!= nil)
	if p.BeratBadan != nil && *p.BeratBadan <= 0 {
		return fmt.Errorf("berat_badan harus lebih dari 0")
	}
	if p.TinggiBadan != nil && *p.TinggiBadan <= 0 {
		return fmt.Errorf("tinggi_badan harus lebih dari 0")
	}
	if p.LingkarKepala != nil && *p.LingkarKepala <= 0 {
		return fmt.Errorf("lingkar_kepala harus lebih dari 0")
	}
	if p.LingkarLengan != nil && *p.LingkarLengan <= 0 {
		return fmt.Errorf("lingkar_lengan harus lebih dari 0")
	}
	if p.LingkarDada != nil && *p.LingkarDada <= 0 {
		return fmt.Errorf("lingkar_dada harus lebih dari 0")
	}

	asiFlags := map[string]*int16{
		"asi_bulan_0":      p.ASIBulan0,
		"asi_bulan_1":      p.ASIBulan1,
		"asi_bulan_2":      p.ASIBulan2,
		"asi_bulan_3":      p.ASIBulan3,
		"asi_bulan_4":      p.ASIBulan4,
		"asi_bulan_5":      p.ASIBulan5,
		"asi_bulan_6":      p.ASIBulan6,
		"vit_biru":         p.VitBiru,
		"vit_merah":        p.VitMerah,
		"pitting_edema":    p.PittingEdema,
		"kelas_ibu_balita": p.KelasIbuBalita,
	}

	for fieldName, val := range asiFlags {
		if val != nil && *val != 0 && *val != 1 {
			return fmt.Errorf("%s harus bernilai 0 atau 1", fieldName)
		}
	}

	return nil
}

func ValidateWilayahKerja(org string, data any) error {

	return nil
}

func ValidatePemeriksaanFaskes(p *types.PemeriksaanFaskes) error {
	if p == nil {
		return fmt.Errorf("payload kosong")
	}

	if !IsFilled(p.Anak.IdSatusehat) && !IsFilled(p.Anak.Nik) {
		return fmt.Errorf("anak.satusehat_id atau anak.nik wajib diisi salah satu")
	}
	if IsFilled(p.Anak.Nik) && !IsDigitsLen(*p.Anak.Nik, 16) {
		return fmt.Errorf("anak.nik harus 16 digit angka")
	}

	// faskes tidak menentukan id; pencocokan lewat satusehat id
	if err := ValidateKunjunganFaskes(*p.Kunjungan.ToPayload()); err != nil {
		return err
	}

	for i, o := range p.Observasi {
		if err := ValidateObservasiFaskes(o); err != nil {
			return fmt.Errorf("observasi[%d]: %s", i, err.Error())
		}
	}
	for i, d := range p.Diagnosa {
		if err := ValidateDiagnosaFaskes(d); err != nil {
			return fmt.Errorf("diagnosa[%d]: %s", i, err.Error())
		}
	}
	for i, l := range p.Layanan {
		if err := ValidateLayananFaskes(l); err != nil {
			return fmt.Errorf("layanan[%d]: %s", i, err.Error())
		}
	}
	for i, r := range p.Rujukan {
		if err := ValidateRujukanFaskes(r); err != nil {
			return fmt.Errorf("rujukan[%d]: %s", i, err.Error())
		}
	}
	for i, e := range p.Episode {
		if err := ValidateEpisodeFaskes(e); err != nil {
			return fmt.Errorf("episode[%d]: %s", i, err.Error())
		}
	}

	return nil
}

func ValidateKunjunganFaskes(kunjungan types.KunjunganPayload) error {
	if !IsFilled(kunjungan.SatusehatId) {
		return fmt.Errorf("kunjungan.id_satusehat wajib dikirim, kirim setelah data diterima SatuSehat")
	}
	if !IsStrFilled(kunjungan.TanggalPengukuran) {
		return fmt.Errorf("kunjungan.tanggal_pengukuran wajib dikirim")
	}
	if !IsValidDate(kunjungan.TanggalPengukuran) {
		return fmt.Errorf("kunjungan.tanggal_pengukuran (%q) harus berformat YYYY-MM-DD", kunjungan.TanggalPengukuran)
	}
	if kunjungan.TanggalSelesai != nil && !IsValidDate(*kunjungan.TanggalSelesai) {
		return fmt.Errorf("kunjungan.tanggal_selesai (%q) harus berformat YYYY-MM-DD", *kunjungan.TanggalSelesai)
	}

	return nil
}

func ValidateDiagnosaFaskes(diagnosa types.Diagnosa) error {
	if !IsFilled(diagnosa.IdSatusehat) {
		return fmt.Errorf("id_satusehat wajib dikirim, kirim setelah data diterima SatuSehat")
	}
	if !IsStrFilled(diagnosa.Kode) {
		return fmt.Errorf("kode wajib dikirim")
	}
	if diagnosa.Jenis != "" && diagnosa.Jenis != DiagnosaDiagnosis && diagnosa.Jenis != DiagnosaAlergi {
		return fmt.Errorf("jenis harus %q atau %q", DiagnosaDiagnosis, DiagnosaAlergi)
	}

	return nil
}

// Validator per resource jalur faskes. id_satusehat diwajibkan di semua:
// ia satu-satunya kunci pencocokan, dan tanpa itu baris yang tersimpan tidak
// akan pernah bisa diperbarui maupun dicegah berganda saat kiriman ulang.
//
// Jalur stream tidak memakai validator ini dan tidak terpengaruh: entry yang
// tidak mendapat resourceID dari SatuSehat memang sudah dilewati sebelum
// sampai ke pemetaan.

func ValidateObservasiFaskes(o types.Observasi) error {
	if !IsFilled(o.IdSatusehat) {
		return fmt.Errorf("id_satusehat wajib dikirim, kirim setelah data diterima SatuSehat")
	}
	if !IsStrFilled(o.Kode) {
		return fmt.Errorf("kode wajib dikirim")
	}
	if !IsStrFilled(o.System) {
		return fmt.Errorf("system wajib dikirim, supaya kodenya bisa dipetakan")
	}

	for i, c := range o.Component {
		if err := ValidateObservasiFaskes(c); err != nil {
			return fmt.Errorf("component[%d]: %s", i, err.Error())
		}
	}

	return nil
}

func ValidateLayananFaskes(l types.Layanan) error {
	if !IsFilled(l.IdSatusehat) {
		return fmt.Errorf("id_satusehat wajib dikirim, kirim setelah data diterima SatuSehat")
	}
	if !IsStrFilled(l.Jenis) {
		return fmt.Errorf("jenis wajib dikirim")
	}

	switch l.Jenis {
	case entity.LayananProcedure, entity.LayananMedicationDispense,
		entity.LayananNutritionOrder, entity.LayananImmunization, entity.LayananServiceRequest:
	default:
		return fmt.Errorf("jenis %q tidak dikenali", l.Jenis)
	}

	return nil
}

// jenis boleh kosong: arahnya disimpulkan dari faskes asal dan tujuan
func ValidateRujukanFaskes(r types.Rujukan) error {
	if !IsFilled(r.IdSatusehat) {
		return fmt.Errorf("id_satusehat wajib dikirim, kirim setelah data diterima SatuSehat")
	}

	switch r.Jenis {
	case "", entity.RujukanKeluar, entity.RujukBalik, entity.RujukanInternal:
	default:
		return fmt.Errorf("jenis harus %q, %q, atau %q", entity.RujukanKeluar, entity.RujukBalik, entity.RujukanInternal)
	}

	return nil
}

func ValidateEpisodeFaskes(e types.Episode) error {
	if !IsFilled(e.IdSatusehat) {
		return fmt.Errorf("id_satusehat wajib dikirim, kirim setelah data diterima SatuSehat")
	}

	return nil
}
