package fhir_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/helper/satusehat"
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

func entry(t *testing.T, id, content string) types2.BundleEntry {
	t.Helper()
	var r struct {
		ResourceType string `json:"resourceType"`
	}
	_ = json.Unmarshal([]byte(content), &r)
	return types2.BundleEntry{
		BundleEntry: fhir.BundleEntry{Resource: json.RawMessage(content)},
		Base:        &types2.BaseResource{Resource: fhir.Resource{Id: &id}, ResourceType: &r.ResourceType},
	}
}

func obs(t *testing.T, value string) fhir.Observation {
	t.Helper()
	var o fhir.Observation
	require.NoError(t, json.Unmarshal([]byte(`{"resourceType":"Observation","status":"final","code":{"text":"x"},`+value+`}`), &o))
	return o
}

func text(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

func TestNilaiObservationSemuaBentukValue(t *testing.T) {
	cases := []struct {
		value string
		angka *float64
		teks  string
		kode  string
	}{
		{`"valueQuantity":{"value":7.5,"code":"kg"}`, ptr(7.5), "<nil>", "<nil>"},
		{`"valueQuantity":{"unit":"kg"}`, nil, "<nil>", "<nil>"},
		{`"valueCodeableConcept":{"coding":[{"system":"s","code":"K","display":"D"}]}`, nil, "<nil>", "K"},
		{`"valueCodeableConcept":{"text":"hanya teks"}`, nil, "hanya teks", "<nil>"},
		{`"valueString":"baik"`, nil, "baik", "<nil>"},
		{`"valueBoolean":true`, ptr(1), "<nil>", "<nil>"},
		{`"valueBoolean":false`, ptr(0), "<nil>", "<nil>"},
		{`"valueInteger":3`, ptr(3), "<nil>", "<nil>"},
		{`"valueTime":"10:00:00"`, nil, "10:00:00", "<nil>"},
		{`"valueDateTime":"2026-01-02"`, nil, "2026-01-02", "<nil>"},
		{`"valueRange":{"low":{"value":1},"high":{"value":2}}`, nil, "1 - 2", "<nil>"},
		{`"valueRange":{"high":{"value":2}}`, nil, "- 2", "<nil>"},
		{`"valueRange":{}`, nil, "-", "<nil>"},
		{`"valueRatio":{"numerator":{"value":1},"denominator":{"value":4}}`, nil, "1:4", "<nil>"},
		{`"valueRatio":{}`, nil, ":", "<nil>"},
		{`"valuePeriod":{"start":"2026-01-01","end":"2026-01-05"}`, nil, "2026-01-01 s/d 2026-01-05", "<nil>"},
	}
	for _, k := range cases {
		n := satusehat.NilaiDariObservation(obs(t, k.value))
		assert.Equal(t, k.angka, n.Angka, k.value)
		assert.Equal(t, k.teks, text(n.Teks), k.value)
		assert.Equal(t, k.kode, text(n.Kode), k.value)
	}

	n := satusehat.NilaiDariObservation(obs(t, `"valueQuantity":{"value":7.5,"code":"kg"}`))
	assert.Equal(t, "kg", text(n.Satuan), "satuan jatuh ke code kalau unit kosong")

	c := satusehat.NilaiDariComponent(fhir.ObservationComponent{ValueString: ptrS("komp")})
	assert.Equal(t, "komp", text(c.Teks))
}

func ptr(f float64) *float64 { return &f }
func ptrS(s string) *string  { return &s }

func TestKodeUtamaMendahulukanKodeUkuran(t *testing.T) {
	cc := fhir.CodeableConcept{Coding: []fhir.Coding{
		{Code: ptrS("tanpa-system")},
		{System: ptrS("x"), Code: ptrS("lain")},
		{System: ptrS(satusehat.SystemLOINC), Code: ptrS("29463-7"), Display: ptrS("BB")},
	}}
	sys, code, disp := satusehat.KodeUtama(cc)
	assert.Equal(t, [3]string{satusehat.SystemLOINC, "29463-7", "BB"}, [3]string{sys, code, disp})

	sys, code, _ = satusehat.KodeUtama(fhir.CodeableConcept{Coding: []fhir.Coding{{System: ptrS("x")}, {Code: ptrS("A")}}})
	assert.Equal(t, [2]string{"", "A"}, [2]string{sys, code})

	_, code, disp = satusehat.KodeUtama(fhir.CodeableConcept{Text: ptrS("teks")})
	assert.Equal(t, [2]string{"", "teks"}, [2]string{code, disp})
}

func TestPeranOrangtua(t *testing.T) {
	rel := func(code ...string) []fhir.CodeableConcept {
		var cs []fhir.Coding
		for _, k := range code {
			if k == "" {
				cs = append(cs, fhir.Coding{})
				continue
			}
			cs = append(cs, fhir.Coding{Code: ptrS(k)})
		}
		return []fhir.CodeableConcept{{Coding: cs}}
	}
	laki, perempuan, other := fhir.AdministrativeGenderMale, fhir.AdministrativeGenderFemale, fhir.AdministrativeGenderOther

	assert.Equal(t, "ayah", satusehat.PeranOrangtua(rel("", "fth"), &perempuan), "relationship menang atas gender")
	assert.Equal(t, "ibu", satusehat.PeranOrangtua(rel("MTH"), &laki))
	assert.Equal(t, "ayah", satusehat.PeranOrangtua(rel("X"), &laki))
	assert.Equal(t, "ibu", satusehat.PeranOrangtua(nil, &perempuan))
	assert.Equal(t, "", satusehat.PeranOrangtua(nil, &other))
	assert.Equal(t, "", satusehat.PeranOrangtua(nil, nil))
}

func TestPencarianIdentitas(t *testing.T) {
	home, work := fhir.AddressUseHome, fhir.AddressUseWork
	assert.Equal(t, "rumah", satusehat.FindAlamatText([]fhir.Address{{}, {Use: &work, Line: []string{"kantor"}}, {Use: &home, Line: []string{"rumah"}}}))
	assert.Equal(t, "kantor", satusehat.FindAlamatText([]fhir.Address{{Line: []string{""}}, {Use: &work, Line: []string{"kantor"}}, {Line: []string{"lain"}}}))
	assert.Equal(t, "", satusehat.FindAlamatText(nil))

	official, usual := fhir.NameUseOfficial, fhir.NameUseUsual
	assert.Equal(t, "Resmi", satusehat.FindName([]fhir.HumanName{{}, {Use: &usual, Text: ptrS("Panggilan")}, {Use: &official, Text: ptrS("Resmi")}}))
	assert.Equal(t, "Panggilan", satusehat.FindName([]fhir.HumanName{{Text: ptrS("")}, {Use: &usual, Text: ptrS("Panggilan")}, {Text: ptrS("Lain")}}))

	phone, email := fhir.ContactPointSystemPhone, fhir.ContactPointSystemEmail
	assert.Equal(t, "0811", satusehat.FindTelepon([]fhir.ContactPoint{{}, {System: &email, Value: ptrS("a@b")}, {System: &phone, Value: ptrS("0811")}}))
	assert.Equal(t, "", satusehat.FindTelepon([]fhir.ContactPoint{{System: &email, Value: ptrS("a@b")}}))

	assert.Equal(t, "35", satusehat.FindNik([]fhir.Identifier{{}, {System: ptrS(satusehat.SystemNIK)}, {System: ptrS(satusehat.SystemNIK), Value: ptrS("35")}}))

	assert.Nil(t, satusehat.ReferenceID(nil))
	assert.Nil(t, satusehat.ReferenceID(&fhir.Reference{Reference: ptrS("")}))
	assert.Nil(t, satusehat.ReferenceID(&fhir.Reference{Reference: ptrS("Patient/")}))
	assert.Equal(t, "abc", *satusehat.ReferenceID(&fhir.Reference{Reference: ptrS("urn:uuid:abc")}))

	assert.Nil(t, satusehat.RefOrganisasi(nil))
	assert.Nil(t, satusehat.RefOrganisasi(&fhir.Reference{Reference: ptrS("Practitioner/1")}))
	assert.Equal(t, "10", *satusehat.RefOrganisasi(&fhir.Reference{Reference: ptrS("Organization/10")}))
}

func TestMapperGagalSaatResourceRusak(t *testing.T) {
	rusak := func(tipe string) types2.BundleEntry { return entry(t, "x", `{"resourceType":"`+tipe+`","status":123}`) }

	_, err := satusehat.PatientToAnak(entry(t, "x", `{"resourceType":"Patient","gender":1}`))
	assert.Error(t, err)
	_, _, err = satusehat.RelatedPersonToOrangtua(entry(t, "x", `{"resourceType":"RelatedPerson","gender":1}`))
	assert.Error(t, err)
	_, _, _, err = satusehat.EncounterToKunjungan(rusak("Encounter"))
	assert.Error(t, err)
	_, _, _, _, err = satusehat.ObservationToObservasi(rusak("Observation"))
	assert.Error(t, err)
	_, _, _, err = satusehat.ConditionToDiagnosa(entry(t, "x", `{"resourceType":"Condition","subject":1}`))
	assert.Error(t, err)
	_, _, _, err = satusehat.AllergyIntoleranceToDiagnosa(entry(t, "x", `{"resourceType":"AllergyIntolerance","patient":1}`))
	assert.Error(t, err)
	_, _, _, err = satusehat.ProcedureToLayanan(rusak("Procedure"))
	assert.Error(t, err)
	_, _, _, err = satusehat.MedicationDispenseToLayanan(rusak("MedicationDispense"))
	assert.Error(t, err)
	_, _, _, err = satusehat.NutritionOrderToLayanan(rusak("NutritionOrder"))
	assert.Error(t, err)
	_, _, _, err = satusehat.ImmunizationToLayanan(rusak("Immunization"))
	assert.Error(t, err)
	_, _, _, err = satusehat.ServiceRequestToRujukan(rusak("ServiceRequest"))
	assert.Error(t, err)
	_, _, _, err = satusehat.EpisodeOfCareToEpisode(rusak("EpisodeOfCare"))
	assert.Error(t, err)
}

func TestDiagnosaTanpaKodeDitolak(t *testing.T) {
	_, _, _, err := satusehat.ConditionToDiagnosa(entry(t, "c", `{"resourceType":"Condition","subject":{"reference":"Patient/P"}}`))
	assert.ErrorContains(t, err, "tanpa kode")
	_, _, _, err = satusehat.AllergyIntoleranceToDiagnosa(entry(t, "a", `{"resourceType":"AllergyIntolerance","patient":{"reference":"Patient/P"}}`))
	assert.ErrorContains(t, err, "tanpa kode")
}

func TestKolomUkurLingkarTubuh(t *testing.T) {
	assert.Equal(t, "lingkar_kepala", satusehat.KolomUkuran(satusehat.SystemLOINC, "9843-4"))
	assert.Equal(t, "lingkar_lengan", satusehat.KolomUkuran(satusehat.SystemSNOMED, "284473002"))
}

func TestWaktuDanTanggalCadangan(t *testing.T) {
	k, _, _, err := satusehat.EncounterToKunjungan(entry(t, "e", `{"resourceType":"Encounter","status":"finished","class":{"code":"AMB"},"period":{"end":"2026-02-03T10:00:00Z"}}`))
	require.NoError(t, err)
	assert.Equal(t, "2026-02-03", k.TanggalPengukuran, "tanpa start, jatuh ke end")

	k, _, _, _ = satusehat.EncounterToKunjungan(entry(t, "e", `{"resourceType":"Encounter","status":"finished","class":{"code":"AMB"},"basedOn":[{"display":"x"},{"reference":"CarePlan/1"},{"reference":"ServiceRequest/sr-1"}]}`))
	assert.Equal(t, "", k.TanggalPengukuran)
	assert.Equal(t, "sr-1", *k.RefRujukan)

	for value, want := range map[string]string{
		`"effectivePeriod":{"start":"2026-01-01"}`:  "2026-01-01",
		`"effectiveInstant":"2026-01-02T00:00:00Z"`: "2026-01-02T00:00:00Z",
		`"issued":"2026-01-03T00:00:00Z"`:           "2026-01-03T00:00:00Z",
		`"valueString":"x"`:                         "<nil>",
	} {
		h, _, _, _, err := satusehat.ObservationToObservasi(entry(t, "o", `{"resourceType":"Observation","status":"final","code":{"text":"x"},`+value+`}`))
		require.NoError(t, err)
		assert.Equal(t, want, text(h.Induk.Tanggal), value)
	}

	l, _, _, err := satusehat.ProcedureToLayanan(entry(t, "p", `{"resourceType":"Procedure","status":"completed","subject":{"reference":"Patient/P"},"performedPeriod":{"start":"2026-01-04"},"category":{"coding":[{"code":"K"}]}}`))
	require.NoError(t, err)
	assert.Equal(t, "2026-01-04", text(l.Tanggal))
	assert.Equal(t, "K", text(l.Kategori))

	l, _, _, err = satusehat.MedicationDispenseToLayanan(entry(t, "m", `{"resourceType":"MedicationDispense","status":"completed","medicationCodeableConcept":{"coding":[{"code":"OB"}]},"whenPrepared":"2026-01-05","quantity":{"value":10,"unit":"tab"}}`))
	require.NoError(t, err)
	assert.Equal(t, "2026-01-05", text(l.Tanggal))
	assert.Equal(t, 10.0, *l.Jumlah)

	for content, code := range map[string]string{
		`"supplement":[{"type":{"coding":[{"code":"SUP"}]}}]`:              "SUP",
		`"enteralFormula":{"baseFormulaType":{"coding":[{"code":"ENT"}]}}`: "ENT",
		`"oralDiet":{"type":[{"coding":[{"code":"ORAL"}]}]}`:               "ORAL",
		`"supplement":[{}]`: "<nil>",
	} {
		l, _, _, err := satusehat.NutritionOrderToLayanan(entry(t, "n", `{"resourceType":"NutritionOrder","status":"active","intent":"order","patient":{"reference":"Patient/P"},"dateTime":"2026-01-06",`+content+`}`))
		require.NoError(t, err, content)
		assert.Equal(t, code, text(l.Kode), content)
	}
}

func TestServiceRequestKategoriDanAlasan(t *testing.T) {
	cases := []struct {
		isi, jenis, alasan, tanggal string
	}{
		{`"category":[{"coding":[{},{"code":"3457005"}]}],"reasonCode":[{"coding":[{"code":"E45","display":"Stunting"}]}],"occurrencePeriod":{"start":"2026-01-01"}`, "rujukan", "Stunting", "2026-01-01"},
		{`"category":[{"coding":[{"code":"3457005"}]},{"coding":[{"code":"SR000007"}]}],"reasonCode":[{"text":"teks alasan"}],"authoredOn":"2026-01-02"`, "rujuk_balik", "teks alasan", "2026-01-02"},
		{`"reasonCode":[{"coding":[{"code":"X"}]}],"reasonReference":[{"reference":"Condition/1"},{"display":"dari ref"}]`, "", "dari ref", "<nil>"},
		{`"priority":"urgent"`, "", "<nil>", "<nil>"},
	}
	for _, k := range cases {
		r, _, _, err := satusehat.ServiceRequestToRujukan(entry(t, "sr", `{"resourceType":"ServiceRequest","status":"active","intent":"order","subject":{"reference":"Patient/P"},`+k.isi+`}`))
		require.NoError(t, err, k.isi)
		assert.Equal(t, k.jenis, r.Jenis, k.isi)
		assert.Equal(t, k.alasan, text(r.Alasan), k.isi)
		assert.Equal(t, k.tanggal, text(r.Tanggal), k.isi)
	}
}

func TestEpisodeTanpaTipeDanPeriode(t *testing.T) {
	e, _, org, err := satusehat.EpisodeOfCareToEpisode(entry(t, "ep", `{"resourceType":"EpisodeOfCare","status":"active","patient":{"reference":"Patient/P"}}`))
	require.NoError(t, err)
	assert.Nil(t, e.Kode)
	assert.Nil(t, e.Mulai)
	assert.Nil(t, org)
}

func TestBobotDanUrutanEntry(t *testing.T) {
	assert.Equal(t, -1, satusehat.BobotEntry(json.RawMessage(`{rusak`)))
	assert.Equal(t, satusehat.BobotDefault, satusehat.BobotResource("Tidakdikenal"))
	biasa := satusehat.BobotEntry(json.RawMessage(`{"resourceType":"Observation"}`))
	turunan := satusehat.BobotEntry(json.RawMessage(`{"resourceType":"Observation","derivedFrom":[{"reference":"Observation/1"}]}`))
	assert.Equal(t, biasa+1, turunan, "observasi turunan diproses setelah induknya")

	list := []types2.BundleEntry{
		entry(t, "1", `{"resourceType":"Observation","derivedFrom":[{"reference":"Observation/2"}]}`),
		entry(t, "2", `{"resourceType":"Observation"}`),
		entry(t, "3", `{"resourceType":"Patient"}`),
	}
	satusehat.UrutkanEntry(list)
	assert.Equal(t, []string{"3", "2", "1"}, []string{*list[0].Base.Id, *list[1].Base.Id, *list[2].Base.Id})
}

func TestResourceRealDipakaiLangsung(t *testing.T) {
	e := entry(t, "e", `{}`)
	start := "2025-01-01T08:00:00Z"
	e.Base.ResourceReal = fhir.Encounter{Period: &fhir.Period{Start: &start}}
	k, _, _, err := satusehat.EncounterToKunjungan(e)
	require.NoError(t, err)
	assert.Equal(t, "2025-01-01", k.TanggalPengukuran)
}
