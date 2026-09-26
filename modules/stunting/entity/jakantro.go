package entity

import (
	"context"
	"reflect"
	"time"

	"github.com/uptrace/bun"
)

const (
	DiagnosaDiagnosis = "diagnosis"
	DiagnosaAlergi    = "alergi"
)

const (
	LayananProcedure          = "procedure"
	LayananMedicationDispense = "medication_dispense"
	LayananNutritionOrder     = "nutrition_order"
	LayananImmunization       = "immunization"
	LayananServiceRequest     = "service_request"
)

const (
	RujukanKeluar = "rujukan"
	RujukBalik    = "rujuk_balik"
	// tanpa faskes tujuan: permintaan lab/radiologi atau kontrol di faskes sama
	RujukanInternal = "internal"
)

type Entity interface {
	TableName() string
	GetID() string
	GetPkName() string
}

// created_at/updated_at snake_case, ikut DDL. Hapus diwakili status = 0.
type Faskes struct {
	bun.BaseModel `bun:"table:stunting.faskes,alias:f"`

	ID          string     `bun:"id,pk,type:varchar(36)" json:"id"`
	IDInduk     *string    `bun:"id_induk,type:varchar(36)" json:"id_induk"`
	SatusehatID *string    `bun:"satusehat_id,type:varchar(36),nullzero,unique" json:"satusehat_id"`
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

type Orangtua struct {
	bun.BaseModel `bun:"table:stunting.orangtua,alias:o"`

	ID           string     `bun:"id,pk,type:varchar(36)" json:"id"`
	SatusehatId  string     `bun:"satusehat_id,type:varchar(36),default:null" json:"satusehat_id"`
	IDPosyandu   *string    `bun:"id_posyandu,type:varchar(36),nullzero" json:"id_posyandu"`
	NoKK         string     `bun:"no_kk,type:varchar(16),nullzero,unique" json:"no_kk"`
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
	bun.BaseModel `bun:"table:stunting.anak,alias:a"`

	ID           string     `bun:"id,pk,type:varchar(36)" json:"id"`
	SatusehatId  string     `bun:"satusehat_id,type:varchar(36),default:null" json:"satusehat_id"`
	IDOrangtua   string     `bun:"id_orangtua,type:varchar(36),nullzero" json:"id_orangtua"`
	Nama         string     `bun:"nama,type:varchar(255),notnull" json:"nama"`
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
	Kesehatan []*Kesehatan `bun:"rel:has-many,join:id=id_anak" json:"kesehatan,omitempty"`
	Episode   []*Episode   `bun:"rel:has-many,join:id=id_anak" json:"episode,omitempty"`
	Observasi []*Observasi `bun:"rel:has-many,join:id=id_anak" json:"observasi,omitempty"`
	Diagnosa  []*Diagnosa  `bun:"rel:has-many,join:id=id_anak" json:"diagnosa,omitempty"`
	Layanan   []*Layanan   `bun:"rel:has-many,join:id=id_anak" json:"layanan,omitempty"`
	Rujukan   []*Rujukan   `bun:"rel:has-many,join:id=id_anak" json:"rujukan,omitempty"`
}

type Kunjungan struct {
	bun.BaseModel `bun:"table:stunting.kunjungan,alias:kj"`

	ID                 string     `bun:"id,pk,type:varchar(36)" json:"id"`
	IDAnak             string     `bun:"id_anak,type:varchar(36),notnull" json:"id_anak"`
	TanggalPengukuran  time.Time  `bun:"tanggal_pengukuran,type:date,notnull" json:"tanggal_pengukuran"`
	TanggalSelesai     *time.Time `bun:"tanggal_selesai,type:date,nullzero" json:"tanggal_selesai"`
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
	IDFaskes           *string    `bun:"id_faskes,type:varchar(36),nullzero" json:"id_faskes"`
	SatusehatId        *string    `bun:"satusehat_id,type:varchar(36),nullzero" json:"satusehat_id"`
	IDEpisode          *string    `bun:"id_episode,type:varchar(36),nullzero" json:"id_episode"`
	RefEpisode         *string    `bun:"ref_episode,type:varchar(36),nullzero" json:"ref_episode"`
	IDRujukan          *string    `bun:"id_rujukan,type:varchar(36),nullzero" json:"id_rujukan"`
	RefRujukan         *string    `bun:"ref_rujukan,type:varchar(36),nullzero" json:"ref_rujukan"`
	Stunting           *int       `bun:"stunting,type:smallint,nullzero" json:"stunting"`

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
	SourceData        *string    `bun:"source_data,type:char(36)" json:"source_data"`

	Anak *Anak `bun:"rel:belongs-to,join:id_anak=id" json:"anak,omitempty"`
}

type Episode struct {
	bun.BaseModel `bun:"table:stunting.episode,alias:ep"`

	ID          string     `bun:"id,pk,type:varchar(36)" json:"id"`
	IDAnak      string     `bun:"id_anak,type:varchar(36),notnull" json:"id_anak"`
	IDFaskes    *string    `bun:"id_faskes,type:varchar(36),nullzero" json:"id_faskes"`
	SatusehatId *string    `bun:"satusehat_id,type:varchar(36),nullzero,unique" json:"satusehat_id"`
	System      *string    `bun:"system,type:varchar(255)" json:"system"`
	Kode        *string    `bun:"kode,type:varchar(50)" json:"kode"`
	Display     *string    `bun:"display,type:varchar(255)" json:"display"`
	Status      *string    `bun:"status,type:varchar(30)" json:"status"`
	Mulai       *time.Time `bun:"mulai,type:date" json:"mulai"`
	Selesai     *time.Time `bun:"selesai,type:date" json:"selesai"`
	CreatedAt   time.Time  `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   *time.Time `bun:"updatedAt" json:"updated_at"`
	DeletedAt   *time.Time `bun:"deletedAt,soft_delete" json:"deleted_at,omitempty"`

	Anak   *Anak   `bun:"rel:belongs-to,join:id_anak=id" json:"anak,omitempty"`
	Faskes *Faskes `bun:"rel:belongs-to,join:id_faskes=id" json:"faskes,omitempty"`
}

type Observasi struct {
	bun.BaseModel `bun:"table:stunting.observasi,alias:ob"`

	ID              string     `bun:"id,pk,type:varchar(36)" json:"id"`
	IDAnak          string     `bun:"id_anak,type:varchar(36),notnull" json:"id_anak"`
	IDKunjungan     *string    `bun:"id_kunjungan,type:varchar(36),nullzero" json:"id_kunjungan"`
	IDInduk         *string    `bun:"id_induk,type:varchar(36),nullzero" json:"id_induk"`
	RefEncounter    *string    `bun:"ref_encounter,type:varchar(36),nullzero" json:"ref_encounter"`
	SatusehatId     *string    `bun:"satusehat_id,type:varchar(36),nullzero,unique" json:"satusehat_id"`
	System          string     `bun:"system,type:varchar(255),notnull" json:"system"`
	Kode            string     `bun:"kode,type:varchar(50),notnull" json:"kode"`
	Display         *string    `bun:"display,type:varchar(255)" json:"display"`
	Kategori        *string    `bun:"kategori,type:varchar(50)" json:"kategori"`
	NilaiAngka      *float64   `bun:"nilai_angka,type:numeric(12,4)" json:"nilai_angka"`
	Satuan          *string    `bun:"satuan,type:varchar(50)" json:"satuan"`
	NilaiTeks       *string    `bun:"nilai_teks,type:text" json:"nilai_teks"`
	NilaiKode       *string    `bun:"nilai_kode,type:varchar(50)" json:"nilai_kode"`
	NilaiKodeSystem *string    `bun:"nilai_kode_system,type:varchar(255)" json:"nilai_kode_system"`
	NilaiDisplay    *string    `bun:"nilai_display,type:varchar(255)" json:"nilai_display"`
	Interpretasi    *string    `bun:"interpretasi,type:varchar(50)" json:"interpretasi"`
	Tanggal         *time.Time `bun:"tanggal" json:"tanggal"`
	CreatedAt       time.Time  `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       *time.Time `bun:"updatedAt" json:"updated_at"`
	DeletedAt       *time.Time `bun:"deletedAt,soft_delete" json:"deleted_at,omitempty"`

	Anak      *Anak        `bun:"rel:belongs-to,join:id_anak=id" json:"anak,omitempty"`
	Kunjungan *Kunjungan   `bun:"rel:belongs-to,join:id_kunjungan=id" json:"kunjungan,omitempty"`
	Induk     *Observasi   `bun:"rel:belongs-to,join:id_induk=id" json:"induk,omitempty"`
	Component []*Observasi `bun:"rel:has-many,join:id=id_induk" json:"component,omitempty"`
}

type Diagnosa struct {
	bun.BaseModel `bun:"table:stunting.diagnosa,alias:d"`

	ID                 string     `bun:"id,pk,type:varchar(36)" json:"id"`
	IDAnak             string     `bun:"id_anak,type:varchar(36),notnull" json:"id_anak"`
	IDKunjungan        *string    `bun:"id_kunjungan,type:varchar(36),nullzero" json:"id_kunjungan"`
	RefEncounter       *string    `bun:"ref_encounter,type:varchar(36),nullzero" json:"ref_encounter"`
	SatusehatId        *string    `bun:"satusehat_id,type:varchar(36),nullzero,unique" json:"satusehat_id"`
	Jenis              string     `bun:"jenis,type:varchar(20),notnull" json:"jenis"`
	System             string     `bun:"system,type:varchar(255),notnull" json:"system"`
	Kode               string     `bun:"kode,type:varchar(50),notnull" json:"kode"`
	Display            *string    `bun:"display,type:varchar(255)" json:"display"`
	Kategori           *string    `bun:"kategori,type:varchar(50)" json:"kategori"`
	Kritikalitas       *string    `bun:"kritikalitas,type:varchar(30)" json:"kritikalitas"`
	ClinicalStatus     *string    `bun:"clinical_status,type:varchar(30)" json:"clinical_status"`
	VerificationStatus *string    `bun:"verification_status,type:varchar(30)" json:"verification_status"`
	Onset              *time.Time `bun:"onset,type:date" json:"onset"`
	TanggalCatat       *time.Time `bun:"tanggal_catat,type:date" json:"tanggal_catat"`
	CreatedAt          time.Time  `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt          *time.Time `bun:"updatedAt" json:"updated_at"`
	DeletedAt          *time.Time `bun:"deletedAt,soft_delete" json:"deleted_at,omitempty"`

	Anak      *Anak      `bun:"rel:belongs-to,join:id_anak=id" json:"anak,omitempty"`
	Kunjungan *Kunjungan `bun:"rel:belongs-to,join:id_kunjungan=id" json:"kunjungan,omitempty"`
}

type Layanan struct {
	bun.BaseModel `bun:"table:stunting.layanan,alias:l"`

	ID           string     `bun:"id,pk,type:varchar(36)" json:"id"`
	IDAnak       string     `bun:"id_anak,type:varchar(36),notnull" json:"id_anak"`
	IDKunjungan  *string    `bun:"id_kunjungan,type:varchar(36),nullzero" json:"id_kunjungan"`
	RefEncounter *string    `bun:"ref_encounter,type:varchar(36),nullzero" json:"ref_encounter"`
	SatusehatId  *string    `bun:"satusehat_id,type:varchar(36),nullzero,unique" json:"satusehat_id"`
	Jenis        string     `bun:"jenis,type:varchar(30),notnull" json:"jenis"`
	System       *string    `bun:"system,type:varchar(255)" json:"system"`
	Kode         *string    `bun:"kode,type:varchar(50)" json:"kode"`
	Display      *string    `bun:"display,type:varchar(255)" json:"display"`
	Kategori     *string    `bun:"kategori,type:varchar(50)" json:"kategori"`
	Status       *string    `bun:"status,type:varchar(30)" json:"status"`
	Jumlah       *float64   `bun:"jumlah,type:numeric(12,4)" json:"jumlah"`
	Satuan       *string    `bun:"satuan,type:varchar(50)" json:"satuan"`
	Tanggal      *time.Time `bun:"tanggal" json:"tanggal"`
	Catatan      *string    `bun:"catatan,type:text" json:"catatan"`
	CreatedAt    time.Time  `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    *time.Time `bun:"updatedAt" json:"updated_at"`
	DeletedAt    *time.Time `bun:"deletedAt,soft_delete" json:"deleted_at,omitempty"`

	Anak      *Anak      `bun:"rel:belongs-to,join:id_anak=id" json:"anak,omitempty"`
	Kunjungan *Kunjungan `bun:"rel:belongs-to,join:id_kunjungan=id" json:"kunjungan,omitempty"`
}

type Rujukan struct {
	bun.BaseModel `bun:"table:stunting.rujukan,alias:rj"`

	ID              string     `bun:"id,pk,type:varchar(36)" json:"id"`
	IDAnak          string     `bun:"id_anak,type:varchar(36),notnull" json:"id_anak"`
	IDKunjungan     *string    `bun:"id_kunjungan,type:varchar(36),nullzero" json:"id_kunjungan"`
	RefEncounter    *string    `bun:"ref_encounter,type:varchar(36),nullzero" json:"ref_encounter"`
	SatusehatId     *string    `bun:"satusehat_id,type:varchar(36),nullzero,unique" json:"satusehat_id"`
	Jenis           string     `bun:"jenis,type:varchar(20),notnull" json:"jenis"`
	IDFaskesAsal    *string    `bun:"id_faskes_asal,type:varchar(36),nullzero" json:"id_faskes_asal"`
	IDFaskesTujuan  *string    `bun:"id_faskes_tujuan,type:varchar(36),nullzero" json:"id_faskes_tujuan"`
	RefFaskesAsal   *string    `bun:"ref_faskes_asal,type:varchar(36),nullzero" json:"ref_faskes_asal"`
	RefFaskesTujuan *string    `bun:"ref_faskes_tujuan,type:varchar(36),nullzero" json:"ref_faskes_tujuan"`
	System          *string    `bun:"system,type:varchar(255)" json:"system"`
	Kode            *string    `bun:"kode,type:varchar(50)" json:"kode"`
	Display         *string    `bun:"display,type:varchar(255)" json:"display"`
	Status          *string    `bun:"status,type:varchar(30)" json:"status"`
	Prioritas       *string    `bun:"prioritas,type:varchar(20)" json:"prioritas"`
	Alasan          *string    `bun:"alasan,type:text" json:"alasan"`
	Tanggal         *time.Time `bun:"tanggal" json:"tanggal"`
	CreatedAt       time.Time  `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       *time.Time `bun:"updatedAt" json:"updated_at"`
	DeletedAt       *time.Time `bun:"deletedAt,soft_delete" json:"deleted_at,omitempty"`

	Anak         *Anak      `bun:"rel:belongs-to,join:id_anak=id" json:"anak,omitempty"`
	Kunjungan    *Kunjungan `bun:"rel:belongs-to,join:id_kunjungan=id" json:"kunjungan,omitempty"`
	FaskesAsal   *Faskes    `bun:"rel:belongs-to,join:id_faskes_asal=id" json:"faskes_asal,omitempty"`
	FaskesTujuan *Faskes    `bun:"rel:belongs-to,join:id_faskes_tujuan=id" json:"faskes_tujuan,omitempty"`
}

type Encounter struct {
	SatusehatId string `bun:"satusehat_id,pk,type:char(36),default:null" json:"satusehat_id"`
	Anak        string `bun:"anak,type:char(36),default:null" json:"anak"`
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
	return nil
}

func (e *Orangtua) Override(new Orangtua, force bool) {
	skipFields := map[string]bool{
		"BaseModel": true,
		"ID":        true,
		"CreatedAt": true,
		"DeletedAt": true,
		"DeletedBy": true,
		"Anak":      true,
	}

	timpaField(e, new, skipFields, force)
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
	return nil
}

func (e *Anak) Override(new Anak, force bool) {
	skipFields := map[string]bool{
		"BaseModel": true,
		"ID":        true,
		"CreatedAt": true,
		"DeletedAt": true,
		"DeletedBy": true,
		"Orangtua":  true,
		"Kunjungan": true,
		"Kesehatan": true,
		"Episode":   true,
		"Observasi": true,
		"Diagnosa":  true,
		"Layanan":   true,
		"Rujukan":   true,
	}

	timpaField(e, new, skipFields, force)
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
	}

	timpaField(e, new, skipFields, force)
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
		"SourceData": true,
		"Anak":       true,
	}

	timpaField(e, new, skipFields, force)
}

func (e Episode) TableName() string {
	return "stunting.episode"
}

func (e Episode) GetPkName() string {
	return "ep.id"
}

func (e *Episode) Override(new Episode, force bool) {
	skipFields := map[string]bool{
		"BaseModel":   true,
		"ID":          true,
		"SatusehatId": true,
		"IDAnak":      true,
		"IDFaskes":    true,
		"CreatedAt":   true,
		"DeletedAt":   true,
		"Anak":        true,
		"Faskes":      true,
	}

	timpaField(e, new, skipFields, force)
}

func (e Observasi) TableName() string {
	return "stunting.observasi"
}

func (e Observasi) GetPkName() string {
	return "ob.id"
}

func (m *Observasi) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.UpdateQuery); ok {
		now := time.Now()
		m.UpdatedAt = &now
	}
	return nil
}

func (e *Observasi) Override(new Observasi, force bool) {
	skipFields := map[string]bool{
		"BaseModel":   true,
		"ID":          true,
		"SatusehatId": true,
		"IDAnak":      true,
		"IDKunjungan": true,
		"IDInduk":     true,
		"CreatedAt":   true,
		"DeletedAt":   true,
		"Anak":        true,
		"Kunjungan":   true,
		"Induk":       true,
		"Component":   true,
	}

	timpaField(e, new, skipFields, force)
}

func (e Diagnosa) TableName() string {
	return "stunting.diagnosa"
}

func (e Diagnosa) GetPkName() string {
	return "d.id"
}

func (m *Diagnosa) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.UpdateQuery); ok {
		now := time.Now()
		m.UpdatedAt = &now
	}
	return nil
}

func (e *Diagnosa) Override(new Diagnosa, force bool) {
	skipFields := map[string]bool{
		"BaseModel":    true,
		"ID":           true,
		"SatusehatId":  true,
		"CreatedAt":    true,
		"DeletedAt":    true,
		"DeletedBy":    true,
		"Anak":         true,
		"Jenis":        true,
		"Kunjungan":    true,
		"TanggalCatat": true,
	}

	timpaField(e, new, skipFields, force)
}

func (e Layanan) TableName() string {
	return "stunting.layanan"
}

func (e Layanan) GetPkName() string {
	return "l.id"
}

func (m *Layanan) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.UpdateQuery); ok {
		now := time.Now()
		m.UpdatedAt = &now
	}
	return nil
}

func (e *Layanan) Override(new Layanan, force bool) {
	skipFields := map[string]bool{
		"BaseModel":   true,
		"ID":          true,
		"SatusehatId": true,
		"IDAnak":      true,
		"IDKunjungan": true,
		"CreatedAt":   true,
		"DeletedAt":   true,
		"Anak":        true,
		"Kunjungan":   true,
	}

	timpaField(e, new, skipFields, force)
}

func (e Rujukan) TableName() string {
	return "stunting.rujukan"
}

func (e Rujukan) GetPkName() string {
	return "rj.id"
}

func (e *Rujukan) Override(new Rujukan, force bool) {
	skipFields := map[string]bool{
		"BaseModel":      true,
		"ID":             true,
		"SatusehatId":    true,
		"IDAnak":         true,
		"IDKunjungan":    true,
		"IDFaskesAsal":   true,
		"IDFaskesTujuan": true,
		"CreatedAt":      true,
		"DeletedAt":      true,
		"Anak":           true,
		"Kunjungan":      true,
		"FaskesAsal":     true,
		"FaskesTujuan":   true,
	}

	timpaField(e, new, skipFields, force)
}

func timpaField(tujuan any, new any, skipFields map[string]bool, force bool) {
	valExisting := reflect.ValueOf(tujuan).Elem()
	valNew := reflect.ValueOf(new)

	for i := 0; i < valNew.NumField(); i++ {
		fieldName := valNew.Type().Field(i).Name
		if skipFields[fieldName] {
			continue
		}

		fieldNew := valNew.Field(i)
		fieldExisting := valExisting.FieldByName(fieldName)

		if !fieldExisting.IsValid() || !fieldExisting.CanSet() {
			continue
		}

		if (!fieldNew.IsZero() && fieldNew.Interface() != fieldExisting.Interface()) || force {
			fieldExisting.Set(fieldNew)
		}
	}
}
