package satusehat

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
)

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

func isiStr(s string) *string {
	if !utils.IsStrFilled(s) {
		return nil
	}
	return &s
}

func WaktuDariTanggal(s string) *time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

// pakai ResourceReal, entry.Resource masih berisi urn:uuid
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

func str(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func StrKosong(s *string) string { return str(s) }
