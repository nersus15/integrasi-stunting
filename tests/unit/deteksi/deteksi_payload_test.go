package deteksi_test

import (
	"encoding/json"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/validasi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func detect(t *testing.T, body string) (validasi.JenisPayload, *types.KunjunganPayload, error) {
	t.Helper()

	var p types.KunjunganNestedPayload
	require.NoError(t, json.Unmarshal([]byte(body), &p), "body harus json valid")
	return validasi.DeteksiJenisPayload(&p)
}

const kunjunganNested = `"kunjungan":{"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16","berat_badan":8.5}`

// --- POSITIVE CASE ---

func TestDeteksi_BentukSah(t *testing.T) {
	tests := []struct {
		nama string
		mau  validasi.JenisPayload
		body string
	}{
		{
			nama: "datalengkap 3 key",
			mau:  validasi.JenisRegistrasiLengkap,
			body: `{"orangtua":{"id":"O1","no_kk":"3515","nik":"351"},"anak":{"id":"A1","id_orangtua":"O1","nama":"Ahmad","nik":"352"},` + kunjunganNested + `}`,
		},
		{
			nama: "anak baru 3 key, orangtua hanya id",
			mau:  validasi.JenisAnakBaru,
			body: `{"orangtua":{"id":"O1"},"anak":{"id":"A1","id_orangtua":"O1","nama":"Ahmad","nik":"352"},` + kunjunganNested + `}`,
		},
		{
			nama: "anak baru 2 key",
			mau:  validasi.JenisAnakBaru,
			body: `{"anak":{"id":"A1","id_orangtua":"O1","nama":"Ahmad","nik":"352"},` + kunjunganNested + `}`,
		},
		{
			nama: "kunjungan saja 3 key",
			mau:  validasi.JenisKunjunganSaja,
			body: `{"orangtua":{"id":"O1"},"anak":{"id":"A1","id_orangtua":"O1"},` + kunjunganNested + `}`,
		},
		{
			nama: "kunjungan saja 2 key",
			mau:  validasi.JenisKunjunganSaja,
			body: `{"anak":{"id":"A1","id_orangtua":"O1"},` + kunjunganNested + `}`,
		},
		{
			nama: "kunjungan saja 1 key nested",
			mau:  validasi.JenisKunjunganSaja,
			body: `{` + kunjunganNested + `}`,
		},
		{
			nama: "kunjungan saja bentuk flat",
			mau:  validasi.JenisKunjunganSaja,
			body: `{"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16","berat_badan":8.5}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			jenis, kunjungan, err := detect(t, tc.body)

			require.NoError(t, err)
			assert.Equal(t, tc.mau, jenis)
			require.NotNil(t, kunjungan, "data kunjungan harus ikut dikembalikan")
			assert.Equal(t, "K1", kunjungan.Id)
			assert.Equal(t, "A1", kunjungan.IDAnak)
		})
	}
}

func TestDeteksi_FlatDanNestedMenghasilkanKunjunganYangSama(t *testing.T) {
	_, flat, err := detect(t, `{"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16","berat_badan":8.5}`)
	require.NoError(t, err)
	_, nested, err := detect(t, `{`+kunjunganNested+`}`)
	require.NoError(t, err)

	assert.Equal(t, nested.Id, flat.Id)
	assert.Equal(t, nested.IDAnak, flat.IDAnak)
	assert.Equal(t, nested.TanggalPengukuran, flat.TanggalPengukuran)
	require.NotNil(t, flat.BeratBadan)
	require.NotNil(t, nested.BeratBadan)
	assert.InDelta(t, *nested.BeratBadan, *flat.BeratBadan, 0.001)
}

// --- NEGATIVE CASE ---

func TestDeteksi_FlatTidakBolehDisertaiKeyLain(t *testing.T) {
	flat := `"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16"`

	tests := []struct {
		nama string
		body string
	}{
		{"flat + key kunjungan", `{` + flat + `,` + kunjunganNested + `}`},
		{"flat + key anak", `{` + flat + `,"anak":{"id":"A1","id_orangtua":"O1","nama":"Ahmad"}}`},
		{"flat + key orangtua", `{` + flat + `,"orangtua":{"id":"O1","nik":"351"}}`},
		{"flat + ketiga key", `{` + flat + `,"orangtua":{"id":"O1","nik":"351"},"anak":{"id":"A1","id_orangtua":"O1","nama":"Ahmad"},` + kunjunganNested + `}`},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			_, _, err := detect(t, tc.body)

			require.Error(t, err, "bentuk flat murni hanya untuk kunjungan saja")
			assert.Contains(t, err.Error(), "flat")
		})
	}
}

func TestDeteksi_PayloadCacat(t *testing.T) {
	tests := []struct {
		nama string
		body string
	}{
		{"payload kosong", `{}`},
		{"tanpa kunjungan", `{"orangtua":{"id":"O1","nik":"351"},"anak":{"id":"A1","id_orangtua":"O1","nama":"Ahmad"}}`},
		{"kunjungan tanpa id", `{"kunjungan":{"id_anak":"A1","tanggal_pengukuran":"2025-02-16"}}`},
		{"kunjungan tanpa id_anak", `{"kunjungan":{"id":"K1","tanggal_pengukuran":"2025-02-16"}}`},
		{"kunjungan tanpa tanggal", `{"kunjungan":{"id":"K1","id_anak":"A1"}}`},
		{"tanggal bukan YYYY-MM-DD", `{"kunjungan":{"id":"K1","id_anak":"A1","tanggal_pengukuran":"16-02-2025"}}`},
		{"flat tanpa id", `{"id_anak":"A1","tanggal_pengukuran":"2025-02-16"}`},
		{"flat tanpa id_anak", `{"id":"K1","tanggal_pengukuran":"2025-02-16"}`},
		{"anak tanpa id", `{"anak":{"id_orangtua":"O1","nama":"Ahmad"},` + kunjunganNested + `}`},
		{"anak tanpa id_orangtua", `{"anak":{"id":"A1","nama":"Ahmad"},` + kunjunganNested + `}`},
		{"anak.id beda dengan kunjungan.id_anak", `{"anak":{"id":"A9","id_orangtua":"O1"},` + kunjunganNested + `}`},
		{"orangtua tanpa anak", `{"orangtua":{"id":"O1","nik":"351"},` + kunjunganNested + `}`},
		{"orangtua tanpa id", `{"orangtua":{"nik":"351"},"anak":{"id":"A1","id_orangtua":"O1","nama":"Ahmad"},` + kunjunganNested + `}`},
		{"orangtua.id beda dengan anak.id_orangtua", `{"orangtua":{"id":"O9"},"anak":{"id":"A1","id_orangtua":"O1","nama":"Ahmad"},` + kunjunganNested + `}`},
		{"orangtua baru untuk anak lama", `{"orangtua":{"id":"O1","nik":"351"},"anak":{"id":"A1","id_orangtua":"O1"},` + kunjunganNested + `}`},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			jenis, kunjungan, err := detect(t, tc.body)

			require.Error(t, err)
			assert.NotEmpty(t, err.Error())
			assert.Empty(t, jenis)
			assert.Nil(t, kunjungan)
		})
	}
}

func TestDeteksi_PayloadNil(t *testing.T) {
	jenis, kunjungan, err := validasi.DeteksiJenisPayload(nil)

	require.Error(t, err)
	assert.Empty(t, jenis)
	assert.Nil(t, kunjungan)
}
