package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Kunjungan struct {
	bun.BaseModel `bun:"table:stunting.kunjungan,alias:kj"`

	ID                string     `bun:"id,pk,type:varchar(36)" json:"id"`
	IDAnak            string     `bun:"id_anak,type:varchar(36),notnull" json:"id_anak"`
	TanggalPengukuran time.Time  `bun:"tanggal_pengukuran,type:date,notnull" json:"tanggal_pengukuran"`
	TanggalSelesai    *time.Time `bun:"tanggal_selesai,type:date,nullzero" json:"tanggal_selesai"`
	CaraUkur          *string    `bun:"cara_ukur,type:varchar(50)" json:"cara_ukur"`
	BeratBadan        *float64   `bun:"berat_badan,type:double precision" json:"berat_badan"`
	TinggiBadan       *float64   `bun:"tinggi_badan,type:double precision" json:"tinggi_badan"`
	LingkarLengan     *float64   `bun:"lingkar_lengan,type:double precision" json:"lingkar_lengan"`
	LingkarKepala     *float64   `bun:"lingkar_kepala,type:double precision" json:"lingkar_kepala"`
	LingkarDada       *float64   `bun:"lingkar_dada,type:double precision" json:"lingkar_dada"`
	ASIBulan0         *int16     `bun:"asi_bulan_0,type:smallint" json:"asi_bulan_0"`
	ASIBulan1         *int16     `bun:"asi_bulan_1,type:smallint" json:"asi_bulan_1"`
	ASIBulan2         *int16     `bun:"asi_bulan_2,type:smallint" json:"asi_bulan_2"`
	ASIBulan3         *int16     `bun:"asi_bulan_3,type:smallint" json:"asi_bulan_3"`
	ASIBulan4         *int16     `bun:"asi_bulan_4,type:smallint" json:"asi_bulan_4"`
	ASIBulan5         *int16     `bun:"asi_bulan_5,type:smallint" json:"asi_bulan_5"`
	ASIBulan6         *int16     `bun:"asi_bulan_6,type:smallint" json:"asi_bulan_6"`
	VitBiru           *int16     `bun:"vit_biru,type:smallint" json:"vit_biru"`
	VitMerah          *int16     `bun:"vit_merah,type:smallint" json:"vit_merah"`
	PittingEdema      *int16     `bun:"pitting_edema,type:smallint" json:"pitting_edema"`
	KelasIbuBalita    *int16     `bun:"kelas_ibu_balita,type:smallint" json:"kelas_ibu_balita"`
	StatusBBU         *string    `bun:"status_bbu,type:varchar(255)" json:"status_bbu"`
	StatusTBU         *string    `bun:"status_tbu,type:varchar(255)" json:"status_tbu"`
	StatusBBTB        *string    `bun:"status_bbtb,type:varchar(255)" json:"status_bbtb"`
	ZScoreBBU         *float64   `bun:"zscore_bbu,type:numeric(5,2)" json:"zscore_bbu"`
	ZScoreTBU         *float64   `bun:"zscore_tbu,type:numeric(5,2)" json:"zscore_tbu"`
	ZScoreBBTB        *float64   `bun:"zscore_bbtb,type:numeric(5,2)" json:"zscore_bbtb"`
	CreatedAt         time.Time  `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt         *time.Time `bun:"updatedAt" json:"updated_at"`
	DeletedAt         *time.Time `bun:"deletedAt,soft_delete" json:"deleted_at,omitempty"`
	SourceData        *string    `bun:"source_data,type:char(36)" json:"source_data"`
	UpdatedBy         *string    `bun:"updated_by,type:varchar(100)" json:"updated_by"`
	DeletedBy         *string    `bun:"deleted_by,type:varchar(100)" json:"deleted_by"`
	IDFaskes          *string    `bun:"id_faskes,type:varchar(36),nullzero" json:"id_faskes"`
	SatusehatId       *string    `bun:"satusehat_id,type:varchar(36),nullzero" json:"satusehat_id"`
	IDEpisode         *string    `bun:"id_episode,type:varchar(36),nullzero" json:"id_episode"`
	RefEpisode        *string    `bun:"ref_episode,type:varchar(36),nullzero" json:"ref_episode"`
	IDRujukan         *string    `bun:"id_rujukan,type:varchar(36),nullzero" json:"id_rujukan"`
	RefRujukan        *string    `bun:"ref_rujukan,type:varchar(36),nullzero" json:"ref_rujukan"`
	Stunting          *int       `bun:"stunting,type:smallint,nullzero" json:"stunting"`

	Anak    *Anak    `bun:"rel:belongs-to,join:id_anak=id" json:"anak,omitempty"`
	Faskes  *Faskes  `bun:"rel:belongs-to,join:id_faskes=id" json:"faskes,omitempty"`
	Episode *Episode `bun:"rel:belongs-to,join:id_episode=id" json:"episode,omitempty"`
	// rujukan yang dipenuhi kunjungan ini, beda dari Rujukan yang diterbitkan di sini
	AtasRujukan *Rujukan     `bun:"rel:belongs-to,join:id_rujukan=id" json:"atas_rujukan,omitempty"`
	Diagnosa    []*Diagnosa  `bun:"rel:has-many,join:id=id_kunjungan" json:"diagnosa,omitempty"`
	Observasi   []*Observasi `bun:"rel:has-many,join:id=id_kunjungan" json:"observasi,omitempty"`
	Layanan     []*Layanan   `bun:"rel:has-many,join:id=id_kunjungan" json:"layanan,omitempty"`
	Rujukan     []*Rujukan   `bun:"rel:has-many,join:id=id_kunjungan" json:"rujukan,omitempty"`
}

type Encounter struct {
	SatusehatId string `bun:"satusehat_id,pk,type:char(36),default:null" json:"satusehat_id"`
	Anak        string `bun:"anak,type:char(36),default:null" json:"anak"`
}

func (e Kunjungan) TableName() string {
	return "stunting.kunjungan"
}

func (e Kunjungan) GetPkName() string {
	return "kj.id"
}

func (m *Kunjungan) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.UpdateQuery); ok {
		now := time.Now()
		m.UpdatedAt = &now
	}
	return nil
}

func (e *Kunjungan) Override(new Kunjungan, force bool) {
	skipFields := map[string]bool{
		"BaseModel":   true,
		"ID":          true,
		"SatusehatId": true,
		"IDAnak":      true,
		"IDFaskes":    true,
		"IDEpisode":   true,
		"IDRujukan":   true,
		"CreatedAt":   true,
		"DeletedAt":   true,
		"DeletedBy":   true,
		"SourceData":  true,
		"Anak":        true,
		"Faskes":      true,
		"Episode":     true,
		"Rujukan":     true,
		"AtasRujukan": true,
		"Observasi":   true,
		"Diagnosa":    true,
		"Layanan":     true,
		"RefEpisode":  true,
		"RefRujukan":  true,
		"Stunting":    true,
		"UpdatedAt":   true,
	}

	timpaField(e, new, skipFields, force)
}
