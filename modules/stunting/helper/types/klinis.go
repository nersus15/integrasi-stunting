package types

import (
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
)

type Observasi struct {
	Id           string     `json:"id"`
	IdAnak       string     `json:"id_anak"`
	IdKunjungan  *string    `json:"id_kunjungan"`
	IdInduk      *string    `json:"id_induk"`
	RefEncounter *string    `json:"ref_encounter"`
	IdSatusehat  *string    `json:"id_satusehat"`
	System       string     `json:"system"`
	Kode         string     `json:"kode"`
	Display      *string    `json:"display"`
	Kategori     *string    `json:"kategori"`
	NilaiAngka   *float64   `json:"nilai_angka"`
	Satuan       *string    `json:"satuan"`
	NilaiTeks    *string    `json:"nilai_teks"`
	NilaiKode    *string    `json:"nilai_kode"`
	NilaiSystem  *string    `json:"nilai_kode_system"`
	NilaiDisplay *string    `json:"nilai_display"`
	Interpretasi *string    `json:"interpretasi"`
	Tanggal      *string    `json:"tanggal"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`

	Component []Observasi `json:"component,omitempty"`
}

type Diagnosa struct {
	Id                 string     `json:"id"`
	IdAnak             string     `json:"id_anak"`
	IdKunjungan        *string    `json:"id_kunjungan"`
	RefEncounter       *string    `json:"ref_encounter"`
	IdSatusehat        *string    `json:"id_satusehat"`
	Jenis              string     `json:"jenis"`
	System             string     `json:"system"`
	Kode               string     `json:"kode"`
	Display            *string    `json:"display"`
	Kategori           *string    `json:"kategori"`
	Kritikalitas       *string    `json:"kritikalitas"`
	ClinicalStatus     *string    `json:"clinical_status"`
	VerificationStatus *string    `json:"verification_status"`
	Onset              *string    `json:"onset"`
	TanggalCatat       *string    `json:"tanggal_catat"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at"`
}

type Layanan struct {
	Id           string     `json:"id"`
	IdAnak       string     `json:"id_anak"`
	IdKunjungan  *string    `json:"id_kunjungan"`
	RefEncounter *string    `json:"ref_encounter"`
	IdSatusehat  *string    `json:"id_satusehat"`
	Jenis        string     `json:"jenis"`
	System       *string    `json:"system"`
	Kode         *string    `json:"kode"`
	Display      *string    `json:"display"`
	Kategori     *string    `json:"kategori"`
	Status       *string    `json:"status"`
	Jumlah       *float64   `json:"jumlah"`
	Satuan       *string    `json:"satuan"`
	Tanggal      *string    `json:"tanggal"`
	Catatan      *string    `json:"catatan"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type Rujukan struct {
	Id              string     `json:"id"`
	IdAnak          string     `json:"id_anak"`
	IdKunjungan     *string    `json:"id_kunjungan"`
	RefEncounter    *string    `json:"ref_encounter"`
	IdSatusehat     *string    `json:"id_satusehat"`
	Jenis           string     `json:"jenis"`
	IdFaskesAsal    *string    `json:"id_faskes_asal"`
	IdFaskesTujuan  *string    `json:"id_faskes_tujuan"`
	RefFaskesAsal   *string    `json:"ref_faskes_asal"`
	RefFaskesTujuan *string    `json:"ref_faskes_tujuan"`
	System          *string    `json:"system"`
	Kode            *string    `json:"kode"`
	Display         *string    `json:"display"`
	Status          *string    `json:"status"`
	Prioritas       *string    `json:"prioritas"`
	Alasan          *string    `json:"alasan"`
	Tanggal         *string    `json:"tanggal"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

type RujukanDetail struct {
	Rujukan
	FaskesAsal   *Faskes `json:"faskes_asal"`
	FaskesTujuan *Faskes `json:"faskes_tujuan"`
}

type Episode struct {
	Id          string     `json:"id"`
	IdAnak      string     `json:"id_anak"`
	IdFaskes    *string    `json:"id_faskes"`
	RefFaskes   *string    `json:"ref_faskes"`
	IdSatusehat *string    `json:"id_satusehat"`
	System      *string    `json:"system"`
	Kode        *string    `json:"kode"`
	Display     *string    `json:"display"`
	Status      *string    `json:"status"`
	Mulai       *string    `json:"mulai"`
	Selesai     *string    `json:"selesai"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

func (ob *Observasi) FromEntity(e *entity.Observasi) *Observasi {
	if e == nil {
		return nil
	}

	o := &Observasi{
		Id:           e.ID,
		IdAnak:       e.IDAnak,
		IdKunjungan:  e.IDKunjungan,
		IdInduk:      e.IDInduk,
		RefEncounter: e.RefEncounter,
		IdSatusehat:  e.SatusehatId,
		System:       e.System,
		Kode:         e.Kode,
		Display:      e.Display,
		Kategori:     e.Kategori,
		NilaiAngka:   e.NilaiAngka,
		Satuan:       e.Satuan,
		NilaiTeks:    e.NilaiTeks,
		NilaiKode:    e.NilaiKode,
		NilaiSystem:  e.NilaiKodeSystem,
		NilaiDisplay: e.NilaiDisplay,
		Interpretasi: e.Interpretasi,
		Tanggal:      teksWaktu(e.Tanggal),
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}

	for _, c := range e.Component {
		var tmp *Observasi
		if t := tmp.FromEntity(c); t != nil {
			o.Component = append(o.Component, *t)
		}
	}

	return o
}

func (p *Observasi) ToEntity() *entity.Observasi {
	if p == nil {
		return nil
	}

	return &entity.Observasi{
		ID:              p.Id,
		IDAnak:          p.IdAnak,
		IDKunjungan:     p.IdKunjungan,
		IDInduk:         p.IdInduk,
		RefEncounter:    p.RefEncounter,
		SatusehatId:     p.IdSatusehat,
		System:          p.System,
		Kode:            p.Kode,
		Display:         p.Display,
		Kategori:        p.Kategori,
		NilaiAngka:      p.NilaiAngka,
		Satuan:          p.Satuan,
		NilaiTeks:       p.NilaiTeks,
		NilaiKode:       p.NilaiKode,
		NilaiKodeSystem: p.NilaiSystem,
		NilaiDisplay:    p.NilaiDisplay,
		Interpretasi:    p.Interpretasi,
		Tanggal:         waktuDariString(p.Tanggal),
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func (d *Diagnosa) FromEntity(e *entity.Diagnosa) *Diagnosa {
	if e == nil {
		return nil
	}
	return &Diagnosa{
		Id:                 e.ID,
		IdAnak:             e.IDAnak,
		IdKunjungan:        e.IDKunjungan,
		RefEncounter:       e.RefEncounter,
		IdSatusehat:        e.SatusehatId,
		Jenis:              e.Jenis,
		System:             e.System,
		Kode:               e.Kode,
		Display:            e.Display,
		Kategori:           e.Kategori,
		Kritikalitas:       e.Kritikalitas,
		ClinicalStatus:     e.ClinicalStatus,
		VerificationStatus: e.VerificationStatus,
		Onset:              teksTanggal(e.Onset),
		TanggalCatat:       teksTanggal(e.TanggalCatat),
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
}

func (p *Diagnosa) ToEntity() *entity.Diagnosa {
	if p == nil {
		return nil
	}

	jenis := p.Jenis
	if jenis == "" {
		jenis = entity.DiagnosaDiagnosis
	}

	return &entity.Diagnosa{
		ID:                 p.Id,
		IDAnak:             p.IdAnak,
		IDKunjungan:        p.IdKunjungan,
		RefEncounter:       p.RefEncounter,
		SatusehatId:        p.IdSatusehat,
		Jenis:              jenis,
		System:             p.System,
		Kode:               p.Kode,
		Display:            p.Display,
		Kategori:           p.Kategori,
		Kritikalitas:       p.Kritikalitas,
		ClinicalStatus:     p.ClinicalStatus,
		VerificationStatus: p.VerificationStatus,
		Onset:              waktuDariString(p.Onset),
		TanggalCatat:       waktuDariString(p.TanggalCatat),
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

func (l *Layanan) FromEntity(e *entity.Layanan) *Layanan {
	if e == nil {
		return nil
	}
	return &Layanan{
		Id:           e.ID,
		IdAnak:       e.IDAnak,
		IdKunjungan:  e.IDKunjungan,
		RefEncounter: e.RefEncounter,
		IdSatusehat:  e.SatusehatId,
		Jenis:        e.Jenis,
		System:       e.System,
		Kode:         e.Kode,
		Display:      e.Display,
		Kategori:     e.Kategori,
		Status:       e.Status,
		Jumlah:       e.Jumlah,
		Satuan:       e.Satuan,
		Tanggal:      teksWaktu(e.Tanggal),
		Catatan:      e.Catatan,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func (p *Layanan) ToEntity() *entity.Layanan {
	if p == nil {
		return nil
	}

	return &entity.Layanan{
		ID:           p.Id,
		IDAnak:       p.IdAnak,
		IDKunjungan:  p.IdKunjungan,
		RefEncounter: p.RefEncounter,
		SatusehatId:  p.IdSatusehat,
		Jenis:        p.Jenis,
		System:       p.System,
		Kode:         p.Kode,
		Display:      p.Display,
		Kategori:     p.Kategori,
		Status:       p.Status,
		Jumlah:       p.Jumlah,
		Satuan:       p.Satuan,
		Tanggal:      waktuDariString(p.Tanggal),
		Catatan:      p.Catatan,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func (r *Rujukan) FromEntity(e *entity.Rujukan) *Rujukan {
	if e == nil {
		return nil
	}
	return &Rujukan{
		Id:              e.ID,
		IdAnak:          e.IDAnak,
		IdKunjungan:     e.IDKunjungan,
		RefEncounter:    e.RefEncounter,
		IdSatusehat:     e.SatusehatId,
		Jenis:           e.Jenis,
		IdFaskesAsal:    e.IDFaskesAsal,
		IdFaskesTujuan:  e.IDFaskesTujuan,
		RefFaskesAsal:   e.RefFaskesAsal,
		RefFaskesTujuan: e.RefFaskesTujuan,
		System:          e.System,
		Kode:            e.Kode,
		Display:         e.Display,
		Status:          e.Status,
		Prioritas:       e.Prioritas,
		Alasan:          e.Alasan,
		Tanggal:         teksWaktu(e.Tanggal),
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

func (p *Rujukan) ToEntity() *entity.Rujukan {
	if p == nil {
		return nil
	}

	return &entity.Rujukan{
		ID:              p.Id,
		IDAnak:          p.IdAnak,
		IDKunjungan:     p.IdKunjungan,
		RefEncounter:    p.RefEncounter,
		SatusehatId:     p.IdSatusehat,
		Jenis:           p.Jenis,
		IDFaskesAsal:    p.IdFaskesAsal,
		IDFaskesTujuan:  p.IdFaskesTujuan,
		RefFaskesAsal:   p.RefFaskesAsal,
		RefFaskesTujuan: p.RefFaskesTujuan,
		System:          p.System,
		Kode:            p.Kode,
		Display:         p.Display,
		Status:          p.Status,
		Prioritas:       p.Prioritas,
		Alasan:          p.Alasan,
		Tanggal:         waktuDariString(p.Tanggal),
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func (p *RujukanDetail) FromEntity(entity *entity.Rujukan) *RujukanDetail {
	if entity == nil {
		return nil
	}
	var faskes *Faskes

	return &RujukanDetail{
		Rujukan: Rujukan{
			Id:              entity.ID,
			IdAnak:          entity.IDAnak,
			IdKunjungan:     entity.IDKunjungan,
			RefEncounter:    entity.RefEncounter,
			IdSatusehat:     entity.SatusehatId,
			Jenis:           entity.Jenis,
			IdFaskesAsal:    entity.IDFaskesAsal,
			IdFaskesTujuan:  entity.IDFaskesTujuan,
			RefFaskesAsal:   entity.RefFaskesAsal,
			RefFaskesTujuan: entity.RefFaskesTujuan,
			System:          entity.System,
			Kode:            entity.Kode,
			Display:         entity.Display,
			Status:          entity.Status,
			Prioritas:       entity.Prioritas,
			Alasan:          entity.Alasan,
			Tanggal:         teksWaktu(entity.Tanggal),
			CreatedAt:       entity.CreatedAt,
			UpdatedAt:       entity.UpdatedAt,
		},
		FaskesAsal:   faskes.FromEntity(entity.FaskesAsal),
		FaskesTujuan: faskes.FromEntity(entity.FaskesTujuan),
	}
}

func (ep *Episode) FromEntity(e *entity.Episode) *Episode {
	if e == nil {
		return nil
	}
	return &Episode{
		Id:          e.ID,
		IdAnak:      e.IDAnak,
		IdFaskes:    e.IDFaskes,
		IdSatusehat: e.SatusehatId,
		System:      e.System,
		Kode:        e.Kode,
		Display:     e.Display,
		Status:      e.Status,
		Mulai:       teksTanggal(e.Mulai),
		Selesai:     teksTanggal(e.Selesai),
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func (p *Episode) ToEntity() *entity.Episode {
	if p == nil {
		return nil
	}

	return &entity.Episode{
		ID:          p.Id,
		IDAnak:      p.IdAnak,
		IDFaskes:    p.IdFaskes,
		SatusehatId: p.IdSatusehat,
		System:      p.System,
		Kode:        p.Kode,
		Display:     p.Display,
		Status:      p.Status,
		Mulai:       waktuDariString(p.Mulai),
		Selesai:     waktuDariString(p.Selesai),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func daftarObservasi(list []*entity.Observasi) []Observasi {
	return daftar(list, (*Observasi)(nil).FromEntity)
}

func daftarDiagnosa(list []*entity.Diagnosa) []Diagnosa {
	return daftar(list, (*Diagnosa)(nil).FromEntity)
}

func daftarLayanan(list []*entity.Layanan) []Layanan { return daftar(list, (*Layanan)(nil).FromEntity) }

func daftarRujukan(list []*entity.Rujukan) []Rujukan { return daftar(list, (*Rujukan)(nil).FromEntity) }

func daftarEpisode(list []*entity.Episode) []Episode { return daftar(list, (*Episode)(nil).FromEntity) }
