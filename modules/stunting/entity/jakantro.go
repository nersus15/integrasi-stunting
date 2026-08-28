package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Entity interface {
	TableName() string
	GetID() string
	GetPkName() string
}

type Orangtua struct {
	bun.BaseModel `bun:"table:jakantro.orangtua,alias:o"`

	ID           string     `bun:"id,pk,type:char(36)" json:"id"`
	SatusehatId  string     `bun:"satusehat_id,type:char(36),default:null" json:"satusehat_id"`
	IDPosyandu   string     `bun:"id_posyandu,type:char(36),notnull" json:"id_posyandu"`
	NoKK         string     `bun:"no_kk,type:varchar(16),notnull,unique" json:"no_kk"`
	NamaAyah     string     `bun:"nama_ayah,type:varchar(255),notnull" json:"nama_ayah"`
	NamaIbu      string     `bun:"nama_ibu,type:varchar(255),notnull" json:"nama_ibu"`
	NIK          string     `bun:"nik,type:varchar(16),notnull,unique" json:"nik"`
	Telepon      string     `bun:"telepon,type:varchar(255),notnull" json:"telepon"`
	RT           string     `bun:"rt,type:char(3),notnull" json:"rt"`
	RW           string     `bun:"rw,type:char(3),notnull" json:"rw"`
	Alamat       string     `bun:"alamat,type:text,notnull" json:"alamat"`
	KIA          int16      `bun:"kia,type:smallint,notnull" json:"kia"`
	CreatedAt    time.Time  `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    *time.Time `bun:"updatedAt" json:"updated_at"`
	DeletedAt    *time.Time `bun:"deletedAt,soft_delete" json:"deleted_at,omitempty"`
	SourceData   *string    `bun:"source_data,type:char(36)" json:"source_data"`
	UsiaHamil    *int16     `bun:"usia_hamil,type:smallint" json:"usia_hamil"`
	KIABayiKecil *int16     `bun:"kia_bayi_kecil,type:smallint" json:"kia_bayi_kecil"`
	UpdatedBy    *string    `bun:"updated_by,type:varchar(100)" json:"updated_by"`
	DeletedBy    *string    `bun:"deleted_by,type:varchar(100)" json:"deleted_by"`

	Anak []*Anak `bun:"rel:has-many,join:id=id_orangtua" json:"anak,omitempty"`
}

type Anak struct {
	bun.BaseModel `bun:"table:jakantro.anak,alias:a"`

	ID          string `bun:"id,pk,type:char(36)" json:"id"`
	SatusehatId string `bun:"satusehat_id,type:char(36),default:null" json:"satusehat_id"`
	IDOrangtua  string `bun:"id_orangtua,type:char(36),notnull" json:"id_orangtua"`
	Nama        string `bun:"nama,type:varchar(255),notnull" json:"nama"`
	// nik anak opsional: nullzero membuat string kosong tersimpan sebagai NULL,
	// sehingga banyak anak tanpa nik tidak saling bentrok di UNIQUE
	NIK          *string    `bun:"nik,type:varchar(16),nullzero,unique" json:"nik"`
	TanggalLahir time.Time  `bun:"tanggal_lahir,type:date,notnull" json:"tanggal_lahir"`
	JenisKelamin string     `bun:"jenis_kelamin,type:char(1),notnull" json:"jenis_kelamin"`
	AnakKe       int16      `bun:"anak_ke,type:smallint,notnull" json:"anak_ke"`
	IMD          int16      `bun:"imd,type:smallint,notnull" json:"imd"`
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
}

type Kesehatan struct {
	bun.BaseModel `bun:"table:jakantro.kesehatan,alias:k"`

	ID                string     `bun:"id,pk,type:char(36)" json:"id"`
	IDAnak            string     `bun:"id_anak,type:char(36),notnull" json:"id_anak"`
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
	SourceData        *string    `bun:"source_data,type:char(36)" json:"source_data"`

	Anak *Anak `bun:"rel:belongs-to,join:id_anak=id" json:"anak,omitempty"`
}

type Kunjungan struct {
	bun.BaseModel `bun:"table:jakantro.kunjungan,alias:kj"`

	ID                 string     `bun:"id,pk,type:char(36)" json:"id"`
	IDAnak             string     `bun:"id_anak,type:char(36),notnull" json:"id_anak"`
	TanggalPengukuran  time.Time  `bun:"tanggal_pengukuran,type:date,notnull" json:"tanggal_pengukuran"`
	CaraUkur           *string    `bun:"cara_ukur,type:varchar(50)" json:"cara_ukur"`
	BeratBadan         *float64   `bun:"berat_badan,type:double precision" json:"berat_badan"`
	TinggiBadan        *float64   `bun:"tinggi_badan,type:double precision" json:"tinggi_badan"`
	LingkarLengan      *float64   `bun:"lingkar_lengan,type:double precision" json:"lingkar_lengan"`
	LingkarKepala      *float64   `bun:"lingkar_kepala,type:double precision" json:"lingkar_kepala"`
	LingkarDada        *float64   `bun:"lingkar_dada,type:double precision" json:"lingkar_dada"`
	ASIBulan0          *int16     `bun:"asi_bulan_0,type:smallint" json:"asi_bulan_0"`
	ASIBulan1          *int16     `bun:"asi_bulan_1,type:smallint" json:"asi_bulan_1"`
	ASIBulan2          *int16     `bun:"asi_bulan_2,type:smallint" json:"asi_bulan_2"`
	ASIBulan3          *int16     `bun:"asi_bulan_3,type:smallint" json:"asi_bulan_3"`
	ASIBulan4          *int16     `bun:"asi_bulan_4,type:smallint" json:"asi_bulan_4"`
	ASIBulan5          *int16     `bun:"asi_bulan_5,type:smallint" json:"asi_bulan_5"`
	ASIBulan6          *int16     `bun:"asi_bulan_6,type:smallint" json:"asi_bulan_6"`
	VitBiru            *int16     `bun:"vit_biru,type:smallint" json:"vit_biru"`
	VitMerah           *int16     `bun:"vit_merah,type:smallint" json:"vit_merah"`
	PittingEdema       *int16     `bun:"pitting_edema,type:smallint" json:"pitting_edema"`
	KelasIbuBalita     *int16     `bun:"kelas_ibu_balita,type:smallint" json:"kelas_ibu_balita"`
	StatusBBUSigizi    *string    `bun:"status_bbu_sigizi,type:varchar(255)" json:"status_bbu_sigizi"`
	StatusTBUSigizi    *string    `bun:"status_tbu_sigizi,type:varchar(255)" json:"status_tbu_sigizi"`
	StatusBBTBSigizi   *string    `bun:"status_bbtb_sigizi,type:varchar(255)" json:"status_bbtb_sigizi"`
	ZScoreBBUSigizi    *float64   `bun:"zscore_bbu_sigizi,type:numeric(5,2)" json:"zscore_bbu_sigizi"`
	ZScoreTBUSigizi    *float64   `bun:"zscore_tbu_sigizi,type:numeric(5,2)" json:"zscore_tbu_sigizi"`
	ZScoreBBTBSigizi   *float64   `bun:"zscore_bbtb_sigizi,type:numeric(5,2)" json:"zscore_bbtb_sigizi"`
	StatusBBUWhoAntro  *string    `bun:"status_bbu_whoantro,type:varchar(255)" json:"status_bbu_whoantro"`
	StatusTBUWhoAntro  *string    `bun:"status_tbu_whoantro,type:varchar(255)" json:"status_tbu_whoantro"`
	StatusBBTBWhoAntro *string    `bun:"status_bbtb_whoantro,type:varchar(255)" json:"status_bbtb_whoantro"`
	ZScoreBBUWhoAntro  *float64   `bun:"zscore_bbu_whoantro,type:numeric(5,2)" json:"zscore_bbu_whoantro"`
	ZScoreTBUWhoAntro  *float64   `bun:"zscore_tbu_whoantro,type:numeric(5,2)" json:"zscore_tbu_whoantro"`
	ZScoreBBTBWhoAntro *float64   `bun:"zscore_bbtb_whoantro,type:numeric(5,2)" json:"zscore_bbtb_whoantro"`
	CreatedAt          time.Time  `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt          *time.Time `bun:"updatedAt" json:"updated_at"`
	DeletedAt          *time.Time `bun:"deletedAt,soft_delete" json:"deleted_at,omitempty"`
	SourceData         *string    `bun:"source_data,type:char(36)" json:"source_data"`
	UpdatedBy          *string    `bun:"updated_by,type:varchar(100)" json:"updated_by"`
	DeletedBy          *string    `bun:"deleted_by,type:varchar(100)" json:"deleted_by"`

	Anak *Anak `bun:"rel:belongs-to,join:id_anak=id" json:"anak,omitempty"`
}

type Encounter struct {
	SatusehatId string `bun:"satusehat_id,pk,type:char(36),default:null" json:"satusehat_id"`
	Anak        string `bun:"anak,type:char(36),default:null" json:"anak"`
}

func (m *Orangtua) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.UpdateQuery); ok {
		now := time.Now()
		m.UpdatedAt = &now
	}
	return nil
}

func (m *Anak) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.UpdateQuery); ok {
		now := time.Now()
		m.UpdatedAt = &now
	}
	return nil
}

func (m *Kesehatan) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.UpdateQuery); ok {
		now := time.Now()
		m.UpdatedAt = &now
	}
	return nil
}

func (m *Kunjungan) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.UpdateQuery); ok {
		now := time.Now()
		m.UpdatedAt = &now
	}
	return nil
}

func (e Kunjungan) TableName() string {
	return "jakantro.kunjungan"
}

func (e Kunjungan) GetPkName() string {
	return "kj.id"
}

func (e Orangtua) TableName() string {
	return "jakantro.orangtua"
}

func (e Orangtua) GetPkName() string {
	return "o.id"
}

func (e Anak) TableName() string {
	return "jakantro.anak"
}

func (e Anak) GetPkName() string {
	return "a.id"
}
