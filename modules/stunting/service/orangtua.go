package service

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
)

func (s *StuntingService) CreateOrangTua(orangtua *entity.Orangtua, orgid string) (*types.Orangtua, error) {
	if orgid != pembuatJakantro {
		orangtua.ID = uuid.NewString()
	}
	asalPemanggil(orgid, orangtua.SourceData).orangtua(orangtua)
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
	if err := s.Akses(utils.StrPtr(orgid), AksiUbah, sasaranOrangtua(lama)); err != nil {
		return nil, err
	}

	switch {
	case utils.StrPtr(orgid) != pembuatJakantro:
		orangtua.IDPosyandu = nil
	case !dibuatOleh(lama.CreatedBy, pembuatJakantro):
		orangtua.SatusehatId, orangtua.NIK = "", ""
	default:
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
	// created_by tidak pernah berubah, jadi aman dibaca dari cache
	lama, err := s.Repository.FindOrangTuaById(id, true)
	if err != nil {
		return exceptions.Classify(err)
	}
	if err := s.Akses(orgid, AksiHapus, sasaranOrangtua(lama)); err != nil {
		return err
	}
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
