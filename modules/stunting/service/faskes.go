package service

import (
	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/satusehat"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/helper/validasi"
)

func (s *StuntingService) UpdateAnakFaskes(orgid string, data *entity.Anak) (*types.Anak, error) {
	if data == nil || (!utils.IsStrFilled(data.ID) && !utils.IsStrFilled(data.SatusehatId)) {
		return nil, exceptions.KolomWajib.Messagef("Id atau id_satusehat wajib dikirim salah astu")
	}

	return s.UpdateAnakById(data, utils.PtrStr(orgid, true))
}

func (s *StuntingService) UpdateOrangtuaFaskes(orgid string, data *entity.Orangtua) (*types.Orangtua, error) {
	if data == nil || (!utils.IsStrFilled(data.ID) && !utils.IsStrFilled(data.SatusehatId)) {
		return nil, exceptions.KolomWajib.Messagef("Id atau id_satusehat wajib dikirim salah astu")
	}

	return s.UpdateOrangTuaById(data, utils.PtrStr(orgid, true))
}
func (s *StuntingService) UpdateKunjunganBySatusehatId(orgid string, data *entity.Kunjungan) (*types.Kunjungan, error) {
	if data == nil {
		return nil, errSatusehatWajib()
	}
	if err := s.syaratUbahFaskes(orgid, "kunjungan", data.SatusehatId); err != nil {
		return nil, err
	}
	gabungan, err := s.Repository.UpdateKunjunganBySatusehatId(data, false)

	if err != nil {
		return nil, err
	}

	var res *types.Kunjungan
	res = res.FromEntity(gabungan)

	return res, nil
}

func (s *StuntingService) UpdateObservasiBySatusehatId(orgid string, data *entity.Observasi) (*types.Observasi, error) {
	if data == nil {
		return nil, errSatusehatWajib()
	}
	if err := s.syaratUbahFaskes(orgid, "observasi", data.SatusehatId); err != nil {
		return nil, err
	}
	gabungan, err := s.Repository.UpdateObservasiBySatusehatId(data, false)

	if err != nil {
		return nil, err
	}
	if err := s.perbaruiUkurKunjungan(*data.SatusehatId); err != nil {
		return nil, err
	}

	var res *types.Observasi
	return res.FromEntity(gabungan), nil
}

// kolom ukur di kunjungan ikut diperbarui
func (s *StuntingService) perbaruiUkurKunjungan(satusehatId string) error {
	o, err := s.Repository.FindObservasiBySatusehatId(satusehatId)
	if err != nil || o == nil || !utils.IsFilled(o.IDKunjungan) {
		return err
	}

	ukur := &types.Kunjungan{}
	angkatKeKunjungan(ukur, []satusehat.HasilObservasi{satusehat.HasilDariObservasi(*(*types.Observasi)(nil).FromEntity(o), nil)})

	return s.Repository.UpdateUkurKunjungan(*o.IDKunjungan, satusehat.KolomUkurTerisi(ukur))
}

func (s *StuntingService) UpdateDiagnosaBySatusehatId(orgid string, data *entity.Diagnosa) (*types.Diagnosa, error) {
	if data == nil {
		return nil, errSatusehatWajib()
	}
	if err := s.syaratUbahFaskes(orgid, "diagnosa", data.SatusehatId); err != nil {
		return nil, err
	}
	// jenis tidak bisa diubah, jadi enum dicek terhadap jenis yang tersimpan
	lama, err := s.Repository.FindDiagnosaBySatusehatId(*data.SatusehatId)
	if err != nil {
		return nil, err
	}
	if err := validasi.ValidateEnumDiagnosa(lama.Jenis, data.Kategori, data.Kritikalitas, data.ClinicalStatus, data.VerificationStatus); err != nil {
		return nil, exceptions.Validasi.WithMessage(err.Error(), err)
	}

	gabungan, err := s.Repository.UpdateDiagnosaBySatusehatId(data, false)

	if err != nil {
		return nil, err
	}

	var res *types.Diagnosa
	return res.FromEntity(gabungan), nil
}

func (s *StuntingService) UpdateLayananBySatusehatId(orgid string, data *entity.Layanan) (*types.Layanan, error) {
	if data == nil {
		return nil, errSatusehatWajib()
	}
	if err := s.syaratUbahFaskes(orgid, "layanan", data.SatusehatId); err != nil {
		return nil, err
	}
	lama, err := s.Repository.FindLayananBySatusehatId(*data.SatusehatId)
	if err != nil {
		return nil, err
	}
	if err := validasi.ValidateStatusLayanan(lama.Jenis, data.Status); err != nil {
		return nil, exceptions.Validasi.WithMessage(err.Error(), err)
	}
	gabungan, err := s.Repository.UpdateLayananBySatusehatId(data, false)

	if err != nil {
		return nil, err
	}

	var res *types.Layanan

	return res.FromEntity(gabungan), nil
}

func (s *StuntingService) UpdateRujukanBySatusehatId(orgid string, data *entity.Rujukan) (*types.Rujukan, error) {
	if data == nil {
		return nil, errSatusehatWajib()
	}
	if err := s.syaratUbahFaskes(orgid, "rujukan", data.SatusehatId); err != nil {
		return nil, err
	}

	gabungan, err := s.Repository.UpdateRujukanBySatusehatId(data, false)

	if err != nil {
		return nil, err
	}

	var res *types.Rujukan

	return res.FromEntity(gabungan), nil
}

func (s *StuntingService) UpdateEpisodeBySatusehatId(orgid string, data *entity.Episode) (*types.Episode, error) {
	if data == nil {
		return nil, errSatusehatWajib()
	}
	if err := s.syaratUbahFaskes(orgid, "episode", data.SatusehatId); err != nil {
		return nil, err
	}
	gabungan, err := s.Repository.UpdateEpisodeBySatusehatId(data, false)

	if err != nil {
		return nil, err
	}

	var res *types.Episode

	return res.FromEntity(gabungan), nil
}

func errSatusehatWajib() error { return exceptions.KolomWajib.Messagef("id_satusehat wajib dikirim") }

// hanya faskes pemilik, lewat id_satusehat
func (s *StuntingService) syaratUbahFaskes(orgid, jenis string, satusehatId *string) error {
	if satusehatId == nil {
		return errSatusehatWajib()
	}
	return s.VerifyAccessForUpdateFaskes(orgid, jenis, *satusehatId)
}
