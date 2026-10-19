package types

import (
	"time"
)

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
type KunjunganFaskesPayload struct {
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
	Asi               *int16     `json:"asi"`
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
	UpdatedBy         *string    `json:"updated_by"`
	DeletedBy         *string    `json:"deleted_by"`
	CreatedAt         *time.Time `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
	Stunting          *int       `json:"stunting"`
}
