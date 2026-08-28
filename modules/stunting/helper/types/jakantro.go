package types

import (
	"strings"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
)

type Orangtua struct {
	Id           string     `json:"id"`
	IdSatusehat  *string    `json:"id_satusehat"`
	IdPosyandu   string     `json:"id_posyandu"`
	NoKk         string     `json:"no_kk"`
	NamaAyah     string     `json:"nama_ayah"`
	NamaIbu      string     `json:"nama_ibu"`
	Nik          string     `json:"nik"`
	Telepon      string     `json:"telepon"`
	Rt           string     `json:"rt"`
	Rw           string     `json:"rw"`
	Alamat       string     `json:"alamat"`
	Kia          int16      `json:"kia"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
	SourceData   *string    `json:"source_data"`
	UsiaHamil    *int16     `json:"usia_hamil"`
	KiaBayiKecil *int16     `json:"kia_bayi_kecil"`
	UpdatedBy    *string    `json:"updated_by"`
	DeletedBy    *string    `json:"deleted_by"`
}

type Anak struct {
	Id           string     `json:"id"`
	IdSatusehat  *string    `json:"id_satusehat"`
	IdOrangtua   string     `json:"id_orangtua"`
	Nama         string     `json:"nama"`
	Nik          *string    `json:"nik"`
	TanggalLahir string     `json:"tanggal_lahir"`
	JenisKelamin string     `json:"jenis_kelamin"`
	AnakKe       int16      `json:"anak_ke"`
	Imd          int16      `json:"imd"`
	BbLahir      float32    `json:"bb_lahir"`
	TbLahir      float32    `json:"tb_lahir"`
	LkLahir      float32    `json:"lk_lahir"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
	SourceData   string     `json:"source_data"`
	StatusAktif  *string    `json:"status_aktif"`
	UpdatedBy    *string    `json:"updated_by"`
	DeletedBy    *string    `json:"deleted_by"`
}

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
}

// 3. Struct Kunjungan
type Kunjungan struct {
	Id                 string     `json:"id"`
	IdAnak             string     `json:"id_anak"`
	TanggalPengukuran  string     `json:"tanggal_pengukuran"`
	CaraUkur           *string    `json:"cara_ukur"`
	BeratBadan         *float64   `json:"berat_badan"`
	TinggiBadan        *float64   `json:"tinggi_badan"`
	LingkarLengan      *float64   `json:"lingkar_lengan"`
	LingkarKepala      *float64   `json:"lingkar_kepala"`
	LingkarDada        *float64   `json:"lingkar_dada"`
	AsiBulan0          *int16     `json:"asi_bulan_0"`
	AsiBulan1          *int16     `json:"asi_bulan_1"`
	AsiBulan2          *int16     `json:"asi_bulan_2"`
	AsiBulan3          *int16     `json:"asi_bulan_3"`
	AsiBulan4          *int16     `json:"asi_bulan_4"`
	AsiBulan5          *int16     `json:"asi_bulan_5"`
	AsiBulan6          *int16     `json:"asi_bulan_6"`
	VitBiru            *int16     `json:"vit_biru"`
	VitMerah           *int16     `json:"vit_merah"`
	PittingEdema       *int16     `json:"pitting_edema"`
	KelasIbuBalita     *int16     `json:"kelas_ibu_balita"`
	StatusBbuSigizi    *string    `json:"status_bbu_sigizi"`
	StatusTbuSigizi    *string    `json:"status_tbu_sigizi"`
	StatusBbtbSigizi   *string    `json:"status_bbtb_sigizi"`
	ZscoreBbuSigizi    *float64   `json:"zscore_bbu_sigizi"`
	ZscoreTbuSigizi    *float64   `json:"zscore_tbu_sigizi"`
	ZscoreBbtbSigizi   *float64   `json:"zscore_bbtb_sigizi"`
	StatusBbuWhoantro  *string    `json:"status_bbu_whoantro"`
	StatusTbuWhoantro  *string    `json:"status_tbu_whoantro"`
	StatusBbtbWhoantro *string    `json:"status_bbtb_whoantro"`
	ZscoreBbuWhoantro  *float64   `json:"zscore_bbu_whoantro"`
	ZscoreTbuWhoantro  *float64   `json:"zscore_tbu_whoantro"`
	ZscoreBbtbWhoantro *float64   `json:"zscore_bbtb_whoantro"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at"`
	SourceData         *string    `json:"source_data"`
	UpdatedBy          *string    `json:"updated_by"`
	DeletedBy          *string    `json:"deleted_by"`
}

type KunjunganAnak struct {
	IdOrangtua *string   `json:"id_orangtua"`
	IdAnak     *string   `json:"id_anak"`
	Orangtua   *Orangtua `json:"orangtua"`
	Anak       *Anak     `json:"anak"`
	Kunjungan  Kunjungan `json:"kunjungan"`
}

type KunjunganAnakArray struct {
	// Orangtua  Orangtua    `json:"orangtua"`
	Anak      Anak        `json:"anak"`
	Kunjungan []Kunjungan `json:"kunjungan"`
}

type ListAnak struct {
	Orangtua
	Anak []Anak `json:"anak"`
}

type OrangtuaPayload struct {
	Id           string     `json:"id"`
	IdPosyandu   string     `json:"id_posyandu"`
	NoKk         string     `json:"no_kk"`
	NamaAyah     string     `json:"nama_ayah"`
	NamaIbu      string     `json:"nama_ibu"`
	Nik          string     `json:"nik"`
	Telepon      string     `json:"telepon"`
	Rt           string     `json:"rt"`
	Rw           string     `json:"rw"`
	Alamat       string     `json:"alamat"`
	Kia          int16      `json:"kia"`
	SourceData   *string    `json:"source_data"`
	UsiaHamil    *int16     `json:"usia_hamil"`
	KiaBayiKecil *int16     `json:"kia_bayi_kecil"`
	UpdatedBy    *string    `json:"updated_by"`
	DeletedBy    *string    `json:"deleted_by"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

type AnakPayload struct {
	Id           string     `json:"id"`
	IDOrangtua   string     `json:"id_orangtua"`
	Nama         string     `json:"nama"`
	NIK          *string    `json:"nik"`
	TanggalLahir string     `json:"tanggal_lahir"` // format YYYY-MM-DD
	JenisKelamin string     `json:"jenis_kelamin"`
	AnakKe       int16      `json:"anak_ke"`
	IMD          int16      `json:"imd"`
	BBLahir      float32    `json:"bb_lahir"`
	TBLahir      float32    `json:"tb_lahir"`
	LKLahir      float32    `json:"lk_lahir"`
	SourceData   string     `json:"source_data"`
	StatusAktif  *string    `json:"status_aktif"`
	UpdatedBy    *string    `json:"updated_by"`
	DeletedBy    *string    `json:"deleted_by"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
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
	CreatedAt         *time.Time `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
}
type KunjunganPayload struct {
	Id                 string     `json:"id"`
	IDAnak             string     `json:"id_anak"`
	TanggalPengukuran  string     `json:"tanggal_pengukuran"`
	CaraUkur           *string    `json:"cara_ukur"`
	BeratBadan         *float64   `json:"berat_badan"`
	TinggiBadan        *float64   `json:"tinggi_badan"`
	LingkarLengan      *float64   `json:"lingkar_lengan"`
	LingkarKepala      *float64   `json:"lingkar_kepala"`
	LingkarDada        *float64   `json:"lingkar_dada"`
	ASIBulan0          *int16     `json:"asi_bulan_0"`
	ASIBulan1          *int16     `json:"asi_bulan_1"`
	ASIBulan2          *int16     `json:"asi_bulan_2"`
	ASIBulan3          *int16     `json:"asi_bulan_3"`
	ASIBulan4          *int16     `json:"asi_bulan_4"`
	ASIBulan5          *int16     `json:"asi_bulan_5"`
	ASIBulan6          *int16     `json:"asi_bulan_6"`
	VitBiru            *int16     `json:"vit_biru"`
	VitMerah           *int16     `json:"vit_merah"`
	PittingEdema       *int16     `json:"pitting_edema"`
	KelasIbuBalita     *int16     `json:"kelas_ibu_balita"`
	StatusBBUSigizi    *string    `json:"status_bbu_sigizi"`
	StatusTBUSigizi    *string    `json:"status_tbu_sigizi"`
	StatusBBTBSigizi   *string    `json:"status_bbtb_sigizi"`
	ZScoreBBUSigizi    *float64   `json:"zscore_bbu_sigizi"`
	ZScoreTBUSigizi    *float64   `json:"zscore_tbu_sigizi"`
	ZScoreBBTBSigizi   *float64   `json:"zscore_bbtb_sigizi"`
	StatusBBUWhoAntro  *string    `json:"status_bbu_whoantro"`
	StatusTBUWhoAntro  *string    `json:"status_tbu_whoantro"`
	StatusBBTBWhoAntro *string    `json:"status_bbtb_whoantro"`
	ZScoreBBUWhoAntro  *float64   `json:"zscore_bbu_whoantro"`
	ZScoreTBUWhoAntro  *float64   `json:"zscore_tbu_whoantro"`
	ZScoreBBTBWhoAntro *float64   `json:"zscore_bbtb_whoantro"`
	SourceData         *string    `json:"source_data"`
	UpdatedBy          *string    `json:"updated_by"`
	DeletedBy          *string    `json:"deleted_by"`
	CreatedAt          *time.Time `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at"`
}

type KunjunganNestedPayload struct {
	Orangtua  *OrangtuaPayload  `json:"orangtua"`
	Anak      *AnakPayload      `json:"anak"`
	Kunjungan *KunjunganPayload `json:"kunjungan"`

	// Kunjungan only without nested
	KunjunganPayload
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
	}

	// timestamp dari jakantro; kalau tidak dikirim, database yang mengisi
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}
	e.UpdatedAt = p.UpdatedAt
	e.DeletedAt = p.DeletedAt

	return e
}

func (p *KunjunganPayload) ToEntity() *entity.Kunjungan {
	if p == nil {
		return nil
	}

	tPengukuran, _ := time.Parse("2006-01-02", p.TanggalPengukuran)
	e := &entity.Kunjungan{
		ID:                 p.Id,
		IDAnak:             p.IDAnak,
		TanggalPengukuran:  tPengukuran,
		CaraUkur:           p.CaraUkur,
		BeratBadan:         p.BeratBadan,
		TinggiBadan:        p.TinggiBadan,
		LingkarLengan:      p.LingkarLengan,
		LingkarKepala:      p.LingkarKepala,
		LingkarDada:        p.LingkarDada,
		ASIBulan0:          p.ASIBulan0,
		ASIBulan1:          p.ASIBulan1,
		ASIBulan2:          p.ASIBulan2,
		ASIBulan3:          p.ASIBulan3,
		ASIBulan4:          p.ASIBulan4,
		ASIBulan5:          p.ASIBulan5,
		ASIBulan6:          p.ASIBulan6,
		VitBiru:            p.VitBiru,
		VitMerah:           p.VitMerah,
		PittingEdema:       p.PittingEdema,
		KelasIbuBalita:     p.KelasIbuBalita,
		StatusBBUSigizi:    p.StatusBBUSigizi,
		StatusTBUSigizi:    p.StatusTBUSigizi,
		StatusBBTBSigizi:   p.StatusBBTBSigizi,
		ZScoreBBUSigizi:    p.ZScoreBBUSigizi,
		ZScoreTBUSigizi:    p.ZScoreTBUSigizi,
		ZScoreBBTBSigizi:   p.ZScoreBBTBSigizi,
		StatusBBUWhoAntro:  p.StatusBBUWhoAntro,
		StatusTBUWhoAntro:  p.StatusTBUWhoAntro,
		StatusBBTBWhoAntro: p.StatusBBTBWhoAntro,
		ZScoreBBUWhoAntro:  p.ZScoreBBUWhoAntro,
		ZScoreTBUWhoAntro:  p.ZScoreTBUWhoAntro,
		ZScoreBBTBWhoAntro: p.ZScoreBBTBWhoAntro,
		SourceData:         p.SourceData,
		UpdatedBy:          p.UpdatedBy,
		DeletedBy:          p.DeletedBy,
	}

	// timestamp dari jakantro; kalau tidak dikirim, database yang mengisi
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}
	e.UpdatedAt = p.UpdatedAt
	e.DeletedAt = p.DeletedAt

	return e
}

func (p *AnakPayload) ToEntity() *entity.Anak {
	if p == nil {
		return nil
	}

	tLahir, _ := time.Parse("2006-01-02", p.TanggalLahir)
	var nik *string
	if p.NIK != nil && strings.TrimSpace(*p.NIK) == "" {
		nik = nil
	} else {
		nik = p.NIK
	}

	e := &entity.Anak{
		ID:           p.Id,
		IDOrangtua:   p.IDOrangtua,
		Nama:         p.Nama,
		NIK:          nik,
		TanggalLahir: tLahir,
		JenisKelamin: p.JenisKelamin,
		AnakKe:       p.AnakKe,
		IMD:          p.IMD,
		BBLahir:      p.BBLahir,
		TBLahir:      p.TBLahir,
		LKLahir:      p.LKLahir,
		SourceData:   p.SourceData,
		StatusAktif:  p.StatusAktif,
		UpdatedBy:    p.UpdatedBy,
		DeletedBy:    p.DeletedBy,
	}

	// timestamp dari jakantro; kalau tidak dikirim, database yang mengisi
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}
	e.UpdatedAt = p.UpdatedAt
	e.DeletedAt = p.DeletedAt

	return e
}

func (p *OrangtuaPayload) ToEntity() *entity.Orangtua {
	if p == nil {
		return nil
	}

	e := &entity.Orangtua{
		ID:           p.Id,
		IDPosyandu:   p.IdPosyandu,
		NoKK:         p.NoKk,
		NamaAyah:     p.NamaAyah,
		NamaIbu:      p.NamaIbu,
		NIK:          p.Nik,
		Telepon:      p.Telepon,
		RT:           p.Rt,
		RW:           p.Rw,
		Alamat:       p.Alamat,
		KIA:          p.Kia,
		SourceData:   p.SourceData,
		UsiaHamil:    p.UsiaHamil,
		KIABayiKecil: p.KiaBayiKecil,
		UpdatedBy:    p.UpdatedBy,
		DeletedBy:    p.DeletedBy,
	}

	// timestamp dari jakantro; kalau tidak dikirim, database yang mengisi
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}
	e.UpdatedAt = p.UpdatedAt
	e.DeletedAt = p.DeletedAt

	return e
}

func (o *Orangtua) FromEntity(e *entity.Orangtua) *Orangtua {
	if e == nil {
		return nil
	}

	// id satusehat diisi puskesmas di database, jadi null untuk anak yang
	// didaftarkan jakantro
	var idSatusehat *string
	if e.SatusehatId != "" {
		idSatusehat = &e.SatusehatId
	}

	return &Orangtua{
		Id:           e.ID,
		IdSatusehat:  idSatusehat,
		IdPosyandu:   e.IDPosyandu,
		NoKk:         e.NoKK,
		NamaAyah:     e.NamaAyah,
		NamaIbu:      e.NamaIbu,
		Nik:          e.NIK,
		Telepon:      e.Telepon,
		Rt:           e.RT,
		Rw:           e.RW,
		Alamat:       e.Alamat,
		Kia:          e.KIA,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
		DeletedAt:    e.DeletedAt,
		SourceData:   e.SourceData,
		UsiaHamil:    e.UsiaHamil,
		KiaBayiKecil: e.KIABayiKecil,
		UpdatedBy:    e.UpdatedBy,
		DeletedBy:    e.DeletedBy,
	}
}

func (a *Anak) FromEntity(e *entity.Anak) *Anak {
	if e == nil {
		return nil
	}

	var idSatusehat *string
	var nik *string
	if e.SatusehatId != "" {
		idSatusehat = &e.SatusehatId
	}

	if e.NIK != nil && strings.TrimSpace(*e.NIK) == "" {
		nik = nil
	} else {
		nik = e.NIK
	}

	return &Anak{
		Id:           e.ID,
		IdSatusehat:  idSatusehat,
		IdOrangtua:   e.IDOrangtua,
		Nama:         e.Nama,
		Nik:          nik,
		TanggalLahir: e.TanggalLahir.Format("2006-01-02"),
		JenisKelamin: e.JenisKelamin,
		AnakKe:       e.AnakKe,
		Imd:          e.IMD,
		BbLahir:      e.BBLahir,
		TbLahir:      e.TBLahir,
		LkLahir:      e.LKLahir,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
		DeletedAt:    e.DeletedAt,
		SourceData:   e.SourceData,
		StatusAktif:  e.StatusAktif,
		UpdatedBy:    e.UpdatedBy,
		DeletedBy:    e.DeletedBy,
	}
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
	}
}

func (kj *Kunjungan) FromEntity(e *entity.Kunjungan) *Kunjungan {
	if e == nil {
		return nil
	}
	return &Kunjungan{
		Id:                 e.ID,
		IdAnak:             e.IDAnak,
		TanggalPengukuran:  e.TanggalPengukuran.Format("2006-01-02"),
		CaraUkur:           e.CaraUkur,
		BeratBadan:         e.BeratBadan,
		TinggiBadan:        e.TinggiBadan,
		LingkarLengan:      e.LingkarLengan,
		LingkarKepala:      e.LingkarKepala,
		LingkarDada:        e.LingkarDada,
		AsiBulan0:          e.ASIBulan0,
		AsiBulan1:          e.ASIBulan1,
		AsiBulan2:          e.ASIBulan2,
		AsiBulan3:          e.ASIBulan3,
		AsiBulan4:          e.ASIBulan4,
		AsiBulan5:          e.ASIBulan5,
		AsiBulan6:          e.ASIBulan6,
		VitBiru:            e.VitBiru,
		VitMerah:           e.VitMerah,
		PittingEdema:       e.PittingEdema,
		KelasIbuBalita:     e.KelasIbuBalita,
		StatusBbuSigizi:    e.StatusBBUSigizi,
		StatusTbuSigizi:    e.StatusTBUSigizi,
		StatusBbtbSigizi:   e.StatusBBTBSigizi,
		ZscoreBbuSigizi:    e.ZScoreBBUSigizi,
		ZscoreTbuSigizi:    e.ZScoreTBUSigizi,
		ZscoreBbtbSigizi:   e.ZScoreBBTBSigizi,
		StatusBbuWhoantro:  e.StatusBBUWhoAntro,
		StatusTbuWhoantro:  e.StatusTBUWhoAntro,
		StatusBbtbWhoantro: e.StatusBBTBWhoAntro,
		ZscoreBbuWhoantro:  e.ZScoreBBUWhoAntro,
		ZscoreTbuWhoantro:  e.ZScoreTBUWhoAntro,
		ZscoreBbtbWhoantro: e.ZScoreBBTBWhoAntro,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
		DeletedAt:          e.DeletedAt,
		SourceData:         e.SourceData,
		UpdatedBy:          e.UpdatedBy,
		DeletedBy:          e.DeletedBy,
	}
}

func (la *ListAnak) FromEntity(e *entity.Orangtua) *ListAnak {
	anak := make([]Anak, 0)

	for _, a := range e.Anak {
		var temp *Anak
		if an := temp.FromEntity(a); an != nil {
			anak = append(anak, *an)
		}

	}
	var temp *Orangtua
	orangtua := temp.FromEntity(e)

	return &ListAnak{
		Orangtua: *orangtua,
		Anak:     anak,
	}
}

func (k *KunjunganAnakArray) FromEntity(e *entity.Anak) *KunjunganAnakArray {
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
