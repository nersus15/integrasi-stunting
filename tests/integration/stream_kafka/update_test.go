package stream_kafka_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupRiwayat(t *testing.T) {
	t.Helper()
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "03-kafka-relatedperson.json", "04-kafka-puskesmas.json", "05-kafka-rs.json")
}

func sendPUT(t *testing.T, trx string, resource map[string]any, id string) {
	t.Helper()
	wiring.Kafka.Run(context.Background(), putEnvelope(t, trx, resource, id))
	assert.Zero(t, count(t, `SELECT count(*) FROM kafka.transactions WHERE group_id = $1`, testGroup),
		"PUT %s tidak boleh gagal", id)
}

func TestPUTEncounterMemperbaruiTanggal(t *testing.T) {
	setupRiwayat(t)

	r := copyResource(t, "04-kafka-puskesmas.json", 1)
	r["period"] = map[string]any{"start": "2026-02-06T01:00:00+00:00", "end": "2026-02-07T03:00:00+00:00"}
	r["episodeOfCare"] = []any{map[string]any{"reference": "EpisodeOfCare/ep-stunting-sim1"}}
	sendPUT(t, "put-enc", r, "enc-pkm-01")

	assert.Equal(t, "2026-02-06", scalar[string](t, `SELECT to_char(tanggal_pengukuran,'YYYY-MM-DD') FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`))
	assert.Equal(t, "2026-02-07", scalar[string](t, `SELECT to_char(tanggal_selesai,'YYYY-MM-DD') FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`))
	assert.Equal(t, "7.4", scalar[string](t, `SELECT berat_badan::text FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`),
		"kolom lain tidak ikut tertimpa")
}

func TestPUTObservationMemperbaruiNilai(t *testing.T) {
	setupRiwayat(t)

	r := copyResource(t, "04-kafka-puskesmas.json", 2)
	r["valueQuantity"].(map[string]any)["value"] = 7.6
	r["encounter"] = map[string]any{"reference": "Encounter/enc-pkm-01"}
	sendPUT(t, "put-obs", r, "obs-pkm-bb-01")

	assert.Equal(t, "7.6", scalar[string](t, `SELECT nilai_angka::float8::text FROM stunting.observasi WHERE satusehat_id = 'obs-pkm-bb-01'`))
	assert.Equal(t, 1, count(t, `SELECT count(*) FROM stunting.observasi WHERE satusehat_id = 'obs-pkm-bb-01' AND id_kunjungan IS NOT NULL`))
}

func TestPUTConditionMemperbaruiDiagnosa(t *testing.T) {
	setupRiwayat(t)

	r := copyResource(t, "04-kafka-puskesmas.json", 4)
	r["code"].(map[string]any)["coding"].([]any)[0].(map[string]any)["display"] = "diubah lewat PUT"
	r["encounter"] = map[string]any{"reference": "Encounter/enc-pkm-01"}
	sendPUT(t, "put-cond", r, "cond-pkm-01")

	assert.Equal(t, "diubah lewat PUT", scalar[string](t, `SELECT display FROM stunting.diagnosa WHERE satusehat_id = 'cond-pkm-01'`))
}

func TestPUTServiceRequestTidakMenghapusArahRujukan(t *testing.T) {
	setupRiwayat(t)

	r := copyResource(t, "04-kafka-puskesmas.json", 5)
	r["status"] = "completed"
	r["encounter"] = map[string]any{"reference": "Encounter/enc-pkm-01"}
	sendPUT(t, "put-sr", r, "sr-rujuk-01")

	assert.Equal(t, "completed", scalar[string](t, `SELECT status FROM stunting.rujukan WHERE satusehat_id = 'sr-rujuk-01'`))
	assert.Equal(t, "rujukan", scalar[string](t, `SELECT jenis FROM stunting.rujukan WHERE satusehat_id = 'sr-rujuk-01'`))
}

func TestPUTEpisodeOfCareMenutupEpisode(t *testing.T) {
	setupRiwayat(t)

	r := copyResource(t, "04-kafka-puskesmas.json", 0)
	r["status"] = "finished"
	r["period"] = map[string]any{"start": "2026-02-05T00:00:00+00:00", "end": "2026-04-01T00:00:00+00:00"}
	sendPUT(t, "put-ep", r, "ep-stunting-sim1")

	assert.Equal(t, "finished", scalar[string](t, `SELECT status FROM stunting.episode WHERE satusehat_id = 'ep-stunting-sim1'`))
	assert.Equal(t, "2026-04-01", scalar[string](t, `SELECT to_char(selesai,'YYYY-MM-DD') FROM stunting.episode WHERE satusehat_id = 'ep-stunting-sim1'`))
}

func TestPUTLayanan(t *testing.T) {
	setupRiwayat(t)

	a := load(t, "05-kafka-rs.json")
	resp := a["response"].([]any)
	for i := 4; i <= 5; i++ {
		r := a.entry(i)["resource"].(map[string]any)
		id := resp[i].(map[string]any)["resourceID"].(string)
		r["status"] = "entered-in-error"
		r["encounter"] = map[string]any{"reference": "Encounter/enc-rs-01"}
		r["context"] = map[string]any{"reference": "Encounter/enc-rs-01"}
		sendPUT(t, "put-lay-"+id, r, id)

		assert.Equal(t, "entered-in-error", scalar[string](t, `SELECT status FROM stunting.layanan WHERE satusehat_id = $1`, id), id)
	}
}

func TestPUTResourceYangBelumTersimpanDiabaikan(t *testing.T) {
	setupRiwayat(t)

	cases := []struct {
		berkas string
		i      int
	}{
		{"04-kafka-puskesmas.json", 0},
		{"04-kafka-puskesmas.json", 1},
		{"04-kafka-puskesmas.json", 2},
		{"04-kafka-puskesmas.json", 4},
		{"04-kafka-puskesmas.json", 5},
		{"05-kafka-rs.json", 4},
	}
	for _, k := range cases {
		r := copyResource(t, k.berkas, k.i)
		sendPUT(t, "put-baru", r, "belum-ada-"+r["resourceType"].(string))
	}

	for _, table := range []string{"episode", "kunjungan", "observasi", "diagnosa", "rujukan", "layanan"} {
		assert.Zero(t, count(t, `SELECT count(*) FROM stunting.`+table+` WHERE satusehat_id LIKE 'belum-ada-%'`), table)
	}
}

func TestPUTPatientDanRelatedPersonTanpaDataLamaDiabaikan(t *testing.T) {
	setupJakantro(t)

	p := copyResource(t, "02-kafka-patient.json", 0)
	p["identifier"] = []any{map[string]any{"system": "https://fhir.kemkes.go.id/id/nik", "value": "3175049999990001"}}
	sendPUT(t, "put-pat", p, "P-TIDAK-ADA")

	o := copyResource(t, "03-kafka-relatedperson.json", 0)
	o["identifier"] = []any{map[string]any{"system": "https://fhir.kemkes.go.id/id/nik", "value": "3175049999990002"}}
	sendPUT(t, "put-rp", o, "RP-TIDAK-ADA")

	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.anak WHERE nik = '3175049999990001'`))
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.orangtua WHERE nik = '3175049999990002'`))
}

func TestPUTPatientMelengkapiAnakJakantro(t *testing.T) {
	setupJakantro(t)

	p := copyResource(t, "02-kafka-patient.json", 0)
	p["name"] = []any{map[string]any{"use": "official", "text": "Siti Aminah Putri"}}
	sendPUT(t, "put-pat-2", p, ihsAnak)

	assert.Equal(t, "Siti Aminah Putri", scalar[string](t, `SELECT nama FROM stunting.anak WHERE nik = $1`, nikAnak))
	assert.Equal(t, ihsAnak, scalar[string](t, `SELECT satusehat_id FROM stunting.anak WHERE nik = $1`, nikAnak))
}

func TestPUTServiceRequestTanpaKategoriMenghitungUlangArah(t *testing.T) {
	setupRiwayat(t)
	send(t, "06-kafka-rujuk-balik.json")

	cases := []struct {
		berkas, id, harap string
		i                 int
	}{
		{"04-kafka-puskesmas.json", "sr-rujuk-01", "rujukan", 5},
		{"06-kafka-rujuk-balik.json", "sr-balik-01", "rujuk_balik", 0},
	}
	for _, k := range cases {
		r := copyResource(t, k.berkas, k.i)
		delete(r, "category")
		r["status"] = "completed"
		sendPUT(t, "put-tanpa-kategori-"+k.id, r, k.id)

		assert.Equal(t, k.harap+"/completed", scalar[string](t, `SELECT jenis || '/' || status FROM stunting.rujukan WHERE satusehat_id = $1`, k.id), k.id)
	}

	r := copyResource(t, "04-kafka-puskesmas.json", 5)
	delete(r, "category")
	delete(r, "requester")
	delete(r, "performer")
	sendPUT(t, "put-tanpa-faskes", r, "sr-rujuk-01")
	assert.Equal(t, "rujukan", scalar[string](t, `SELECT jenis FROM stunting.rujukan WHERE satusehat_id = 'sr-rujuk-01'`),
		"faskes asal dan tujuan diambil dari yang tersimpan")

	r = copyResource(t, "04-kafka-puskesmas.json", 5)
	delete(r, "category")
	sendPUT(t, "put-belum-ada", r, "sr-belum-ada")
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.rujukan WHERE satusehat_id = 'sr-belum-ada'`))
}
