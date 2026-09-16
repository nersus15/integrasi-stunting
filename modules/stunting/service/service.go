package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

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

func (s *StuntingService) UpdateOrangTuaById(orangtua *entity.Orangtua, updatedBy *string) (*types.Orangtua, error) {
	// Cari dulu
	o, err := s.FindOrangTua(&orangtua.ID, &orangtua.NIK, &orangtua.NoKK, nil, nil)

	if err != nil || o == nil {
		if o == nil || errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.TidakDitemukan.Messagef("Data anak tidak ditemukan")
		} else {
			return nil, err
		}
	}

	orangtua_db := o.ToPayload().ToEntity()
	orangtua_db.Override(*orangtua, false)

	orangtua_db.UpdatedAt = new(time.Now())
	orangtua_db.UpdatedBy = updatedBy

	return s.Repository.UpdateOrangTuaById(orangtua_db, nil, false)
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

func (s *StuntingService) UpdateAnakById(anak *entity.Anak, updatedBy *string) (*types.Anak, error) {
	// cari dulu
	a, err := s.FindAnak(&anak.ID, anak.NIK, &anak.IDOrangtua, &anak.AnakKe)
	if err != nil || a == nil {
		return nil, exceptions.TidakDitemukan.Messagef("Data anak tidak ditemukan")
	}

	anak_db := a.ToPayload().ToEntity()
	anak_db.Override(*anak, false)

	anak_db.UpdatedAt = new(time.Now())
	anak_db.UpdatedBy = updatedBy

	return s.Repository.UpdateAnakById(anak_db, nil, false)
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

func (s *StuntingService) VerifyAccess(idposyandu *string, orgid *string) error {

	if !utils.IsFilled(orgid) {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses", nil)
	}

	// null berarti belum terikat posyandu, beda dari posyandu tidak dikenal
	if !utils.IsFilled(idposyandu) {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, orangtua belum terikat posyandu", nil)
	}

	posyandu, err := s.Repository.FindPosyanduDetail(*idposyandu)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Posyandu Tidak ditemukan", nil)
		}
		return exceptions.Forbidden.WithMessage(err.Error(), err)
	}

	if posyandu == nil {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Posyandu Tidak ditemukan", nil)
	}

	// jika puskesmas/RS
	if utils.IsFilled(orgid) && *orgid != "jakantro" {
		if posyandu.Puskesmas == nil {
			return exceptions.Forbidden.WithMessage("Tidak bisa verifikasi akses, posyandu tidak terhubung dengan puskesmas", nil)
		}
		valid := false
		satusehatid := posyandu.Puskesmas.SatusehatId
		induk := posyandu.Puskesmas.Induk
		if satusehatid != nil {
			if *satusehatid == *orgid {
				valid = true
			}
		} else if induk != nil {
			if induk.SatusehatId != nil && *induk.SatusehatId == *orgid {
				valid = true
			}
		}

		if !valid {
			if !utils.IsFilled(posyandu.Puskesmas.Wilayah) {
				return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Wilayah Faskes posyandu kosong", nil)
			}

			// Cari wilayah kerja dari faskes user
			faskesSaya, err := s.Repository.FindFaskesByOrgid(*orgid)

			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Faskes user Tidak ditemukan", nil)
				}
				return exceptions.Forbidden.WithMessage(err.Error(), err)
			}

			if faskesSaya == nil {
				return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Faskes user Tidak ditemukan", nil)
			}

			if slices.Contains([]string{"PUSTU", "PUSKESMAS"}, strings.ToUpper(faskesSaya.Jenis)) {
				if !utils.IsFilled(faskesSaya.Wilayah) {
					return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Wilayah Faskes user kosong", nil)
				}
			} else {
				// Untuk RS, jangan bandingkan wilayah kerja tapi rujukan (saat ini belum ada)

			}

			levelWilayah := utils.LevelWilayah(posyandu.Puskesmas.Wilayah)

			switch levelWilayah {
			case 3: // kecamatan
				if utils.Substr(*faskesSaya.Wilayah, 0, 9)+"00000" == *posyandu.Puskesmas.Wilayah {
					valid = true
				}
			case 4:
				if utils.Substr(*faskesSaya.Wilayah, 0, 9)+"0000" == utils.Substr(*posyandu.Puskesmas.Wilayah, 0, 9)+"0000" {
					valid = true
				}
			}

		}

		if !valid {
			return exceptions.Forbidden.WithMessage("Tidak bisa akses data", nil)
		}
	}

	return nil
}

func (s *StuntingService) VerifyAccessByIdOrangtua(idorangtua string, orgid *string) error {
	// Ini pembungkus kecil untuk anak, karena di dto anak tidak ada id posyandu
	if !utils.IsStrFilled(idorangtua) {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, Id orang tua kosong", nil)
	}

	// cari orangtua
	orangtua, err := s.Repository.FindOrangTuaById(idorangtua)
	if err != nil {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses", nil)
	}
	if orangtua == nil {
		return exceptions.Forbidden.WithMessage("Tidak Bisa Verifikasi Akses, orangtua tidak ditemukan", nil)
	}

	return s.VerifyAccess(orangtua.IdPosyandu, orgid)
}
