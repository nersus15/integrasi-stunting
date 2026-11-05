package stream_kafka_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryRusakDilewatiTanpaMenggagalkanBundle(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "04-kafka-puskesmas.json")

	rusak := func(id, tipe, field string) res {
		return res{id, map[string]any{"resourceType": tipe, field: 12345}}
	}
	sendBundle(t, "uji-rusak", ihsAnak,
		res{"tanpa-tipe", map[string]any{"status": "x"}},
		rusak("r-pat", "Patient", "gender"),
		rusak("r-rp", "RelatedPerson", "gender"),
		rusak("r-enc", "Encounter", "status"),
		rusak("r-obs", "Observation", "status"),
		rusak("r-cond", "Condition", "subject"),
		rusak("r-proc", "Procedure", "status"),
		rusak("r-sr", "ServiceRequest", "status"),
		rusak("r-ep", "EpisodeOfCare", "status"),
		rusak("r-alg", "AllergyIntolerance", "patient"),
		observasi("obs-uji-selamat", ihsAnak, "enc-pkm-01", "8302-2", 72.6, "OI000011"),
	)

	assert.Zero(t, failedKafkaCount(t))
	assert.Equal(t, 1, count(t, `SELECT count(*) FROM stunting.observasi WHERE satusehat_id = 'obs-uji-selamat'`),
		"entry yang sah tetap tersimpan")
}

func TestEntryTanpaResourceIDMenggagalkanPesan(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json")

	a := envelope{}
	require.NoError(t, json.Unmarshal(bundle(t, "uji-sebagian", ihsAnak,
		encounter("enc-uji-sebagian", ihsAnak, "100011"),
		observasi("obs-uji-gagal", ihsAnak, "enc-uji-sebagian", "8302-2", 70, "OI000011"),
		observasi("obs-uji-sebagian", ihsAnak, "enc-uji-sebagian", "8302-2", 70, "OI000011"),
	), &a))
	a["response"].([]any)[1] = map[string]any{"status": "400 Bad Request"}
	wiring.Kafka.Run(context.Background(), a.bytes(t))

	assert.Equal(t, 1, failedKafkaCount(t))
	assert.Contains(t, scalar[string](t, `SELECT error::text FROM kafka.transactions WHERE transaction_id = 'uji-sebagian'`), "resourceID")
	for _, q := range []string{
		`SELECT count(*) FROM stunting.kunjungan WHERE satusehat_id = 'enc-uji-sebagian'`,
		`SELECT count(*) FROM stunting.observasi WHERE satusehat_id IN ('obs-uji-gagal', 'obs-uji-sebagian')`,
	} {
		assert.Zero(t, count(t, q), "tidak ada yang tersimpan dari bundle yang gagal")
	}
}

func TestArahRujukanDisimpulkanDariFaskes(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "04-kafka-puskesmas.json")

	sr := func(id, enc, asal, tujuan string) res {
		r := map[string]any{
			"resourceType": "ServiceRequest", "status": "active", "intent": "order",
			"subject":    ref("Patient/" + ihsAnak),
			"encounter":  ref("Encounter/" + enc),
			"authoredOn": "2026-03-02",
		}
		if asal != "" {
			r["requester"] = ref("Organization/" + asal)
		}
		if tujuan != "" {
			r["performer"] = []any{ref("Organization/" + tujuan)}
		}
		return res{id, r}
	}

	sendBundle(t, "uji-arah", ihsAnak,
		sr("sr-uji-tanpa-tujuan", "enc-pkm-01", "100011", ""),
		sr("sr-uji-sama", "enc-pkm-01", "100011", "100011"),
		sr("sr-uji-balik", "enc-pkm-01", "200022", "100011"),
		sr("sr-uji-keluar", "enc-pkm-01", "100011", "999999"),
	)

	for id, jenis := range map[string]string{
		"sr-uji-tanpa-tujuan": "internal",
		"sr-uji-sama":         "internal",
		"sr-uji-balik":        "rujuk_balik",
		"sr-uji-keluar":       "rujukan",
	} {
		assert.Equal(t, jenis, scalar[string](t, `SELECT jenis FROM stunting.rujukan WHERE satusehat_id = $1`, id), id)
	}
}

func TestTandaStuntingDariKomponenDanDiagnosa(t *testing.T) {
	cases := map[string]res{
		"interpretasi komponen": {"obs-uji-k1", map[string]any{
			"resourceType": "Observation", "status": "final", "code": coding("http://loinc.org", "x"),
			"subject": ref("Patient/" + ihsAnak), "encounter": ref("Encounter/enc-uji-tanda"),
			"component": []any{map[string]any{"code": coding("http://loinc.org", "y"), "valueString": "n",
				"interpretation": []any{coding("http://terminology.kemkes.go.id/CodeSystem/clinical-term", "OI000011")}}},
		}},
		"nilai kode komponen": {"obs-uji-k2", map[string]any{
			"resourceType": "Observation", "status": "final", "code": coding("http://loinc.org", "x"),
			"subject": ref("Patient/" + ihsAnak), "encounter": ref("Encounter/enc-uji-tanda"),
			"component": []any{map[string]any{"code": coding("http://loinc.org", "y"),
				"valueCodeableConcept": coding("http://snomed.info/sct", "444000005")}},
		}},
		"nilai kode induk": {"obs-uji-k3", map[string]any{
			"resourceType": "Observation", "status": "final", "code": coding("http://loinc.org", "x"),
			"subject": ref("Patient/" + ihsAnak), "encounter": ref("Encounter/enc-uji-tanda"),
			"valueCodeableConcept": coding("http://snomed.info/sct", "444000005"),
		}},
		"diagnosa": {"cond-uji-k4", map[string]any{
			"resourceType": "Condition", "code": coding("http://hl7.org/fhir/sid/icd-10", "E45"),
			"subject": ref("Patient/" + ihsAnak), "encounter": ref("Encounter/enc-uji-tanda"),
		}},
	}

	for name, r := range cases {
		t.Run(name, func(t *testing.T) {
			setupJakantro(t)
			send(t, "02-kafka-patient.json")
			sendBundle(t, "uji-tanda", ihsAnak, encounter("enc-uji-tanda", ihsAnak, "100011"), r)
			assert.Equal(t, "1", scalar[string](t, `SELECT stunting::text FROM stunting.kunjungan WHERE satusehat_id = 'enc-uji-tanda'`))
		})
	}
}

func TestTurunanTanpaEncounterMencariKunjunganLewatReferensi(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "04-kafka-puskesmas.json")

	sendBundle(t, "uji-diag-saja", ihsAnak, res{"cond-uji-saja", map[string]any{
		"resourceType": "Condition", "code": coding("http://hl7.org/fhir/sid/icd-10", "E45"),
		"subject": ref("Patient/" + ihsAnak), "encounter": ref("Encounter/enc-pkm-01"),
	}})
	sendBundle(t, "uji-layanan-saja", ihsAnak, res{"proc-uji-saja", map[string]any{
		"resourceType": "Procedure", "status": "completed", "code": coding("http://snomed.info/sct", "1"),
		"subject": ref("Patient/" + ihsAnak), "encounter": ref("Encounter/enc-pkm-01"),
	}})

	for _, q := range []string{
		`SELECT count(*) FROM stunting.diagnosa d JOIN stunting.kunjungan k ON k.id = d.id_kunjungan WHERE d.satusehat_id = 'cond-uji-saja' AND k.satusehat_id = 'enc-pkm-01'`,
		`SELECT count(*) FROM stunting.layanan l JOIN stunting.kunjungan k ON k.id = l.id_kunjungan WHERE l.satusehat_id = 'proc-uji-saja' AND k.satusehat_id = 'enc-pkm-01'`,
	} {
		assert.Equal(t, 1, count(t, q))
	}
}

func TestPUTAllergyIntolerance(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "04-kafka-puskesmas.json")

	alergi := map[string]any{
		"resourceType": "AllergyIntolerance", "code": coding("http://snomed.info/sct", "91935009"),
		"patient": ref("Patient/" + ihsAnak), "encounter": ref("Encounter/enc-pkm-01"), "criticality": "low",
	}
	sendBundle(t, "uji-alergi", ihsAnak, res{"alg-uji-put", alergi})
	alergi["criticality"] = "high"
	sendPUT(t, "uji-alergi-put", alergi, "alg-uji-put")

	assert.Equal(t, "high", scalar[string](t, `SELECT kritikalitas FROM stunting.diagnosa WHERE satusehat_id = 'alg-uji-put'`))
}

func TestRelatedPersonMemperbaruiOrangtuaLama(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json")

	sendBundle(t, "uji-rp-ubah", ihsAnak, res{"RP-UJI-UBAH", map[string]any{
		"resourceType": "RelatedPerson",
		"patient":      ref("Patient/" + ihsAnak),
		"identifier":   []any{map[string]any{"system": "https://fhir.kemkes.go.id/id/nik", "value": nikIbu}},
		"gender":       "female",
		"name":         []any{map[string]any{"text": "Rina M."}},
		"telecom":      []any{map[string]any{"system": "phone", "value": "0877"}},
	}})

	assert.Zero(t, failedKafkaCount(t))
	assert.Equal(t, "Rina M./0877/RP-UJI-UBAH", scalar[string](t,
		`SELECT nama_ibu || '/' || telepon || '/' || satusehat_id FROM stunting.orangtua WHERE nik = $1`, nikIbu))
}

func TestRelatedPersonBaruTidakValidDanTanpaAnak(t *testing.T) {
	cleanupNewData()
	t.Cleanup(cleanupNewData)

	sendBundle(t, "uji-rp-tanpa-peran", "", res{"RP-UJI-NP", map[string]any{
		"resourceType": "RelatedPerson",
		"identifier":   []any{map[string]any{"system": "https://fhir.kemkes.go.id/id/nik", "value": nikIbuBr}},
		"name":         []any{map[string]any{"text": "Tanpa Peran"}},
	}})
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.orangtua WHERE nik = $1`, nikIbuBr),
		"tanpa nama ayah/ibu tidak lolos validasi")

	sendBundle(t, "uji-rp-tanpa-anak", "", res{"RP-UJI-TA", map[string]any{
		"resourceType": "RelatedPerson",
		"identifier":   []any{map[string]any{"system": "https://fhir.kemkes.go.id/id/nik", "value": nikIbuBr}},
		"gender":       "female",
		"name":         []any{map[string]any{"text": "Ibu Sendiri"}},
	}})
	assert.Equal(t, "Ibu Sendiri", scalar[string](t, `SELECT nama_ibu FROM stunting.orangtua WHERE nik = $1`, nikIbuBr),
		"tetap tersimpan walau belum ada anak yang dihubungkan")
	assert.Zero(t, failedKafkaCount(t))
}

func TestSimpanPemeriksaanFaskesJalurGagal(t *testing.T) {
	setupJakantro(t)
	stream := wiring.Stream

	errorCode := func(err error) int { return exceptions.Classify(err).ErrorCode }
	anakBaru := types.Anak{Nik: ptr(nikBaru), TanggalLahir: "2025-01-10", Nama: "B", JenisKelamin: "L"}
	stunting := []types.Diagnosa{{IdSatusehat: ptr("cond-uji-pf"), Kode: "E45", Jenis: entity.DiagnosaDiagnosis}}
	dasar := func() *types.PemeriksaanFaskes {
		return &types.PemeriksaanFaskes{Anak: anakBaru,
			Kunjungan: types.Kunjungan{IdSatusehat: ptr("enc-uji-pf"), TanggalPengukuran: "2026-03-01"}}
	}
	t.Cleanup(cleanupNewData)

	_, err := stream.SimpanPemeriksaanFaskes(nil, "100011")
	assert.Equal(t, exceptions.BodyRusak.ErrorCode, errorCode(err))
	_, err = stream.SimpanPemeriksaanFaskes(dasar(), "")
	assert.Equal(t, exceptions.Forbidden.ErrorCode, errorCode(err))
	_, err = stream.SimpanPemeriksaanFaskes(dasar(), "tidak-terdaftar")
	assert.Equal(t, exceptions.Forbidden.ErrorCode, errorCode(err))

	_, err = stream.SimpanPemeriksaanFaskes(dasar(), "100011")
	assert.Equal(t, exceptions.TidakDisimpan.ErrorCode, errorCode(err), "anak baru tanpa tanda stunting")

	tua := dasar()
	tua.Anak.TanggalLahir = "2019-01-01"
	tua.Diagnosa = stunting
	_, err = stream.SimpanPemeriksaanFaskes(tua, "100011")
	assert.Equal(t, exceptions.TidakDisimpan.ErrorCode, errorCode(err), "lewat 5 tahun")

	salah := dasar()
	salah.Anak.TanggalLahir = "bukan"
	salah.Diagnosa = stunting
	_, err = stream.SimpanPemeriksaanFaskes(salah, "100011")
	assert.Equal(t, exceptions.Validasi.ErrorCode, errorCode(err))

	old := dasar()
	old.Anak = types.Anak{Nik: ptr(nikAnak)}
	_, err = stream.SimpanPemeriksaanFaskes(old, "100011")
	assert.Equal(t, exceptions.TidakDisimpan.ErrorCode, errorCode(err), "anak lama tanpa riwayat stunting")

	meluap := dasar()
	meluap.Anak = types.Anak{Nik: ptr(nikAnak)}
	meluap.Diagnosa = stunting
	meluap.Observasi = []types.Observasi{{IdSatusehat: ptr("obs-uji-luap"), System: "s", Kode: "k", NilaiAngka: ptrF(1e12)}}
	_, err = stream.SimpanPemeriksaanFaskes(meluap, "100011")
	assert.Equal(t, exceptions.Constraint.ErrorCode, errorCode(err), "nilai di luar batas kolom")
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.kunjungan WHERE satusehat_id = 'enc-uji-pf'`), "batal utuh")
}

func TestPastikanDanCariAnak(t *testing.T) {
	_, err := wiring.Stream.PastikanAnak(nil, nil)
	assert.Error(t, err)

	a := &types.Anak{Id: "sudah-ada"}
	result, err := wiring.Stream.PastikanAnak(a, nil)
	require.NoError(t, err)
	assert.Same(t, a, result)

	assert.Nil(t, wiring.Stream.CariAnak(nil))
	assert.Nil(t, wiring.Stream.CariAnak(&types.Anak{IdSatusehat: ptr("P-TIDAK-ADA"), Nik: ptr("3175049999999999")}))
}

func ptr(s string) *string    { return &s }
func ptrF(f float64) *float64 { return &f }
