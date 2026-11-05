package update_endpoint_test

import (
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func obsZscore(id, code string, value float64, interpretasi string) map[string]any {
	o := map[string]any{
		"id_satusehat": id,
		"system":       "http://snomed.info/sct",
		"kode":         code,
		"nilai_angka":  value,
		"satuan":       "{Zscore}",
	}
	if interpretasi != "" {
		o["interpretasi"] = interpretasi
	}
	return o
}

func addObservasi(o map[string]any) func(p map[string]any) {
	return func(p map[string]any) {
		p["observasi"] = append(p["observasi"].([]any), o)
	}
}

func zscoreLangsung(t *testing.T) [6]string {
	t.Helper()
	var r [6]string
	err := db.QueryRow(`
		SELECT coalesce(zscore_bbu::text,'NULL'), coalesce(status_bbu,'NULL'),
		       coalesce(zscore_tbu::text,'NULL'), coalesce(status_tbu,'NULL'),
		       coalesce(zscore_bbtb::text,'NULL'), coalesce(status_bbtb,'NULL')
		FROM stunting.kunjungan WHERE satusehat_id = 'enc-langsung-01'`).
		Scan(&r[0], &r[1], &r[2], &r[3], &r[4], &r[5])
	require.NoError(t, err)
	return r
}

func TestPemeriksaanZscoreDariObservasi(t *testing.T) {
	setupRiwayat(t)

	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA,
		pemeriksaan(t, addObservasi(obsZscore("obs-langsung-z-tbu", "1153604005", -2.41, "444000005"))))
	require.Equal(t, 201, j.Kode, "%v", j.Isi)

	assert.Equal(t, [6]string{"NULL", "NULL", "-2.41", "Pendek", "NULL", "NULL"}, zscoreLangsung(t))
}

func TestPemeriksaanZscoreDariKunjungan(t *testing.T) {
	setupRiwayat(t)

	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, pemeriksaan(t, func(p map[string]any) {
		k := p["kunjungan"].(map[string]any)
		k["zscore_bbu"], k["status_bbu"] = -3.05, "Berat Badan Sangat Kurang"
	}))
	require.Equal(t, 201, j.Kode, "%v", j.Isi)

	assert.Equal(t, [6]string{"-3.05", "Berat Badan Sangat Kurang", "NULL", "NULL", "NULL", "NULL"}, zscoreLangsung(t))
}

func TestPemeriksaanZscoreIndeksBerbedaDariKeduanya(t *testing.T) {
	setupRiwayat(t)

	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, pemeriksaan(t, func(p map[string]any) {
		k := p["kunjungan"].(map[string]any)
		k["zscore_bbu"], k["status_bbu"] = -3.05, "Berat Badan Sangat Kurang"
		addObservasi(obsZscore("obs-langsung-z-tbu", "1153604005", -2.41, "444000005"))(p)
	}))
	require.Equal(t, 201, j.Kode, "%v", j.Isi)

	assert.Equal(t, [6]string{"-3.05", "Berat Badan Sangat Kurang", "-2.41", "Pendek", "NULL", "NULL"}, zscoreLangsung(t))
}

func TestPemeriksaanZscoreDiKeduanyaDitolak(t *testing.T) {
	setupRiwayat(t)

	cases := map[string]func(p map[string]any){
		"zscore_tbu di kunjungan dan TB/U di observasi": func(p map[string]any) {
			p["kunjungan"].(map[string]any)["zscore_tbu"] = -2.41
			addObservasi(obsZscore("obs-langsung-z-tbu", "1153604005", -2.41, "444000005"))(p)
		},
		"status_tbu saja di kunjungan dan PB/U di observasi": func(p map[string]any) {
			p["kunjungan"].(map[string]any)["status_tbu"] = "Pendek"
			addObservasi(obsZscore("obs-langsung-z-pbu", "1153590000", -2.41, ""))(p)
		},
		"zscore_bbtb di kunjungan dan BB/TB di observasi": func(p map[string]any) {
			p["kunjungan"].(map[string]any)["zscore_bbtb"] = 1.2
			addObservasi(obsZscore("obs-langsung-z-bbtb", "1153600001", 1.5, "OI000004"))(p)
		},
	}
	for name, modify := range cases {
		j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, pemeriksaan(t, modify))
		assert.Equal(t, 422, j.Kode, "%s: %v", name, j.Isi)
		assert.Equal(t, 1003, j.Galat, name)
	}
	assert.Zero(t, scalar[int](t, `SELECT count(*) FROM stunting.kunjungan WHERE satusehat_id = 'enc-langsung-01'`))
	assert.Zero(t, scalar[int](t, `SELECT count(*) FROM stunting.observasi WHERE satusehat_id LIKE 'obs-langsung-%'`))
}

func TestPUTFaskesObservasiMemperbaruiKunjungan(t *testing.T) {
	setupRiwayat(t)

	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA,
		pemeriksaan(t, addObservasi(obsZscore("obs-langsung-z-tbu", "1153604005", -2.41, "444000005"))))
	require.Equal(t, 201, j.Kode, "%v", j.Isi)

	j = request(t, "PUT", "/faskes/observasi", testutil.KeyFaskesA,
		map[string]any{"id_satusehat": "obs-langsung-z-tbu", "system": "http://snomed.info/sct", "kode": "1153604005", "nilai_angka": -1.52, "interpretasi": "17489000"})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, [6]string{"NULL", "NULL", "-1.52", "Normal", "NULL", "NULL"}, zscoreLangsung(t))

	j = request(t, "PUT", "/faskes/observasi", testutil.KeyFaskesA,
		map[string]any{"id_satusehat": "obs-langsung-bb", "system": "http://loinc.org", "kode": "29463-7", "nilai_angka": 8.5})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, "8.5/75.2", scalar[string](t,
		`SELECT berat_badan::text || '/' || tinggi_badan::text FROM stunting.kunjungan WHERE satusehat_id = 'enc-langsung-01'`))
}
