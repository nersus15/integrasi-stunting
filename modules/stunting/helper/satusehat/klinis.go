package satusehat

import (
	"fmt"
	"strings"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
)

const (
	LayananProcedure          = "procedure"
	LayananMedicationDispense = "medication_dispense"
	LayananNutritionOrder     = "nutrition_order"
	LayananImmunization       = "immunization"
)

// Playbook Rujukan v6.1: arah rujukan dibaca dari ServiceRequest.category
const (
	KategoriRujukanPasien = "3457005"  // SNOMED, Patient referral
	KategoriRujukBalik    = "SR000007" // kemkes clinical-term, Rujuk balik
)

func ConditionToDiagnosa(entry types2.BundleEntry) (*types.Diagnosa, *string, *string, error) {
	r, err := resourceDari[fhir.Condition](entry)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Gagal memetakan resource Condition: %s", err.Error())
	}

	sys, kode, disp := KodeUtama(nilaiCC(r.Code))
	if !utils.IsStrFilled(kode) {
		return nil, nil, nil, fmt.Errorf("Condition tanpa kode diagnosis")
	}

	d := &types.Diagnosa{
		IdSatusehat:        entry.Base.Id,
		Jenis:              entity.DiagnosaDiagnosis,
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
	if !utils.IsStrFilled(kode) {
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
	if !utils.IsStrFilled(onset) && r.OnsetPeriod != nil {
		onset = str(r.OnsetPeriod.Start)
	}

	d := &types.Diagnosa{
		IdSatusehat:        entry.Base.Id,
		Jenis:              entity.DiagnosaAlergi,
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
	if !utils.IsStrFilled(waktu) && r.PerformedPeriod != nil {
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
	if !utils.IsStrFilled(waktu) {
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

// E45 = nutritional stunting; interpretasi TB/U jadi cadangan bila tidak ada Condition
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
	if !utils.IsStrFilled(waktu) && r.OccurrencePeriod != nil {
		waktu = str(r.OccurrencePeriod.Start)
	}
	if !utils.IsStrFilled(waktu) {
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

// SR000007 rujuk balik, 3457005 rujukan pasien, kosong = bukan rujukan antar faskes
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

// reasonCode tetap dibaca, tidak semua pengirim memakai reasonReference
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

// hanya Organization, id Practitioner bukan id faskes
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
