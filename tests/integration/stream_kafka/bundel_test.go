package stream_kafka_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	nikBaru  = "3175049900000001"
	ihsBaru  = "P-UJI-000001"
	nikIbuBr = "3175049900000002"
)

type res struct {
	id string
	r  map[string]any
}

func bundle(t *testing.T, trx, patient string, content ...res) []byte {
	t.Helper()
	var entries, answers []any
	for _, x := range content {
		tipe, _ := x.r["resourceType"].(string)
		entries = append(entries, map[string]any{
			"fullUrl":  "urn:uuid:" + x.id,
			"resource": x.r,
			"request":  map[string]any{"method": "POST", "url": tipe},
		})
		answers = append(answers, map[string]any{"status": "201 Created", "resourceID": x.id, "resourceType": tipe})
	}
	return envelope{
		"method":        "POST",
		"transactionId": trx,
		"patient":       patient,
		"input":         map[string]any{"resourceType": "Bundle", "type": "transaction", "entry": entries},
		"response":      answers,
	}.bytes(t)
}

func sendBundle(t *testing.T, trx, patient string, content ...res) {
	t.Helper()
	wiring.Kafka.Run(context.Background(), bundle(t, trx, patient, content...))
}

func coding(system, code string) map[string]any {
	return map[string]any{"coding": []any{map[string]any{"system": system, "code": code, "display": code}}}
}

func ref(r string) map[string]any { return map[string]any{"reference": r} }

func pasien(ihs, nik, lahir string) res {
	r := map[string]any{
		"resourceType": "Patient",
		"identifier":   []any{map[string]any{"system": "https://fhir.kemkes.go.id/id/nik", "value": nik}},
		"name":         []any{map[string]any{"use": "official", "text": "Anak Uji"}},
		"gender":       "male",
	}
	if lahir != "" {
		r["birthDate"] = lahir
	}
	return res{ihs, r}
}

func encounter(id, ihs, org string) res {
	return res{id, map[string]any{
		"resourceType":    "Encounter",
		"status":          "finished",
		"class":           map[string]any{"system": "http://terminology.hl7.org/CodeSystem/v3-ActCode", "code": "AMB"},
		"subject":         ref("Patient/" + ihs),
		"period":          map[string]any{"start": "2026-03-01T01:00:00+00:00"},
		"serviceProvider": ref("Organization/" + org),
	}}
}

func observasi(id, ihs, enc, loinc string, value float64, interp string) res {
	r := map[string]any{
		"resourceType":      "Observation",
		"status":            "final",
		"code":              coding("http://loinc.org", loinc),
		"subject":           ref("Patient/" + ihs),
		"encounter":         ref("Encounter/" + enc),
		"effectiveDateTime": "2026-03-01T01:10:00+00:00",
		"valueQuantity":     map[string]any{"value": value, "unit": "cm"},
	}
	if interp != "" {
		r["interpretation"] = []any{coding("http://terminology.kemkes.go.id/CodeSystem/clinical-term", interp)}
	}
	return res{id, r}
}

func cleanupNewData() {
	_, _ = db.Exec(`DELETE FROM stunting.anak WHERE nik = $1 OR satusehat_id = $2`, nikBaru, ihsBaru)
	_, _ = db.Exec(`DELETE FROM stunting.orangtua WHERE nik = $1`, nikIbuBr)
	_, _ = db.Exec(`DELETE FROM stunting.episode WHERE satusehat_id LIKE 'ep-uji-%'`)
	cleanup()
}

func failedKafkaCount(t *testing.T) int {
	return count(t, `SELECT count(*) FROM kafka.transactions WHERE group_id = $1`, testGroup)
}

func TestAnakBaruLengkapDariSatuSehat(t *testing.T) {
	cleanupNewData()
	t.Cleanup(cleanupNewData)

	anakKe := res{"obs-uji-anakke", map[string]any{
		"resourceType": "Observation", "status": "final",
		"code":         coding("http://snomed.info/sct", "365475008"),
		"subject":      ref("Patient/" + ihsBaru),
		"valueInteger": 2,
	}}
	lk := observasi("obs-uji-lk", ihsBaru, "enc-uji-1", "9843-4", 44.5, "")
	tb := observasi("obs-uji-tb", ihsBaru, "enc-uji-1", "8306-3", 70.1, "OI000011")
	tb.r["component"] = []any{map[string]any{
		"code":           coding("http://loinc.org", "8302-2"),
		"valueQuantity":  map[string]any{"value": 70.1, "unit": "cm"},
		"interpretation": []any{coding("http://snomed.info/sct", "444000005")},
	}}
	alergi := res{"alg-uji-1", map[string]any{
		"resourceType": "AllergyIntolerance",
		"code":         coding("http://snomed.info/sct", "91935009"),
		"category":     []any{"food"},
		"criticality":  "high",
		"patient":      ref("Patient/" + ihsBaru),
		"encounter":    ref("Encounter/enc-uji-1"),
		"onsetPeriod":  map[string]any{"start": "2026-02-01"},
		"recordedDate": "2026-03-01",
	}}
	imun := res{"imm-uji-1", map[string]any{
		"resourceType":       "Immunization",
		"status":             "completed",
		"vaccineCode":        coding("http://sys-ids.kemkes.go.id/kfa", "VG001"),
		"patient":            ref("Patient/" + ihsBaru),
		"encounter":          ref("Encounter/enc-uji-1"),
		"occurrenceDateTime": "2026-03-01",
		"doseQuantity":       map[string]any{"value": 0.5, "unit": "mL"},
	}}
	ibu := res{"RP-UJI-1", map[string]any{
		"resourceType": "RelatedPerson",
		"patient":      ref("Patient/" + ihsBaru),
		"identifier":   []any{map[string]any{"system": "https://fhir.kemkes.go.id/id/nik", "value": nikIbuBr}},
		"relationship": []any{coding("http://terminology.hl7.org/CodeSystem/v3-RoleCode", "MTH")},
		"name":         []any{map[string]any{"text": "Ibu Uji"}},
		"telecom":      []any{map[string]any{"system": "phone", "value": "0811"}},
		"address":      []any{map[string]any{"text": "Jl. Uji"}},
	}}
	other := []res{
		{"qr-uji-1", map[string]any{"resourceType": "QuestionnaireResponse", "status": "completed"}},
		{"comp-uji-1", map[string]any{"resourceType": "Composition", "status": "final"}},
	}

	content := append([]res{pasien(ihsBaru, nikBaru, "2025-01-10"), ibu, encounter("enc-uji-1", ihsBaru, "100011"), anakKe, lk, tb, alergi, imun}, other...)
	sendBundle(t, "uji-anak-baru", ihsBaru, content...)

	require.Zero(t, failedKafkaCount(t))
	assert.Equal(t, "2/"+ihsBaru, scalar[string](t, `SELECT anak_ke || '/' || satusehat_id FROM stunting.anak WHERE nik = $1`, nikBaru))
	assert.Equal(t, "Ibu Uji/RP-UJI-1", scalar[string](t, `SELECT nama_ibu || '/' || satusehat_id FROM stunting.orangtua WHERE nik = $1`, nikIbuBr))
	assert.Equal(t, 1, count(t, `SELECT count(*) FROM stunting.anak a JOIN stunting.orangtua o ON o.id = a.id_orangtua WHERE a.nik = $1 AND o.nik = $2`, nikBaru, nikIbuBr))
	assert.Equal(t, "44.5/70.1/telentang/1", scalar[string](t,
		`SELECT lingkar_kepala || '/' || tinggi_badan || '/' || cara_ukur || '/' || stunting FROM stunting.kunjungan WHERE satusehat_id = 'enc-uji-1'`))
	assert.Equal(t, "alergi/high", scalar[string](t, `SELECT jenis || '/' || kritikalitas FROM stunting.diagnosa WHERE satusehat_id = 'alg-uji-1'`))
	assert.Equal(t, "immunization/0.5", scalar[string](t, `SELECT jenis || '/' || jumlah::float8 FROM stunting.layanan WHERE satusehat_id = 'imm-uji-1'`))
	assert.Equal(t, 1, count(t, `SELECT count(*) FROM stunting.observasi WHERE id_induk = (SELECT id FROM stunting.observasi WHERE satusehat_id = 'obs-uji-tb')`))
}

func TestPatientDiluarSasaranAtauTidakValid(t *testing.T) {
	cleanupNewData()
	t.Cleanup(cleanupNewData)

	sendBundle(t, "uji-tua", ihsBaru, pasien(ihsBaru, nikBaru, "2019-01-01"))
	assert.Zero(t, failedKafkaCount(t), "lewat 5 tahun bukan kegagalan")
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.anak WHERE nik = $1`, nikBaru))

	sendBundle(t, "uji-masa-depan", ihsBaru, pasien(ihsBaru, nikBaru, "2099-01-01"))
	sendBundle(t, "uji-tanpa-lahir", ihsBaru, pasien(ihsBaru, nikBaru, ""))
	assert.Equal(t, 2, failedKafkaCount(t))
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.anak WHERE nik = $1`, nikBaru))
}

func TestPatientTanpaGenderDanLahirMelengkapiAnakLama(t *testing.T) {
	setupJakantro(t)

	p := pasien(ihsAnak, nikAnak, "")
	delete(p.r, "gender")
	sendBundle(t, "uji-minimal", ihsAnak, p)

	assert.Zero(t, failedKafkaCount(t))
	assert.Equal(t, ihsAnak+"/2024-05-20/P", scalar[string](t,
		`SELECT satusehat_id || '/' || to_char(tanggal_lahir,'YYYY-MM-DD') || '/' || jenis_kelamin FROM stunting.anak WHERE nik = $1`, nikAnak))
}

func TestRelatedPersonTanpaNikDilewati(t *testing.T) {
	setupJakantro(t)

	sendBundle(t, "uji-rp-tanpa-nik", ihsAnak, res{"RP-UJI-X", map[string]any{
		"resourceType": "RelatedPerson",
		"patient":      ref("Patient/" + ihsAnak),
		"gender":       "female",
		"name":         []any{map[string]any{"text": "Tanpa NIK"}},
	}})

	assert.Zero(t, failedKafkaCount(t))
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.orangtua WHERE satusehat_id = 'RP-UJI-X'`))
}

func TestDataMedisTidakMemenuhiSyarat(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json")

	sendBundle(t, "uji-normal", ihsAnak,
		encounter("enc-uji-normal", ihsAnak, "100011"),
		res{"obs-uji-normal", map[string]any{
			"resourceType": "Observation", "status": "final",
			"code":                 coding("http://loinc.org", "8302-2"),
			"subject":              ref("Patient/" + ihsAnak),
			"encounter":            ref("Encounter/enc-uji-normal"),
			"valueCodeableConcept": coding("http://snomed.info/sct", "17489000"),
		}})
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.kunjungan WHERE satusehat_id = 'enc-uji-normal'`),
		"tanpa tanda stunting dan riwayat NULL, tidak disimpan")

	sendBundle(t, "uji-anak-hantu", "P-HANTU",
		observasi("obs-uji-hantu", "P-HANTU", "enc-hantu", "8302-2", 60, "OI000011"))
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.observasi WHERE satusehat_id = 'obs-uji-hantu'`))
	assert.Zero(t, failedKafkaCount(t))
}

func TestTurunanLebihDuluLaluDisapuKunjunganDanEpisode(t *testing.T) {
	cleanupNewData()
	setupJakantro(t)
	t.Cleanup(cleanupNewData)
	send(t, "02-kafka-patient.json", "04-kafka-puskesmas.json")

	sendBundle(t, "uji-obs-dulu", ihsAnak, observasi("obs-uji-dulu", ihsAnak, "enc-uji-telat", "8302-2", 73, "OI000011"))
	assert.Equal(t, 1, count(t, `SELECT count(*) FROM stunting.observasi WHERE satusehat_id = 'obs-uji-dulu' AND id_kunjungan IS NULL`))

	enc := encounter("enc-uji-telat", ihsAnak, "100011")
	enc.r["episodeOfCare"] = []any{ref("EpisodeOfCare/ep-uji-telat")}
	sendBundle(t, "uji-enc-telat", ihsAnak, enc)
	assert.Equal(t, 1, count(t, `SELECT count(*) FROM stunting.observasi o JOIN stunting.kunjungan k ON k.id = o.id_kunjungan
		WHERE o.satusehat_id = 'obs-uji-dulu' AND k.satusehat_id = 'enc-uji-telat'`), "observasi disapu ke kunjungan")
	assert.Equal(t, "NULL", scalar[string](t, `SELECT coalesce(stunting::text,'NULL') FROM stunting.kunjungan WHERE satusehat_id = 'enc-uji-telat'`))

	sendBundle(t, "uji-ep-telat", ihsAnak, res{"ep-uji-telat", map[string]any{
		"resourceType":         "EpisodeOfCare",
		"status":               "active",
		"type":                 []any{coding("http://terminology.hl7.org/CodeSystem/episodeofcare-type", "hacc")},
		"patient":              ref("Patient/" + ihsAnak),
		"managingOrganization": ref("Organization/100011"),
		"period":               map[string]any{"start": "2026-03-01"},
	}})
	assert.Equal(t, "ep-uji-telat", scalar[string](t, `SELECT e.satusehat_id FROM stunting.kunjungan k JOIN stunting.episode e ON e.id = k.id_episode WHERE k.satusehat_id = 'enc-uji-telat'`))

	sendBundle(t, "uji-status-menyusul", ihsAnak, observasi("obs-uji-susul", ihsAnak, "enc-uji-telat", "8302-2", 73, "OI000011"))
	assert.Equal(t, "1", scalar[string](t, `SELECT stunting::text FROM stunting.kunjungan WHERE satusehat_id = 'enc-uji-telat'`),
		"status ditetapkan saat data klinis tiba")
	assert.Zero(t, failedKafkaCount(t))
}

func TestRujukanTanpaRequester(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "04-kafka-puskesmas.json")

	sr := func(id, enc string) res {
		return res{id, map[string]any{
			"resourceType": "ServiceRequest", "status": "active", "intent": "order",
			"subject":    ref("Patient/" + ihsAnak),
			"encounter":  ref("Encounter/" + enc),
			"performer":  []any{ref("Organization/200022")},
			"authoredOn": "2026-03-02",
		}}
	}

	sendBundle(t, "uji-sr-dengan-enc", ihsAnak, encounter("enc-uji-sr", ihsAnak, "100011"),
		observasi("obs-uji-sr", ihsAnak, "enc-uji-sr", "8302-2", 73, "OI000011"), sr("sr-uji-1", "enc-uji-sr"))
	sendBundle(t, "uji-sr-saja", ihsAnak, sr("sr-uji-2", "enc-pkm-01"))

	for _, id := range []string{"sr-uji-1", "sr-uji-2"} {
		assert.Equal(t, "rujukan/100011/200022", scalar[string](t, `SELECT r.jenis || '/' || fa.satusehat_id || '/' || ft.satusehat_id
			FROM stunting.rujukan r JOIN stunting.faskes fa ON fa.id = r.id_faskes_asal JOIN stunting.faskes ft ON ft.id = r.id_faskes_tujuan
			WHERE r.satusehat_id = $1`, id), id)
	}
	assert.Zero(t, failedKafkaCount(t))
}
