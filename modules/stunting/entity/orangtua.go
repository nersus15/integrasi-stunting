package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Orangtua struct {
	bun.BaseModel `bun:"table:stunting.orangtua,alias:o"`

	ID           string     `bun:"id,pk,type:varchar(36)" json:"id"`
	SatusehatId  string     `bun:"satusehat_id,type:varchar(36),nullzero,default:null" json:"id_satusehat"`
	IDPosyandu   *string    `bun:"id_posyandu,type:varchar(36),nullzero" json:"id_posyandu"`
	NoKK         string     `bun:"no_kk,type:varchar(16),nullzero,unique" json:"no_kk"`
	NamaAyah     string     `bun:"nama_ayah,type:varchar(255),notnull" json:"nama_ayah"`
	NamaIbu      string     `bun:"nama_ibu,type:varchar(255),notnull" json:"nama_ibu"`
	NIK          string     `bun:"nik,type:varchar(16),notnull,unique" json:"nik"`
	Telepon      string     `bun:"telepon,type:varchar(255),notnull" json:"telepon"`
	RT           string     `bun:"rt,type:char(3),notnull" json:"rt"`
	RW           string     `bun:"rw,type:char(3),notnull" json:"rw"`
	Alamat       string     `bun:"alamat,type:text,notnull" json:"alamat"`
	KIA          *int16     `bun:"kia,type:smallint,notnull" json:"kia"`
	CreatedAt    time.Time  `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    *time.Time `bun:"updatedAt" json:"updated_at"`
	DeletedAt    *time.Time `bun:"deletedAt,soft_delete" json:"deleted_at,omitempty"`
	SourceData   *string    `bun:"source_data,type:varchar(50)" json:"source_data"`
	UsiaHamil    *int16     `bun:"usia_hamil,type:smallint" json:"usia_hamil"`
	KIABayiKecil *int16     `bun:"kia_bayi_kecil,type:smallint" json:"kia_bayi_kecil"`
	CreatedBy    *string    `bun:"created_by,type:varchar(100)" json:"created_by"`
	UpdatedBy    *string    `bun:"updated_by,type:varchar(100)" json:"updated_by"`
	DeletedBy    *string    `bun:"deleted_by,type:varchar(100)" json:"deleted_by"`

	Anak []*Anak `bun:"rel:has-many,join:id=id_orangtua" json:"anak,omitempty"`
}

func (e Orangtua) TableName() string {
	return "stunting.orangtua"
}

func (e Orangtua) GetPkName() string {
	return "o.id"
}

func (m *Orangtua) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.UpdateQuery); ok {
		now := time.Now()
		m.UpdatedAt = &now
	}
	if _, ok := query.(*bun.InsertQuery); ok && m.KIA == nil {
		m.KIA = new(int16)
	}
	return nil
}

func (e *Orangtua) Override(new Orangtua, force bool) {
	skipFields := map[string]bool{
		"BaseModel":  true,
		"ID":         true,
		"CreatedAt":  true,
		"DeletedAt":  true,
		"DeletedBy":  true,
		"CreatedBy":  true,
		"SourceData": true,
		"Anak":       true,
	}

	timpaField(e, new, skipFields, force)
}
