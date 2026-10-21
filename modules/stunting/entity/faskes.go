package entity

import (
	"time"

	"github.com/uptrace/bun"
)

// created_at/updated_at snake_case, ikut DDL. Hapus diwakili status = 0.
type Faskes struct {
	bun.BaseModel `bun:"table:stunting.faskes,alias:f"`

	ID          string     `bun:"id,pk,type:varchar(36)" json:"id"`
	IDInduk     *string    `bun:"id_induk,type:varchar(36)" json:"id_induk"`
	SatusehatID *string    `bun:"satusehat_id,type:varchar(36),nullzero,unique" json:"id_satusehat"`
	Nama        string     `bun:"nama,type:varchar(100),notnull" json:"nama"`
	Jenis       string     `bun:"jenis,type:varchar(36),notnull" json:"jenis"`
	Wilayah     *string    `bun:"wilayah,type:varchar(13)" json:"wilayah"`
	Alamat      *string    `bun:"alamat,type:text" json:"alamat"`
	NomorTelpon *string    `bun:"nomor_telpon,type:varchar(36)" json:"nomor_telpon"`
	Email       *string    `bun:"email,type:varchar(72)" json:"email"`
	Status      *int16     `bun:"status,type:smallint" json:"status"`
	CreatedAt   time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   *time.Time `bun:"updated_at" json:"updated_at"`

	// tanpa foreign key di DDL, relasi ini hanya untuk pembacaan
	Induk *Faskes `bun:"rel:belongs-to,join:id_induk=id" json:"induk,omitempty"`

	Posyandu []*Posyandu `bun:"rel:has-many,join:id=id_puskesmas" json:"posyandu,omitempty"`
}

type Posyandu struct {
	bun.BaseModel `bun:"table:stunting.posyandu,alias:p"`

	ID          string     `bun:"id,pk,type:varchar(36)" json:"id"`
	IDPuskesmas *string    `bun:"id_puskesmas,type:varchar(36)" json:"id_puskesmas"`
	Nama        string     `bun:"nama,type:varchar(255),notnull" json:"nama"`
	Telepon     *string    `bun:"telepon,type:varchar(255)" json:"telepon"`
	Alamat      *string    `bun:"alamat,type:varchar(255)" json:"alamat"`
	IDKelurahan *string    `bun:"id_kelurahan,type:varchar(36)" json:"id_kelurahan"`
	RT          string     `bun:"rt,type:varchar(3),notnull" json:"rt"`
	RW          string     `bun:"rw,type:varchar(3),notnull" json:"rw"`
	NamaPic     *string    `bun:"nama_pic,type:varchar(100)" json:"nama_pic"`
	CreatedAt   time.Time  `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   *time.Time `bun:"updatedAt" json:"updated_at"`
	DeletedAt   *time.Time `bun:"deletedAt,soft_delete" json:"deleted_at,omitempty"`

	Puskesmas *Faskes `bun:"rel:belongs-to,join:id_puskesmas=id" json:"puskesmas,omitempty"`
}

func (e Faskes) TableName() string {
	return "stunting.faskes"
}

func (e Faskes) GetPkName() string {
	return "f.id"
}

func (e Posyandu) TableName() string {
	return "stunting.posyandu"
}

func (e Posyandu) GetPkName() string {
	return "p.id"
}
