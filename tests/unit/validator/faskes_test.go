package validator_test

import (
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/satusehat"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/validasi"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
)

func validPemeriksaan() *types.PemeriksaanFaskes {
	return &types.PemeriksaanFaskes{
		Anak:      types.Anak{Nik: testutil.Ptr("3175049900000001")},
		Kunjungan: types.Kunjungan{IdSatusehat: testutil.Ptr("enc"), TanggalPengukuran: "2026-01-01"},
		Observasi: []types.Observasi{{IdSatusehat: testutil.Ptr("o"), System: "s", Kode: "k",
			Component: []types.Observasi{{IdSatusehat: testutil.Ptr("oc"), System: "s", Kode: "k"}}}},
		Diagnosa: []types.Diagnosa{{IdSatusehat: testutil.Ptr("d"), Kode: "E45", Jenis: entity.DiagnosaAlergi}},
		Layanan:  []types.Layanan{{IdSatusehat: testutil.Ptr("l"), Jenis: "procedure"}},
		Rujukan:  []types.Rujukan{{IdSatusehat: testutil.Ptr("r"), Jenis: "rujuk_balik"}},
		Episode:  []types.Episode{{IdSatusehat: testutil.Ptr("e")}},
	}
}

func TestValidatePemeriksaanFaskes_Valid(t *testing.T) {
	assert.NoError(t, validasi.ValidatePemeriksaanFaskes(validPemeriksaan()))

	p := validPemeriksaan()
	p.Anak = types.Anak{IdSatusehat: testutil.Ptr("P1")}
	assert.NoError(t, validasi.ValidatePemeriksaanFaskes(p), "satusehat_id saja cukup")
}

func TestValidatePemeriksaanFaskes_Cacat(t *testing.T) {
	cases := map[string]struct {
		ubah  func(p *types.PemeriksaanFaskes)
		pesan string
	}{
		"tanpa identitas anak": {func(p *types.PemeriksaanFaskes) { p.Anak = types.Anak{} }, "anak.satusehat_id atau anak.nik"},
		"nik anak pendek":      {func(p *types.PemeriksaanFaskes) { p.Anak.Nik = testutil.Ptr("123") }, "anak.nik"},
		"kunjungan tanpa id":   {func(p *types.PemeriksaanFaskes) { p.Kunjungan.IdSatusehat = nil }, "kunjungan.id_satusehat"},
		"kunjungan tanpa tgl":  {func(p *types.PemeriksaanFaskes) { p.Kunjungan.TanggalPengukuran = "" }, "tanggal_pengukuran wajib"},
		"kunjungan tgl salah":  {func(p *types.PemeriksaanFaskes) { p.Kunjungan.TanggalPengukuran = "1/1/2026" }, "harus berformat"},
		"tgl selesai salah":    {func(p *types.PemeriksaanFaskes) { p.Kunjungan.TanggalSelesai = testutil.Ptr("besok") }, "tanggal_selesai"},
		"observasi tanpa kode": {func(p *types.PemeriksaanFaskes) { p.Observasi[0].Kode = "" }, "observasi[0]: kode"},
		"observasi tanpa sys":  {func(p *types.PemeriksaanFaskes) { p.Observasi[0].System = "" }, "observasi[0]: system"},
		"komponen tanpa id":    {func(p *types.PemeriksaanFaskes) { p.Observasi[0].Component[0].IdSatusehat = nil }, "component[0]"},
		"diagnosa tanpa kode":  {func(p *types.PemeriksaanFaskes) { p.Diagnosa[0].Kode = "" }, "diagnosa[0]: kode"},
		"diagnosa jenis asing": {func(p *types.PemeriksaanFaskes) { p.Diagnosa[0].Jenis = "lain" }, "diagnosa[0]: jenis"},
		"layanan tanpa jenis":  {func(p *types.PemeriksaanFaskes) { p.Layanan[0].Jenis = "" }, "layanan[0]: jenis wajib"},
		"layanan jenis asing":  {func(p *types.PemeriksaanFaskes) { p.Layanan[0].Jenis = "pijat" }, "tidak dikenali"},
		"rujukan jenis asing":  {func(p *types.PemeriksaanFaskes) { p.Rujukan[0].Jenis = "keluar" }, "rujukan[0]"},
		"rujukan tanpa id":     {func(p *types.PemeriksaanFaskes) { p.Rujukan[0].IdSatusehat = nil }, "rujukan[0]"},
		"episode tanpa id":     {func(p *types.PemeriksaanFaskes) { p.Episode[0].IdSatusehat = nil }, "episode[0]"},
		"layanan tanpa id":     {func(p *types.PemeriksaanFaskes) { p.Layanan[0].IdSatusehat = nil }, "layanan[0]"},
		"diagnosa tanpa id":    {func(p *types.PemeriksaanFaskes) { p.Diagnosa[0].IdSatusehat = nil }, "diagnosa[0]"},
	}
	for name, k := range cases {
		t.Run(name, func(t *testing.T) {
			p := validPemeriksaan()
			k.ubah(p)
			err := validasi.ValidatePemeriksaanFaskes(p)
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), k.pesan)
			}
		})
	}

	assert.Error(t, validasi.ValidatePemeriksaanFaskes(nil))
}

func TestValidateOrangtuaStream(t *testing.T) {
	valid := types.OrangtuaPayload{Id: "o", Nik: "3175049900000002", NamaIbu: "Ibu"}
	assert.NoError(t, validasi.ValidateOrangtuaStream(valid))

	cases := map[string]func(p *types.OrangtuaPayload){
		"tanpa id":        func(p *types.OrangtuaPayload) { p.Id = "" },
		"tanpa nik":       func(p *types.OrangtuaPayload) { p.Nik = "" },
		"nik pendek":      func(p *types.OrangtuaPayload) { p.Nik = "317" },
		"tanpa nama sama": func(p *types.OrangtuaPayload) { p.NamaIbu = "" },
	}
	for name, modify := range cases {
		p := valid
		modify(&p)
		assert.Error(t, validasi.ValidateOrangtuaStream(p), name)
	}
}

func TestValidateKesehatan_TanggalSalahFormat(t *testing.T) {
	p := testutil.KesehatanPayloadValid()
	p.TanggalPemantauan = "16/02/2025"
	assert.ErrorContains(t, validasi.ValidateKesehatan(p), "YYYY-MM-DD")
}

func TestValidateZscoreSatuSumber(t *testing.T) {
	z, label := -2.41, "Pendek"
	obsTBU := types.Observasi{IdSatusehat: testutil.Ptr("o-tbu"), System: satusehat.SystemSNOMED, Kode: "1153604005", NilaiAngka: &z}

	p := validPemeriksaan()
	p.Observasi = append(p.Observasi, obsTBU)
	assert.NoError(t, validasi.ValidateZscoreSatuSumber(p), "hanya di observasi")
	assert.NoError(t, validasi.ValidatePemeriksaanFaskes(p))

	p = validPemeriksaan()
	p.Kunjungan.ZscoreTbu, p.Kunjungan.StatusTbu = &z, &label
	assert.NoError(t, validasi.ValidateZscoreSatuSumber(p), "hanya di kunjungan")

	p = validPemeriksaan()
	p.Kunjungan.ZscoreBbu = &z
	p.Observasi = append(p.Observasi, obsTBU)
	assert.NoError(t, validasi.ValidateZscoreSatuSumber(p), "indeks berbeda")

	p = validPemeriksaan()
	p.Kunjungan.ZscoreTbu = &z
	p.Observasi = append(p.Observasi, obsTBU)
	assert.NoError(t, validasi.ValidatePemeriksaanFaskes(p), "setiap field sah, yang salah susunannya")
	err := validasi.ValidateZscoreSatuSumber(p)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "zscore_tbu")
	}

	p = validPemeriksaan()
	p.Kunjungan.StatusTbu = &label
	p.Observasi = append(p.Observasi, obsTBU)
	assert.Error(t, validasi.ValidateZscoreSatuSumber(p), "status saja juga bentrok")
}
