package validasi

import (
	"fmt"
	"slices"
	"strings"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
)

// value set FHIR R4 resource asalnya
var (
	kategoriDiagnosa = map[string][]string{
		entity.DiagnosaDiagnosis: {"problem-list-item", "encounter-diagnosis"},
		entity.DiagnosaAlergi:    {"food", "medication", "environment", "biologic"},
	}
	clinicalStatus = map[string][]string{
		entity.DiagnosaDiagnosis: {"active", "recurrence", "relapse", "inactive", "remission", "resolved"},
		entity.DiagnosaAlergi:    {"active", "inactive", "resolved"},
	}
	verificationStatus = map[string][]string{
		entity.DiagnosaDiagnosis: {"unconfirmed", "provisional", "differential", "confirmed", "refuted", "entered-in-error"},
		entity.DiagnosaAlergi:    {"unconfirmed", "confirmed", "refuted", "entered-in-error"},
	}
	kritikalitas     = []string{"low", "high", "unable-to-assess"}
	statusPermintaan = []string{"draft", "active", "on-hold", "revoked", "completed", "entered-in-error", "unknown"}
	statusLayanan    = map[string][]string{
		entity.LayananProcedure:          {"preparation", "in-progress", "not-done", "on-hold", "stopped", "completed", "entered-in-error", "unknown"},
		entity.LayananMedicationDispense: {"preparation", "in-progress", "cancelled", "on-hold", "completed", "entered-in-error", "stopped", "declined", "unknown"},
		entity.LayananNutritionOrder:     statusPermintaan,
		entity.LayananServiceRequest:     statusPermintaan,
		entity.LayananImmunization:       {"completed", "entered-in-error", "not-done"},
	}
	prioritasRujukan = []string{"routine", "urgent", "asap", "stat"}
	statusEpisode    = []string{"planned", "waitlist", "active", "onhold", "finished", "cancelled", "entered-in-error"}
)

func cekEnum(field string, nilai *string, sah []string) error {
	if !utils.IsFilled(nilai) || slices.Contains(sah, *nilai) {
		return nil
	}
	return fmt.Errorf("%s %q tidak dikenali, pilih salah satu: %s", field, *nilai, strings.Join(sah, ", "))
}

// jenis kosong berarti diagnosis
func ValidateEnumDiagnosa(jenis string, kategori, kritis, clinical, verification *string) error {
	if jenis == "" {
		jenis = entity.DiagnosaDiagnosis
	}
	for _, c := range []struct {
		field string
		nilai *string
		sah   []string
	}{
		{"kategori", kategori, kategoriDiagnosa[jenis]},
		{"kritikalitas", kritis, kritikalitas},
		{"clinical_status", clinical, clinicalStatus[jenis]},
		{"verification_status", verification, verificationStatus[jenis]},
	} {
		if err := cekEnum(c.field, c.nilai, c.sah); err != nil {
			return err
		}
	}
	return nil
}

func ValidateStatusLayanan(jenis string, status *string) error {
	return cekEnum("status", status, statusLayanan[jenis])
}
