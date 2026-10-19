package validasi

import (
	"fmt"
	"strings"

	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
)

func ValidateCariOrangtua(id, nik, nokk, namaAyah, namaIbu string) error {
	if !utils.IsStrFilled(id) && !utils.IsStrFilled(nik) && !utils.IsStrFilled(nokk) &&
		!utils.IsStrFilled(namaAyah) && !utils.IsStrFilled(namaIbu) {
		return fmt.Errorf("sebutkan minimal satu kata kunci (id | nik | nokk | nama_ayah | nama_ibu)")
	}

	return nil
}

func ValidateCariAnak(id string) error {
	if utils.IsStrFilled(id) {
		return nil
	}

	return fmt.Errorf("Id Anak Tidak Valid")
}
func ValidateOrangtua(p types.OrangtuaPayload) error {
	// id datang dari jakantro, bukan digenerate server
	return pertama(wajib(p.Id, "id"), wajib(p.IdPosyandu, "id_posyandu"), cekNoKk(p), cekNamaOrangtua(p, "harus diisi"),
		cekNikOrangtua(p), wajib(p.Telepon, "telepon"), wajib(p.Rt, "rt"), wajib(p.Rw, "rw"), wajib(p.Alamat, "alamat"),
		cekFlagOrangtua(p))
}

func ValidateOrangtuaFaskes(p types.OrangtuaPayload) error {
	return pertama(cekNamaOrangtua(p, "harus diisi"), cekNoKk(p), cekNikOrangtua(p), wajib(p.Telepon, "telepon"), cekFlagOrangtua(p))
}

func ValidateAnakUpdate(p types.AnakPayload) error {
	if !utils.IsStrFilled(p.Id) {
		return fmt.Errorf("id tidak boleh kosong")
	}
	if utils.IsFilled(p.NIK) && strings.TrimSpace(*p.NIK) != "" && !utils.IsDigitsLen(*p.NIK, 16) {
		return fmt.Errorf("nik harus 16 digit angka")
	}
	if utils.IsStrFilled(p.TanggalLahir) && !utils.IsValidDate(p.TanggalLahir) {
		return fmt.Errorf("tanggal_lahir harus berformat YYYY-MM-DD")
	}
	if jk := strings.ToUpper(strings.TrimSpace(p.JenisKelamin)); jk != "" && jk != "L" && jk != "P" {
		return fmt.Errorf("jenis_kelamin harus 'L' atau 'P'")
	}
	if p.AnakKe < 0 {
		return fmt.Errorf("anak_ke tidak boleh negatif")
	}
	if p.IMD != nil && *p.IMD != 0 && *p.IMD != 1 {
		return fmt.Errorf("imd harus bernilai 0 atau 1")
	}
	if p.BBLahir < 0 || p.TBLahir < 0 || p.LKLahir < 0 {
		return fmt.Errorf("bb_lahir, tb_lahir, dan lk_lahir tidak boleh negatif")
	}

	return nil
}

// versi longgar untuk data dari Kafka
func ValidateOrangtuaUpdate(p types.OrangtuaPayload) error {
	return pertama(wajib(p.Id, "id"), cekNoKk(p), cekNamaOrangtua(p, "harus diisi"), cekNikOrangtua(p), cekFlagOrangtua(p))
}

func ValidateOrangtuaStream(p types.OrangtuaPayload) error {
	return pertama(wajib(p.Id, "id"), cekNikOrangtua(p), cekNamaOrangtua(p, "harus terisi salah satu"))
}

// identitas anak, berlaku untuk semua sumber data
func validateAnakDasar(p types.AnakPayload) error {
	if !utils.IsStrFilled(p.Id) {
		return fmt.Errorf("id tidak boleh kosong")
	}
	if !utils.IsStrFilled(p.Nama) {
		return fmt.Errorf("nama tidak boleh kosong")
	}

	if utils.IsFilled(p.NIK) && !utils.IsDigitsLen(*p.NIK, 16) {
		return fmt.Errorf("nik harus 16 digit angka")
	}

	if !utils.IsStrFilled(p.TanggalLahir) {
		return fmt.Errorf("tanggal_lahir tidak boleh kosong")
	} else if !utils.IsValidDate(p.TanggalLahir) {
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
	return pertama(validateAnakDasar(p), wajib(p.IDOrangtua, "id_orangtua"), cekLahirAnak(p), wajib(p.SourceData, "source_data"))
}

func ValidateAnakFaskes(p types.AnakPayload) error {
	return pertama(validateAnakDasar(p), cekLahirAnak(p))
}

// versi longgar untuk data dari Kafka
func ValidateAnakStream(p types.AnakPayload) error {
	return validateAnakDasar(p)
}

// pertama mengembalikan galat pertama sesuai urutan pemeriksaan
func pertama(galat ...error) error {
	for _, g := range galat {
		if g != nil {
			return g
		}
	}
	return nil
}

func wajib(nilai, kolom string) error {
	if !utils.IsStrFilled(nilai) {
		return fmt.Errorf("%s tidak boleh kosong", kolom)
	}
	return nil
}

func cekNikOrangtua(p types.OrangtuaPayload) error {
	if !utils.IsStrFilled(p.Nik) {
		return fmt.Errorf("nik tidak boleh kosong")
	}
	if !utils.IsDigitsLen(p.Nik, 16) {
		return fmt.Errorf("nik harus 16 digit angka")
	}
	return nil
}

func cekNoKk(p types.OrangtuaPayload) error {
	if utils.IsStrFilled(p.NoKk) && !utils.IsDigitsLen(p.NoKk, 16) {
		return fmt.Errorf("no_kk harus 16 digit angka")
	}
	return nil
}

func cekNamaOrangtua(p types.OrangtuaPayload, pesan string) error {
	if !utils.IsStrFilled(p.NamaAyah) && !utils.IsStrFilled(p.NamaIbu) {
		return fmt.Errorf("nama_ayah atau nama_ibu %s", pesan)
	}
	return nil
}

func cekFlagOrangtua(p types.OrangtuaPayload) error {
	if p.Kia != nil && *p.Kia != 0 && *p.Kia != 1 {
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

// data lahir wajib lengkap untuk anak yang didaftarkan jakantro atau faskes
func cekLahirAnak(p types.AnakPayload) error {
	if p.AnakKe <= 0 {
		return fmt.Errorf("anak_ke harus lebih dari 0")
	}
	if p.IMD != nil && *p.IMD != 0 && *p.IMD != 1 {
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
	return nil
}
