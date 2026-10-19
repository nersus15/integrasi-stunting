package types

import (
	"strings"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
)

type Anak struct {
	Id           string     `json:"id"`
	IdSatusehat  *string    `json:"id_satusehat"`
	IdOrangtua   string     `json:"id_orangtua"`
	Nama         string     `json:"nama"`
	Nik          *string    `json:"nik"`
	TanggalLahir string     `json:"tanggal_lahir"`
	JenisKelamin string     `json:"jenis_kelamin"`
	AnakKe       int16      `json:"anak_ke"`
	Imd          int16      `json:"imd"`
	BbLahir      float32    `json:"bb_lahir"`
	TbLahir      float32    `json:"tb_lahir"`
	LkLahir      float32    `json:"lk_lahir"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
	SourceData   string     `json:"source_data"`
	StatusAktif  *string    `json:"status_aktif"`
	UpdatedBy    *string    `json:"updated_by"`
	DeletedBy    *string    `json:"deleted_by"`
}

type ListAnak struct {
	Orangtua
	Anak []Anak `json:"anak"`
}

type AnakPayload struct {
	Id           string     `json:"id"`
	SatusehatId  *string    `json:"satusehat_id"`
	IDOrangtua   string     `json:"id_orangtua"`
	Nama         string     `json:"nama"`
	NIK          *string    `json:"nik"`
	TanggalLahir string     `json:"tanggal_lahir"` // format YYYY-MM-DD
	JenisKelamin string     `json:"jenis_kelamin"`
	AnakKe       int16      `json:"anak_ke"`
	IMD          *int16     `json:"imd"`
	BBLahir      float32    `json:"bb_lahir"`
	TBLahir      float32    `json:"tb_lahir"`
	LKLahir      float32    `json:"lk_lahir"`
	SourceData   string     `json:"source_data"`
	StatusAktif  *string    `json:"status_aktif"`
	UpdatedBy    *string    `json:"updated_by"`
	DeletedBy    *string    `json:"deleted_by"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

func (a *Anak) FromEntity(e *entity.Anak) *Anak {
	if e == nil {
		return nil
	}

	var idSatusehat *string
	var nik *string
	if e.SatusehatId != "" {
		idSatusehat = &e.SatusehatId
	}

	if e.NIK != nil && strings.TrimSpace(*e.NIK) == "" {
		nik = nil
	} else {
		nik = e.NIK
	}

	return &Anak{
		Id:           e.ID,
		IdSatusehat:  idSatusehat,
		IdOrangtua:   e.IDOrangtua,
		Nama:         e.Nama,
		Nik:          nik,
		TanggalLahir: e.TanggalLahir.Format("2006-01-02"),
		JenisKelamin: e.JenisKelamin,
		AnakKe:       e.AnakKe,
		Imd:          nilaiInt16(e.IMD),
		BbLahir:      e.BBLahir,
		TbLahir:      e.TBLahir,
		LkLahir:      e.LKLahir,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
		DeletedAt:    e.DeletedAt,
		SourceData:   e.SourceData,
		StatusAktif:  e.StatusAktif,
		UpdatedBy:    e.UpdatedBy,
		DeletedBy:    e.DeletedBy,
	}
}

func (p *Anak) ToPayload() *AnakPayload {
	if p == nil {
		return nil
	}

	var nik *string
	if p.Nik != nil && strings.TrimSpace(*p.Nik) == "" {
		nik = nil
	} else {
		nik = p.Nik
	}

	e := &AnakPayload{
		Id:           p.Id,
		SatusehatId:  p.IdSatusehat,
		IDOrangtua:   p.IdOrangtua,
		Nama:         p.Nama,
		NIK:          nik,
		TanggalLahir: p.TanggalLahir,
		JenisKelamin: p.JenisKelamin,
		AnakKe:       p.AnakKe,
		IMD:          &p.Imd,
		BBLahir:      p.BbLahir,
		TBLahir:      p.TbLahir,
		LKLahir:      p.LkLahir,
		SourceData:   p.SourceData,
		StatusAktif:  p.StatusAktif,
		UpdatedBy:    p.UpdatedBy,
		DeletedBy:    p.DeletedBy,
		CreatedAt:    &p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		DeletedAt:    p.DeletedAt,
	}

	return e
}

func (p *AnakPayload) ToEntity() *entity.Anak {
	if p == nil {
		return nil
	}

	tLahir, _ := time.Parse("2006-01-02", p.TanggalLahir)
	var nik *string
	if p.NIK != nil && strings.TrimSpace(*p.NIK) == "" {
		nik = nil
	} else {
		nik = p.NIK
	}

	// jakantro tidak mengirim satusehat_id, kolomnya ditulis null
	satusehat := ""
	if p.SatusehatId != nil {
		satusehat = strings.TrimSpace(*p.SatusehatId)
	}

	e := &entity.Anak{
		ID:           p.Id,
		SatusehatId:  satusehat,
		IDOrangtua:   p.IDOrangtua,
		Nama:         p.Nama,
		NIK:          nik,
		TanggalLahir: tLahir,
		JenisKelamin: p.JenisKelamin,
		AnakKe:       p.AnakKe,
		IMD:          p.IMD,
		BBLahir:      p.BBLahir,
		TBLahir:      p.TBLahir,
		LKLahir:      p.LKLahir,
		SourceData:   p.SourceData,
		StatusAktif:  p.StatusAktif,
		UpdatedBy:    p.UpdatedBy,
		DeletedBy:    p.DeletedBy,
	}

	// timestamp dari jakantro; kalau tidak dikirim, database yang mengisi
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}
	e.UpdatedAt = p.UpdatedAt
	e.DeletedAt = p.DeletedAt

	return e
}

func (la *ListAnak) FromEntity(e *entity.Orangtua) *ListAnak {
	if e == nil {
		return nil
	}

	anak := make([]Anak, 0)

	for _, a := range e.Anak {
		var temp *Anak
		if an := temp.FromEntity(a); an != nil {
			anak = append(anak, *an)
		}

	}
	var temp *Orangtua
	orangtua := temp.FromEntity(e)

	return &ListAnak{
		Orangtua: *orangtua,
		Anak:     anak,
	}
}
