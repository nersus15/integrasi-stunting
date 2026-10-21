package types

import (
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
)

type Kunjungan struct {
	Id                string     `json:"id"`
	IdAnak            string     `json:"id_anak"`
	TanggalPengukuran string     `json:"tanggal_pengukuran"`
	TanggalSelesai    *string    `json:"tanggal_selesai"`
	CaraUkur          *string    `json:"cara_ukur"`
	BeratBadan        *float64   `json:"berat_badan"`
	TinggiBadan       *float64   `json:"tinggi_badan"`
	LingkarLengan     *float64   `json:"lingkar_lengan"`
	LingkarKepala     *float64   `json:"lingkar_kepala"`
	LingkarDada       *float64   `json:"lingkar_dada"`
	ASIBulan0         *int16     `json:"asi_bulan_0"`
	ASIBulan1         *int16     `json:"asi_bulan_1"`
	ASIBulan2         *int16     `json:"asi_bulan_2"`
	ASIBulan3         *int16     `json:"asi_bulan_3"`
	ASIBulan4         *int16     `json:"asi_bulan_4"`
	ASIBulan5         *int16     `json:"asi_bulan_5"`
	ASIBulan6         *int16     `json:"asi_bulan_6"`
	VitBiru           *int16     `json:"vit_biru"`
	VitMerah          *int16     `json:"vit_merah"`
	PittingEdema      *int16     `json:"pitting_edema"`
	KelasIbuBalita    *int16     `json:"kelas_ibu_balita"`
	StatusBbu         *string    `json:"status_bbu"`
	StatusTbu         *string    `json:"status_tbu"`
	StatusBbtb        *string    `json:"status_bbtb"`
	ZscoreBbu         *float64   `json:"zscore_bbu"`
	ZscoreTbu         *float64   `json:"zscore_tbu"`
	ZscoreBbtb        *float64   `json:"zscore_bbtb"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
	SourceData        *string    `json:"source_data"`
	CreatedBy         *string    `json:"created_by"`
	UpdatedBy         *string    `json:"updated_by"`
	DeletedBy         *string    `json:"deleted_by"`
	IdFaskes          *string    `json:"id_faskes"`
	IdSatusehat       *string    `json:"id_satusehat"`
	IdEpisode         *string    `json:"id_episode"`
	RefEpisode        *string    `json:"ref_episode"`
	IdRujukan         *string    `json:"id_rujukan"`
	RefRujukan        *string    `json:"ref_rujukan"`
	Stunting          *int       `json:"stunting"`
}

type KunjunganAnak struct {
	IdOrangtua *string   `json:"id_orangtua"`
	IdAnak     *string   `json:"id_anak"`
	Orangtua   *Orangtua `json:"orangtua"`
	Anak       *Anak     `json:"anak"`
	Kunjungan  Kunjungan `json:"kunjungan"`
}

type KunjunganAnakArray struct {
	Anak      Anak        `json:"anak"`
	Kunjungan []Kunjungan `json:"kunjungan"`
}

type KunjunganPayload struct {
	Id                string     `json:"id"`
	SatusehatId       *string    `json:"id_satusehat"`
	IDAnak            string     `json:"id_anak"`
	IdFaskes          *string    `json:"id_faskes"`
	TanggalPengukuran string     `json:"tanggal_pengukuran"`
	TanggalSelesai    *string    `json:"tanggal_selesai"`
	IdEpisode         *string    `json:"id_episode"`
	RefEpisode        *string    `json:"ref_episode"`
	IdRujukan         *string    `json:"id_rujukan"`
	RefRujukan        *string    `json:"ref_rujukan"`
	CaraUkur          *string    `json:"cara_ukur"`
	BeratBadan        *float64   `json:"berat_badan"`
	TinggiBadan       *float64   `json:"tinggi_badan"`
	LingkarLengan     *float64   `json:"lingkar_lengan"`
	LingkarKepala     *float64   `json:"lingkar_kepala"`
	LingkarDada       *float64   `json:"lingkar_dada"`
	ASIBulan0         *int16     `json:"asi_bulan_0"`
	ASIBulan1         *int16     `json:"asi_bulan_1"`
	ASIBulan2         *int16     `json:"asi_bulan_2"`
	ASIBulan3         *int16     `json:"asi_bulan_3"`
	ASIBulan4         *int16     `json:"asi_bulan_4"`
	ASIBulan5         *int16     `json:"asi_bulan_5"`
	ASIBulan6         *int16     `json:"asi_bulan_6"`
	VitBiru           *int16     `json:"vit_biru"`
	VitMerah          *int16     `json:"vit_merah"`
	PittingEdema      *int16     `json:"pitting_edema"`
	KelasIbuBalita    *int16     `json:"kelas_ibu_balita"`
	StatusBBU         *string    `json:"status_bbu"`
	StatusTBU         *string    `json:"status_tbu"`
	StatusBBTB        *string    `json:"status_bbtb"`
	ZScoreBBU         *float64   `json:"zscore_bbu"`
	ZScoreTBU         *float64   `json:"zscore_tbu"`
	ZScoreBBTB        *float64   `json:"zscore_bbtb"`
	SourceData        *string    `json:"source_data"`
	CreatedBy         *string    `json:"created_by"`
	UpdatedBy         *string    `json:"updated_by"`
	DeletedBy         *string    `json:"deleted_by"`
	CreatedAt         *time.Time `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
	Stunting          *int       `json:"stunting"`
}

type KunjunganNestedPayload struct {
	Orangtua  *OrangtuaPayload  `json:"orangtua"`
	Anak      *AnakPayload      `json:"anak"`
	Kunjungan *KunjunganPayload `json:"kunjungan"`

	// Kunjungan only without nested
	KunjunganPayload
}

func (kj *Kunjungan) FromEntity(e *entity.Kunjungan) *Kunjungan {
	if e == nil {
		return nil
	}
	return &Kunjungan{
		Id:                e.ID,
		IdAnak:            e.IDAnak,
		TanggalPengukuran: e.TanggalPengukuran.Format("2006-01-02"),
		TanggalSelesai:    teksTanggal(e.TanggalSelesai),
		CaraUkur:          e.CaraUkur,
		BeratBadan:        e.BeratBadan,
		TinggiBadan:       e.TinggiBadan,
		LingkarLengan:     e.LingkarLengan,
		LingkarKepala:     e.LingkarKepala,
		LingkarDada:       e.LingkarDada,
		ASIBulan0:         e.ASIBulan0,
		ASIBulan1:         e.ASIBulan1,
		ASIBulan2:         e.ASIBulan2,
		ASIBulan3:         e.ASIBulan3,
		ASIBulan4:         e.ASIBulan4,
		ASIBulan5:         e.ASIBulan5,
		ASIBulan6:         e.ASIBulan6,
		VitBiru:           e.VitBiru,
		VitMerah:          e.VitMerah,
		PittingEdema:      e.PittingEdema,
		KelasIbuBalita:    e.KelasIbuBalita,
		StatusBbu:         e.StatusBBU,
		StatusTbu:         e.StatusTBU,
		StatusBbtb:        e.StatusBBTB,
		ZscoreBbu:         e.ZScoreBBU,
		ZscoreTbu:         e.ZScoreTBU,
		ZscoreBbtb:        e.ZScoreBBTB,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
		DeletedAt:         e.DeletedAt,
		SourceData:        e.SourceData,
		CreatedBy:         e.CreatedBy,
		UpdatedBy:         e.UpdatedBy,
		DeletedBy:         e.DeletedBy,
		IdFaskes:          e.IDFaskes,
		IdSatusehat:       e.SatusehatId,
		IdEpisode:         e.IDEpisode,
		RefEpisode:        e.RefEpisode,
		IdRujukan:         e.IDRujukan,
		RefRujukan:        e.RefRujukan,
		Stunting:          e.Stunting,
	}
}

func (p *Kunjungan) ToPayload() *KunjunganPayload {
	if p == nil {
		return nil
	}

	return &KunjunganPayload{
		Id:                p.Id,
		IDAnak:            p.IdAnak,
		SatusehatId:       p.IdSatusehat,
		IdFaskes:          p.IdFaskes,
		CreatedAt:         &p.CreatedAt,
		DeletedAt:         p.DeletedAt,
		TanggalPengukuran: p.TanggalPengukuran,
		TanggalSelesai:    p.TanggalSelesai,
		IdEpisode:         p.IdEpisode,
		RefEpisode:        p.RefEpisode,
		IdRujukan:         p.IdRujukan,
		RefRujukan:        p.RefRujukan,
		CaraUkur:          p.CaraUkur,
		BeratBadan:        p.BeratBadan,
		TinggiBadan:       p.TinggiBadan,
		LingkarLengan:     p.LingkarLengan,
		LingkarKepala:     p.LingkarKepala,
		LingkarDada:       p.LingkarDada,
		ASIBulan0:         p.ASIBulan0,
		ASIBulan1:         p.ASIBulan1,
		ASIBulan2:         p.ASIBulan2,
		ASIBulan3:         p.ASIBulan3,
		ASIBulan4:         p.ASIBulan4,
		ASIBulan5:         p.ASIBulan5,
		ASIBulan6:         p.ASIBulan6,
		VitBiru:           p.VitBiru,
		VitMerah:          p.VitMerah,
		PittingEdema:      p.PittingEdema,
		KelasIbuBalita:    p.KelasIbuBalita,
		StatusBBU:         p.StatusBbu,
		StatusTBU:         p.StatusTbu,
		StatusBBTB:        p.StatusBbtb,
		ZScoreBBU:         p.ZscoreBbu,
		ZScoreTBU:         p.ZscoreTbu,
		ZScoreBBTB:        p.ZscoreBbtb,
		SourceData:        p.SourceData,
		CreatedBy:         p.CreatedBy,
		UpdatedBy:         p.UpdatedBy,
		DeletedBy:         p.DeletedBy,
		UpdatedAt:         p.UpdatedAt,
		Stunting:          p.Stunting,
	}
}

func (p *KunjunganPayload) ToEntity() *entity.Kunjungan {
	if p == nil {
		return nil
	}

	tPengukuran, _ := time.Parse("2006-01-02", p.TanggalPengukuran)
	e := &entity.Kunjungan{
		ID:                p.Id,
		IDAnak:            p.IDAnak,
		TanggalPengukuran: tPengukuran,
		CaraUkur:          p.CaraUkur,
		BeratBadan:        p.BeratBadan,
		TinggiBadan:       p.TinggiBadan,
		LingkarLengan:     p.LingkarLengan,
		LingkarKepala:     p.LingkarKepala,
		LingkarDada:       p.LingkarDada,
		ASIBulan0:         p.ASIBulan0,
		ASIBulan1:         p.ASIBulan1,
		ASIBulan2:         p.ASIBulan2,
		ASIBulan3:         p.ASIBulan3,
		ASIBulan4:         p.ASIBulan4,
		ASIBulan5:         p.ASIBulan5,
		ASIBulan6:         p.ASIBulan6,
		VitBiru:           p.VitBiru,
		VitMerah:          p.VitMerah,
		PittingEdema:      p.PittingEdema,
		KelasIbuBalita:    p.KelasIbuBalita,
		StatusBBU:         p.StatusBBU,
		StatusTBU:         p.StatusTBU,
		StatusBBTB:        p.StatusBBTB,
		ZScoreBBU:         p.ZScoreBBU,
		ZScoreTBU:         p.ZScoreTBU,
		ZScoreBBTB:        p.ZScoreBBTB,
		SourceData:        p.SourceData,
		CreatedBy:         p.CreatedBy,
		UpdatedBy:         p.UpdatedBy,
		DeletedBy:         p.DeletedBy,
		Stunting:          p.Stunting,
		SatusehatId:       p.SatusehatId,
		IDFaskes:          p.IdFaskes,
		TanggalSelesai:    waktuDariString(p.TanggalSelesai),
		IDEpisode:         p.IdEpisode,
		RefEpisode:        p.RefEpisode,
		IDRujukan:         p.IdRujukan,
		RefRujukan:        p.RefRujukan,
	}

	// timestamp dari jakantro; kalau tidak dikirim, database yang mengisi
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}
	e.UpdatedAt = p.UpdatedAt
	e.DeletedAt = p.DeletedAt

	return e
}

func (k *KunjunganAnakArray) FromEntity(e *entity.Anak) *KunjunganAnakArray {
	if e == nil {
		return nil
	}

	var a *Anak
	kunjungan := make([]Kunjungan, 0)

	anak := a.FromEntity(e)

	for _, k := range e.Kunjungan {
		var tmp *Kunjungan
		if t := tmp.FromEntity(k); t != nil {
			kunjungan = append(kunjungan, *tmp.FromEntity(k))
		}
	}
	return &KunjunganAnakArray{
		Anak:      *anak,
		Kunjungan: kunjungan,
	}
}

// selalu tidak-nil supaya JSON-nya [] bukan null
func daftarKunjungan(list []*entity.Kunjungan) []Kunjungan {
	return daftar(list, (*Kunjungan)(nil).FromEntity)
}
