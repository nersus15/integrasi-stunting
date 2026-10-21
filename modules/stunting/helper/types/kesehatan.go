package types

import (
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
)

type Kesehatan struct {
	Id                string     `json:"id"`
	IdAnak            string     `json:"id_anak"`
	TanggalPemantauan string     `json:"tanggal_pemantauan"`
	TbcBatuk          *int16     `json:"tbc_batuk"`
	TbcDemam          *int16     `json:"tbc_demam"`
	TbcBb             *int16     `json:"tbc_bb"`
	TbcKontak         *int16     `json:"tbc_kontak"`
	LayananAsiEks     *int16     `json:"layanan_asi_eks"`
	LayananMpasi      *int16     `json:"layanan_mpasi"`
	LayananImunisasi  *int16     `json:"layanan_imunisasi"`
	LayananVitA       *int16     `json:"layanan_vit_a"`
	LayananObatCacing *int16     `json:"layanan_obat_cacing"`
	LayananMtPangan   *int16     `json:"layanan_mt_pangan"`
	PenyuluhanEdukasi *int16     `json:"penyuluhan_edukasi"`
	PenyuluhanRujukan *int16     `json:"penyuluhan_rujukan"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
	SourceData        *string    `json:"source_data"`
	CreatedBy         *string    `json:"created_by"`
	UpdatedBy         *string    `json:"updated_by"`
	DeletedBy         *string    `json:"deleted_by"`
}

type KesehatanAnak struct {
	IdOrangtua *string   `json:"id_orangtua"`
	IdAnak     *string   `json:"id_anak"`
	Orangtua   *Orangtua `json:"orangtua"`
	Anak       *Anak     `json:"anak"`
	Kesehatan  Kesehatan `json:"kesehatan"`
}

type KesehatanAnakArray struct {
	Anak      Anak        `json:"anak"`
	Kesehatan []Kesehatan `json:"kesehatan"`
}

type KesehatanPayload struct {
	Id                string     `json:"id"`
	IDAnak            string     `json:"id_anak"`
	TanggalPemantauan string     `json:"tanggal_pemantauan"` // format YYYY-MM-DD
	TBCBatuk          *int16     `json:"tbc_batuk"`
	TBCDemam          *int16     `json:"tbc_demam"`
	TBCBB             *int16     `json:"tbc_bb"`
	TBCKontak         *int16     `json:"tbc_kontak"`
	LayananASIEks     *int16     `json:"layanan_asi_eks"`
	LayananMPASI      *int16     `json:"layanan_mpasi"`
	LayananImunisasi  *int16     `json:"layanan_imunisasi"`
	LayananVitA       *int16     `json:"layanan_vit_a"`
	LayananObatCacing *int16     `json:"layanan_obat_cacing"`
	LayananMTPangan   *int16     `json:"layanan_mt_pangan"`
	PenyuluhanEdukasi *int16     `json:"penyuluhan_edukasi"`
	PenyuluhanRujukan *int16     `json:"penyuluhan_rujukan"`
	SourceData        *string    `json:"source_data"`
	CreatedBy         *string    `json:"created_by"`
	UpdatedBy         *string    `json:"updated_by"`
	DeletedBy         *string    `json:"deleted_by"`
	CreatedAt         *time.Time `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
}

type KesehatanNestedPayload struct {
	Orangtua  *OrangtuaPayload  `json:"orangtua"`
	Anak      *AnakPayload      `json:"anak"`
	Kesehatan *KesehatanPayload `json:"kesehatan"`

	// Kesehatan only without nested
	KesehatanPayload
}

func (k *Kesehatan) FromEntity(e *entity.Kesehatan) *Kesehatan {
	if e == nil {
		return nil
	}
	return &Kesehatan{
		Id:                e.ID,
		IdAnak:            e.IDAnak,
		TanggalPemantauan: e.TanggalPemantauan.Format("2006-01-02"),
		TbcBatuk:          e.TBCBatuk,
		TbcDemam:          e.TBCDemam,
		TbcBb:             e.TBCBB,
		TbcKontak:         e.TBCKontak,
		LayananAsiEks:     e.LayananASIEks,
		LayananMpasi:      e.LayananMPASI,
		LayananImunisasi:  e.LayananImunisasi,
		LayananVitA:       e.LayananVitA,
		LayananObatCacing: e.LayananObatCacing,
		LayananMtPangan:   e.LayananMTPangan,
		PenyuluhanEdukasi: e.PenyuluhanEdukasi,
		PenyuluhanRujukan: e.PenyuluhanRujukan,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
		DeletedAt:         e.DeletedAt,
		SourceData:        e.SourceData,
		CreatedBy:         e.CreatedBy,
		UpdatedBy:         e.UpdatedBy,
		DeletedBy:         e.DeletedBy,
	}
}

func (p *Kesehatan) ToPayload() *KesehatanPayload {
	if p == nil {
		return nil
	}

	return &KesehatanPayload{
		Id:                p.Id,
		IDAnak:            p.IdAnak,
		TanggalPemantauan: p.TanggalPemantauan,
		TBCBatuk:          p.TbcBatuk,
		TBCDemam:          p.TbcDemam,
		TBCBB:             p.TbcBb,
		TBCKontak:         p.TbcKontak,
		LayananASIEks:     p.LayananAsiEks,
		LayananMPASI:      p.LayananMpasi,
		LayananImunisasi:  p.LayananImunisasi,
		LayananVitA:       p.LayananVitA,
		LayananObatCacing: p.LayananObatCacing,
		LayananMTPangan:   p.LayananMtPangan,
		PenyuluhanEdukasi: p.PenyuluhanEdukasi,
		PenyuluhanRujukan: p.PenyuluhanRujukan,
		SourceData:        p.SourceData,
		CreatedBy:         p.CreatedBy,
		UpdatedBy:         p.UpdatedBy,
		DeletedBy:         p.DeletedBy,
		CreatedAt:         &p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
		DeletedAt:         p.DeletedAt,
	}
}

func (p *KesehatanPayload) ToEntity() *entity.Kesehatan {
	if p == nil {
		return nil
	}

	tPemantauan, _ := time.Parse("2006-01-02", p.TanggalPemantauan)
	e := &entity.Kesehatan{
		ID:                p.Id,
		IDAnak:            p.IDAnak,
		TanggalPemantauan: tPemantauan,
		TBCBatuk:          p.TBCBatuk,
		TBCDemam:          p.TBCDemam,
		TBCBB:             p.TBCBB,
		TBCKontak:         p.TBCKontak,
		LayananASIEks:     p.LayananASIEks,
		LayananMPASI:      p.LayananMPASI,
		LayananImunisasi:  p.LayananImunisasi,
		LayananVitA:       p.LayananVitA,
		LayananObatCacing: p.LayananObatCacing,
		LayananMTPangan:   p.LayananMTPangan,
		PenyuluhanEdukasi: p.PenyuluhanEdukasi,
		PenyuluhanRujukan: p.PenyuluhanRujukan,
		SourceData:        p.SourceData,
		CreatedBy:         p.CreatedBy,
		UpdatedBy:         p.UpdatedBy,
		DeletedBy:         p.DeletedBy,
	}

	// timestamp dari jakantro; kalau tidak dikirim, database yang mengisi
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}
	e.UpdatedAt = p.UpdatedAt
	e.DeletedAt = p.DeletedAt

	return e
}

func (k *KesehatanAnakArray) FromEntity(e *entity.Anak) *KesehatanAnakArray {
	if e == nil {
		return nil
	}

	var a *Anak
	anak := a.FromEntity(e)

	return &KesehatanAnakArray{
		Anak:      *anak,
		Kesehatan: daftarKesehatan(e.Kesehatan),
	}
}

func daftarKesehatan(list []*entity.Kesehatan) []Kesehatan {
	return daftar(list, (*Kesehatan)(nil).FromEntity)
}
