package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Anak struct {
	bun.BaseModel `bun:"table:stunting.anak,alias:a"`

	ID           string     `bun:"id,pk,type:varchar(36)" json:"id"`
	SatusehatId  string     `bun:"satusehat_id,type:varchar(36),nullzero,default:null" json:"satusehat_id"`
	IDOrangtua   string     `bun:"id_orangtua,type:varchar(36),nullzero" json:"id_orangtua"`
	Nama         string     `bun:"nama,type:varchar(255),notnull" json:"nama"`
	NIK          *string    `bun:"nik,type:varchar(16),nullzero,unique" json:"nik"`
	TanggalLahir time.Time  `bun:"tanggal_lahir,type:date,notnull" json:"tanggal_lahir"`
	JenisKelamin string     `bun:"jenis_kelamin,type:char(1),notnull" json:"jenis_kelamin"`
	AnakKe       int16      `bun:"anak_ke,type:smallint,notnull" json:"anak_ke"`
	IMD          *int16     `bun:"imd,type:smallint,notnull" json:"imd"`
	BBLahir      float32    `bun:"bb_lahir,type:real,notnull" json:"bb_lahir"`
	TBLahir      float32    `bun:"tb_lahir,type:real,notnull" json:"tb_lahir"`
	LKLahir      float32    `bun:"lk_lahir,type:real,notnull" json:"lk_lahir"`
	CreatedAt    time.Time  `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    *time.Time `bun:"updatedAt" json:"updated_at"`
	DeletedAt    *time.Time `bun:"deletedAt,soft_delete" json:"deleted_at,omitempty"`
	SourceData   string     `bun:"source_data,type:char(36),notnull" json:"source_data"`
	StatusAktif  *string    `bun:"status_aktif,type:varchar(20)" json:"status_aktif"`
	UpdatedBy    *string    `bun:"updated_by,type:varchar(100)" json:"updated_by"`
	DeletedBy    *string    `bun:"deleted_by,type:varchar(100)" json:"deleted_by"`

	Orangtua  *Orangtua    `bun:"rel:belongs-to,join:id_orangtua=id" json:"orangtua,omitempty"`
	Kunjungan []*Kunjungan `bun:"rel:has-many,join:id=id_anak" json:"kunjungan,omitempty"`
	Kesehatan []*Kesehatan `bun:"rel:has-many,join:id=id_anak" json:"kesehatan,omitempty"`
	Episode   []*Episode   `bun:"rel:has-many,join:id=id_anak" json:"episode,omitempty"`
	Observasi []*Observasi `bun:"rel:has-many,join:id=id_anak" json:"observasi,omitempty"`
	Diagnosa  []*Diagnosa  `bun:"rel:has-many,join:id=id_anak" json:"diagnosa,omitempty"`
	Layanan   []*Layanan   `bun:"rel:has-many,join:id=id_anak" json:"layanan,omitempty"`
	Rujukan   []*Rujukan   `bun:"rel:has-many,join:id=id_anak" json:"rujukan,omitempty"`
}

func (e Anak) TableName() string {
	return "stunting.anak"
}

func (e Anak) GetPkName() string {
	return "a.id"
}

func (m *Anak) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.UpdateQuery); ok {
		now := time.Now()
		m.UpdatedAt = &now
	}
	if _, ok := query.(*bun.InsertQuery); ok && m.IMD == nil {
		m.IMD = new(int16)
	}
	return nil
}

func (e *Anak) Override(new Anak, force bool) {
	skipFields := map[string]bool{
		"BaseModel":  true,
		"ID":         true,
		"CreatedAt":  true,
		"DeletedAt":  true,
		"DeletedBy":  true,
		"SourceData": true,
		"Orangtua":   true,
		"Kunjungan":  true,
		"Kesehatan":  true,
		"Episode":    true,
		"Observasi":  true,
		"Diagnosa":   true,
		"Layanan":    true,
		"Rujukan":    true,
	}

	timpaField(e, new, skipFields, force)
}
