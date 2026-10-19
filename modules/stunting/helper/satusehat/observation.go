package satusehat

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
)

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

// Quantity.Value json.Number, konversinya bisa gagal
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

// Lampiran 3 terminologi Gizi SatuSehat, satu kode bisa dipakai beberapa indeks
var statusGizi = map[string]map[string]string{
	"zscore_bbu": {
		"OI000007":  "Berat Badan Sangat Kurang",
		"248342006": "Berat Badan Kurang",
		"43664005":  "Berat Badan Normal",
		"OI000010":  "Risiko Berat Badan Lebih",
	},
	"zscore_tbu": {
		"OI000011":  "Sangat Pendek",
		"444000005": "Pendek",
		"17489000":  "Normal",
		"83077003":  "Tinggi",
	},
	"zscore_bbtb": {
		"OI000001":  "Gizi Buruk",
		"248325000": "Gizi Kurang",
		"248324001": "Gizi Baik",
		"OI000004":  "Risiko Gizi Lebih",
		"238131007": "Gizi Lebih",
		"414915002": "Obesitas",
	},
}

func StatusGizi(kolom, kode string) (string, bool) {
	s, ok := statusGizi[kolom][kode]
	return s, ok
}

func statusGiziDari(kolom string, kode *string) string {
	if kode == nil {
		return ""
	}
	s, _ := StatusGizi(kolom, *kode)
	return s
}

// untuk jalur masuk yang tidak lewat FHIR
func HasilDariObservasi(induk types.Observasi, komponen []types.Observasi) HasilObservasi {
	kolom := KolomUkuran(induk.System, induk.Kode)
	h := HasilObservasi{
		Induk:    induk,
		Kolom:    kolom,
		Angka:    induk.NilaiAngka,
		CaraUkur: CaraUkur(induk.System, induk.Kode),
		Status:   statusGiziDari(kolom, induk.Interpretasi),
	}
	for _, c := range komponen {
		if c.Tanggal == nil {
			c.Tanggal = induk.Tanggal
		}
		h.Component = append(h.Component, c)
	}
	return h
}

type HasilObservasi struct {
	Induk     types.Observasi
	Component []types.Observasi
	// Kolom dan Angka terisi kalau kodenya dikenal kolomUkuran, mis. berat_badan.
	Kolom    string
	Angka    *float64
	CaraUkur string
	// label interpretasi, hanya untuk Kolom zscore_*
	Status string
}

func ObservationToObservasi(entry types2.BundleEntry) (*HasilObservasi, *fhir.Observation, *string, *string, error) {
	resource, err := resourceDari[fhir.Observation](entry)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("Gagal memetakan resource Observation: %s", err.Error())
	}

	sys, kode, disp := KodeUtama(resource.Code)
	nilai := NilaiDariObservation(resource)
	kolom := KolomUkuran(sys, kode)
	interpretasi := isiStr(Interpretasi(resource.Interpretation))

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
			Interpretasi: interpretasi,
			Tanggal:      waktuObservasi(resource),
		},
		Kolom:    kolom,
		Angka:    nilai.Angka,
		CaraUkur: CaraUkur(sys, kode),
		Status:   statusGiziDari(kolom, interpretasi),
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
	case utils.IsFilled(o.EffectiveDateTime):
		return o.EffectiveDateTime
	case o.EffectivePeriod != nil && utils.IsFilled(o.EffectivePeriod.Start):
		return o.EffectivePeriod.Start
	case utils.IsFilled(o.EffectiveInstant):
		return o.EffectiveInstant
	case utils.IsFilled(o.Issued):
		return o.Issued
	}
	return nil
}
