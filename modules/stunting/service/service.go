package service

import (
	"encoding/json"

	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/entity"
	exceptions "github.com/nersus15/integrasi/mod-stunting/helper/Exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/repository"
	backgroundworker "github.com/nersus15/lib-background-worker"
	"github.com/webcore-go/webcore/app/core"
)

type StuntingService struct {
	Context    *core.AppContext
	Config     *config.ModuleConfig
	Repository *repository.StuntingRepository
	Token      *string
	Background *backgroundworker.BackgroundWorker
	// DevHttpClient            *http.Client
	// ProdHttpClient           *http.Client
	// DevBackgroundHttpClient  *http.Client // connection pool TERPISAH dari DevHttpClient, khusus trafik background (credential sync & backup forward) supaya tidak berebut slot koneksi dengan request real-time yang sedang ditunggu user
	// ProdBackgroundHttpClient *http.Client
}

func NewStuntingService(wctx *core.AppContext, cfg *config.ModuleConfig, repository *repository.StuntingRepository, background *backgroundworker.BackgroundWorker) *StuntingService {
	return &StuntingService{
		Context:    wctx,
		Config:     cfg,
		Repository: repository,
		Background: background,
		// DevHttpClient:            utils.CreateHttpClient(cfg.Ildki.HttpProxy),
		// ProdHttpClient:           utils.CreateHttpClient(cfg.Ildki.HttpProxy),
		// DevBackgroundHttpClient:  utils.CreateHttpClient(cfg.Ildki.HttpProxy),
		// ProdBackgroundHttpClient: utils.CreateHttpClient(cfg.Ildki.HttpProxy),
	}
}

func (s *StuntingService) CreateOrangTua(orangtua *entity.Orangtua) (*types.Orangtua, error) {
	return s.Repository.CreateOrangTua(orangtua, nil)
}

func (s *StuntingService) FindOrangTua(id *string, nik *string, nokk *string, nama_ayah *string, nama_ibu *string) (*types.Orangtua, error) {
	switch {
	case utils.IsFilled(id):
		return s.Repository.FindOrangTuaById(*id)
	case utils.IsFilled(nik):
		return s.Repository.FindOrangTuaByNik(*nik)
	}

	return nil, utils.ErrTidakDitemukan
}

func (s *StuntingService) CreateAnak(anak *entity.Anak) (*types.Anak, error) {
	return s.Repository.CreateAnak(anak, nil)
}

func (s *StuntingService) FindAnak(id, nik, orangtua *string, urutan *int16) (*types.Anak, error) {
	anak, err := s.Repository.FindAnak(id, nik, orangtua, urutan)

	if err != nil {
		return nil, err
	}

	if anak == nil {
		return nil, utils.ErrTidakDitemukan
	}
	return anak, nil
}

func (s *StuntingService) ListAnak(idorangtua *string) (*types.ListAnak, error) {
	r, err := s.Repository.ListAnak(idorangtua, nil, nil)

	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, utils.ErrTidakDitemukan
	}
	return r, nil
}

func (s *StuntingService) CreateKunjungan(body []byte) (*types.KunjunganAnak, error) {
	res := new(types.KunjunganAnak)

	var anak *types.Anak
	var orangtua *types.Orangtua
	var kunjungan *types.Kunjungan

	var payload *types.KunjunganNestedPayload

	// unmarshall body
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}

	// Deteksi jenis payload
	jenis, payloadKunjungan, err := utils.DeteksiJenisPayload(payload)

	if err != nil {
		return nil, exceptions.BentukPayload.WithMessage(err.Error(), err)
	}

	switch jenis {
	case utils.JenisRegistrasiLengkap, utils.JenisAnakBaru:
		if jenis == utils.JenisRegistrasiLengkap {
			err := utils.ValidateOrangtua(*payload.Orangtua)
			if err != nil {
				return nil, exceptions.Validasi.WithMessage(err.Error(), err)
			}
		}

		err := utils.ValidateAnak(*payload.Anak)
		if err != nil {
			return nil, exceptions.Validasi.WithMessage(err.Error(), err)
		}

		err = utils.ValidateKunjungan(*payload.Kunjungan)
		if err != nil {
			return nil, exceptions.Validasi.WithMessage(err.Error(), err)
		}

		if jenis == utils.JenisAnakBaru {
			payload.Orangtua = nil
		}

		orangtua, anak, kunjungan, err = s.Repository.KunjunganTransaction(payload.Orangtua.ToEntity(), payload.Anak.ToEntity(), payload.Kunjungan.ToEntity())

		if err != nil {
			return nil, exceptions.Classify(err)
		}

	case utils.JenisKunjunganSaja:
		if err := utils.ValidateKunjungan(*payloadKunjungan); err != nil {
			return nil, exceptions.Validasi.WithMessage(err.Error(), err)
		}

		kunjungan, err = s.Repository.CreateKunjungan(payloadKunjungan.ToEntity(), nil)

		if err != nil {
			return nil, exceptions.Classify(err)
		}
	}

	switch jenis {
	case utils.JenisKunjunganSaja:
		res.IdAnak = &payloadKunjungan.IDAnak
	case utils.JenisAnakBaru:
		res.IdAnak = &anak.Id
		res.Anak = anak
	case utils.JenisRegistrasiLengkap:
		res.Orangtua = orangtua
		res.IdOrangtua = &orangtua.Id

		res.IdAnak = &anak.Id
		res.Anak = anak
	}
	res.Kunjungan = *kunjungan
	return res, nil
}

// CreateKesehatan menerima ketujuh bentuk payload yang sama dengan
// CreateKunjungan, hanya entitas terakhirnya kesehatan.
func (s *StuntingService) CreateKesehatan(body []byte) (*types.KesehatanAnak, error) {
	res := new(types.KesehatanAnak)

	var anak *types.Anak
	var orangtua *types.Orangtua
	var kesehatan *types.Kesehatan

	var payload *types.KesehatanNestedPayload

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}

	jenis, payloadKesehatan, err := utils.DeteksiJenisPayloadKesehatan(payload)
	if err != nil {
		return nil, exceptions.BentukPayload.WithMessage(err.Error(), err)
	}

	switch jenis {
	case utils.JenisRegistrasiLengkap, utils.JenisAnakBaru:
		if jenis == utils.JenisRegistrasiLengkap {
			if err := utils.ValidateOrangtua(*payload.Orangtua); err != nil {
				return nil, exceptions.Validasi.WithMessage(err.Error(), err)
			}
		}

		if err := utils.ValidateAnak(*payload.Anak); err != nil {
			return nil, exceptions.Validasi.WithMessage(err.Error(), err)
		}

		if err := utils.ValidateKesehatan(*payload.Kesehatan); err != nil {
			return nil, exceptions.Validasi.WithMessage(err.Error(), err)
		}

		if jenis == utils.JenisAnakBaru {
			payload.Orangtua = nil
		}

		orangtua, anak, kesehatan, err = s.Repository.KesehatanTransaction(
			payload.Orangtua.ToEntity(), payload.Anak.ToEntity(), payload.Kesehatan.ToEntity())

		if err != nil {
			return nil, exceptions.Classify(err)
		}

	case utils.JenisKesehatanSaja:
		if err := utils.ValidateKesehatan(*payloadKesehatan); err != nil {
			return nil, exceptions.Validasi.WithMessage(err.Error(), err)
		}

		kesehatan, err = s.Repository.CreateKesehatan(payloadKesehatan.ToEntity(), nil)

		if err != nil {
			return nil, exceptions.Classify(err)
		}
	}

	switch jenis {
	case utils.JenisKesehatanSaja:
		res.IdAnak = &payloadKesehatan.IDAnak
	case utils.JenisAnakBaru:
		res.IdAnak = &anak.Id
		res.Anak = anak
	case utils.JenisRegistrasiLengkap:
		res.Orangtua = orangtua
		res.IdOrangtua = &orangtua.Id

		res.IdAnak = &anak.Id
		res.Anak = anak
	}

	res.Kesehatan = *kesehatan
	return res, nil
}

func (s *StuntingService) KesehatanByAnak(id string) (*types.KesehatanAnakArray, error) {
	k, err := s.Repository.ListKesehatanAnak(id)
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, utils.ErrTidakDitemukan
	}
	return k, nil
}

func (s *StuntingService) SummaryAnak(id string) (*types.SummaryAnak, error) {
	sum, err := s.Repository.SummaryAnak(id)
	if err != nil {
		return nil, err
	}
	if sum == nil {
		return nil, utils.ErrTidakDitemukan
	}
	return sum, nil
}

func (s *StuntingService) FindKunjunganById(id string) (*types.Kunjungan, error) {
	return s.Repository.FindKunjunganById(id)
}

func (s *StuntingService) FindKesehatanById(id string) (*types.Kesehatan, error) {
	return s.Repository.FindKesehatanById(id)
}

func (s *StuntingService) KunjunganByAnak(id string) (*types.KunjunganAnakArray, error) {
	k, err := s.Repository.ListKunjunganAnak(id)

	if err != nil {
		return nil, err
	}

	if k == nil {
		return nil, utils.ErrTidakDitemukan
	}

	return k, nil
}
