package utils

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"
)

var BobotResourceBundle = map[string]int{
	// prerequisite / master
	"organization":      10,
	"location":          11,
	"healthcareservice": 12,
	"practitioner":      13,
	"practitionerrole":  14,
	"patient":           15,
	"relatedperson":     16,
	"coverage":          17,

	// konteks kunjungan
	"episodeofcare":       20,
	"slot":                21,
	"appointment":         22,
	"appointmentresponse": 23,
	"encounter":           24,

	// data mentah, hanya butuh patient + encounter
	"allergyintolerance":  30,
	"familymemberhistory": 31,
	"medicationstatement": 32,
	"immunization":        33,
	"specimen":            34,
	"observation":         35,
	"imagingstudy":        36,

	// turunan dari data mentah
	"diagnosticreport":      40,
	"questionnaireresponse": 41,

	// kesimpulan klinis
	"condition":          50,
	"clinicalimpression": 51,
	"riskassessment":     52,
	"goal":               53,

	// intervensi
	"procedure":                60,
	"medication":               61,
	"medicationrequest":        62,
	"medicationadministration": 63,
	"medicationdispense":       64,
	"nutritionorder":           65,
	"servicerequest":           66,
	"careplan":                 67,
	"task":                     68,

	// dokumen
	"composition":       70,
	"documentreference": 71,

	// pembiayaan
	"account":                     80,
	"chargeitem":                  81,
	"claim":                       82,
	"claimresponse":               83,
	"invoice":                     84,
	"paymentnotice":               85,
	"paymentreconciliation":       86,
	"coverageeligibilityrequest":  87,
	"coverageeligibilityresponse": 88,
}

func BobotEntry(raw json.RawMessage) int {
	var res struct {
		ResourceType string          `json:"resourceType"`
		DerivedFrom  json.RawMessage `json:"derivedFrom"`
	}

	if err := json.Unmarshal(raw, &res); err != nil {
		return -1
	}

	bobot := BobotResource(res.ResourceType)

	if strings.EqualFold(res.ResourceType, "Observation") {
		var d []json.RawMessage
		if json.Unmarshal(res.DerivedFrom, &d) == nil && len(d) > 0 {
			bobot++
		}
	}
	return bobot
}

// bobot berpasangan dengan entry-nya supaya cukup dihitung sekali
type urutBobot struct {
	entry []types2.BundleEntry
	bobot []int
}

func (u *urutBobot) Len() int           { return len(u.entry) }
func (u *urutBobot) Less(i, j int) bool { return u.bobot[i] < u.bobot[j] }
func (u *urutBobot) Swap(i, j int) {
	u.entry[i], u.entry[j] = u.entry[j], u.entry[i]
	u.bobot[i], u.bobot[j] = u.bobot[j], u.bobot[i]
}

func UrutkanEntry(entries []types2.BundleEntry) {
	bobot := make([]int, len(entries))
	for i := range entries {
		bobot[i] = BobotEntry(entries[i].Resource)
	}
	sort.Stable(&urutBobot{entry: entries, bobot: bobot})
}

const BobotDefault = 999

func BobotResource(resourceType string) int {
	if b, ok := BobotResourceBundle[strings.ToLower(resourceType)]; ok {
		return b
	}
	return BobotDefault
}

const (
	LayananProcedure          = "procedure"
	LayananMedicationDispense = "medication_dispense"
	LayananNutritionOrder     = "nutrition_order"
	LayananImmunization       = "immunization"
)

const (
	DiagnosaDiagnosis = "diagnosis"
	DiagnosaAlergi    = "alergi"
)

// Playbook Rujukan v6.1: rujukan pasien dan rujuk balik ditandai di
// ServiceRequest.category, bukan disimpulkan dari jenis faskes.
const (
	KategoriRujukanPasien = "3457005"  // SNOMED, Patient referral
	KategoriRujukBalik    = "SR000007" // kemkes clinical-term, Rujuk balik
)

const (
	SystemLOINC  = "http://loinc.org"
	SystemSNOMED = "http://snomed.info/sct"
	SystemICD10  = "http://hl7.org/fhir/sid/icd-10"
	SystemKemkes = "http://terminology.kemkes.go.id/CodeSystem/clinical-term"
)

type Ukuran struct {
	System string
	Kode   string
}

// kode -> kolom kunjungan, system bebas. Yang tidak terdaftar tetap tersimpan di observasi.
var kolomUkuran = map[Ukuran]string{
	{SystemLOINC, "29463-7"}: "berat_badan",
	{SystemLOINC, "8302-2"}:  "tinggi_badan",
	{SystemLOINC, "8306-3"}:  "tinggi_badan",
	{SystemLOINC, "8308-9"}:  "tinggi_badan",
	{SystemLOINC, "9843-4"}:  "lingkar_kepala",
}

var caraUkur = map[Ukuran]string{
	{SystemLOINC, "8306-3"}: "telentang",
	{SystemLOINC, "8308-9"}: "berdiri",
}

func KolomUkuran(system, kode string) string { return kolomUkuran[Ukuran{system, kode}] }
func CaraUkur(system, kode string) string    { return caraUkur[Ukuran{system, kode}] }

// yang dikenal kolomUkuran didahulukan, supaya tidak bergantung urutan coding
func KodeUtama(cc fhir.CodeableConcept) (system, kode, display string) {
	for _, c := range cc.Coding {
		if c.System == nil || c.Code == nil {
			continue
		}
		if KolomUkuran(*c.System, *c.Code) != "" {
			return *c.System, *c.Code, str(c.Display)
		}
	}
	for _, c := range cc.Coding {
		if c.Code == nil {
			continue
		}
		return str(c.System), *c.Code, str(c.Display)
	}
	return "", "", str(cc.Text)
}

type NilaiObservasi struct {
	Angka       *float64
	Satuan      *string
	Teks        *string
	Kode        *string
	KodeSystem  *string
	KodeDisplay *string
}

// value[x] Observation dan component bentuknya identik
type nilaiMentah struct {
	Quantity        *fhir.Quantity
	CodeableConcept *fhir.CodeableConcept
	Str             *string
	Boolean         *bool
	Integer         *int
	Range           *fhir.Range
	Ratio           *fhir.Ratio
	Time            *string
	DateTime        *string
	Period          *fhir.Period
}

func NilaiDariObservation(o fhir.Observation) NilaiObservasi {
	return ambilNilai(nilaiMentah{o.ValueQuantity, o.ValueCodeableConcept, o.ValueString,
		o.ValueBoolean, o.ValueInteger, o.ValueRange, o.ValueRatio, o.ValueTime,
		o.ValueDateTime, o.ValuePeriod})
}

func NilaiDariComponent(c fhir.ObservationComponent) NilaiObservasi {
	return ambilNilai(nilaiMentah{c.ValueQuantity, c.ValueCodeableConcept, c.ValueString,
		c.ValueBoolean, c.ValueInteger, c.ValueRange, c.ValueRatio, c.ValueTime,
		c.ValueDateTime, c.ValuePeriod})
}

func ambilNilai(m nilaiMentah) NilaiObservasi {
	var n NilaiObservasi

	switch {
	case m.Quantity != nil:
		n.Angka = angkaQuantity(m.Quantity.Value)
		if s := satuanQuantity(m.Quantity); s != "" {
			n.Satuan = &s
		}
	case m.CodeableConcept != nil:
		sys, kode, disp := KodeUtama(*m.CodeableConcept)
		if kode != "" {
			n.Kode = &kode
		}
		if sys != "" {
			n.KodeSystem = &sys
		}
		if disp != "" {
			n.KodeDisplay = &disp
		}
		if kode == "" && disp != "" {
			n.Teks = &disp
		}
	case m.Str != nil:
		n.Teks = m.Str
	case m.Boolean != nil:
		v := 0.0
		if *m.Boolean {
			v = 1
		}
		n.Angka = &v
	case m.Integer != nil:
		v := float64(*m.Integer)
		n.Angka = &v
	case m.Time != nil:
		n.Teks = m.Time
	case m.DateTime != nil:
		n.Teks = m.DateTime
	case m.Range != nil:
		t := rentangTeks(m.Range)
		n.Teks = &t
	case m.Ratio != nil:
		t := rasioTeks(m.Ratio)
		n.Teks = &t
	case m.Period != nil:
		t := str(m.Period.Start) + " s/d " + str(m.Period.End)
		n.Teks = &t
	}

	return n
}

// Quantity.Value bertipe json.Number, jadi harus dikonversi dan bisa gagal.
func angkaQuantity(v *json.Number) *float64 {
	if v == nil {
		return nil
	}
	f, err := v.Float64()
	if err != nil {
		return nil
	}
	return &f
}

func satuanQuantity(q *fhir.Quantity) string {
	if q.Unit != nil && *q.Unit != "" {
		return *q.Unit
	}
	return str(q.Code)
}

func rentangTeks(r *fhir.Range) string {
	t := ""
	if r.Low != nil && r.Low.Value != nil {
		t += r.Low.Value.String()
	}
	t += " - "
	if r.High != nil && r.High.Value != nil {
		t += r.High.Value.String()
	}
	return strings.TrimSpace(t)
}

func rasioTeks(r *fhir.Ratio) string {
	t := ""
	if r.Numerator != nil && r.Numerator.Value != nil {
		t = r.Numerator.Value.String()
	}
	t += ":"
	if r.Denominator != nil && r.Denominator.Value != nil {
		t += r.Denominator.Value.String()
	}
	return t
}

// kode interpretasi pertama, mis. "L"/"N"/"H" atau SNOMED status gizi
func Interpretasi(list []fhir.CodeableConcept) string {
	for _, cc := range list {
		if _, kode, _ := KodeUtama(cc); kode != "" {
			return kode
		}
	}
	return ""
}

type StatusIndeks struct {
	Kolom string
	Label string
}

// hanya kode yang menunjuk satu indeks. Yang ambigu (mis. 248342006 Underweight,
// dipakai BB/U dan IMT dewasa) ditentukan dari Observation.code induknya.
var statusIndeks = map[Ukuran]StatusIndeks{
	{SystemKemkes, "OI000007"}: {"status_bbu_whoantro", "Berat Badan Sangat Kurang"},
	{SystemKemkes, "OI000010"}: {"status_bbu_whoantro", "Risiko Berat Badan Lebih"},

	{SystemKemkes, "OI000011"}:  {"status_tbu_whoantro", "Sangat Pendek"},
	{SystemSNOMED, "444000005"}: {"status_tbu_whoantro", "Pendek"},
	{SystemSNOMED, "17489000"}:  {"status_tbu_whoantro", "Normal"},
	{SystemSNOMED, "83077003"}:  {"status_tbu_whoantro", "Tinggi"},

	{SystemKemkes, "OI000004"}: {"status_bbtb_whoantro", "Risiko Gizi Lebih"},
}

func StatusGizi(system, kode string) (StatusIndeks, bool) {
	s, ok := statusIndeks[Ukuran{system, kode}]
	return s, ok
}

func PatientToAnak(entry types2.BundleEntry) (*types.Anak, error) {
	var resource fhir.Patient
	raw := entry.Resource

	if err := json.Unmarshal(raw, &resource); err != nil {
		return nil, fmt.Errorf("Gagal memetakan resource Patient ke Anak: %s", err.Error())
	}

	nik := FindNik(resource.Identifier)
	name := FindName(resource.Name)
	gender := strings.ToLower(resource.Gender.Display())
	kelamin := ""

	switch gender {
	case "male":
		kelamin = "L"
	case "female":
		kelamin = "P"
	}

	logger.Info("ID Resource: " + helper.ToLogJSON(entry))
	return &types.Anak{
		IdSatusehat:  entry.Base.Id,
		Nama:         name,
		Nik:          &nik,
		TanggalLahir: *resource.BirthDate,
		JenisKelamin: kelamin,
	}, nil

}

// nilai kedua adalah IHS id anak dari RelatedPerson.patient
func RelatedPersonToOrangtua(entry types2.BundleEntry) (*types.Orangtua, *string, error) {
	var resource fhir.RelatedPerson
	raw := entry.Resource

	if err := json.Unmarshal(raw, &resource); err != nil {
		return nil, nil, fmt.Errorf("Gagal memetakan resource related person ke orangtua: %s", err.Error())
	}

	nik := FindNik(resource.Identifier)
	telepon := FindTelepon(resource.Telecom)
	name := FindName(resource.Name)
	alamat := FindAlamatText(resource.Address)

	namaAyah := ""
	namaIbu := ""

	switch PeranOrangtua(resource.Relationship, resource.Gender) {
	case "ayah":
		namaAyah = name
	case "ibu":
		namaIbu = name
	}

	return &types.Orangtua{
		Nik:         nik,
		IdSatusehat: entry.Base.Id,
		Telepon:     telepon,
		NamaAyah:    namaAyah,
		NamaIbu:     namaIbu,
		Alamat:      alamat,
	}, ReferenceID(&resource.Patient), nil
}

// baca relationship dulu, gender hanya cadangan, "" kalau keduanya diam
func PeranOrangtua(relationship []fhir.CodeableConcept, gender *fhir.AdministrativeGender) string {
	for _, rel := range relationship {
		for _, c := range rel.Coding {
			if c.Code == nil {
				continue
			}
			switch strings.ToUpper(*c.Code) {
			case "FTH", "NFTH", "STPFTH", "FTHFOST", "ADOPTF":
				return "ayah"
			case "MTH", "NMTH", "STPMTH", "MTHFOST", "ADOPTM":
				return "ibu"
			}
		}
	}

	if gender == nil {
		return ""
	}

	switch strings.ToLower(gender.Display()) {
	case "male":
		return "ayah"
	case "female":
		return "ibu"
	}

	return ""
}

// "Patient/P0247..." menjadi "P0247..."
func ReferenceID(ref *fhir.Reference) *string {
	if ref == nil || ref.Reference == nil || *ref.Reference == "" {
		return nil
	}

	r := strings.TrimPrefix(*ref.Reference, "urn:uuid:")
	bagian := strings.Split(r, "/")
	id := bagian[len(bagian)-1]
	if id == "" {
		return nil
	}

	return &id
}

func FindNik(identifiers []fhir.Identifier) string {
	nik := ""

	for _, identifier := range identifiers {
		if identifier.System == nil {
			continue
		}

		if *identifier.System == "https://fhir.kemkes.go.id/id/nik" && identifier.Value != nil {
			nik = *identifier.Value
			break
		}
	}
	return nik
}

func FindTelepon(com []fhir.ContactPoint) string {
	for _, c := range com {
		if c.System == nil || c.Value == nil {
			continue
		}
		if strings.ToLower(c.System.Display()) == "phone" {
			return *c.Value
		}
	}
	return ""
}

// prefer official, jatuh ke nama pertama yang punya text
func FindName(names []fhir.HumanName) string {
	cadangan := ""

	for _, n := range names {
		if n.Text == nil || *n.Text == "" {
			continue
		}
		if n.Use != nil && strings.ToLower(n.Use.Display()) == "official" {
			return *n.Text
		}
		if cadangan == "" {
			cadangan = *n.Text
		}
	}

	return cadangan
}

// prefer home, jatuh ke alamat pertama yang punya line
func FindAlamatText(addresses []fhir.Address) string {
	cadangan := ""

	for _, a := range addresses {
		if len(a.Line) == 0 || a.Line[0] == "" {
			continue
		}
		if a.Use != nil && strings.ToLower(a.Use.Display()) == "home" {
			return a.Line[0]
		}
		if cadangan == "" {
			cadangan = a.Line[0]
		}
	}

	return cadangan
}

func EncounterToKunjungan(entry types2.BundleEntry) (*types.Kunjungan, *string, *string, error) {
	resource, err := resourceDari[fhir.Encounter](entry)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Gagal memetakan resource Encounter ke Kunjungan: %s", err.Error())
	}

	k := &types.Kunjungan{
		IdSatusehat:       entry.Base.Id,
		TanggalPengukuran: tanggalEncounter(resource.Period),
	}

	// rawat inap bisa melintasi beberapa hari; tanpa ini rujukan yang terbit
	// saat pulang terlihat di luar rentang kunjungannya
	if resource.Period != nil {
		k.TanggalSelesai = isiStr(TanggalSaja(str(resource.Period.End)))
	}

	// episode-nya bisa tiba di bundle lain, jadi ref-nya disimpan dulu
	if len(resource.EpisodeOfCare) > 0 {
		k.RefEpisode = ReferenceID(&resource.EpisodeOfCare[0])
	}

	// basedOn diisi faskes tujuan: kunjungan ini memenuhi rujukan tersebut
	k.RefRujukan = refServiceRequest(resource.BasedOn)

	return k, ReferenceID(resource.Subject), ReferenceID(resource.ServiceProvider), nil
}

func tanggalEncounter(p *fhir.Period) string {
	if p == nil {
		return ""
	}
	if IsFilled(p.Start) {
		return TanggalSaja(*p.Start)
	}
	return TanggalSaja(str(p.End))
}

// TanggalSaja memotong bagian waktu dari nilai FHIR dateTime.
func TanggalSaja(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

// untuk jalur masuk yang tidak lewat FHIR
func HasilDariObservasi(induk types.Observasi, komponen []types.Observasi) HasilObservasi {
	h := HasilObservasi{
		Induk:    induk,
		Kolom:    KolomUkuran(induk.System, induk.Kode),
		Angka:    induk.NilaiAngka,
		CaraUkur: CaraUkur(induk.System, induk.Kode),
	}
	h.Component = append(h.Component, komponen...)
	return h
}

type HasilObservasi struct {
	Induk     types.Observasi
	Component []types.Observasi
	// Kolom dan Angka terisi kalau kodenya dikenal kolomUkuran, mis. berat_badan.
	Kolom    string
	Angka    *float64
	CaraUkur string
}

func ObservationToObservasi(entry types2.BundleEntry) (*HasilObservasi, *fhir.Observation, *string, *string, error) {
	resource, err := resourceDari[fhir.Observation](entry)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("Gagal memetakan resource Observation: %s", err.Error())
	}

	sys, kode, disp := KodeUtama(resource.Code)
	nilai := NilaiDariObservation(resource)

	h := &HasilObservasi{
		Induk: types.Observasi{
			IdSatusehat:  entry.Base.Id,
			System:       sys,
			Kode:         kode,
			Display:      isiStr(disp),
			Kategori:     isiStr(kategoriPertama(resource.Category)),
			NilaiAngka:   nilai.Angka,
			Satuan:       nilai.Satuan,
			NilaiTeks:    nilai.Teks,
			NilaiKode:    nilai.Kode,
			NilaiSystem:  nilai.KodeSystem,
			NilaiDisplay: nilai.KodeDisplay,
			Interpretasi: isiStr(Interpretasi(resource.Interpretation)),
			Tanggal:      waktuObservasi(resource),
		},
		Kolom:    KolomUkuran(sys, kode),
		Angka:    nilai.Angka,
		CaraUkur: CaraUkur(sys, kode),
	}

	for _, c := range resource.Component {
		csys, ckode, cdisp := KodeUtama(c.Code)
		cn := NilaiDariComponent(c)
		h.Component = append(h.Component, types.Observasi{
			System:       csys,
			Kode:         ckode,
			Display:      isiStr(cdisp),
			NilaiAngka:   cn.Angka,
			Satuan:       cn.Satuan,
			NilaiTeks:    cn.Teks,
			NilaiKode:    cn.Kode,
			NilaiSystem:  cn.KodeSystem,
			NilaiDisplay: cn.KodeDisplay,
			Interpretasi: isiStr(Interpretasi(c.Interpretation)),
			Tanggal:      h.Induk.Tanggal,
		})
	}

	return h, &resource, ReferenceID(resource.Subject), ReferenceID(resource.Encounter), nil
}

func kategoriPertama(list []fhir.CodeableConcept) string {
	for _, cc := range list {
		if _, kode, _ := KodeUtama(cc); kode != "" {
			return kode
		}
	}
	return ""
}

func waktuObservasi(o fhir.Observation) *string {
	switch {
	case IsFilled(o.EffectiveDateTime):
		return o.EffectiveDateTime
	case o.EffectivePeriod != nil && IsFilled(o.EffectivePeriod.Start):
		return o.EffectivePeriod.Start
	case IsFilled(o.EffectiveInstant):
		return o.EffectiveInstant
	case IsFilled(o.Issued):
		return o.Issued
	}
	return nil
}

func isiStr(s string) *string {
	if !IsStrFilled(s) {
		return nil
	}
	return &s
}

func ConditionToDiagnosa(entry types2.BundleEntry) (*types.Diagnosa, *string, *string, error) {
	r, err := resourceDari[fhir.Condition](entry)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Gagal memetakan resource Condition: %s", err.Error())
	}

	sys, kode, disp := KodeUtama(nilaiCC(r.Code))
	if !IsStrFilled(kode) {
		return nil, nil, nil, fmt.Errorf("Condition tanpa kode diagnosis")
	}

	d := &types.Diagnosa{
		IdSatusehat:        entry.Base.Id,
		Jenis:              DiagnosaDiagnosis,
		System:             sys,
		Kode:               kode,
		Display:            isiStr(disp),
		Kategori:           isiStr(kategoriPertama(r.Category)),
		ClinicalStatus:     isiStr(kodeCC(r.ClinicalStatus)),
		VerificationStatus: isiStr(kodeCC(r.VerificationStatus)),
		Onset:              isiStr(TanggalSaja(str(r.OnsetDateTime))),
		TanggalCatat:       isiStr(TanggalSaja(str(r.RecordedDate))),
	}

	return d, ReferenceID(&r.Subject), ReferenceID(r.Encounter), nil
}

// AllergyIntolerance masuk tabel diagnosa dengan jenis alergi
func AllergyIntoleranceToDiagnosa(entry types2.BundleEntry) (*types.Diagnosa, *string, *string, error) {
	r, err := resourceDari[fhir.AllergyIntolerance](entry)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Gagal memetakan resource AllergyIntolerance: %s", err.Error())
	}

	sys, kode, disp := KodeUtama(nilaiCC(r.Code))
	if !IsStrFilled(kode) {
		return nil, nil, nil, fmt.Errorf("AllergyIntolerance tanpa kode alergen")
	}

	var kategori string
	if len(r.Category) > 0 {
		kategori = r.Category[0].Code()
	}

	var kritis string
	if r.Criticality != nil {
		kritis = r.Criticality.Code()
	}

	onset := str(r.OnsetDateTime)
	if !IsStrFilled(onset) && r.OnsetPeriod != nil {
		onset = str(r.OnsetPeriod.Start)
	}

	d := &types.Diagnosa{
		IdSatusehat:        entry.Base.Id,
		Jenis:              DiagnosaAlergi,
		System:             sys,
		Kode:               kode,
		Display:            isiStr(disp),
		Kategori:           isiStr(kategori),
		Kritikalitas:       isiStr(kritis),
		ClinicalStatus:     isiStr(kodeCC(r.ClinicalStatus)),
		VerificationStatus: isiStr(kodeCC(r.VerificationStatus)),
		Onset:              isiStr(TanggalSaja(onset)),
		TanggalCatat:       isiStr(TanggalSaja(str(r.RecordedDate))),
	}

	return d, ReferenceID(&r.Patient), ReferenceID(r.Encounter), nil
}

func ProcedureToLayanan(entry types2.BundleEntry) (*types.Layanan, *string, *string, error) {
	r, err := resourceDari[fhir.Procedure](entry)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Gagal memetakan resource Procedure: %s", err.Error())
	}

	waktu := str(r.PerformedDateTime)
	if !IsStrFilled(waktu) && r.PerformedPeriod != nil {
		waktu = str(r.PerformedPeriod.Start)
	}

	l := layananDasar(entry, LayananProcedure, nilaiCC(r.Code), r.Status.Code(), waktu)
	l.Kategori = isiStr(kodeCC(r.Category))

	return l, ReferenceID(&r.Subject), ReferenceID(r.Encounter), nil
}

func MedicationDispenseToLayanan(entry types2.BundleEntry) (*types.Layanan, *string, *string, error) {
	r, err := resourceDari[fhir.MedicationDispense](entry)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Gagal memetakan resource MedicationDispense: %s", err.Error())
	}

	waktu := str(r.WhenHandedOver)
	if !IsStrFilled(waktu) {
		waktu = str(r.WhenPrepared)
	}

	l := layananDasar(entry, LayananMedicationDispense, r.MedicationCodeableConcept, r.Status, waktu)
	l.Kategori = isiStr(kodeCC(r.Category))
	if r.Quantity != nil {
		l.Jumlah = angkaQuantity(r.Quantity.Value)
		l.Satuan = isiStr(satuanQuantity(r.Quantity))
	}

	return l, ReferenceID(r.Subject), ReferenceID(r.Context), nil
}

func NutritionOrderToLayanan(entry types2.BundleEntry) (*types.Layanan, *string, *string, error) {
	r, err := resourceDari[fhir.NutritionOrder](entry)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Gagal memetakan resource NutritionOrder: %s", err.Error())
	}

	// tanpa code tunggal; jenis diet ada di oralDiet, supplement, atau enteralFormula
	var cc fhir.CodeableConcept
	switch {
	case r.OralDiet != nil && len(r.OralDiet.Type) > 0:
		cc = r.OralDiet.Type[0]
	case len(r.Supplement) > 0 && r.Supplement[0].Type != nil:
		cc = *r.Supplement[0].Type
	case r.EnteralFormula != nil && r.EnteralFormula.BaseFormulaType != nil:
		cc = *r.EnteralFormula.BaseFormulaType
	}

	l := layananDasar(entry, LayananNutritionOrder, cc, r.Status.Code(), r.DateTime)

	return l, ReferenceID(&r.Patient), ReferenceID(r.Encounter), nil
}

func ImmunizationToLayanan(entry types2.BundleEntry) (*types.Layanan, *string, *string, error) {
	r, err := resourceDari[fhir.Immunization](entry)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Gagal memetakan resource Immunization: %s", err.Error())
	}

	l := layananDasar(entry, LayananImmunization, r.VaccineCode, r.Status.Code(), r.OccurrenceDateTime)
	if r.DoseQuantity != nil {
		l.Jumlah = angkaQuantity(r.DoseQuantity.Value)
		l.Satuan = isiStr(satuanQuantity(r.DoseQuantity))
	}

	return l, ReferenceID(&r.Patient), ReferenceID(r.Encounter), nil
}

func layananDasar(entry types2.BundleEntry, jenis string, cc fhir.CodeableConcept, status, waktu string) *types.Layanan {
	sys, kode, disp := KodeUtama(cc)

	return &types.Layanan{
		IdSatusehat: entry.Base.Id,
		Jenis:       jenis,
		System:      isiStr(sys),
		Kode:        isiStr(kode),
		Display:     isiStr(disp),
		Status:      isiStr(status),
		Tanggal:     isiStr(waktu),
	}
}

func nilaiCC(cc *fhir.CodeableConcept) fhir.CodeableConcept {
	if cc == nil {
		return fhir.CodeableConcept{}
	}
	return *cc
}

func kodeCC(cc *fhir.CodeableConcept) string {
	if cc == nil {
		return ""
	}
	_, kode, _ := KodeUtama(*cc)
	return kode
}

func WaktuDariTanggal(s string) *time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

// pakai ResourceReal hasil SusunUlangBundle; entry.Resource masih versi sebelum urn:uuid ditulis ulang
func resourceDari[T any](entry types2.BundleEntry) (T, error) {
	var kosong T

	if entry.Base != nil && entry.Base.ResourceReal != nil {
		if r, ok := entry.Base.ResourceReal.(T); ok {
			return r, nil
		}
	}

	if err := json.Unmarshal(entry.Resource, &kosong); err != nil {
		return kosong, err
	}
	return kosong, nil
}

// E45 mencantumkan "Nutritional stunting" sebagai inclusion term. Kode
// interpretasi TB/U ikut sebagai cadangan kalau Condition tidak ada.
var kodeStunting = map[Ukuran]bool{
	{SystemICD10, "E45"}:        true,
	{SystemICD10, "R62.5"}:      true,
	{SystemICD10, "R62.51"}:     true,
	{SystemICD10, "R62.52"}:     true,
	{SystemICD10, "R62.8"}:      true,
	{SystemKemkes, "OI000011"}:  true,
	{SystemSNOMED, "444000005"}: true,
}

// untuk nilai yang tersimpan tanpa system, mis. observasi.interpretasi
var kodeStuntingTanpaSystem = map[string]bool{
	"E45": true, "R62.5": true, "R62.51": true, "R62.52": true, "R62.8": true,
	"OI000011": true, "444000005": true,
}

func TandaStunting(system, kode string) bool {
	if kodeStunting[Ukuran{system, kode}] {
		return true
	}
	return kodeStuntingTanpaSystem[strings.ToUpper(strings.TrimSpace(kode))]
}

// arah rujukan ditentukan pemanggil dari jenis faskes, bukan di sini
func ServiceRequestToRujukan(entry types2.BundleEntry) (*types.Rujukan, *string, *string, error) {
	r, err := resourceDari[fhir.ServiceRequest](entry)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Gagal memetakan resource ServiceRequest: %s", err.Error())
	}

	sys, kode, disp := KodeUtama(nilaiCC(r.Code))

	waktu := str(r.OccurrenceDateTime)
	if !IsStrFilled(waktu) && r.OccurrencePeriod != nil {
		waktu = str(r.OccurrencePeriod.Start)
	}
	if !IsStrFilled(waktu) {
		waktu = str(r.AuthoredOn)
	}

	rj := &types.Rujukan{
		IdSatusehat:     entry.Base.Id,
		Jenis:           jenisDariKategori(r.Category),
		System:          isiStr(sys),
		Kode:            isiStr(kode),
		Display:         isiStr(disp),
		Status:          isiStr(r.Status.Code()),
		Alasan:          isiStr(alasanRujukan(r.ReasonCode, r.ReasonReference)),
		Tanggal:         isiStr(waktu),
		RefFaskesAsal:   RefOrganisasi(r.Requester),
		RefFaskesTujuan: organisasiPertama(r.Performer),
	}
	if r.Priority != nil {
		rj.Prioritas = isiStr(r.Priority.Code())
	}

	return rj, ReferenceID(&r.Subject), ReferenceID(r.Encounter), nil
}

// SR000007 menandai rujuk balik, 3457005 menandai rujukan pasien. Kosong
// berarti bukan rujukan antar faskes, arahnya ditentukan pemanggil.
func jenisDariKategori(list []fhir.CodeableConcept) string {
	jenis := ""
	for _, cc := range list {
		for _, c := range cc.Coding {
			if c.Code == nil {
				continue
			}
			switch *c.Code {
			case KategoriRujukBalik:
				return entity.RujukBalik
			case KategoriRujukanPasien:
				jenis = entity.RujukanKeluar
			}
		}
	}
	return jenis
}

// Playbook memakai reasonReference ke Condition; reasonCode tetap dibaca
// karena sebagian pengirim memakainya.
func alasanRujukan(kode []fhir.CodeableConcept, ref []fhir.Reference) string {
	if a := alasanPertama(kode); a != "" {
		return a
	}
	for i := range ref {
		if d := str(ref[i].Display); d != "" {
			return d
		}
	}
	return ""
}

func alasanPertama(list []fhir.CodeableConcept) string {
	for _, cc := range list {
		if _, _, disp := KodeUtama(cc); disp != "" {
			return disp
		}
		if cc.Text != nil {
			return *cc.Text
		}
	}
	return ""
}

// basedOn boleh berisi beberapa referensi; hanya ServiceRequest yang relevan
func refServiceRequest(list []fhir.Reference) *string {
	for i := range list {
		if list[i].Reference == nil {
			continue
		}
		if strings.HasPrefix(strings.TrimPrefix(*list[i].Reference, "urn:uuid:"), "ServiceRequest/") {
			return ReferenceID(&list[i])
		}
	}
	return nil
}

// hanya Organization: requester/performer bisa berisi Practitioner, dan id praktisi bukan id faskes
func RefOrganisasi(ref *fhir.Reference) *string {
	if ref == nil || ref.Reference == nil {
		return nil
	}
	if !strings.HasPrefix(strings.TrimPrefix(*ref.Reference, "urn:uuid:"), "Organization/") {
		return nil
	}
	return ReferenceID(ref)
}

func organisasiPertama(list []fhir.Reference) *string {
	for i := range list {
		if id := RefOrganisasi(&list[i]); id != nil {
			return id
		}
	}
	return nil
}

func EpisodeOfCareToEpisode(entry types2.BundleEntry) (*types.Episode, *string, *string, error) {
	r, err := resourceDari[fhir.EpisodeOfCare](entry)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Gagal memetakan resource EpisodeOfCare: %s", err.Error())
	}

	var sys, kode, disp string
	if len(r.Type) > 0 {
		sys, kode, disp = KodeUtama(r.Type[0])
	}

	e := &types.Episode{
		IdSatusehat: entry.Base.Id,
		System:      isiStr(sys),
		Kode:        isiStr(kode),
		Display:     isiStr(disp),
		Status:      isiStr(r.Status.Code()),
	}
	if r.Period != nil {
		e.Mulai = isiStr(TanggalSaja(str(r.Period.Start)))
		e.Selesai = isiStr(TanggalSaja(str(r.Period.End)))
	}

	return e, ReferenceID(&r.Patient), ReferenceID(r.ManagingOrganization), nil
}

func str(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func StrKosong(s *string) string { return str(s) }

func FindAnakKe(obs *fhir.Observation) *int {
	system := "http://snomed.info/sct"
	code := "365475008"
	var anak_ke *int
	for _, coding := range obs.Code.Coding {
		if coding.System == nil || coding.Code == nil {
			continue
		}
		if str(coding.System) == system && str(coding.Code) == code {
			anak_ke = obs.ValueInteger
			break
		}
	}
	return anak_ke
}
