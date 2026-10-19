package service

import (
	"errors"
	"fmt"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/satusehat"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
)

func (s *StreamService) verifyEncounterPemeriksaan(orgid string, p *types.PemeriksaanFaskes) error {
	encounter := utils.StrPtr(p.Kunjungan.IdSatusehat)
	if err := s.stunting.VerifyEncounterMilik(orgid, encounter, true); err != nil {
		return err
	}

	refs := make([]*string, 0, len(p.Observasi)+len(p.Diagnosa)+len(p.Layanan)+len(p.Rujukan))
	for _, o := range p.Observasi {
		refs = append(refs, o.RefEncounter)
	}
	for _, d := range p.Diagnosa {
		refs = append(refs, d.RefEncounter)
	}
	for _, l := range p.Layanan {
		refs = append(refs, l.RefEncounter)
	}
	for _, r := range p.Rujukan {
		refs = append(refs, r.RefEncounter)
	}

	for _, ref := range refs {
		if !utils.IsFilled(ref) || *ref == encounter {
			continue
		}
		if err := s.stunting.VerifyEncounterMilik(orgid, *ref, false); err != nil {
			return err
		}
	}

	return nil
}

// orgid dari API key, bukan dari body
func (s *StreamService) SimpanPemeriksaanFaskes(p *types.PemeriksaanFaskes, orgid string) (*types.HasilPemeriksaan, error) {
	if p == nil {
		return nil, exceptions.BodyRusak.New(nil)
	}
	if !utils.IsStrFilled(orgid) {
		return nil, exceptions.Forbidden.WithMessage("orgid faskes tidak diketahui", nil)
	}

	faskes, err := s.repository.FindFaskesByOrgid(orgid)
	if err != nil || faskes == nil {
		return nil, exceptions.Forbidden.WithMessage("faskes dengan orgid "+orgid+" belum terdaftar", err)
	}

	for i, r := range p.Rujukan {
		if !utils.IsFilled(r.RefFaskesTujuan) {
			continue
		}
		if _, err := s.repository.FindFaskesByOrgid(*r.RefFaskesTujuan); err != nil {
			var e *exceptions.Error
			if errors.As(err, &e) && e.ErrorCode == exceptions.TidakDitemukan.ErrorCode {
				return nil, exceptions.ReferensiHilang.Messagef("rujukan[%d]: faskes tujuan %s belum terdaftar", i, *r.RefFaskesTujuan)
			}
			return nil, err
		}
	}

	observasi := make([]satusehat.HasilObservasi, 0, len(p.Observasi))
	for _, o := range p.Observasi {
		observasi = append(observasi, satusehat.HasilDariObservasi(o, o.Component))
	}

	// id dan kepemilikan ditentukan sistem
	p.Anak.Id, p.Anak.IdOrangtua = "", ""
	p.Kunjungan.IdEpisode, p.Kunjungan.IdRujukan, p.Kunjungan.Stunting = nil, nil, nil
	for i := range observasi {
		observasi[i].Induk.IdInduk = nil
	}
	for i := range p.Episode {
		p.Episode[i].IdFaskes = nil
	}

	var anakBaru *entity.Anak
	anak := s.CariAnak(&p.Anak)
	if anak == nil {
		if !adaTandaStunting(p.Diagnosa, observasi) {
			return nil, exceptions.TidakDisimpan.Messagef("tidak ada tanda stunting dan anak belum punya riwayat")
		}

		p.Anak.SourceData = fmt.Sprintf("faskes-api:%s", orgid)

		anakBaru, err = siapkanAnakBaru(&p.Anak, nil)
		if err != nil {
			return nil, err
		}
		if anakBaru == nil {
			return nil, exceptions.TidakDisimpan.Messagef("anak bukan sasaran pemantauan, umurnya sudah lewat 5 tahun")
		}
		anak = &p.Anak
	} else if _, err := s.stunting.VerifyAksesAnak(&orgid, anak, utils.PtrStr(anak.Id, true), nil); err != nil {
		return nil, err
	}

	if err := s.verifyEncounterPemeriksaan(orgid, p); err != nil {
		return nil, err
	}

	kunjungan := p.Kunjungan
	kunjungan.IdFaskes = &faskes.ID

	idKunjungan, disimpan, alasan, err := s.simpanDataMedis(anak, anak.IdSatusehat, &kunjungan, &orgid,
		observasi, p.Diagnosa, p.Layanan, p.Rujukan, p.Episode, anakBaru)

	if err != nil {
		return nil, err
	}

	if !disimpan {
		return nil, exceptions.TidakDisimpan.Messagef("%s", alasan)
	}

	return &types.HasilPemeriksaan{
		IdAnak:      anak.Id,
		IdKunjungan: idKunjungan,
	}, nil
}
