package types

import (
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
)

type Orangtua struct {
	Id           string     `json:"id"`
	IdSatusehat  *string    `json:"id_satusehat"`
	IdPosyandu   *string    `json:"id_posyandu"`
	NoKk         string     `json:"no_kk"`
	NamaAyah     string     `json:"nama_ayah"`
	NamaIbu      string     `json:"nama_ibu"`
	Nik          string     `json:"nik"`
	Telepon      string     `json:"telepon"`
	Rt           string     `json:"rt"`
	Rw           string     `json:"rw"`
	Alamat       string     `json:"alamat"`
	Kia          int16      `json:"kia"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
	SourceData   *string    `json:"source_data"`
	CreatedBy    *string    `json:"created_by"`
	UsiaHamil    *int16     `json:"usia_hamil"`
	KiaBayiKecil *int16     `json:"kia_bayi_kecil"`
	UpdatedBy    *string    `json:"updated_by"`
	DeletedBy    *string    `json:"deleted_by"`
}

type OrangtuaPayload struct {
	Id           string     `json:"id"`
	SatusehatId  *string    `json:"id_satusehat"`
	IdPosyandu   string     `json:"id_posyandu"`
	NoKk         string     `json:"no_kk"`
	NamaAyah     string     `json:"nama_ayah"`
	NamaIbu      string     `json:"nama_ibu"`
	Nik          string     `json:"nik"`
	Telepon      string     `json:"telepon"`
	Rt           string     `json:"rt"`
	Rw           string     `json:"rw"`
	Alamat       string     `json:"alamat"`
	Kia          *int16     `json:"kia"`
	SourceData   *string    `json:"source_data"`
	CreatedBy    *string    `json:"created_by"`
	UsiaHamil    *int16     `json:"usia_hamil"`
	KiaBayiKecil *int16     `json:"kia_bayi_kecil"`
	UpdatedBy    *string    `json:"updated_by"`
	DeletedBy    *string    `json:"deleted_by"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

func (o *Orangtua) FromEntity(e *entity.Orangtua) *Orangtua {
	if e == nil {
		return nil
	}

	// null untuk anak yang didaftarkan jakantro
	var idSatusehat *string
	if e.SatusehatId != "" {
		idSatusehat = &e.SatusehatId
	}

	return &Orangtua{
		Id:           e.ID,
		IdSatusehat:  idSatusehat,
		IdPosyandu:   e.IDPosyandu,
		NoKk:         e.NoKK,
		NamaAyah:     e.NamaAyah,
		NamaIbu:      e.NamaIbu,
		Nik:          e.NIK,
		Telepon:      e.Telepon,
		Rt:           e.RT,
		Rw:           e.RW,
		Alamat:       e.Alamat,
		Kia:          nilaiInt16(e.KIA),
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
		DeletedAt:    e.DeletedAt,
		SourceData:   e.SourceData,
		CreatedBy:    e.CreatedBy,
		UsiaHamil:    e.UsiaHamil,
		KiaBayiKecil: e.KIABayiKecil,
		UpdatedBy:    e.UpdatedBy,
		DeletedBy:    e.DeletedBy,
	}
}

func (p *Orangtua) ToPayload() *OrangtuaPayload {
	if p == nil {
		return nil
	}

	idPosyandu := ""
	if p.IdPosyandu != nil {
		idPosyandu = *p.IdPosyandu
	}

	return &OrangtuaPayload{
		SatusehatId:  p.IdSatusehat,
		Id:           p.Id,
		IdPosyandu:   idPosyandu,
		NoKk:         p.NoKk,
		NamaAyah:     p.NamaAyah,
		NamaIbu:      p.NamaIbu,
		Nik:          p.Nik,
		Telepon:      p.Telepon,
		Rt:           p.Rt,
		Rw:           p.Rw,
		Alamat:       p.Alamat,
		Kia:          &p.Kia,
		SourceData:   p.SourceData,
		CreatedBy:    p.CreatedBy,
		UsiaHamil:    p.UsiaHamil,
		KiaBayiKecil: p.KiaBayiKecil,
		UpdatedBy:    p.UpdatedBy,
		DeletedBy:    p.DeletedBy,
		CreatedAt:    &p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		DeletedAt:    p.DeletedAt,
	}
}

func (p *OrangtuaPayload) ToEntity() *entity.Orangtua {
	if p == nil {
		return nil
	}

	var idPosyandu *string
	if p.IdPosyandu != "" {
		idPosyandu = &p.IdPosyandu
	}

	e := &entity.Orangtua{
		ID:           p.Id,
		IDPosyandu:   idPosyandu,
		NoKK:         p.NoKk,
		NamaAyah:     p.NamaAyah,
		NamaIbu:      p.NamaIbu,
		NIK:          p.Nik,
		Telepon:      p.Telepon,
		RT:           p.Rt,
		RW:           p.Rw,
		Alamat:       p.Alamat,
		KIA:          p.Kia,
		SourceData:   p.SourceData,
		CreatedBy:    p.CreatedBy,
		UsiaHamil:    p.UsiaHamil,
		KIABayiKecil: p.KiaBayiKecil,
		UpdatedBy:    p.UpdatedBy,
		DeletedBy:    p.DeletedBy,
	}

	if p.SatusehatId != nil {
		e.SatusehatId = *p.SatusehatId
	}

	// timestamp dari jakantro; kalau tidak dikirim, database yang mengisi
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}
	e.UpdatedAt = p.UpdatedAt
	e.DeletedAt = p.DeletedAt

	return e
}
