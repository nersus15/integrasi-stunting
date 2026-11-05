package stream_kafka_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const encPkm = "urn:uuid:bb000002-0000-0000-0000-000000000002"

func obsZscore(encounter, coding string, value float64, interpretasi string) map[string]any {
	r := map[string]any{
		"resourceType": "Observation",
		"status":       "final",
		"category": []any{map[string]any{"coding": []any{map[string]any{
			"system": "http://terminology.hl7.org/CodeSystem/observation-category", "code": "exam", "display": "Exam",
		}}}},
		"code":              map[string]any{"coding": []any{map[string]any{"system": "http://snomed.info/sct", "code": coding}}},
		"subject":           map[string]any{"reference": "Patient/" + ihsAnak},
		"encounter":         map[string]any{"reference": encounter},
		"effectiveDateTime": "2026-02-05T01:15:00+00:00",
		"valueQuantity":     map[string]any{"value": value, "unit": "{Zscore}", "system": "http://unitsofmeasure.org", "code": "{Zscore}"},
	}
	if interpretasi != "" {
		system := "http://snomed.info/sct"
		if strings.HasPrefix(interpretasi, "OI") {
			system = "http://terminology.kemkes.go.id/CodeSystem/clinical-term"
		}
		r["interpretation"] = []any{map[string]any{"coding": []any{map[string]any{"system": system, "code": interpretasi}}}}
	}
	return r
}

func obsLOINC(encounter, coding string, value float64, satuan string) map[string]any {
	return map[string]any{
		"resourceType":      "Observation",
		"status":            "final",
		"code":              map[string]any{"coding": []any{map[string]any{"system": "http://loinc.org", "code": coding}}},
		"subject":           map[string]any{"reference": "Patient/" + ihsAnak},
		"encounter":         map[string]any{"reference": encounter},
		"effectiveDateTime": "2026-02-05T01:15:00+00:00",
		"valueQuantity":     map[string]any{"value": value, "unit": satuan, "system": "http://unitsofmeasure.org", "code": satuan},
	}
}

func emptyEnvelope(trx string) envelope {
	return envelope{
		"method":        "POST",
		"transactionId": trx,
		"patient":       ihsAnak,
		"input":         map[string]any{"resourceType": "Bundle", "type": "transaction", "entry": []any{}},
		"response":      []any{},
	}
}

func (a envelope) add(r map[string]any, id string) envelope {
	in := a["input"].(map[string]any)
	n := len(in["entry"].([]any))
	in["entry"] = append(in["entry"].([]any), map[string]any{
		"fullUrl":  fmt.Sprintf("urn:uuid:cc000000-0000-0000-0000-%012d", n),
		"resource": r,
		"request":  map[string]any{"method": "POST", "url": "Observation"},
	})
	a["response"] = append(a["response"].([]any), map[string]any{"status": "201 Created", "resourceID": id, "resourceType": "Observation"})
	return a
}

func sendEnvelope(t *testing.T, a envelope) {
	t.Helper()
	wiring.Kafka.Run(context.Background(), a.bytes(t))
	assert.Zero(t, count(t, `SELECT count(*) FROM kafka.transactions WHERE group_id = $1`, testGroup),
		"pesan %v tidak boleh gagal", a["transactionId"])
}

func statusKunjungan(t *testing.T, encounter string) [6]string {
	t.Helper()
	var r [6]string
	err := db.QueryRow(`
		SELECT coalesce(zscore_bbu::text,'NULL'), coalesce(status_bbu,'NULL'),
		       coalesce(zscore_tbu::text,'NULL'), coalesce(status_tbu,'NULL'),
		       coalesce(zscore_bbtb::text,'NULL'), coalesce(status_bbtb,'NULL')
		FROM stunting.kunjungan WHERE satusehat_id = $1`, encounter).
		Scan(&r[0], &r[1], &r[2], &r[3], &r[4], &r[5])
	if err != nil {
		t.Fatal(err)
	}
	return r
}

var kosong = [6]string{"NULL", "NULL", "NULL", "NULL", "NULL", "NULL"}

func TestZscoreDariObservationMengisiKunjungan(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "03-kafka-relatedperson.json")

	a := load(t, "04-kafka-puskesmas.json").
		add(obsZscore(encPkm, "1153593003", -3.05, "OI000007"), "obs-z-bbu").
		add(obsZscore(encPkm, "1153590000", -3.21, "OI000011"), "obs-z-pbu").
		add(obsZscore(encPkm, "1153598007", -1.94, "248324001"), "obs-z-bbpb")
	sendEnvelope(t, a)

	assert.Equal(t, [6]string{
		"-3.05", "Berat Badan Sangat Kurang",
		"-3.21", "Sangat Pendek",
		"-1.94", "Gizi Baik",
	}, statusKunjungan(t, "enc-pkm-01"))
	assert.Equal(t, 3, count(t, `SELECT count(*) FROM stunting.observasi WHERE satusehat_id LIKE 'obs-z-%' AND id_kunjungan IS NOT NULL`))
}

func TestZscoreTinggiDanBeratPerTinggi(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "03-kafka-relatedperson.json")

	a := load(t, "04-kafka-puskesmas.json").
		add(obsZscore(encPkm, "1153604005", 3.12, "83077003"), "obs-z-tbu").
		add(obsZscore(encPkm, "1153600001", 3.33, "414915002"), "obs-z-bbtb")
	sendEnvelope(t, a)

	assert.Equal(t, [6]string{"NULL", "NULL", "3.12", "Tinggi", "3.33", "Obesitas"}, statusKunjungan(t, "enc-pkm-01"))
}

func TestInterpretasiObservationLainTidakMengisiStatus(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "03-kafka-relatedperson.json", "04-kafka-puskesmas.json")

	assert.Equal(t, kosong, statusKunjungan(t, "enc-pkm-01"),
		"OI000007 di berat badan dan OI000011 di tinggi badan bukan observation z-score")
}

func TestZscoreTanpaInterpretasiStatusNull(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "03-kafka-relatedperson.json")

	a := load(t, "04-kafka-puskesmas.json").
		add(obsZscore(encPkm, "1153593003", -2.47, ""), "obs-z-bbu")
	sendEnvelope(t, a)

	assert.Equal(t, [6]string{"-2.47", "NULL", "NULL", "NULL", "NULL", "NULL"}, statusKunjungan(t, "enc-pkm-01"))
}

func TestZscoreIMTDanLingkarKepalaHanyaDiObservasi(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "03-kafka-relatedperson.json")

	a := load(t, "04-kafka-puskesmas.json").
		add(obsZscore(encPkm, "1153596006", 1.2, "248324001"), "obs-z-imtu").
		add(obsZscore(encPkm, "1153594009", 0.5, ""), "obs-z-lku")
	sendEnvelope(t, a)

	assert.Equal(t, kosong, statusKunjungan(t, "enc-pkm-01"))
	assert.Equal(t, 2, count(t, `SELECT count(*) FROM stunting.observasi WHERE satusehat_id IN ('obs-z-imtu','obs-z-lku') AND id_kunjungan IS NOT NULL`))
}

func TestObservationBundleBerikutnyaMemperbaruiKunjungan(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "03-kafka-relatedperson.json", "04-kafka-puskesmas.json")

	enc := "Encounter/enc-pkm-01"
	a := emptyEnvelope("sim1-pkm-01-susulan").
		add(obsLOINC(enc, "29463-7", 7.6, "kg"), "obs-susulan-bb").
		add(obsLOINC(enc, "8308-9", 73, "cm"), "obs-susulan-tb").
		add(obsZscore(enc, "1153593003", -2.23, "248342006"), "obs-susulan-z-bbu").
		add(obsZscore(enc, "1153604005", -2.41, "444000005"), "obs-susulan-z-tbu")
	sendEnvelope(t, a)

	assert.Equal(t, [6]string{"-2.23", "Berat Badan Kurang", "-2.41", "Pendek", "NULL", "NULL"}, statusKunjungan(t, "enc-pkm-01"))
	assert.Equal(t, "7.6/73/berdiri", scalar[string](t,
		`SELECT berat_badan::text || '/' || tinggi_badan::text || '/' || cara_ukur FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`))
	assert.Equal(t, 1, count(t, `SELECT count(*) FROM stunting.kunjungan WHERE id_anak = $1 AND satusehat_id = 'enc-pkm-01'`, idAnakJkt),
		"kunjungan diperbarui, bukan dibuat ulang")
	assert.Equal(t, 4, count(t, `SELECT count(*) FROM stunting.observasi WHERE satusehat_id LIKE 'obs-susulan-%' AND id_kunjungan IS NOT NULL`))
}

func TestEncounterUlangMembawaZscoreMemperbaruiKunjungan(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "03-kafka-relatedperson.json", "04-kafka-puskesmas.json")
	assert.Equal(t, kosong, statusKunjungan(t, "enc-pkm-01"))

	a := load(t, "04-kafka-puskesmas.json").
		add(obsZscore(encPkm, "1153590000", -3.21, "OI000011"), "obs-z-pbu")
	a["transactionId"] = "sim1-pkm-01-ulang"
	sendEnvelope(t, a)

	assert.Equal(t, [6]string{"NULL", "NULL", "-3.21", "Sangat Pendek", "NULL", "NULL"}, statusKunjungan(t, "enc-pkm-01"))
	assert.Equal(t, 1, count(t, `SELECT count(*) FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`))
}

func setupZscore(t *testing.T) {
	t.Helper()
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "03-kafka-relatedperson.json")
	sendEnvelope(t, load(t, "04-kafka-puskesmas.json").
		add(obsZscore(encPkm, "1153604005", -3.21, "OI000011"), "obs-z-tbu"))
}

func kunjunganRow(t *testing.T, encounter string) string {
	t.Helper()
	return scalar[string](t, `SELECT (to_jsonb(k) - 'berat_badan' - 'updatedAt')::text FROM stunting.kunjungan k WHERE satusehat_id = $1`, encounter)
}

func TestPUTObservationZscoreMemperbaruiKunjungan(t *testing.T) {
	setupZscore(t)

	sendPUT(t, "put-z-tbu", obsZscore("Encounter/enc-pkm-01", "1153604005", -1.52, "17489000"), "obs-z-tbu")

	assert.Equal(t, [6]string{"NULL", "NULL", "-1.52", "Normal", "NULL", "NULL"}, statusKunjungan(t, "enc-pkm-01"))
	assert.Equal(t, "-1.52", scalar[string](t, `SELECT nilai_angka::float8::text FROM stunting.observasi WHERE satusehat_id = 'obs-z-tbu'`))
}

func TestPUTObservationUkurMemperbaruiKunjungan(t *testing.T) {
	setupZscore(t)
	before := kunjunganRow(t, "enc-pkm-01")

	r := copyResource(t, "04-kafka-puskesmas.json", 2)
	r["valueQuantity"].(map[string]any)["value"] = 7.6
	r["encounter"] = map[string]any{"reference": "Encounter/enc-pkm-01"}
	sendPUT(t, "put-bb", r, "obs-pkm-bb-01")

	assert.Equal(t, before, kunjunganRow(t, "enc-pkm-01"), "selain berat_badan, baris kunjungan tidak berubah")

	assert.Equal(t, "7.6/72.5", scalar[string](t,
		`SELECT berat_badan::text || '/' || tinggi_badan::text FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`))
	assert.Equal(t, [6]string{"NULL", "NULL", "-3.21", "Sangat Pendek", "NULL", "NULL"}, statusKunjungan(t, "enc-pkm-01"),
		"kolom yang tidak dibawa PUT tidak tersentuh")
}

func TestPUTObservationBelumTersimpanTidakMengubahKunjungan(t *testing.T) {
	setupZscore(t)

	sendPUT(t, "put-z-baru", obsZscore("Encounter/enc-pkm-01", "1153593003", -2.9, "OI000007"), "obs-z-belum-ada")

	assert.Equal(t, [6]string{"NULL", "NULL", "-3.21", "Sangat Pendek", "NULL", "NULL"}, statusKunjungan(t, "enc-pkm-01"))
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.observasi WHERE satusehat_id = 'obs-z-belum-ada'`))
}

func obsLILA(encounter string, value float64) map[string]any {
	r := obsZscore(encounter, "284473002", value, "")
	r["valueQuantity"] = map[string]any{"value": value, "unit": "cm", "system": "http://unitsofmeasure.org", "code": "cm"}
	return r
}

func TestLILAMengisiLingkarLengan(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "03-kafka-relatedperson.json")
	sendEnvelope(t, load(t, "04-kafka-puskesmas.json").add(obsLILA(encPkm, 12.4), "obs-lila"))

	assert.Equal(t, "12.4", scalar[string](t, `SELECT lingkar_lengan::text FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`))

	sendEnvelope(t, emptyEnvelope("sim1-lila-susulan").add(obsLILA("Encounter/enc-pkm-01", 12.9), "obs-lila-susulan"))
	assert.Equal(t, "12.9", scalar[string](t, `SELECT lingkar_lengan::text FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`))
}

func TestZscoreBaruTanpaInterpretasiMengosongkanStatusLama(t *testing.T) {
	setupZscore(t)

	sendPUT(t, "put-z-tanpa-interp", obsZscore("Encounter/enc-pkm-01", "1153604005", -1.52, ""), "obs-z-tbu")
	assert.Equal(t, [6]string{"NULL", "NULL", "-1.52", "NULL", "NULL", "NULL"}, statusKunjungan(t, "enc-pkm-01"),
		"status lama Sangat Pendek tidak boleh menempel pada z-score baru")

	sendEnvelope(t, emptyEnvelope("sim1-z-susulan").
		add(obsZscore("Encounter/enc-pkm-01", "1153593003", -2.23, "248342006"), "obs-z-bbu"))
	sendEnvelope(t, emptyEnvelope("sim1-z-susulan-2").
		add(obsZscore("Encounter/enc-pkm-01", "1153593003", -1.1, ""), "obs-z-bbu-2"))
	assert.Equal(t, [6]string{"-1.10", "NULL", "-1.52", "NULL", "NULL", "NULL"}, statusKunjungan(t, "enc-pkm-01"))
}
