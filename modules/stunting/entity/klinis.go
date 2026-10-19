package entity

import (
	"context"
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
		"UpdatedAt":   true,
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
		"BaseModel":    true,
		"ID":           true,
		"SatusehatId":  true,
		"IDAnak":       true,
		"IDKunjungan":  true,
		"IDInduk":      true,
		"CreatedAt":    true,
		"DeletedAt":    true,
		"Anak":         true,
		"Kunjungan":    true,
		"Induk":        true,
		"Component":    true,
		"RefEncounter": true,
		"UpdatedAt":    true,
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
		"IDAnak":       true,
		"IDKunjungan":  true,
		"RefEncounter": true,
		"UpdatedAt":    true,
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
		"BaseModel":    true,
		"ID":           true,
		"SatusehatId":  true,
		"IDAnak":       true,
		"IDKunjungan":  true,
		"CreatedAt":    true,
		"DeletedAt":    true,
		"Anak":         true,
		"Kunjungan":    true,
		"RefEncounter": true,
		"Jenis":        true,
		"UpdatedAt":    true,
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
		"BaseModel":       true,
		"ID":              true,
		"SatusehatId":     true,
		"IDAnak":          true,
		"IDKunjungan":     true,
		"IDFaskesAsal":    true,
		"IDFaskesTujuan":  true,
		"CreatedAt":       true,
		"DeletedAt":       true,
		"Anak":            true,
		"Kunjungan":       true,
		"FaskesAsal":      true,
		"FaskesTujuan":    true,
		"RefEncounter":    true,
		"RefFaskesAsal":   true,
		"RefFaskesTujuan": true,
		"Jenis":           true,
		"Tanggal":         true,
		"UpdatedAt":       true,
	}

	timpaField(e, new, skipFields, force)
}
