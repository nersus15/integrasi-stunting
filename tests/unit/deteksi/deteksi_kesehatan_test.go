package deteksi_test

import (
	"encoding/json"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/validasi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const kesehatanNested = `"kesehatan":{"id":"S1","id_anak":"A1","tanggal_pemantauan":"2025-02-16"}`

func detectKesehatan(t *testing.T, body string) (validasi.JenisPayload, *types.KesehatanPayload, error) {
	t.Helper()
	var p types.KesehatanNestedPayload
	require.NoError(t, json.Unmarshal([]byte(body), &p))
	return validasi.DeteksiJenisPayloadKesehatan(&p)
}

func TestDeteksiKesehatan_BentukSah(t *testing.T) {
	cases := []struct {
		nama string
		mau  validasi.JenisPayload
		body string
	}{
		{"registrasi lengkap", validasi.JenisRegistrasiLengkap,
			`{"orangtua":{"id":"O1","no_kk":"3515","nik":"351"},"anak":{"id":"A1","id_orangtua":"O1","nama":"Ahmad"},` + kesehatanNested + `}`},
		{"anak baru, orangtua hanya id", validasi.JenisAnakBaru,
			`{"orangtua":{"id":"O1"},"anak":{"id":"A1","id_orangtua":"O1","nik":"352"},` + kesehatanNested + `}`},
		{"anak baru 2 key", validasi.JenisAnakBaru, `{"anak":{"id":"A1","id_orangtua":"O1","nama":"Ahmad"},` + kesehatanNested + `}`},
		{"kesehatan saja 2 key", validasi.JenisKesehatanSaja, `{"anak":{"id":"A1","id_orangtua":"O1"},` + kesehatanNested + `}`},
		{"kesehatan saja 1 key", validasi.JenisKesehatanSaja, `{` + kesehatanNested + `}`},
		{"flat", validasi.JenisKesehatanSaja, `{"id":"S1","id_anak":"A1","tanggal_pemantauan":"2025-02-16"}`},
	}
	for _, k := range cases {
		t.Run(k.nama, func(t *testing.T) {
			jenis, kes, err := detectKesehatan(t, k.body)
			require.NoError(t, err)
			assert.Equal(t, k.mau, jenis)
			assert.Equal(t, "S1", kes.Id)
		})
	}
}

func TestDeteksiKesehatan_PayloadCacat(t *testing.T) {
	cases := map[string]string{
		"flat bercampur nested":         `{"id":"S1","anak":{"id":"A1"}}`,
		"flat tanpa tanggal":            `{"id":"S1","id_anak":"A1"}`,
		"tanpa kesehatan":               `{"anak":{"id":"A1","id_orangtua":"O1"}}`,
		"kesehatan tanpa id":            `{"kesehatan":{"id_anak":"A1","tanggal_pemantauan":"2025-02-16"}}`,
		"kesehatan tanpa id_anak":       `{"kesehatan":{"id":"S1","tanggal_pemantauan":"2025-02-16"}}`,
		"tanggal salah format":          `{"kesehatan":{"id":"S1","id_anak":"A1","tanggal_pemantauan":"16-02-2025"}}`,
		"anak tanpa id":                 `{"anak":{"id_orangtua":"O1"},` + kesehatanNested + `}`,
		"anak tanpa id_orangtua":        `{"anak":{"id":"A1"},` + kesehatanNested + `}`,
		"anak beda id":                  `{"anak":{"id":"A2","id_orangtua":"O1"},` + kesehatanNested + `}`,
		"orangtua tanpa anak":           `{"orangtua":{"id":"O1"},` + kesehatanNested + `}`,
		"orangtua tanpa id":             `{"orangtua":{"nik":"351"},"anak":{"id":"A1","id_orangtua":"O1"},` + kesehatanNested + `}`,
		"orangtua beda id":              `{"orangtua":{"id":"O2"},"anak":{"id":"A1","id_orangtua":"O1"},` + kesehatanNested + `}`,
		"orangtua baru tanpa anak baru": `{"orangtua":{"id":"O1","nik":"351"},"anak":{"id":"A1","id_orangtua":"O1"},` + kesehatanNested + `}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, err := detectKesehatan(t, body)
			assert.Error(t, err)
		})
	}

	_, _, err := validasi.DeteksiJenisPayloadKesehatan(nil)
	assert.Error(t, err)
}
