package zscore_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/helper/satusehat"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testutil.SetupLogger()
	os.Exit(m.Run())
}

func observation(t *testing.T, system, code string, value float64, interpretasi string) *satusehat.HasilObservasi {
	t.Helper()
	r := map[string]any{
		"resourceType":  "Observation",
		"status":        "final",
		"code":          map[string]any{"coding": []any{map[string]any{"system": system, "code": code}}},
		"valueQuantity": map[string]any{"value": value, "unit": "{Zscore}", "system": "http://unitsofmeasure.org", "code": "{Zscore}"},
	}
	if interpretasi != "" {
		sys := satusehat.SystemSNOMED
		if interpretasi[:2] == "OI" {
			sys = satusehat.SystemKemkes
		}
		r["interpretation"] = []any{map[string]any{"coding": []any{map[string]any{"system": sys, "code": interpretasi}}}}
	}
	content, err := json.Marshal(r)
	require.NoError(t, err)

	id, tipe := "obs-z", "Observation"
	h, _, _, _, err := satusehat.ObservationToObservasi(types2.BundleEntry{
		BundleEntry: fhir.BundleEntry{Resource: content},
		Base:        &types2.BaseResource{Resource: fhir.Resource{Id: &id}, ResourceType: &tipe},
	})
	require.NoError(t, err)
	return h
}

func TestKolomZscoreDariKodeObservation(t *testing.T) {
	cases := map[string]string{
		"1153593003": "zscore_bbu",
		"1153590000": "zscore_tbu",
		"1153604005": "zscore_tbu",
		"1153598007": "zscore_bbtb",
		"1153600001": "zscore_bbtb",
		"1153596006": "",
		"1153594009": "",
	}
	for code, kolom := range cases {
		assert.Equal(t, kolom, satusehat.KolomUkuran(satusehat.SystemSNOMED, code), code)
	}
}

func TestStatusGiziPerIndeks(t *testing.T) {
	cases := []struct{ kolom, kode, label string }{
		{"zscore_bbu", "OI000007", "Berat Badan Sangat Kurang"},
		{"zscore_bbu", "248342006", "Berat Badan Kurang"},
		{"zscore_bbu", "43664005", "Berat Badan Normal"},
		{"zscore_bbu", "OI000010", "Risiko Berat Badan Lebih"},

		{"zscore_tbu", "OI000011", "Sangat Pendek"},
		{"zscore_tbu", "444000005", "Pendek"},
		{"zscore_tbu", "17489000", "Normal"},
		{"zscore_tbu", "83077003", "Tinggi"},

		{"zscore_bbtb", "OI000001", "Gizi Buruk"},
		{"zscore_bbtb", "248325000", "Gizi Kurang"},
		{"zscore_bbtb", "248324001", "Gizi Baik"},
		{"zscore_bbtb", "OI000004", "Risiko Gizi Lebih"},
		{"zscore_bbtb", "238131007", "Gizi Lebih"},
		{"zscore_bbtb", "414915002", "Obesitas"},
	}
	for _, k := range cases {
		label, ok := satusehat.StatusGizi(k.kolom, k.kode)
		assert.True(t, ok, "%s %s", k.kolom, k.kode)
		assert.Equal(t, k.label, label, "%s %s", k.kolom, k.kode)
	}
}

func TestStatusGiziKodeIndeksLainDitolak(t *testing.T) {
	cases := [][2]string{
		{"zscore_bbu", "248325000"},
		{"zscore_bbu", "OI000011"},
		{"zscore_tbu", "43664005"},
		{"zscore_tbu", "OI000001"},
		{"zscore_bbtb", "OI000007"},
		{"zscore_bbtb", "444000005"},
		{"berat_badan", "OI000007"},
		{"tinggi_badan", "OI000011"},
		{"", "OI000007"},
		{"zscore_bbu", "tidak-dikenal"},
	}
	for _, k := range cases {
		label, ok := satusehat.StatusGizi(k[0], k[1])
		assert.False(t, ok, "%s %s", k[0], k[1])
		assert.Empty(t, label, "%s %s", k[0], k[1])
	}
}

func TestObservationZscoreTerangkat(t *testing.T) {
	h := observation(t, satusehat.SystemSNOMED, "1153604005", -2.41, "444000005")
	assert.Equal(t, "zscore_tbu", h.Kolom)
	require.NotNil(t, h.Angka)
	assert.Equal(t, -2.41, *h.Angka)
	assert.Equal(t, "Pendek", h.Status)
	assert.Empty(t, h.CaraUkur, "z-score panjang/tinggi tidak menentukan cara ukur")

	h = observation(t, satusehat.SystemSNOMED, "1153590000", -3.21, "OI000011")
	assert.Equal(t, "zscore_tbu", h.Kolom)
	assert.Equal(t, "Sangat Pendek", h.Status)
	assert.Empty(t, h.CaraUkur)

	h = observation(t, satusehat.SystemSNOMED, "1153600001", 0, "248324001")
	assert.Equal(t, "zscore_bbtb", h.Kolom)
	require.NotNil(t, h.Angka, "z-score 0 tetap nilai")
	assert.Equal(t, 0.0, *h.Angka)
	assert.Equal(t, "Gizi Baik", h.Status)
}

func TestObservationZscoreTanpaInterpretasi(t *testing.T) {
	h := observation(t, satusehat.SystemSNOMED, "1153593003", -2.47, "")
	assert.Equal(t, "zscore_bbu", h.Kolom)
	require.NotNil(t, h.Angka)
	assert.Equal(t, -2.47, *h.Angka)
	assert.Empty(t, h.Status)
}

func TestStatusHanyaUntukObservationZscore(t *testing.T) {
	h := observation(t, satusehat.SystemLOINC, "29463-7", 7.4, "OI000007")
	assert.Equal(t, "berat_badan", h.Kolom)
	assert.Empty(t, h.Status, "interpretasi berat badan bukan status BB/U")

	h = observation(t, satusehat.SystemLOINC, "8306-3", 72.5, "OI000011")
	assert.Equal(t, "tinggi_badan", h.Kolom)
	assert.Empty(t, h.Status)

	h = observation(t, satusehat.SystemSNOMED, "1153596006", 1.2, "248324001")
	assert.Empty(t, h.Kolom, "IMT/U tidak punya kolom kunjungan")
	assert.Empty(t, h.Status)

	h = observation(t, satusehat.SystemSNOMED, "1153593003", -2.5, "248325000")
	assert.Equal(t, "zscore_bbu", h.Kolom)
	assert.Empty(t, h.Status, "248325000 milik BB/PB-BB/TB, bukan BB/U")
}

func TestJalurFaskesSamaDenganFHIR(t *testing.T) {
	value, interp := -2.23, "248342006"
	h := satusehat.HasilDariObservasi(types.Observasi{
		System:       satusehat.SystemSNOMED,
		Kode:         "1153593003",
		NilaiAngka:   &value,
		Interpretasi: &interp,
	}, nil)
	assert.Equal(t, "zscore_bbu", h.Kolom)
	assert.Equal(t, "Berat Badan Kurang", h.Status)

	h = satusehat.HasilDariObservasi(types.Observasi{System: satusehat.SystemSNOMED, Kode: "1153593003", NilaiAngka: &value}, nil)
	assert.Equal(t, "zscore_bbu", h.Kolom)
	assert.Empty(t, h.Status)
}
