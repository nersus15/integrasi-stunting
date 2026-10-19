package service

import (
	"encoding/json"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/validasi"
)

func (s *StuntingService) UpdateKesehatanById(payload *types.KesehatanPayload, id string) (*types.Kesehatan, error) {
	lama, err := s.Repository.FindKesehatanById(id)
	if err != nil {
		return nil, exceptions.Classify(err)
	}

	lama.Override(*payload.ToEntity(), false)
	e, err := s.Repository.UpdateKesehatanById(id, lama, false)
	return keDTO((*types.Kesehatan).FromEntity, e, err)
}

// bentuk payload sama dengan CreateKunjungan, entitas terakhirnya kesehatan
func (s *StuntingService) CreateKesehatan(body []byte) (*types.KesehatanAnak, error) {
	res := new(types.KesehatanAnak)

	var anak *entity.Anak
	var orangtua *entity.Orangtua
	var kesehatan *entity.Kesehatan

	var payload *types.KesehatanNestedPayload

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}
	if payload != nil {
		validasi.AbaikanFieldSistem(payload)
		bersihkanPayloadJakantro(payload.Orangtua, payload.Anak)
	}

	jenis, payloadKesehatan, err := validasi.DeteksiJenisPayloadKesehatan(payload)
	if err != nil {
		return nil, exceptions.BentukPayload.WithMessage(err.Error(), err)
	}

	switch jenis {
	case validasi.JenisRegistrasiLengkap, validasi.JenisAnakBaru:
		if jenis == validasi.JenisRegistrasiLengkap {
			if err := validasi.ValidateOrangtua(*payload.Orangtua); err != nil {
				return nil, exceptions.Validasi.WithMessage(err.Error(), err)
			}
		}

		if err := validasi.ValidateAnak(*payload.Anak); err != nil {
			return nil, exceptions.Validasi.WithMessage(err.Error(), err)
		}

		if err := validasi.ValidateKesehatan(*payload.Kesehatan); err != nil {
			return nil, exceptions.Validasi.WithMessage(err.Error(), err)
		}

		if jenis == validasi.JenisAnakBaru {
			payload.Orangtua = nil
		}

		orangtua, anak, kesehatan, err = s.Repository.KesehatanTransaction(
			payload.Orangtua.ToEntity(), payload.Anak.ToEntity(), payload.Kesehatan.ToEntity())

		if err != nil {
			return nil, exceptions.Classify(err)
		}

	case validasi.JenisKesehatanSaja:
		if err := validasi.ValidateKesehatan(*payloadKesehatan); err != nil {
			return nil, exceptions.Validasi.WithMessage(err.Error(), err)
		}

		kesehatan, err = s.Repository.CreateKesehatan(payloadKesehatan.ToEntity(), nil)

		if err != nil {
			return nil, exceptions.Classify(err)
		}
	}

	if jenis == validasi.JenisKesehatanSaja {
		res.IdAnak = &payloadKesehatan.IDAnak
	} else {
		res.IdOrangtua, res.IdAnak, res.Orangtua, res.Anak = identitasRespons(jenis, orangtua, anak)
	}
	res.Kesehatan = *(*types.Kesehatan)(nil).FromEntity(kesehatan)
	return res, nil
}

func (s *StuntingService) KesehatanByAnak(id string) (*types.KesehatanAnakArray, error) {
	e, err := s.Repository.ListKesehatanAnak(id)
	return keDTO((*types.KesehatanAnakArray).FromEntity, e, err)
}

func (s *StuntingService) FindKesehatanById(id string) (*types.Kesehatan, error) {
	e, err := s.Repository.FindKesehatanById(id)
	return keDTO((*types.Kesehatan).FromEntity, e, err)
}
