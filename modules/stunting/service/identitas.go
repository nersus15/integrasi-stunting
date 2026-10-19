package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/helper/validasi"
	"github.com/webcore-go/webcore/infra/logger"
)

// nil tanpa error berarti bukan sasaran pemantauan, bukan kegagalan
func (s *StreamService) PastikanAnak(anak *types.Anak, anakKe *int) (*types.Anak, error) {
	if anak == nil {
		return nil, exceptions.Validasi.Messagef("data anak kosong")
	}
	if utils.IsStrFilled(anak.Id) {
		return anak, nil
	}

	e, err := siapkanAnakBaru(anak, anakKe)
	if err != nil || e == nil {
		return nil, err
	}
	baru, err := s.repository.CreateAnak(e, nil)
	return keDTO((*types.Anak).FromEntity, baru, err)
}

// nil tanpa error berarti anak sudah lewat 5 tahun
func siapkanAnakBaru(anak *types.Anak, anakKe *int) (*entity.Anak, error) {
	bd, err := time.Parse("2006-01-02", anak.TanggalLahir)
	if err != nil {
		return nil, exceptions.Validasi.Messagef("anak.tanggal_lahir tidak valid: %q", anak.TanggalLahir)
	}
	if bd.After(time.Now()) {
		return nil, exceptions.Validasi.Messagef("anak.tanggal_lahir belum terjadi")
	}
	if !time.Now().Before(bd.AddDate(5, 0, 0)) {
		logger.Info(fmt.Sprintf("PastikanAnak: %s sudah lewat 5 tahun, dilewati", utils.Nilai(anak.IdSatusehat)))
		return nil, nil
	}

	anak.Id = uuid.NewString()
	p := anak.ToPayload()
	if anakKe != nil {
		p.AnakKe = int16(*anakKe)
	}
	if err := validasi.ValidateAnakStream(*p); err != nil {
		return nil, exceptions.Validasi.WithMessage("anak."+err.Error(), err)
	}
	return p.ToEntity(), nil
}

func (s *StreamService) CariAnak(a *types.Anak) *types.Anak {
	if a == nil {
		return nil
	}
	if utils.IsFilled(a.IdSatusehat) {
		if k := anakDTO(s.repository.FindAnakBySatusehatId(*a.IdSatusehat, false)); k != nil {
			return k
		}
	}
	if utils.IsFilled(a.Nik) {
		if k := anakDTO(s.repository.FindAnak(nil, a.Nik, nil, nil, false)); k != nil {
			return k
		}
	}
	return nil
}

func (s *StreamService) CariOrangtua(o *types.Orangtua) *types.Orangtua {
	if o == nil {
		return nil
	}

	if utils.IsFilled(o.IdSatusehat) {
		if k := orangtuaDTO(s.repository.FindOrangtuaBySatusehatId(*o.IdSatusehat, false)); k != nil {
			return k
		}
	}

	if utils.IsStrFilled(o.Nik) {
		if k := orangtuaDTO(s.repository.FindOrangTuaByNik(o.Nik, false)); k != nil {
			return k
		}
	}

	return nil
}

// repository mengembalikan entity, pipeline FHIR memakai types
func anakDTO(e *entity.Anak, err error) *types.Anak {
	if err != nil || e == nil {
		return nil
	}
	return (*types.Anak)(nil).FromEntity(e)
}

func orangtuaDTO(e *entity.Orangtua, err error) *types.Orangtua {
	if err != nil || e == nil {
		return nil
	}
	return (*types.Orangtua)(nil).FromEntity(e)
}

// penghubungan butuh id hasil create, jadi berurutan
func (s *StreamService) simpanOrangtua(orangtua *types.Orangtua, perluUpdate bool, idAnakSatusehat *string) {
	idOrangtua := orangtua.Id

	switch {
	case !utils.IsStrFilled(idOrangtua):
		// belum terdaftar di database, buat baru
		orangtua.Id = uuid.NewString()

		p := orangtua.ToPayload()
		if err := validasi.ValidateOrangtuaStream(*p); err != nil {
			logger.Error("ProsesTransaksiFHIR:CreateOrangtua validasi => " + err.Error())
			return
		}

		e := p.ToEntity()
		if orangtua.IdSatusehat != nil {
			e.SatusehatId = *orangtua.IdSatusehat
		}

		o, err := s.repository.CreateOrangTua(e, nil)
		if err != nil {
			logger.Error("ProsesTransaksiFHIR:CreateOrangtua => " + err.Error())
			return
		}
		idOrangtua = o.ID

	case perluUpdate:
		// sudah ada tapi satusehat_id-nya masih kosong
		e := orangtua.ToPayload().ToEntity()
		if orangtua.IdSatusehat != nil {
			e.SatusehatId = *orangtua.IdSatusehat
		}

		if _, err := s.repository.UpdateOrangTuaById(e, nil, true); err != nil {
			logger.Error("ProsesTransaksiFHIR:UpdateOrangtua => " + err.Error())
		}
	}

	if !utils.IsFilled(idAnakSatusehat) {
		return
	}

	n, err := s.repository.HubungkanAnakKeOrangTua(*idAnakSatusehat, idOrangtua, nil)
	if err != nil {
		logger.Error("ProsesTransaksiFHIR:HubungkanAnakKeOrangTua => " + err.Error())
		return
	}

	if n == 0 {
		logger.Info(fmt.Sprintf("ProsesTransaksiFHIR: anak %s tidak dihubungkan ke orangtua %s (belum ada atau sudah punya orangtua)", *idAnakSatusehat, idOrangtua))
	}
}

// FHIR menimpa jakantro, hanya untuk field yang terisi di resource
func timpaAnakDariFHIR(db *types.Anak, fhirData *types.Anak) bool {
	berubah := false

	timpaStr := func(tujuan *string, nilai string) {
		if utils.IsStrFilled(nilai) && *tujuan != nilai {
			*tujuan = nilai
			berubah = true
		}
	}

	timpaStr(&db.Nama, fhirData.Nama)
	timpaStr(&db.TanggalLahir, fhirData.TanggalLahir)
	timpaStr(&db.JenisKelamin, fhirData.JenisKelamin)

	if utils.IsFilled(fhirData.Nik) && (!utils.IsFilled(db.Nik) || *db.Nik != *fhirData.Nik) {
		db.Nik = fhirData.Nik
		berubah = true
	}

	if utils.IsFilled(fhirData.IdSatusehat) &&
		(!utils.IsFilled(db.IdSatusehat) || *db.IdSatusehat != *fhirData.IdSatusehat) {
		db.IdSatusehat = fhirData.IdSatusehat
		berubah = true
	}

	return berubah
}

// FHIR menimpa jakantro, hanya untuk field yang terisi di resource
func timpaDariFHIR(db *types.Orangtua, fhirData *types.Orangtua) bool {
	berubah := false

	timpaStr := func(tujuan *string, nilai string) {
		if utils.IsStrFilled(nilai) && *tujuan != nilai {
			*tujuan = nilai
			berubah = true
		}
	}

	timpaStr(&db.NamaAyah, fhirData.NamaAyah)
	timpaStr(&db.NamaIbu, fhirData.NamaIbu)
	timpaStr(&db.Telepon, fhirData.Telepon)
	timpaStr(&db.Alamat, fhirData.Alamat)

	if utils.IsFilled(fhirData.IdSatusehat) &&
		(!utils.IsFilled(db.IdSatusehat) || *db.IdSatusehat != *fhirData.IdSatusehat) {
		db.IdSatusehat = fhirData.IdSatusehat
		berubah = true
	}

	return berubah
}
