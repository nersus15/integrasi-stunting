package satusehat

import (
	"fmt"

	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
)

func EncounterToKunjungan(entry types2.BundleEntry) (*types.Kunjungan, *string, *string, error) {
	resource, err := resourceDari[fhir.Encounter](entry)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Gagal memetakan resource Encounter ke Kunjungan: %s", err.Error())
	}

	k := &types.Kunjungan{
		IdSatusehat:       entry.Base.Id,
		TanggalPengukuran: tanggalEncounter(resource.Period),
	}

	if resource.Period != nil {
		k.TanggalSelesai = isiStr(TanggalSaja(str(resource.Period.End)))
	}
	if len(resource.EpisodeOfCare) > 0 {
		k.RefEpisode = ReferenceID(&resource.EpisodeOfCare[0])
	}

	k.RefRujukan = refServiceRequest(resource.BasedOn)

	return k, ReferenceID(resource.Subject), ReferenceID(resource.ServiceProvider), nil
}

func tanggalEncounter(p *fhir.Period) string {
	if p == nil {
		return ""
	}
	if utils.IsFilled(p.Start) {
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
