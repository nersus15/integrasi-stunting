package satusehat

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"
)

const SystemNIK = "https://fhir.kemkes.go.id/id/nik"

func PatientToAnak(entry types2.BundleEntry) (*types.Anak, error) {
	var resource fhir.Patient
	raw := entry.Resource

	if err := json.Unmarshal(raw, &resource); err != nil {
		return nil, fmt.Errorf("Gagal memetakan resource Patient ke Anak: %s", err.Error())
	}

	nik := FindNik(resource.Identifier)
	name := FindName(resource.Name)
	kelamin := ""
	if resource.Gender != nil {
		switch strings.ToLower(resource.Gender.Display()) {
		case "male":
			kelamin = "L"
		case "female":
			kelamin = "P"
		}
	}

	logger.Info("ID Resource: " + helper.ToLogJSON(entry))
	return &types.Anak{
		IdSatusehat:  entry.Base.Id,
		Nama:         name,
		Nik:          &nik,
		TanggalLahir: str(resource.BirthDate),
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

func FindNik(identifiers []fhir.Identifier) string {
	nik := ""

	for _, identifier := range identifiers {
		if identifier.System == nil {
			continue
		}

		if *identifier.System == SystemNIK && identifier.Value != nil {
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
