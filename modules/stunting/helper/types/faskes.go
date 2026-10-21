package types

import (
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
)

type Faskes struct {
	Id          string     `json:"id"`
	IdInduk     *string    `json:"id_induk"`
	SatusehatId *string    `json:"id_satusehat"`
	Nama        string     `json:"nama"`
	Jenis       string     `json:"jenis"`
	Wilayah     *string    `json:"wilayah"`
	Alamat      *string    `json:"alamat"`
	NomorTelpon *string    `json:"nomor_telpon"`
	Email       *string    `json:"email"`
	Status      *int16     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type FaskesDetail struct {
	Id          string     `json:"id"`
	Induk       *Faskes    `json:"induk"`
	SatusehatId *string    `json:"id_satusehat"`
	Nama        string     `json:"nama"`
	Jenis       string     `json:"jenis"`
	Wilayah     *string    `json:"wilayah"`
	Alamat      *string    `json:"alamat"`
	NomorTelpon *string    `json:"nomor_telpon"`
	Email       *string    `json:"email"`
	Status      *int16     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type Posyandu struct {
	Id          string     `json:"id"`
	IdPuskesmas *string    `json:"id_puskesmas"`
	Nama        string     `json:"nama"`
	Telepon     *string    `json:"telepon"`
	Alamat      *string    `json:"alamat"`
	IdKelurahan *string    `json:"id_kelurahan"`
	Rt          string     `json:"rt"`
	Rw          string     `json:"rw"`
	NamaPic     *string    `json:"nama_pic"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

// faskes beserta posyandu di bawahnya, field faskes di-flatten ke root
type ListPosyandu struct {
	Faskes
	Posyandu []Posyandu `json:"posyandu"`
}

// posyandu beserta puskesmas induknya, puskesmas bisa nil
type PosyanduDetail struct {
	Posyandu
	Puskesmas *FaskesDetail `json:"puskesmas"`
}

type FaskesPayload struct {
	Id          string     `json:"id"`
	IdInduk     *string    `json:"id_induk"`
	SatusehatId *string    `json:"id_satusehat"`
	Nama        string     `json:"nama"`
	Jenis       string     `json:"jenis"`
	Wilayah     *string    `json:"wilayah"`
	Alamat      *string    `json:"alamat"`
	NomorTelpon *string    `json:"nomor_telpon"`
	Email       *string    `json:"email"`
	Status      *int16     `json:"status"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type PosyanduPayload struct {
	Id          string     `json:"id"`
	IdPuskesmas *string    `json:"id_puskesmas"`
	Nama        string     `json:"nama"`
	Telepon     *string    `json:"telepon"`
	Alamat      *string    `json:"alamat"`
	IdKelurahan *string    `json:"id_kelurahan"`
	Rt          string     `json:"rt"`
	Rw          string     `json:"rw"`
	NamaPic     *string    `json:"nama_pic"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

func (f *Faskes) FromEntity(e *entity.Faskes) *Faskes {
	if e == nil {
		return nil
	}

	return &Faskes{
		Id:          e.ID,
		IdInduk:     e.IDInduk,
		SatusehatId: e.SatusehatID,
		Nama:        e.Nama,
		Jenis:       e.Jenis,
		Wilayah:     e.Wilayah,
		Alamat:      e.Alamat,
		NomorTelpon: e.NomorTelpon,
		Email:       e.Email,
		Status:      e.Status,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func (p *FaskesPayload) ToEntity() *entity.Faskes {
	if p == nil {
		return nil
	}

	e := &entity.Faskes{
		ID:          p.Id,
		IDInduk:     p.IdInduk,
		SatusehatID: p.SatusehatId,
		Nama:        p.Nama,
		Jenis:       p.Jenis,
		Wilayah:     p.Wilayah,
		Alamat:      p.Alamat,
		NomorTelpon: p.NomorTelpon,
		Email:       p.Email,
		Status:      p.Status,
		UpdatedAt:   p.UpdatedAt,
	}

	// Kalau tidak dikirim, database yang mengisi lewat default.
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}

	return e
}

func (f *FaskesDetail) FromEntity(e *entity.Faskes) *FaskesDetail {
	if e == nil {
		return nil
	}
	var induk *Faskes
	return &FaskesDetail{
		Id:          e.ID,
		Induk:       induk.FromEntity(e.Induk),
		SatusehatId: e.SatusehatID,
		Nama:        e.Nama,
		Jenis:       e.Jenis,
		Wilayah:     e.Wilayah,
		Alamat:      e.Alamat,
		NomorTelpon: e.NomorTelpon,
		Email:       e.Email,
		Status:      e.Status,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func (p *Posyandu) FromEntity(e *entity.Posyandu) *Posyandu {
	if e == nil {
		return nil
	}

	return &Posyandu{
		Id:          e.ID,
		IdPuskesmas: e.IDPuskesmas,
		Nama:        e.Nama,
		Telepon:     e.Telepon,
		Alamat:      e.Alamat,
		IdKelurahan: e.IDKelurahan,
		Rt:          e.RT,
		Rw:          e.RW,
		NamaPic:     e.NamaPic,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
		DeletedAt:   e.DeletedAt,
	}
}

func (p *PosyanduPayload) ToEntity() *entity.Posyandu {
	if p == nil {
		return nil
	}

	e := &entity.Posyandu{
		ID:          p.Id,
		IDPuskesmas: p.IdPuskesmas,
		Nama:        p.Nama,
		Telepon:     p.Telepon,
		Alamat:      p.Alamat,
		IDKelurahan: p.IdKelurahan,
		RT:          p.Rt,
		RW:          p.Rw,
		NamaPic:     p.NamaPic,
		UpdatedAt:   p.UpdatedAt,
		DeletedAt:   p.DeletedAt,
	}

	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}

	return e
}

func (l *ListPosyandu) FromEntity(e *entity.Faskes) *ListPosyandu {
	if e == nil {
		return nil
	}

	var f *Faskes
	return &ListPosyandu{
		Faskes:   *f.FromEntity(e),
		Posyandu: daftarPosyandu(e.Posyandu),
	}
}

func (d *PosyanduDetail) FromEntity(e *entity.Posyandu) *PosyanduDetail {
	if e == nil {
		return nil
	}

	var p *Posyandu
	var f *FaskesDetail
	return &PosyanduDetail{
		Posyandu:  *p.FromEntity(e),
		Puskesmas: f.FromEntity(e.Puskesmas),
	}
}

// selalu tidak-nil supaya JSON-nya [] bukan null
func daftarPosyandu(list []*entity.Posyandu) []Posyandu {
	return daftar(list, (*Posyandu)(nil).FromEntity)
}
