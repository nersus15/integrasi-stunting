package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Kesehatan struct {
	bun.BaseModel `bun:"table:stunting.kesehatan,alias:k"`

	ID                string     `bun:"id,pk,type:varchar(36)" json:"id"`
	IDAnak            string     `bun:"id_anak,type:varchar(36),notnull" json:"id_anak"`
	TanggalPemantauan time.Time  `bun:"tanggal_pemantauan,type:date,notnull" json:"tanggal_pemantauan"`
	TBCBatuk          *int16     `bun:"tbc_batuk,type:smallint" json:"tbc_batuk"`
	TBCDemam          *int16     `bun:"tbc_demam,type:smallint" json:"tbc_demam"`
	TBCBB             *int16     `bun:"tbc_bb,type:smallint" json:"tbc_bb"`
	TBCKontak         *int16     `bun:"tbc_kontak,type:smallint" json:"tbc_kontak"`
	LayananASIEks     *int16     `bun:"layanan_asi_eks,type:smallint" json:"layanan_asi_eks"`
	LayananMPASI      *int16     `bun:"layanan_mpasi,type:smallint" json:"layanan_mpasi"`
	LayananImunisasi  *int16     `bun:"layanan_imunisasi,type:smallint" json:"layanan_imunisasi"`
	LayananVitA       *int16     `bun:"layanan_vit_a,type:smallint" json:"layanan_vit_a"`
	LayananObatCacing *int16     `bun:"layanan_obat_cacing,type:smallint" json:"layanan_obat_cacing"`
	LayananMTPangan   *int16     `bun:"layanan_mt_pangan,type:smallint" json:"layanan_mt_pangan"`
	PenyuluhanEdukasi *int16     `bun:"penyuluhan_edukasi,type:smallint" json:"penyuluhan_edukasi"`
	PenyuluhanRujukan *int16     `bun:"penyuluhan_rujukan,type:smallint" json:"penyuluhan_rujukan"`
	CreatedAt         time.Time  `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt         *time.Time `bun:"updatedAt" json:"updated_at"`
	DeletedAt         *time.Time `bun:"deletedAt,soft_delete" json:"deleted_at,omitempty"`
	SourceData        *string    `bun:"source_data,type:varchar(50)" json:"source_data"`
	CreatedBy         *string    `bun:"created_by,type:varchar(100)" json:"created_by"`
	UpdatedBy         *string    `bun:"updated_by,type:varchar(100)" json:"updated_by"`
	DeletedBy         *string    `bun:"deleted_by,type:varchar(100)" json:"deleted_by"`

	Anak *Anak `bun:"rel:belongs-to,join:id_anak=id" json:"anak,omitempty"`
}

func (e Kesehatan) TableName() string {
	return "stunting.kesehatan"
}

func (e Kesehatan) GetPkName() string {
	return "k.id"
}

func (m *Kesehatan) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.UpdateQuery); ok {
		now := time.Now()
		m.UpdatedAt = &now
	}
	return nil
}

func (e *Kesehatan) Override(new Kesehatan, force bool) {
	skipFields := map[string]bool{
		"BaseModel":  true,
		"ID":         true,
		"IDAnak":     true,
		"CreatedAt":  true,
		"DeletedAt":  true,
		"DeletedBy":  true,
		"CreatedBy":  true,
		"SourceData": true,
		"Anak":       true,
	}

	timpaField(e, new, skipFields, force)
}
