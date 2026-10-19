package satusehat

import (
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
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

// kode -> kolom kunjungan; kode lain tetap tersimpan di observasi
var kolomUkuran = map[Ukuran]string{
	{SystemLOINC, "29463-7"}:    "berat_badan",
	{SystemLOINC, "8302-2"}:     "tinggi_badan",
	{SystemLOINC, "8306-3"}:     "tinggi_badan",
	{SystemLOINC, "8308-9"}:     "tinggi_badan",
	{SystemLOINC, "9843-4"}:     "lingkar_kepala",
	{SystemSNOMED, "284473002"}: "lingkar_lengan",

	{SystemSNOMED, "1153593003"}: "zscore_bbu",
	{SystemSNOMED, "1153590000"}: "zscore_tbu",
	{SystemSNOMED, "1153604005"}: "zscore_tbu",
	{SystemSNOMED, "1153598007"}: "zscore_bbtb",
	{SystemSNOMED, "1153600001"}: "zscore_bbtb",
}

var caraUkur = map[Ukuran]string{
	{SystemLOINC, "8306-3"}: "telentang",
	{SystemLOINC, "8308-9"}: "berdiri",
}

func KolomUkuran(system, kode string) string { return kolomUkuran[Ukuran{system, kode}] }
func CaraUkur(system, kode string) string    { return caraUkur[Ukuran{system, kode}] }

func KolomUkurTerisi(k *types.Kunjungan) map[string]any {
	m := map[string]any{}
	for kolom, v := range map[string]*float64{
		"berat_badan": k.BeratBadan, "tinggi_badan": k.TinggiBadan,
		"lingkar_kepala": k.LingkarKepala, "lingkar_lengan": k.LingkarLengan,
	} {
		if v != nil {
			m[kolom] = *v
		}
	}
	if k.CaraUkur != nil {
		m["cara_ukur"] = *k.CaraUkur
	}
	// status ikut z-score pasangannya, nil ditulis NULL
	for _, p := range []struct {
		kolomZ, kolomS string
		z              *float64
		s              *string
	}{
		{"zscore_bbu", "status_bbu", k.ZscoreBbu, k.StatusBbu},
		{"zscore_tbu", "status_tbu", k.ZscoreTbu, k.StatusTbu},
		{"zscore_bbtb", "status_bbtb", k.ZscoreBbtb, k.StatusBbtb},
	} {
		if p.z != nil {
			m[p.kolomZ], m[p.kolomS] = *p.z, p.s
		}
	}
	return m
}

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
