package validator_test

import (
	"strings"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/validasi"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- POSITIVE CASE ---

func TestValidateOrangtua_PayloadValid(t *testing.T) {
	assert.NoError(t, validasi.ValidateOrangtua(testutil.OrangtuaPayloadValid()))
}

func TestValidateOrangtua_FieldOpsionalBolehKosong(t *testing.T) {
	tests := []struct {
		nama string
		ubah func(*types.OrangtuaPayload)
	}{
		{"tanpa usia_hamil", func(p *types.OrangtuaPayload) { p.UsiaHamil = nil }},
		{"tanpa kia_bayi_kecil", func(p *types.OrangtuaPayload) { p.KiaBayiKecil = nil }},
		{"tanpa source_data", func(p *types.OrangtuaPayload) { p.SourceData = nil }},
		{"kia bernilai 0", func(p *types.OrangtuaPayload) { p.Kia = testutil.Ptr(int16(0)) }},
		{"tanpa no_kk", func(p *types.OrangtuaPayload) { p.NoKk = "" }},
		{"hanya nama_ayah", func(p *types.OrangtuaPayload) { p.NamaIbu = "" }},
		{"hanya nama_ibu", func(p *types.OrangtuaPayload) { p.NamaAyah = "" }},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			p := testutil.OrangtuaPayloadValid()
			tc.ubah(&p)
			assert.NoError(t, validasi.ValidateOrangtua(p))
		})
	}
}

func TestValidateAnak_PayloadValid(t *testing.T) {
	assert.NoError(t, validasi.ValidateAnak(testutil.AnakPayloadValid()))
}

func TestValidateAnak_JenisKelaminHurufKecil(t *testing.T) {
	p := testutil.AnakPayloadValid()
	p.JenisKelamin = "l"

	assert.NoError(t, validasi.ValidateAnak(p), "jenis_kelamin tidak peka huruf besar-kecil")
}

func TestValidateKunjungan_PayloadValid(t *testing.T) {
	assert.NoError(t, validasi.ValidateKunjungan(testutil.KunjunganPayloadValid()))
}

func TestValidateKunjungan_PengukuranOpsionalBolehNil(t *testing.T) {
	p := testutil.KunjunganPayloadValid()
	p.BeratBadan, p.TinggiBadan = nil, nil
	p.LingkarLengan, p.LingkarKepala, p.LingkarDada = nil, nil, nil

	assert.NoError(t, validasi.ValidateKunjungan(p), "indikator fisik nullable di skema")
}

func TestValidateKesehatan_PayloadValid(t *testing.T) {
	assert.NoError(t, validasi.ValidateKesehatan(testutil.KesehatanPayloadValid()))
}

// --- NEGATIVE CASE: orangtua ---

func TestValidateOrangtua_FieldWajibKosong(t *testing.T) {
	tests := []struct {
		nama  string
		field string
		ubah  func(*types.OrangtuaPayload)
	}{
		{"tanpa id", "id", func(p *types.OrangtuaPayload) { p.Id = "" }},
		{"tanpa id_posyandu", "id_posyandu", func(p *types.OrangtuaPayload) { p.IdPosyandu = "" }},
		{"tanpa nik", "nik", func(p *types.OrangtuaPayload) { p.Nik = "" }},
		{"tanpa telepon", "telepon", func(p *types.OrangtuaPayload) { p.Telepon = "" }},
		{"tanpa rt", "rt", func(p *types.OrangtuaPayload) { p.Rt = "" }},
		{"tanpa rw", "rw", func(p *types.OrangtuaPayload) { p.Rw = "" }},
		{"tanpa alamat", "alamat", func(p *types.OrangtuaPayload) { p.Alamat = "" }},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			p := testutil.OrangtuaPayloadValid()
			tc.ubah(&p)

			err := validasi.ValidateOrangtua(p)

			require.Error(t, err, "%s wajib diisi", tc.field)
			assert.Contains(t, err.Error(), tc.field,
				"pesan error harus menyebut field yang kurang")
		})
	}
}

func TestValidateOrangtua_NamaAyahDanIbuKosongDitolak(t *testing.T) {
	p := testutil.OrangtuaPayloadValid()
	p.NamaAyah, p.NamaIbu = "", "  "

	err := validasi.ValidateOrangtua(p)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nama_ayah")
	assert.Contains(t, err.Error(), "nama_ibu")
}

func TestValidateOrangtua_NikDanNoKkHarusEnambelasDigit(t *testing.T) {
	tests := []struct {
		nama  string
		nilai string
	}{
		{"terlalu pendek", "35150812"},
		{"terlalu panjang", "35150812038000021"},
		{"berisi huruf", "abcdefghijklmnop"},
		{"berisi spasi", "3515 0812 038000"},
		{"berisi tanda baca", "3515-0812-038000"},
	}

	for _, tc := range tests {
		t.Run("nik "+tc.nama, func(t *testing.T) {
			p := testutil.OrangtuaPayloadValid()
			p.Nik = tc.nilai

			require.Error(t, validasi.ValidateOrangtua(p), "nik harus 16 digit angka")
		})

		t.Run("no_kk "+tc.nama, func(t *testing.T) {
			p := testutil.OrangtuaPayloadValid()
			p.NoKk = tc.nilai

			require.Error(t, validasi.ValidateOrangtua(p), "no_kk harus 16 digit angka")
		})
	}
}

func TestValidateOrangtua_NilaiFlagDiluarNolSatu(t *testing.T) {
	p := testutil.OrangtuaPayloadValid()
	p.Kia = testutil.Ptr(int16(5))
	require.Error(t, validasi.ValidateOrangtua(p))

	p = testutil.OrangtuaPayloadValid()
	p.KiaBayiKecil = testutil.Ptr(int16(2))
	require.Error(t, validasi.ValidateOrangtua(p))

	p = testutil.OrangtuaPayloadValid()
	p.UsiaHamil = testutil.Ptr(int16(-1))
	require.Error(t, validasi.ValidateOrangtua(p))
}

// --- NEGATIVE CASE: anak ---

func TestValidateAnak_FieldWajibKosong(t *testing.T) {
	tests := []struct {
		nama  string
		field string
		ubah  func(*types.AnakPayload)
	}{
		{"tanpa id", "id", func(p *types.AnakPayload) { p.Id = "" }},
		{"tanpa id_orangtua", "id_orangtua", func(p *types.AnakPayload) { p.IDOrangtua = "" }},
		{"tanpa nama", "nama", func(p *types.AnakPayload) { p.Nama = "" }},
		{"tanpa tanggal_lahir", "tanggal_lahir", func(p *types.AnakPayload) { p.TanggalLahir = "" }},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			p := testutil.AnakPayloadValid()
			tc.ubah(&p)

			err := validasi.ValidateAnak(p)

			require.Error(t, err, "%s wajib diisi", tc.field)
			assert.Contains(t, err.Error(), tc.field)
		})
	}
}

// nik anak opsional: kolomnya nullable, jadi payload tanpa nik harus lolos.
// Bila diisi, formatnya tetap wajib 16 digit.
func TestValidateAnak_NikOpsional(t *testing.T) {
	p := testutil.AnakPayloadValid()
	p.NIK = nil
	assert.NoError(t, validasi.ValidateAnak(p), "nik tidak dikirim harus diterima")

	kosong := ""
	p.NIK = &kosong
	assert.NoError(t, validasi.ValidateAnak(p), "nik string kosong harus diterima")

	pendek := "123"
	p.NIK = &pendek
	err := validasi.ValidateAnak(p)
	require.Error(t, err, "nik yang diisi harus 16 digit")
	assert.Contains(t, err.Error(), "nik")
}

func TestValidateAnak_NilaiTidakMasukAkal(t *testing.T) {
	tests := []struct {
		nama string
		ubah func(*types.AnakPayload)
	}{
		{"tanggal_lahir bukan YYYY-MM-DD", func(p *types.AnakPayload) { p.TanggalLahir = "10-05-2024" }},
		{"jenis_kelamin selain L atau P", func(p *types.AnakPayload) { p.JenisKelamin = "X" }},
		{"anak_ke nol", func(p *types.AnakPayload) { p.AnakKe = 0 }},
		{"anak_ke negatif", func(p *types.AnakPayload) { p.AnakKe = -1 }},
		{"imd selain 0 atau 1", func(p *types.AnakPayload) { p.IMD = testutil.Ptr(int16(3)) }},
		{"bb_lahir nol", func(p *types.AnakPayload) { p.BBLahir = 0 }},
		{"tb_lahir negatif", func(p *types.AnakPayload) { p.TBLahir = -1 }},
		{"lk_lahir nol", func(p *types.AnakPayload) { p.LKLahir = 0 }},
		{"nik bukan 16 digit", func(p *types.AnakPayload) { n := "351508"; p.NIK = &n }},
		{"nik 16 huruf", func(p *types.AnakPayload) { n := "abcdefghijklmnop"; p.NIK = &n }},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			p := testutil.AnakPayloadValid()
			tc.ubah(&p)

			require.Error(t, validasi.ValidateAnak(p))
		})
	}
}

// --- NEGATIVE CASE: kunjungan ---

func TestValidateKunjungan_FieldWajibKosong(t *testing.T) {
	tests := []struct {
		nama  string
		field string
		ubah  func(*types.KunjunganPayload)
	}{
		{"tanpa id", "id", func(p *types.KunjunganPayload) { p.Id = "" }},
		{"tanpa id_anak", "id_anak", func(p *types.KunjunganPayload) { p.IDAnak = "" }},
		{"tanpa tanggal_pengukuran", "tanggal_pengukuran", func(p *types.KunjunganPayload) { p.TanggalPengukuran = "" }},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			p := testutil.KunjunganPayloadValid()
			tc.ubah(&p)

			err := validasi.ValidateKunjungan(p)

			require.Error(t, err, "%s wajib diisi", tc.field)
			assert.Contains(t, err.Error(), tc.field)
		})
	}
}

func TestValidateKunjungan_NilaiTidakMasukAkal(t *testing.T) {
	tests := []struct {
		nama string
		ubah func(*types.KunjunganPayload)
	}{
		{"tanggal bukan YYYY-MM-DD", func(p *types.KunjunganPayload) { p.TanggalPengukuran = "16-02-2025" }},
		{"berat_badan nol", func(p *types.KunjunganPayload) { p.BeratBadan = testutil.Ptr(0.0) }},
		{"berat_badan negatif", func(p *types.KunjunganPayload) { p.BeratBadan = testutil.Ptr(-1.0) }},
		{"tinggi_badan nol", func(p *types.KunjunganPayload) { p.TinggiBadan = testutil.Ptr(0.0) }},
		{"lingkar_lengan negatif", func(p *types.KunjunganPayload) { p.LingkarLengan = testutil.Ptr(-1.0) }},
		{"lingkar_kepala nol", func(p *types.KunjunganPayload) { p.LingkarKepala = testutil.Ptr(0.0) }},
		{"lingkar_dada negatif", func(p *types.KunjunganPayload) { p.LingkarDada = testutil.Ptr(-2.0) }},
		{"asi_bulan_0 selain 0/1", func(p *types.KunjunganPayload) { p.ASIBulan0 = testutil.Ptr(int16(5)) }},
		{"vit_biru selain 0/1", func(p *types.KunjunganPayload) { p.VitBiru = testutil.Ptr(int16(2)) }},
		{"pitting_edema selain 0/1", func(p *types.KunjunganPayload) { p.PittingEdema = testutil.Ptr(int16(9)) }},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			p := testutil.KunjunganPayloadValid()
			tc.ubah(&p)

			require.Error(t, validasi.ValidateKunjungan(p))
		})
	}
}

// --- NEGATIVE CASE: kesehatan ---

func TestValidateKesehatan_FieldWajibKosong(t *testing.T) {
	tests := []struct {
		nama  string
		field string
		ubah  func(*types.KesehatanPayload)
	}{
		{"tanpa id", "id", func(p *types.KesehatanPayload) { p.Id = "" }},
		{"tanpa id_anak", "id_anak", func(p *types.KesehatanPayload) { p.IDAnak = "" }},
		{"tanpa tanggal_pemantauan", "tanggal_pemantauan", func(p *types.KesehatanPayload) { p.TanggalPemantauan = "" }},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			p := testutil.KesehatanPayloadValid()
			tc.ubah(&p)

			err := validasi.ValidateKesehatan(p)

			require.Error(t, err, "%s wajib diisi", tc.field)
			assert.Contains(t, err.Error(), tc.field)
		})
	}
}

func TestValidateKesehatan_FlagDiluarNolSatu(t *testing.T) {
	flags := []struct {
		nama string
		ubah func(*types.KesehatanPayload)
	}{
		{"tbc_batuk", func(p *types.KesehatanPayload) { p.TBCBatuk = testutil.Ptr(int16(3)) }},
		{"tbc_demam", func(p *types.KesehatanPayload) { p.TBCDemam = testutil.Ptr(int16(3)) }},
		{"layanan_imunisasi", func(p *types.KesehatanPayload) { p.LayananImunisasi = testutil.Ptr(int16(7)) }},
		{"penyuluhan_rujukan", func(p *types.KesehatanPayload) { p.PenyuluhanRujukan = testutil.Ptr(int16(2)) }},
	}

	for _, tc := range flags {
		t.Run(tc.nama, func(t *testing.T) {
			p := testutil.KesehatanPayloadValid()
			tc.ubah(&p)

			err := validasi.ValidateKesehatan(p)

			require.Error(t, err)
			assert.Contains(t, strings.ToLower(err.Error()), tc.nama)
		})
	}
}

// --- ValidateCariOrangtua ---

func TestValidateCariOrangtua_MinimalSatuKataKunci(t *testing.T) {
	tests := []struct {
		nama                             string
		id, nik, nokk, namaAyah, namaIbu string
	}{
		{"hanya id", testutil.IdOrangtua, "", "", "", ""},
		{"hanya nik", "", testutil.NikOrangtua, "", "", ""},
		{"hanya nokk", "", "", testutil.NoKK, "", ""},
		{"hanya nama_ayah", "", "", "", "Budi Santoso", ""},
		{"hanya nama_ibu", "", "", "", "", "Siti Aminah"},
		{"id dan nik sekaligus", testutil.IdOrangtua, testutil.NikOrangtua, "", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			assert.NoError(t, validasi.ValidateCariOrangtua(tc.id, tc.nik, tc.nokk, tc.namaAyah, tc.namaIbu))
		})
	}
}

func TestValidateCariOrangtua_TanpaKataKunci(t *testing.T) {
	tests := []struct {
		nama                             string
		id, nik, nokk, namaAyah, namaIbu string
	}{
		{"semua kosong", "", "", "", "", ""},
		{"id hanya spasi", "   ", "", "", "", ""},
		{"nik hanya tab", "", "\t", "", "", ""},
		{"semua hanya spasi", " ", " ", " ", " ", " "},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			err := validasi.ValidateCariOrangtua(tc.id, tc.nik, tc.nokk, tc.namaAyah, tc.namaIbu)

			require.Error(t, err, "nilai yang hanya berisi spasi dianggap tidak diisi")
			assert.Contains(t, err.Error(), "kata kunci")
		})
	}
}
