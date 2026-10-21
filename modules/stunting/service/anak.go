package service

import (
	"time"

	"github.com/google/uuid"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
)

func (s *StuntingService) CreateAnak(anak *entity.Anak, orgid string) (*types.Anak, error) {
	if err := s.AksesOrangtua(orgid, AksiUbah, anak.IDOrangtua); err != nil {
		return nil, err
	}
	if orgid != pembuatJakantro {
		anak.ID = uuid.NewString()
	}
	asalPemanggil(orgid, &anak.SourceData).anak(anak)
	e, err := s.Repository.CreateAnak(anak, nil)
	return keDTO((*types.Anak).FromEntity, e, err)
}

func (s *StuntingService) UpdateAnakById(anak *entity.Anak, orgid *string) (*types.Anak, error) {
	lama, err := s.Repository.FindAnak(&anak.ID, anak.NIK, &anak.IDOrangtua, &anak.AnakKe, false)
	if err != nil || lama == nil {
		return nil, exceptions.TidakDitemukan.Messagef("Data anak tidak ditemukan")
	}
	if err := s.Akses(utils.StrPtr(orgid), AksiUbah, sasaranAnak(lama)); err != nil {
		return nil, err
	}

	switch {
	case utils.StrPtr(orgid) != pembuatJakantro:
		anak.IDOrangtua = ""
	case !dibuatOleh(lama.CreatedBy, pembuatJakantro):
		// anak buatan faskes/SatuSehat: jakantro tidak boleh mengubah nik atau memindah keluarga
		anak.SatusehatId, anak.NIK, anak.IDOrangtua = "", nil, ""
	default:
		anak.SatusehatId = ""
	}

	ortuLama := lama.IDOrangtua
	lama.Override(*anak, false)
	lama.UpdatedAt = new(time.Now())
	lama.UpdatedBy = orgid

	res, err := s.Repository.UpdateAnakById(lama, nil, false)
	if err == nil && ortuLama != lama.IDOrangtua {
		s.Repository.LupakanAksesPerubahanOrangtua(lama.ID, ortuLama, lama.IDOrangtua)
	}
	return keDTO((*types.Anak).FromEntity, res, err)
}

func (s *StuntingService) FindAnak(id, nik, orangtua *string, urutan *int16) (*types.Anak, error) {
	e, err := s.Repository.FindAnak(id, nik, orangtua, urutan, true)
	return keDTO((*types.Anak).FromEntity, e, err)
}

func (s *StuntingService) ListAnak(idorangtua *string) (*types.ListAnak, error) {
	e, err := s.Repository.ListAnak(idorangtua, nil, nil)
	return keDTO((*types.ListAnak).FromEntity, e, err)
}

func (s *StuntingService) SummaryAnak(id string) (*types.SummaryAnak, error) {
	e, err := s.Repository.SummaryAnak(id)
	return keDTO((*types.SummaryAnak).FromEntity, e, err)
}

func (s *StuntingService) DeleteAnakByID(id string, orgid string) error {
	// created_by tidak pernah berubah, jadi aman dibaca dari cache
	lama, err := s.Repository.FindAnak(&id, nil, nil, nil, true)
	if err != nil {
		return exceptions.Classify(err)
	}
	if err := s.Akses(orgid, AksiHapus, sasaranAnak(lama)); err != nil {
		return err
	}
	if err := s.Repository.DeleteAnakByID(id, orgid); err != nil {
		return err
	}
	// kunci id dihapus seketika karena dipakai cek akses; kunci lain menyusul di job latar
	s.Repository.LupakanAnak(&entity.Anak{ID: id})

	s.Background.Run("delete_cache_anak", func() {
		if anak, err := s.Repository.FindDeletedAnak(&id, nil, nil, nil); err == nil && anak != nil {
			s.Repository.LupakanAnak(anak)
			s.Repository.LupakanAksesPerubahanOrangtua(anak.ID, anak.IDOrangtua)
		}
	})
	return nil
}
