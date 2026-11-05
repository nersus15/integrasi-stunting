package optimasi_query_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- 5B: sapuan turunan yang tiba lebih dulu, empat tabel dalam satu statement ---

const encSapu = "enc-uji-opt-sapu"

func pemeriksaanSapu(t *testing.T) map[string]any {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(testutil.RepoRoot(), "simulasi", "faskes-langsung", "01-pemeriksaan.json"))
	require.NoError(t, err)
	var p map[string]any
	require.NoError(t, json.Unmarshal(content, &p))
	p["kunjungan"].(map[string]any)["id_satusehat"] = encSapu
	return p
}

func TestPemeriksaanMenyapuEmpatTabelDalamSatuStatement(t *testing.T) {
	setupRiwayat(t)
	a := testutil.Scenario1IdAnak
	exec(t, `INSERT INTO stunting.observasi (id, id_anak, ref_encounter, system, kode, display, nilai_angka, satuan, tanggal)
		VALUES ('uji-opt-sapu-ob', $1, $2, 'http://loinc.org', '29463-7', 'Body weight', 8.1, 'kg', '2026-04-10T02:00:00Z')`, a, encSapu)
	exec(t, `INSERT INTO stunting.diagnosa (id, id_anak, ref_encounter, system, kode, display, tanggal_catat)
		VALUES ('uji-opt-sapu-dg', $1, $2, 'http://hl7.org/fhir/sid/icd-10', 'E45', 'Nutritional stunting', '2026-04-10')`, a, encSapu)
	exec(t, `INSERT INTO stunting.layanan (id, id_anak, ref_encounter, jenis, system, kode, display, status, tanggal)
		VALUES ('uji-opt-sapu-ly', $1, $2, 'procedure', 'http://snomed.info/sct', '441041000124100', 'Counseling about nutrition', 'completed', '2026-04-10T03:00:00Z')`, a, encSapu)
	exec(t, `INSERT INTO stunting.rujukan (id, id_anak, ref_encounter, jenis, status, prioritas, tanggal)
		VALUES ('uji-opt-sapu-rj', $1, $2, 'rujukan', 'active', 'routine', '2026-04-10T04:00:00Z')`, a, encSapu)

	var j jawaban
	q := counter.Measure(func() {
		j = request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, pemeriksaanSapu(t))
	})
	require.Equal(t, 201, j.Kode, "%s", j.Mentah)

	expectQuery(t, q, `^\s*(UPDATE|WITH)\b.*ref_encounter\s*=`, 1, "sapuan turunan menggantung")

	assert.Equal(t, 4, scalar[int](t, `SELECT count(*) FROM (
			SELECT id_kunjungan FROM stunting.observasi WHERE id = 'uji-opt-sapu-ob'
			UNION ALL SELECT id_kunjungan FROM stunting.diagnosa WHERE id = 'uji-opt-sapu-dg'
			UNION ALL SELECT id_kunjungan FROM stunting.layanan WHERE id = 'uji-opt-sapu-ly'
			UNION ALL SELECT id_kunjungan FROM stunting.rujukan WHERE id = 'uji-opt-sapu-rj') x
		WHERE id_kunjungan = (SELECT id FROM stunting.kunjungan WHERE satusehat_id = $1)`, encSapu),
		"keempat turunan menggantung tertempel ke kunjungan baru")
}

// --- 5C: cek akses faskes membaca orangtua lewat cache ---

func TestCekAksesFaskesAnakMemakaiCacheOrangtua(t *testing.T) {
	setupRiwayat(t)
	path := "/anak/" + testutil.Scenario1IdAnak
	require.Equal(t, 200, request(t, "GET", path, testutil.KeyFaskesA, nil).Kode, "mengisi cache")

	var j jawaban
	q := counter.Measure(func() { j = request(t, "GET", path, testutil.KeyFaskesA, nil) })
	require.Equal(t, 200, j.Kode, "%s", j.Mentah)
	expectQuery(t, q, selectOrangtua, 0, "SELECT orangtua pada GET kedua")
}

func TestCekAksesFaskesOrangtuaMemakaiCache(t *testing.T) {
	setupRiwayat(t)
	path := "/orangtua/" + testutil.Scenario1IdOrtu
	require.Equal(t, 200, request(t, "GET", path, testutil.KeyFaskesA, nil).Kode, "mengisi cache")

	var j jawaban
	q := counter.Measure(func() { j = request(t, "GET", path, testutil.KeyFaskesA, nil) })
	require.Equal(t, 200, j.Kode, "%s", j.Mentah)
	expectQuery(t, q, selectOrangtua, 0, "SELECT orangtua pada GET kedua")
}

// cache tidak boleh membuat akses bertahan setelah orangtua pindah ke posyandu di luar wilayah
func TestCekAksesFaskesTercabutSetelahPindahPosyandu(t *testing.T) {
	setupBelumDirujuk(t)
	setupPosyanduLuar(t)

	pathAnak, pathOrtu := "/anak/"+testutil.Scenario1IdAnak, "/orangtua/"+testutil.Scenario1IdOrtu
	for range 2 {
		require.Equal(t, 200, request(t, "GET", pathAnak, testutil.KeyFaskesA, nil).Kode)
		require.Equal(t, 200, request(t, "GET", pathOrtu, testutil.KeyFaskesA, nil).Kode)
	}

	j := request(t, "PUT", "/orangtua", testutil.KeyJakantro, map[string]any{
		"id": testutil.Scenario1IdOrtu, "nik": testutil.Scenario1NikIbu, "nama_ibu": "Rina Marlina", "id_posyandu": "UJI-OPT-POS",
	})
	require.Equal(t, 200, j.Kode, "%s", j.Mentah)

	assert.Equal(t, 403, request(t, "GET", pathAnak, testutil.KeyFaskesA, nil).Kode, "anak")
	assert.Equal(t, 403, request(t, "GET", pathOrtu, testutil.KeyFaskesA, nil).Kode, "orangtua")
}
