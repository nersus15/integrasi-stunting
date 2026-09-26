package types

import (
	"strings"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
)

type Orangtua struct {
	Id           string     `json:"id"`
	IdSatusehat  *string    `json:"id_satusehat"`
	IdPosyandu   *string    `json:"id_posyandu"`
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

type Kunjungan struct {
	Id                 string     `json:"id"`
	IdAnak             string     `json:"id_anak"`
	TanggalPengukuran  string     `json:"tanggal_pengukuran"`
	TanggalSelesai     *string    `json:"tanggal_selesai"`
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
	IdFaskes           *string    `json:"id_faskes"`
	IdSatusehat        *string    `json:"id_satusehat"`
	IdEpisode          *string    `json:"id_episode"`
	RefEpisode         *string    `json:"ref_episode"`
	IdRujukan          *string    `json:"id_rujukan"`
	RefRujukan         *string    `json:"ref_rujukan"`
	Stunting           *int       `json:"stunting"`
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

type Faskes struct {
	Id          string     `json:"id"`
	IdInduk     *string    `json:"id_induk"`
	SatusehatId *string    `json:"satusehat_id"`
	Nama        string     `json:"nama"`
	Jenis       string     `json:"jenis"`
	Wilayah     *string    `json:"wilayah"`
	Alamat      *string    `json:"alamat"`
	NomorTelpon *string    `json:"nomor_telpon"`
	Email       *string    `json:"email"`
	Status      *int16     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type FaskesDetail struct {
	Id          string     `json:"id"`
	Induk       *Faskes    `json:"induk"`
	SatusehatId *string    `json:"satusehat_id"`
	Nama        string     `json:"nama"`
	Jenis       string     `json:"jenis"`
	Wilayah     *string    `json:"wilayah"`
	Alamat      *string    `json:"alamat"`
	NomorTelpon *string    `json:"nomor_telpon"`
	Email       *string    `json:"email"`
	Status      *int16     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type Posyandu struct {
	Id          string     `json:"id"`
	IdPuskesmas *string    `json:"id_puskesmas"`
	Nama        string     `json:"nama"`
	Telepon     *string    `json:"telepon"`
	Alamat      *string    `json:"alamat"`
	IdKelurahan *string    `json:"id_kelurahan"`
	Rt          string     `json:"rt"`
	Rw          string     `json:"rw"`
	NamaPic     *string    `json:"nama_pic"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

// faskes beserta posyandu di bawahnya, field faskes di-flatten ke root
type ListPosyandu struct {
	Faskes
	Posyandu []Posyandu `json:"posyandu"`
}

// posyandu beserta puskesmas induknya, puskesmas bisa nil
type PosyanduDetail struct {
	Posyandu
	Puskesmas *FaskesDetail `json:"puskesmas"`
}

type ListAnak struct {
	Orangtua
	Anak []Anak `json:"anak"`
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

type SummaryAnak struct {
	Anak      Anak               `json:"anak"`
	Episode   []Episode          `json:"episode"`
	Kunjungan []KunjunganRiwayat `json:"kunjungan"`
	Kesehatan []Kesehatan        `json:"kesehatan"`

	// diulang datar lintas kunjungan supaya jejaknya bisa dibaca sekali lihat
	Layanan []Layanan        `json:"layanan"`
	Rujukan []RujukanRiwayat `json:"rujukan"`

	Menggantung Menggantung `json:"menggantung"`
}

type KunjunganRiwayat struct {
	Kunjungan
	Faskes  *Faskes  `json:"faskes"`
	Episode *Episode `json:"episode"`
	// rujukan yang dipenuhi kunjungan ini, beda dari Rujukan yang diterbitkan di sini
	AtasRujukan *Rujukan    `json:"atas_rujukan"`
	Observasi   []Observasi `json:"observasi"`
	Diagnosa    []Diagnosa  `json:"diagnosa"`
	Layanan     []Layanan   `json:"layanan"`
	Rujukan     []Rujukan   `json:"rujukan"`
}

// Tindakan kosong berarti rujukan belum ditindaklanjuti.
type RujukanRiwayat struct {
	Rujukan
	Tindakan []KunjunganRingkas `json:"tindakan"`
}

type KunjunganRingkas struct {
	Id                string  `json:"id"`
	TanggalPengukuran string  `json:"tanggal_pengukuran"`
	IdFaskes          *string `json:"id_faskes"`
	Faskes            *Faskes `json:"faskes"`
}

// sudah tiba tapi Encounter-nya belum
type Menggantung struct {
	Observasi []Observasi `json:"observasi"`
	Diagnosa  []Diagnosa  `json:"diagnosa"`
	Layanan   []Layanan   `json:"layanan"`
	Rujukan   []Rujukan   `json:"rujukan"`
}

// riwayat satu anak dari posyandu, puskesmas, dan RS
type PemeriksaanFaskes struct {
	Anak      Anak        `json:"anak"`
	Kunjungan Kunjungan   `json:"kunjungan"`
	Observasi []Observasi `json:"observasi"`
	Diagnosa  []Diagnosa  `json:"diagnosa"`
	Layanan   []Layanan   `json:"layanan"`
	Rujukan   []Rujukan   `json:"rujukan"`
	Episode   []Episode   `json:"episode"`
}

type HasilPemeriksaan struct {
	IdAnak      string  `json:"id_anak"`
	IdKunjungan *string `json:"id_kunjungan"`
}

type OrangtuaPayload struct {
	Id           string     `json:"id"`
	SatusehatId  *string    `json:"satusehat_id"`
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
	SatusehatId  *string    `json:"satusehat_id"`
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

type KunjunganPayload struct {
	Id                 string     `json:"id"`
	SatusehatId        *string    `json:"id_satusehat"`
	IDAnak             string     `json:"id_anak"`
	IdFaskes           *string    `json:"id_faskes"`
	TanggalPengukuran  string     `json:"tanggal_pengukuran"`
	TanggalSelesai     *string    `json:"tanggal_selesai"`
	IdEpisode          *string    `json:"id_episode"`
	RefEpisode         *string    `json:"ref_episode"`
	IdRujukan          *string    `json:"id_rujukan"`
	RefRujukan         *string    `json:"ref_rujukan"`
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
	Stunting           *int       `json:"stunting"`
}

type KunjunganNestedPayload struct {
	Orangtua  *OrangtuaPayload  `json:"orangtua"`
	Anak      *AnakPayload      `json:"anak"`
	Kunjungan *KunjunganPayload `json:"kunjungan"`

	// Kunjungan only without nested
	KunjunganPayload
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

type KesehatanNestedPayload struct {
	Orangtua  *OrangtuaPayload  `json:"orangtua"`
	Anak      *AnakPayload      `json:"anak"`
	Kesehatan *KesehatanPayload `json:"kesehatan"`

	// Kesehatan only without nested
	KesehatanPayload
}

type FaskesPayload struct {
	Id          string     `json:"id"`
	IdInduk     *string    `json:"id_induk"`
	SatusehatId *string    `json:"satusehat_id"`
	Nama        string     `json:"nama"`
	Jenis       string     `json:"jenis"`
	Wilayah     *string    `json:"wilayah"`
	Alamat      *string    `json:"alamat"`
	NomorTelpon *string    `json:"nomor_telpon"`
	Email       *string    `json:"email"`
	Status      *int16     `json:"status"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type PosyanduPayload struct {
	Id          string     `json:"id"`
	IdPuskesmas *string    `json:"id_puskesmas"`
	Nama        string     `json:"nama"`
	Telepon     *string    `json:"telepon"`
	Alamat      *string    `json:"alamat"`
	IdKelurahan *string    `json:"id_kelurahan"`
	Rt          string     `json:"rt"`
	Rw          string     `json:"rw"`
	NamaPic     *string    `json:"nama_pic"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

func (o *Orangtua) FromEntity(e *entity.Orangtua) *Orangtua {
	if e == nil {
		return nil
	}

	// null untuk anak yang didaftarkan jakantro
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

func (p *Orangtua) ToPayload() *OrangtuaPayload {
	if p == nil {
		return nil
	}

	idPosyandu := ""
	if p.IdPosyandu != nil {
		idPosyandu = *p.IdPosyandu
	}

	return &OrangtuaPayload{
		SatusehatId:  p.IdSatusehat,
		Id:           p.Id,
		IdPosyandu:   idPosyandu,
		NoKk:         p.NoKk,
		NamaAyah:     p.NamaAyah,
		NamaIbu:      p.NamaIbu,
		Nik:          p.Nik,
		Telepon:      p.Telepon,
		Rt:           p.Rt,
		Rw:           p.Rw,
		Alamat:       p.Alamat,
		Kia:          p.Kia,
		SourceData:   p.SourceData,
		UsiaHamil:    p.UsiaHamil,
		KiaBayiKecil: p.KiaBayiKecil,
		UpdatedBy:    p.UpdatedBy,
		DeletedBy:    p.DeletedBy,
		CreatedAt:    &p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		DeletedAt:    p.DeletedAt,
	}
}

func (p *OrangtuaPayload) ToEntity() *entity.Orangtua {
	if p == nil {
		return nil
	}

	var idPosyandu *string
	if p.IdPosyandu != "" {
		idPosyandu = &p.IdPosyandu
	}

	e := &entity.Orangtua{
		ID:           p.Id,
		IDPosyandu:   idPosyandu,
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

	if p.SatusehatId != nil {
		e.SatusehatId = *p.SatusehatId
	}

	// timestamp dari jakantro; kalau tidak dikirim, database yang mengisi
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}
	e.UpdatedAt = p.UpdatedAt
	e.DeletedAt = p.DeletedAt

	return e
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

func (p *Anak) ToPayload() *AnakPayload {
	if p == nil {
		return nil
	}

	var nik *string
	if p.Nik != nil && strings.TrimSpace(*p.Nik) == "" {
		nik = nil
	} else {
		nik = p.Nik
	}

	e := &AnakPayload{
		Id:           p.Id,
		SatusehatId:  p.IdSatusehat,
		IDOrangtua:   p.IdOrangtua,
		Nama:         p.Nama,
		NIK:          nik,
		TanggalLahir: p.TanggalLahir,
		JenisKelamin: p.JenisKelamin,
		AnakKe:       p.AnakKe,
		IMD:          p.Imd,
		BBLahir:      p.BbLahir,
		TBLahir:      p.TbLahir,
		LKLahir:      p.LkLahir,
		SourceData:   p.SourceData,
		StatusAktif:  p.StatusAktif,
		UpdatedBy:    p.UpdatedBy,
		DeletedBy:    p.DeletedBy,
		CreatedAt:    &p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		DeletedAt:    p.DeletedAt,
	}

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

	// jakantro tidak mengirim satusehat_id, kolomnya ditulis null
	satusehat := ""
	if p.SatusehatId != nil {
		satusehat = strings.TrimSpace(*p.SatusehatId)
	}

	e := &entity.Anak{
		ID:           p.Id,
		SatusehatId:  satusehat,
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

func (kj *Kunjungan) FromEntity(e *entity.Kunjungan) *Kunjungan {
	if e == nil {
		return nil
	}
	return &Kunjungan{
		Id:                 e.ID,
		IdAnak:             e.IDAnak,
		TanggalPengukuran:  e.TanggalPengukuran.Format("2006-01-02"),
		TanggalSelesai:     teksTanggal(e.TanggalSelesai),
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
		IdFaskes:           e.IDFaskes,
		IdSatusehat:        e.SatusehatId,
		IdEpisode:          e.IDEpisode,
		RefEpisode:         e.RefEpisode,
		IdRujukan:          e.IDRujukan,
		RefRujukan:         e.RefRujukan,
		Stunting:           e.Stunting,
	}
}

func (p *Kunjungan) ToPayload() *KunjunganPayload {
	if p == nil {
		return nil
	}

	return &KunjunganPayload{
		Id:                 p.Id,
		IDAnak:             p.IdAnak,
		SatusehatId:        p.IdSatusehat,
		IdFaskes:           p.IdFaskes,
		CreatedAt:          &p.CreatedAt,
		DeletedAt:          p.DeletedAt,
		TanggalPengukuran:  p.TanggalPengukuran,
		TanggalSelesai:     p.TanggalSelesai,
		IdEpisode:          p.IdEpisode,
		RefEpisode:         p.RefEpisode,
		IdRujukan:          p.IdRujukan,
		RefRujukan:         p.RefRujukan,
		CaraUkur:           p.CaraUkur,
		BeratBadan:         p.BeratBadan,
		TinggiBadan:        p.TinggiBadan,
		LingkarLengan:      p.LingkarLengan,
		LingkarKepala:      p.LingkarKepala,
		LingkarDada:        p.LingkarDada,
		ASIBulan0:          p.AsiBulan0,
		ASIBulan1:          p.AsiBulan1,
		ASIBulan2:          p.AsiBulan2,
		ASIBulan3:          p.AsiBulan3,
		ASIBulan4:          p.AsiBulan4,
		ASIBulan5:          p.AsiBulan5,
		ASIBulan6:          p.AsiBulan6,
		VitBiru:            p.VitBiru,
		VitMerah:           p.VitMerah,
		PittingEdema:       p.PittingEdema,
		KelasIbuBalita:     p.KelasIbuBalita,
		StatusBBUSigizi:    p.StatusBbuSigizi,
		StatusTBUSigizi:    p.StatusTbuSigizi,
		StatusBBTBSigizi:   p.StatusBbtbSigizi,
		ZScoreBBUSigizi:    p.ZscoreBbuSigizi,
		ZScoreTBUSigizi:    p.ZscoreTbuSigizi,
		ZScoreBBTBSigizi:   p.ZscoreBbtbSigizi,
		StatusBBUWhoAntro:  p.StatusBbuWhoantro,
		StatusTBUWhoAntro:  p.StatusTbuWhoantro,
		StatusBBTBWhoAntro: p.StatusBbtbWhoantro,
		ZScoreBBUWhoAntro:  p.ZscoreBbuWhoantro,
		ZScoreTBUWhoAntro:  p.ZscoreTbuWhoantro,
		ZScoreBBTBWhoAntro: p.ZscoreBbtbWhoantro,
		SourceData:         p.SourceData,
		UpdatedBy:          p.UpdatedBy,
		DeletedBy:          p.DeletedBy,
		UpdatedAt:          p.UpdatedAt,
		Stunting:           p.Stunting,
	}
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
		Stunting:           p.Stunting,
		SatusehatId:        p.SatusehatId,
		IDFaskes:           p.IdFaskes,
		TanggalSelesai:     waktuDariString(p.TanggalSelesai),
		IDEpisode:          p.IdEpisode,
		RefEpisode:         p.RefEpisode,
		IDRujukan:          p.IdRujukan,
		RefRujukan:         p.RefRujukan,
	}

	// timestamp dari jakantro; kalau tidak dikirim, database yang mengisi
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}
	e.UpdatedAt = p.UpdatedAt
	e.DeletedAt = p.DeletedAt

	return e
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
	}

	// timestamp dari jakantro; kalau tidak dikirim, database yang mengisi
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}
	e.UpdatedAt = p.UpdatedAt
	e.DeletedAt = p.DeletedAt

	return e
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

func (f *Faskes) FromEntity(e *entity.Faskes) *Faskes {
	if e == nil {
		return nil
	}

	return &Faskes{
		Id:          e.ID,
		IdInduk:     e.IDInduk,
		SatusehatId: e.SatusehatID,
		Nama:        e.Nama,
		Jenis:       e.Jenis,
		Wilayah:     e.Wilayah,
		Alamat:      e.Alamat,
		NomorTelpon: e.NomorTelpon,
		Email:       e.Email,
		Status:      e.Status,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func (p *FaskesPayload) ToEntity() *entity.Faskes {
	if p == nil {
		return nil
	}

	e := &entity.Faskes{
		ID:          p.Id,
		IDInduk:     p.IdInduk,
		SatusehatID: p.SatusehatId,
		Nama:        p.Nama,
		Jenis:       p.Jenis,
		Wilayah:     p.Wilayah,
		Alamat:      p.Alamat,
		NomorTelpon: p.NomorTelpon,
		Email:       p.Email,
		Status:      p.Status,
		UpdatedAt:   p.UpdatedAt,
	}

	// Kalau tidak dikirim, database yang mengisi lewat default.
	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}

	return e
}

func (f *FaskesDetail) FromEntity(e *entity.Faskes) *FaskesDetail {
	if e == nil {
		return nil
	}
	var induk *Faskes
	return &FaskesDetail{
		Id:          e.ID,
		Induk:       induk.FromEntity(e.Induk),
		SatusehatId: e.SatusehatID,
		Nama:        e.Nama,
		Jenis:       e.Jenis,
		Wilayah:     e.Wilayah,
		Alamat:      e.Alamat,
		NomorTelpon: e.NomorTelpon,
		Email:       e.Email,
		Status:      e.Status,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func (p *Posyandu) FromEntity(e *entity.Posyandu) *Posyandu {
	if e == nil {
		return nil
	}

	return &Posyandu{
		Id:          e.ID,
		IdPuskesmas: e.IDPuskesmas,
		Nama:        e.Nama,
		Telepon:     e.Telepon,
		Alamat:      e.Alamat,
		IdKelurahan: e.IDKelurahan,
		Rt:          e.RT,
		Rw:          e.RW,
		NamaPic:     e.NamaPic,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
		DeletedAt:   e.DeletedAt,
	}
}

func (p *PosyanduPayload) ToEntity() *entity.Posyandu {
	if p == nil {
		return nil
	}

	e := &entity.Posyandu{
		ID:          p.Id,
		IDPuskesmas: p.IdPuskesmas,
		Nama:        p.Nama,
		Telepon:     p.Telepon,
		Alamat:      p.Alamat,
		IDKelurahan: p.IdKelurahan,
		RT:          p.Rt,
		RW:          p.Rw,
		NamaPic:     p.NamaPic,
		UpdatedAt:   p.UpdatedAt,
		DeletedAt:   p.DeletedAt,
	}

	if p.CreatedAt != nil {
		e.CreatedAt = *p.CreatedAt
	}

	return e
}

func (l *ListPosyandu) FromEntity(e *entity.Faskes) *ListPosyandu {
	if e == nil {
		return nil
	}

	var f *Faskes
	return &ListPosyandu{
		Faskes:   *f.FromEntity(e),
		Posyandu: daftarPosyandu(e.Posyandu),
	}
}

func (d *PosyanduDetail) FromEntity(e *entity.Posyandu) *PosyanduDetail {
	if e == nil {
		return nil
	}

	var p *Posyandu
	var f *FaskesDetail
	return &PosyanduDetail{
		Posyandu:  *p.FromEntity(e),
		Puskesmas: f.FromEntity(e.Puskesmas),
	}
}

func (la *ListAnak) FromEntity(e *entity.Orangtua) *ListAnak {
	if e == nil {
		return nil
	}

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

func (s *SummaryAnak) FromEntity(e *entity.Anak) *SummaryAnak {
	if e == nil {
		return nil
	}

	var a *Anak
	anak := a.FromEntity(e)

	// disaring di sini, jadi tidak perlu query kedua
	semuaLayanan := daftarLayanan(e.Layanan)
	semuaRujukan := daftarRujukan(e.Rujukan)

	// tindak lanjut dikumpulkan dari kunjungan yang sudah dimuat, tanpa query lagi
	tindakan := map[string][]KunjunganRingkas{}
	for _, k := range e.Kunjungan {
		if k == nil || k.IDRujukan == nil {
			continue
		}
		var f *Faskes
		tindakan[*k.IDRujukan] = append(tindakan[*k.IDRujukan], KunjunganRingkas{
			Id:                k.ID,
			TanggalPengukuran: k.TanggalPengukuran.Format("2006-01-02"),
			IdFaskes:          k.IDFaskes,
			Faskes:            f.FromEntity(k.Faskes),
		})
	}

	riwayatRujukan := make([]RujukanRiwayat, 0, len(semuaRujukan))
	for _, r := range semuaRujukan {
		riwayatRujukan = append(riwayatRujukan, RujukanRiwayat{
			Rujukan:  r,
			Tindakan: append([]KunjunganRingkas{}, tindakan[r.Id]...),
		})
	}

	res := &SummaryAnak{
		Anak:      *anak,
		Episode:   daftarEpisode(e.Episode),
		Kunjungan: make([]KunjunganRiwayat, 0, len(e.Kunjungan)),
		Kesehatan: daftarKesehatan(e.Kesehatan),
		Layanan:   semuaLayanan,
		Rujukan:   riwayatRujukan,
		Menggantung: Menggantung{
			Observasi: belumBerkunjungObservasi(daftarObservasi(e.Observasi)),
			Diagnosa:  belumBerkunjungDiagnosa(daftarDiagnosa(e.Diagnosa)),
			Layanan:   belumBerkunjungLayanan(semuaLayanan),
			Rujukan:   belumBerkunjungRujukan(semuaRujukan),
		},
	}

	for _, k := range e.Kunjungan {
		var tmp *Kunjungan
		dasar := tmp.FromEntity(k)
		if dasar == nil {
			continue
		}

		var f *Faskes
		var ep *Episode
		var rj *Rujukan
		res.Kunjungan = append(res.Kunjungan, KunjunganRiwayat{
			Kunjungan:   *dasar,
			Faskes:      f.FromEntity(k.Faskes),
			Episode:     ep.FromEntity(k.Episode),
			AtasRujukan: rj.FromEntity(k.AtasRujukan),
			Observasi:   daftarObservasi(k.Observasi),
			Diagnosa:    daftarDiagnosa(k.Diagnosa),
			Layanan:     daftarLayanan(k.Layanan),
			Rujukan:     daftarRujukan(k.Rujukan),
		})
	}

	return res
}

func daftarObservasi(list []*entity.Observasi) []Observasi {
	out := make([]Observasi, 0, len(list))
	for _, e := range list {
		var tmp *Observasi
		if t := tmp.FromEntity(e); t != nil {
			out = append(out, *t)
		}
	}
	return out
}

func daftarDiagnosa(list []*entity.Diagnosa) []Diagnosa {
	out := make([]Diagnosa, 0, len(list))
	for _, e := range list {
		var tmp *Diagnosa
		if t := tmp.FromEntity(e); t != nil {
			out = append(out, *t)
		}
	}
	return out
}

func daftarLayanan(list []*entity.Layanan) []Layanan {
	out := make([]Layanan, 0, len(list))
	for _, e := range list {
		var tmp *Layanan
		if t := tmp.FromEntity(e); t != nil {
			out = append(out, *t)
		}
	}
	return out
}

func daftarRujukan(list []*entity.Rujukan) []Rujukan {
	out := make([]Rujukan, 0, len(list))
	for _, e := range list {
		var tmp *Rujukan
		if t := tmp.FromEntity(e); t != nil {
			out = append(out, *t)
		}
	}
	return out
}

func daftarEpisode(list []*entity.Episode) []Episode {
	out := make([]Episode, 0, len(list))
	for _, e := range list {
		var tmp *Episode
		if t := tmp.FromEntity(e); t != nil {
			out = append(out, *t)
		}
	}
	return out
}

// selalu tidak-nil supaya JSON-nya [] bukan null
func daftarKunjungan(list []*entity.Kunjungan) []Kunjungan {
	out := make([]Kunjungan, 0, len(list))
	for _, e := range list {
		var tmp *Kunjungan
		if t := tmp.FromEntity(e); t != nil {
			out = append(out, *t)
		}
	}
	return out
}

func daftarKesehatan(list []*entity.Kesehatan) []Kesehatan {
	out := make([]Kesehatan, 0, len(list))
	for _, e := range list {
		var tmp *Kesehatan
		if t := tmp.FromEntity(e); t != nil {
			out = append(out, *t)
		}
	}
	return out
}

// selalu tidak-nil supaya JSON-nya [] bukan null
func daftarPosyandu(list []*entity.Posyandu) []Posyandu {
	out := make([]Posyandu, 0, len(list))
	for _, e := range list {
		var tmp *Posyandu
		if t := tmp.FromEntity(e); t != nil {
			out = append(out, *t)
		}
	}
	return out
}

func belumBerkunjungObservasi(list []Observasi) []Observasi {
	out := make([]Observasi, 0)
	for _, v := range list {
		if v.IdKunjungan == nil {
			out = append(out, v)
		}
	}
	return out
}

func belumBerkunjungDiagnosa(list []Diagnosa) []Diagnosa {
	out := make([]Diagnosa, 0)
	for _, v := range list {
		if v.IdKunjungan == nil {
			out = append(out, v)
		}
	}
	return out
}

func belumBerkunjungLayanan(list []Layanan) []Layanan {
	out := make([]Layanan, 0)
	for _, v := range list {
		if v.IdKunjungan == nil {
			out = append(out, v)
		}
	}
	return out
}

func belumBerkunjungRujukan(list []Rujukan) []Rujukan {
	out := make([]Rujukan, 0)
	for _, v := range list {
		if v.IdKunjungan == nil {
			out = append(out, v)
		}
	}
	return out
}

// kebalikan waktuDariString, supaya bentuk JSON-nya sama dengan yang masuk
func teksWaktu(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func teksTanggal(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

// waktuDariString menerima dateTime FHIR maupun tanggal saja.
func waktuDariString(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, *s); err == nil {
			return &t
		}
	}
	return nil
}
