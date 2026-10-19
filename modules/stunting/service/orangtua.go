package service

import (
	"database/sql"
	"errors"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
)

func (s *StuntingService) CreateOrangTua(orangtua *entity.Orangtua) (*types.Orangtua, error) {
	e, err := s.Repository.CreateOrangTua(orangtua, nil)
	return keDTO((*types.Orangtua).FromEntity, e, err)
}

func (s *StuntingService) UpdateOrangTuaById(orangtua *entity.Orangtua, orgid *string) (*types.Orangtua, error) {
	lama, err := s.cariOrangtua(&orangtua.ID, &orangtua.NIK)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, exceptions.ErrTidakDitemukan) {
			return nil, exceptions.TidakDitemukan.Messagef("Data orangtua tidak ditemukan")
		}
		return nil, err
	}

	if utils.StrPtr(orgid) != "jakantro" {
		orangtua.IDPosyandu = nil
	} else {
		orangtua.SatusehatId = ""
	}

	lama.Override(*orangtua, false)
	lama.UpdatedAt = new(time.Now())
	lama.UpdatedBy = orgid

	e, err := s.Repository.UpdateOrangTuaById(lama, nil, false)
	return keDTO((*types.Orangtua).FromEntity, e, err)
}

func (s *StuntingService) FindOrangTua(id *string, nik *string, nokk *string, nama_ayah *string, nama_ibu *string) (*types.Orangtua, error) {
	e, err := s.cariOrangtua(id, nik)
	return keDTO((*types.Orangtua).FromEntity, e, err)
}

func (s *StuntingService) cariOrangtua(id, nik *string) (*entity.Orangtua, error) {
	switch {
	case utils.IsFilled(id):
		return s.Repository.FindOrangTuaById(*id, true)
	case utils.IsFilled(nik):
		return s.Repository.FindOrangTuaByNik(*nik, true)
	}
	return nil, exceptions.ErrTidakDitemukan
}

func (s *StuntingService) DeleteOrangtuaByID(id, orgid string) error {
	if err := s.Repository.DeleteOrangtuaByID(id, orgid); err != nil {
		return err
	}
	s.Repository.LupakanOrangtua(&entity.Orangtua{ID: id})

	s.Background.Run("delete_cache_orangtua", func() {
		if orangtua, err := s.Repository.FindDeletedOrangtua(id); err == nil && orangtua != nil {
			s.Repository.LupakanOrangtua(orangtua)
		}
	})

	return nil
}
