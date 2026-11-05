package update_endpoint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPUTFaskesPemilikBerhasil(t *testing.T) {
	setupRiwayat(t)

	cases := []struct {
		resource string
		body     map[string]any
		cek      string
		harap    string
	}{
		{"kunjungan", map[string]any{"id_satusehat": "enc-pkm-01", "tanggal_pengukuran": "2026-02-06", "tanggal_selesai": "2026-02-06"},
			`SELECT to_char(tanggal_pengukuran,'YYYY-MM-DD') || '/' || berat_badan FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`, "2026-02-06/7.4"},
		{"observasi", map[string]any{"id_satusehat": "obs-pkm-bb-01", "system": "http://loinc.org", "kode": "29463-7", "nilai_angka": 7.5},
			`SELECT nilai_angka::float8::text || '/' || coalesce(interpretasi,'') FROM stunting.observasi WHERE satusehat_id = 'obs-pkm-bb-01'`, "7.5/OI000007"},
		{"diagnosa", map[string]any{"id_satusehat": "cond-pkm-01", "kode": "E45", "display": "revisi"},
			`SELECT display || '/' || kode FROM stunting.diagnosa WHERE satusehat_id = 'cond-pkm-01'`, "revisi/E45"},
		{"layanan", map[string]any{"id_satusehat": "proc-pkm-02", "jenis": "procedure", "catatan": "catatan baru"},
			`SELECT catatan || '/' || jenis FROM stunting.layanan WHERE satusehat_id = 'proc-pkm-02'`, "catatan baru/procedure"},
		{"rujukan", map[string]any{"id_satusehat": "sr-rujuk-01", "status": "completed"},
			`SELECT status || '/' || jenis FROM stunting.rujukan WHERE satusehat_id = 'sr-rujuk-01'`, "completed/rujukan"},
		{"episode", map[string]any{"id_satusehat": "ep-stunting-sim1", "status": "finished", "selesai": "2026-04-01T00:00:00Z"},
			`SELECT status || '/' || to_char(selesai,'YYYY-MM-DD') FROM stunting.episode WHERE satusehat_id = 'ep-stunting-sim1'`, "finished/2026-04-01"},
	}

	for _, k := range cases {
		t.Run(k.resource, func(t *testing.T) {
			j := request(t, "PUT", "/faskes/"+k.resource, testutil.KeyFaskesA, k.body)
			if !assert.Equal(t, 200, j.Kode, "%v", j.Isi) {
				return
			}
			assert.Equal(t, k.body["id_satusehat"], j.Isi["id_satusehat"], "respons dari baris gabungan")
			assert.Equal(t, k.harap, scalar[string](t, k.cek))
		})
	}
}

func TestPUTFaskesMilikFaskesLainDitolak(t *testing.T) {
	setupRiwayat(t)

	cases := []struct {
		kunci, resource string
		body            map[string]any
	}{
		{testutil.KeyFaskesA, "layanan", map[string]any{"id_satusehat": "proc-rs-01", "jenis": "procedure"}},
		{testutil.KeyFaskesA, "observasi", map[string]any{"id_satusehat": "obs-rs-bb-01", "system": "http://loinc.org", "kode": "29463-7"}},
		{testutil.KeyFaskesA, "diagnosa", map[string]any{"id_satusehat": "cond-rs-01", "kode": "E43"}},
		{testutil.KeyFaskesA, "rujukan", map[string]any{"id_satusehat": "sr-balik-01"}},
		{testutil.KeyFaskesA, "kunjungan", map[string]any{"id_satusehat": "enc-rs-01", "tanggal_pengukuran": "2026-02-20"}},
		{keyRS, "kunjungan", map[string]any{"id_satusehat": "enc-pkm-01", "tanggal_pengukuran": "2026-02-05"}},
		{keyRS, "episode", map[string]any{"id_satusehat": "ep-stunting-sim1"}},
	}
	for _, k := range cases {
		j := request(t, "PUT", "/faskes/"+k.resource, k.kunci, k.body)
		assert.Equal(t, 403, j.Kode, "%s %v", k.resource, k.body["id_satusehat"])
		assert.Equal(t, 1008, j.Galat, "%s %v", k.resource, j.Isi)
	}

	j := request(t, "PUT", "/faskes/layanan", keyRS, map[string]any{"id_satusehat": "proc-rs-01", "jenis": "procedure", "status": "completed"})
	assert.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, "completed", scalar[string](t, `SELECT status FROM stunting.layanan WHERE satusehat_id = 'proc-rs-01'`))
}

func TestPUTFaskesBelumTersimpan404(t *testing.T) {
	setupRiwayat(t)

	for resource, body := range map[string]map[string]any{
		"kunjungan": {"id_satusehat": "tidak-ada", "tanggal_pengukuran": "2026-02-05"},
		"observasi": {"id_satusehat": "tidak-ada", "system": "http://loinc.org", "kode": "29463-7"},
		"diagnosa":  {"id_satusehat": "tidak-ada", "kode": "E45"},
		"layanan":   {"id_satusehat": "tidak-ada", "jenis": "procedure"},
		"rujukan":   {"id_satusehat": "tidak-ada"},
		"episode":   {"id_satusehat": "tidak-ada"},
	} {
		j := request(t, "PUT", "/faskes/"+resource, testutil.KeyFaskesA, body)
		assert.Equal(t, 404, j.Kode, "%s %v", resource, j.Isi)
		assert.Equal(t, 2001, j.Galat, resource)
	}
}

func TestPUTFaskesPayloadSalah(t *testing.T) {
	for _, resource := range []string{"kunjungan", "observasi", "diagnosa", "layanan", "rujukan", "episode"} {
		j := request(t, "PUT", "/faskes/"+resource, testutil.KeyFaskesA, map[string]any{"kode": "x"})
		assert.Equal(t, 422, j.Kode, "%s tanpa id_satusehat", resource)
		assert.Equal(t, 1002, j.Galat, resource)

		j = request(t, "PUT", "/faskes/"+resource, testutil.KeyFaskesA, `{rusak`)
		assert.Equal(t, 400, j.Kode, "%s body rusak", resource)
		assert.Equal(t, 1001, j.Galat, resource)
	}

	j := request(t, "PUT", "/faskes/hantu", testutil.KeyFaskesA, map[string]any{})
	assert.Equal(t, 404, j.Kode)
}

func TestPUTFaskesHakAkses(t *testing.T) {
	body := map[string]any{"id_satusehat": "enc-pkm-01", "tanggal_pengukuran": "2026-02-05"}

	for name, key := range map[string]string{
		"jakantro":   testutil.KeyJakantro,
		"tanpa grup": keyWithoutGroup,
		"grup lain":  kunciGrupLain,
	} {
		j := request(t, "PUT", "/faskes/kunjungan", key, body)
		assert.Equal(t, 403, j.Kode, name)
	}

	j := request(t, "PUT", "/faskes/kunjungan", "", body)
	assert.Equal(t, 401, j.Kode)
}

func TestPUTFaskesDataDariEndpointPemeriksaan(t *testing.T) {
	setupRiwayat(t)

	content, err := os.ReadFile(filepath.Join(testutil.RepoRoot(), "simulasi", "faskes-langsung", "01-pemeriksaan.json"))
	require.NoError(t, err)
	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, string(content))
	require.Equal(t, 201, j.Kode, "%v", j.Isi)

	cases := map[string]map[string]any{
		"observasi": {"id_satusehat": "obs-langsung-bb", "system": "http://loinc.org", "kode": "29463-7", "nilai_angka": 8.5},
		"diagnosa":  {"id_satusehat": "cond-langsung-01", "kode": "E45", "display": "lewat endpoint"},
		"layanan":   {"id_satusehat": "proc-langsung-01", "jenis": "procedure", "status": "completed"},
	}
	for resource, body := range cases {
		j := request(t, "PUT", "/faskes/"+resource, testutil.KeyFaskesA, body)
		assert.Equal(t, 200, j.Kode, "%s: pemilik dari kunjungan: %v", resource, j.Isi)

		j = request(t, "PUT", "/faskes/"+resource, keyRS, body)
		assert.Equal(t, 403, j.Kode, "%s: pemilik dari cache", resource)
	}
}
