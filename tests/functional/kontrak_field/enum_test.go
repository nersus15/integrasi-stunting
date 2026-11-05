package kontrak_field_test

import (
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func diagnosa(jenis string, content map[string]any) map[string]any {
	d := map[string]any{"id_satusehat": "cond-kontrak-02", "jenis": jenis, "system": "http://snomed.info/sct", "kode": "227493005"}
	if jenis == "diagnosis" {
		d["system"], d["kode"] = "http://hl7.org/fhir/sid/icd-10", "R62.8"
	}
	for k, v := range content {
		d[k] = v
	}
	return d
}

func TestEnumFaskesDitolakKalauDiLuarValueSet(t *testing.T) {
	setup(t)

	cases := map[string]struct {
		key string
		isi map[string]any
	}{
		"kategori diagnosis":            {"diagnosa", diagnosa("diagnosis", map[string]any{"kategori": "food"})},
		"kategori alergi":               {"diagnosa", diagnosa("alergi", map[string]any{"kategori": "encounter-diagnosis"})},
		"kritikalitas":                  {"diagnosa", diagnosa("alergi", map[string]any{"kritikalitas": "sedang"})},
		"clinical_status diagnosis":     {"diagnosa", diagnosa("diagnosis", map[string]any{"clinical_status": "aktif"})},
		"clinical_status alergi":        {"diagnosa", diagnosa("alergi", map[string]any{"clinical_status": "remission"})},
		"verification_status diagnosis": {"diagnosa", diagnosa("diagnosis", map[string]any{"verification_status": "pasti"})},
		"verification_status alergi":    {"diagnosa", diagnosa("alergi", map[string]any{"verification_status": "provisional"})},
		"status procedure":              {"layanan", map[string]any{"id_satusehat": "lay-kontrak-01", "jenis": "procedure", "status": "active"}},
		"status immunization":           {"layanan", map[string]any{"id_satusehat": "lay-kontrak-01", "jenis": "immunization", "status": "in-progress"}},
		"status medication_dispense":    {"layanan", map[string]any{"id_satusehat": "lay-kontrak-01", "jenis": "medication_dispense", "status": "selesai"}},
		"status nutrition_order":        {"layanan", map[string]any{"id_satusehat": "lay-kontrak-01", "jenis": "nutrition_order", "status": "not-done"}},
		"status service_request":        {"layanan", map[string]any{"id_satusehat": "lay-kontrak-01", "jenis": "service_request", "status": "finished"}},
		"status rujukan":                {"rujukan", map[string]any{"id_satusehat": "sr-kontrak-01", "jenis": "internal", "status": "selesai"}},
		"prioritas rujukan":             {"rujukan", map[string]any{"id_satusehat": "sr-kontrak-01", "jenis": "internal", "prioritas": "segera"}},
		"status episode":                {"episode", map[string]any{"id_satusehat": "ep-kontrak-01", "status": "aktif"}},
	}
	for name, k := range cases {
		p := newPemeriksaan()
		add(p, k.key, k.isi)
		j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, p)
		assert.Equal(t, 422, j.Kode, "%s: %v", name, j.Isi)
		assert.Equal(t, 1002, j.Galat, name)
	}
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.anak WHERE satusehat_id = $1`, ihsAnakBaru))
}

func TestEnumFaskesDalamValueSetDiterima(t *testing.T) {
	setup(t)

	p := newPemeriksaan()
	add(p, "diagnosa", diagnosa("diagnosis", map[string]any{"id_satusehat": "cond-kontrak-03", "kategori": "problem-list-item",
		"clinical_status": "remission", "verification_status": "differential"}))
	add(p, "diagnosa", diagnosa("alergi", map[string]any{"kategori": "biologic", "kritikalitas": "unable-to-assess",
		"clinical_status": "inactive", "verification_status": "unconfirmed"}))
	add(p, "layanan", map[string]any{"id_satusehat": "lay-kontrak-01", "jenis": "procedure", "status": "not-done", "kategori": "bebas-apa-saja"})
	add(p, "layanan", map[string]any{"id_satusehat": "lay-kontrak-02", "jenis": "immunization", "status": "not-done"})
	add(p, "layanan", map[string]any{"id_satusehat": "lay-kontrak-03", "jenis": "medication_dispense", "status": "declined"})
	add(p, "layanan", map[string]any{"id_satusehat": "lay-kontrak-04", "jenis": "nutrition_order", "status": "revoked"})
	add(p, "layanan", map[string]any{"id_satusehat": "lay-kontrak-05", "jenis": "service_request", "status": "draft"})
	add(p, "rujukan", map[string]any{"id_satusehat": "sr-kontrak-01", "jenis": "internal", "status": "on-hold", "prioritas": "asap"})
	add(p, "episode", map[string]any{"id_satusehat": "ep-kontrak-01", "status": "waitlist"})
	add(p, "observasi", map[string]any{"id_satusehat": "obs-kontrak-01", "system": "http://loinc.org", "kode": "29463-7", "nilai_angka": 7.9, "kategori": "bebas-juga"})
	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, p)
	assert.Equal(t, 201, j.Kode, "%v", j.Isi)
}

func TestEnumDivalidasiSaatPUTMenurutJenisTersimpan(t *testing.T) {
	setup(t)

	p := newPemeriksaan()
	add(p, "diagnosa", diagnosa("alergi", nil))
	add(p, "layanan", map[string]any{"id_satusehat": "lay-kontrak-01", "jenis": "immunization", "status": "completed"})
	require.Equal(t, 201, request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, p).Kode)

	alergi := map[string]any{"id_satusehat": "cond-kontrak-02", "system": "http://snomed.info/sct", "kode": "227493005"}
	alergi["kategori"] = "food"
	j := request(t, "PUT", "/faskes/diagnosa", testutil.KeyFaskesA, alergi)
	assert.Equal(t, 200, j.Kode, "kategori alergi tanpa mengirim jenis: %v", j.Isi)

	alergi["kategori"] = "encounter-diagnosis"
	j = request(t, "PUT", "/faskes/diagnosa", testutil.KeyFaskesA, alergi)
	assert.Equal(t, 422, j.Kode, "%v", j.Isi)
	assert.Equal(t, 1002, j.Galat)

	j = request(t, "PUT", "/faskes/layanan", testutil.KeyFaskesA, map[string]any{"id_satusehat": "lay-kontrak-01", "jenis": "nutrition_order", "status": "active"})
	assert.Equal(t, 422, j.Kode, "yang tersimpan immunization, jenis kiriman tidak dipakai: %v", j.Isi)
	assert.Equal(t, 1002, j.Galat)

	j = request(t, "PUT", "/faskes/layanan", testutil.KeyFaskesA, map[string]any{"id_satusehat": "lay-kontrak-01", "jenis": "immunization", "status": "entered-in-error"})
	assert.Equal(t, 200, j.Kode, "%v", j.Isi)

	for resource, body := range map[string]map[string]any{
		"rujukan": {"id_satusehat": "sr-x", "status": "selesai"},
		"episode": {"id_satusehat": "ep-x", "status": "aktif"},
	} {
		j := request(t, "PUT", "/faskes/"+resource, testutil.KeyFaskesA, body)
		assert.Equal(t, 422, j.Kode, "%s: %v", resource, j.Isi)
		assert.Equal(t, 1002, j.Galat, resource)
	}
}
