package service

import (
	"encoding/json"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/helper/validasi"
)

func (s *StuntingService) CreateKunjungan(body []byte) (*types.KunjunganAnak, error) {
	res := new(types.KunjunganAnak)

	var anak *entity.Anak
	var orangtua *entity.Orangtua
	var kunjungan *entity.Kunjungan

	var payload *types.KunjunganNestedPayload

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}
	if payload != nil {
		validasi.AbaikanFieldSistem(payload)
		bersihkanPayloadJakantro(payload.Orangtua, payload.Anak)
		bersihkanKunjunganJakantro(payload.Kunjungan)
		bersihkanKunjunganJakantro(&payload.KunjunganPayload)
	}

	jenis, payloadKunjungan, err := validasi.DeteksiJenisPayload(payload)

	if err != nil {
		return nil, exceptions.BentukPayload.WithMessage(err.Error(), err)
	}

	switch jenis {
	case validasi.JenisRegistrasiLengkap, validasi.JenisAnakBaru:
		if jenis == validasi.JenisRegistrasiLengkap {
			err := validasi.ValidateOrangtua(*payload.Orangtua)
			if err != nil {
				return nil, exceptions.Validasi.WithMessage(err.Error(), err)
			}
		}

		err := validasi.ValidateAnak(*payload.Anak)
		if err != nil {
			return nil, exceptions.Validasi.WithMessage(err.Error(), err)
		}

		err = validasi.ValidateKunjungan(*payload.Kunjungan)
		if err != nil {
			return nil, exceptions.Validasi.WithMessage(err.Error(), err)
		}

		if jenis == validasi.JenisAnakBaru {
			payload.Orangtua = nil
		}

		orangtua, anak, kunjungan, err = s.Repository.KunjunganTransaction(payload.Orangtua.ToEntity(), payload.Anak.ToEntity(), payload.Kunjungan.ToEntity())

		if err != nil {
			return nil, exceptions.Classify(err)
		}

	case validasi.JenisKunjunganSaja:
		if err := validasi.ValidateKunjungan(*payloadKunjungan); err != nil {
			return nil, exceptions.Validasi.WithMessage(err.Error(), err)
		}

		kunjungan, err = s.Repository.CreateKunjungan(payloadKunjungan.ToEntity(), nil)

		if err != nil {
			return nil, exceptions.Classify(err)
		}
	}

	if jenis == validasi.JenisKunjunganSaja {
		res.IdAnak = &payloadKunjungan.IDAnak
	} else {
		res.IdOrangtua, res.IdAnak, res.Orangtua, res.Anak = identitasRespons(jenis, orangtua, anak)
	}
	res.Kunjungan = *(*types.Kunjungan)(nil).FromEntity(kunjungan)
	return res, nil
}

func (s *StuntingService) UpdateKunjunganById(payload *types.KunjunganPayload, id string) (*types.Kunjungan, error) {
	lama, err := s.Repository.FindKunjunganById(id)
	if err != nil {
		return nil, exceptions.Classify(err)
	}

	if lama.IDFaskes != nil {
		// berarti kunjungan ini bukan dari jakantro, dan jakantro tidak boleh update
		return nil, exceptions.Forbidden.Messagef("Kunjungan dengan id:%s dibuat oleh faskes:%s. jakantro tidak boleh update", id, utils.StrPtr(lama.IDFaskes))
	}

	lama.Override(*payload.ToEntity(), false)
	e, err := s.Repository.UpdateKunjunganById(id, lama, false)
	return keDTO((*types.Kunjungan).FromEntity, e, err)
}

func (s *StuntingService) FindKunjunganById(id string) (*types.Kunjungan, error) {
	e, err := s.Repository.FindKunjunganById(id)
	return keDTO((*types.Kunjungan).FromEntity, e, err)
}

func (s *StuntingService) KunjunganByAnak(id string) (*types.KunjunganAnakArray, error) {
	e, err := s.Repository.ListKunjunganAnak(id)
	return keDTO((*types.KunjunganAnakArray).FromEntity, e, err)
}

// satusehat_id hanya diisi jalur SatuSehat dan faskes
func bersihkanPayloadJakantro(orangtua *types.OrangtuaPayload, anak *types.AnakPayload) {
	if orangtua != nil {
		orangtua.SatusehatId = nil
	}
	if anak != nil {
		anak.SatusehatId = nil
	}
}

// kepemilikan faskes, episode, rujukan, dan status stunting milik jalur faskes
func bersihkanKunjunganJakantro(k *types.KunjunganPayload) {
	if k == nil {
		return
	}
	k.SatusehatId, k.IdFaskes, k.Stunting = nil, nil, nil
	k.IdEpisode, k.RefEpisode, k.IdRujukan, k.RefRujukan = nil, nil, nil, nil
}

// bagian identitas respons registrasi: orangtua hanya ikut bila didaftarkan sekalian
func identitasRespons(jenis validasi.JenisPayload, orangtua *entity.Orangtua, anak *entity.Anak) (idOrangtua, idAnak *string, o *types.Orangtua, a *types.Anak) {
	if jenis == validasi.JenisRegistrasiLengkap {
		idOrangtua, o = &orangtua.ID, (*types.Orangtua)(nil).FromEntity(orangtua)
	}
	return idOrangtua, &anak.ID, o, (*types.Anak)(nil).FromEntity(anak)
}
