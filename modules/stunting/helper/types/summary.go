package types

import (
	"github.com/nersus15/integrasi/mod-stunting/entity"
)

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

func belumBerkunjungObservasi(list []Observasi) []Observasi {
	return belumBerkunjung(list, func(v Observasi) *string { return v.IdKunjungan })
}

func belumBerkunjungDiagnosa(list []Diagnosa) []Diagnosa {
	return belumBerkunjung(list, func(v Diagnosa) *string { return v.IdKunjungan })
}

func belumBerkunjungLayanan(list []Layanan) []Layanan {
	return belumBerkunjung(list, func(v Layanan) *string { return v.IdKunjungan })
}

func belumBerkunjungRujukan(list []Rujukan) []Rujukan {
	return belumBerkunjung(list, func(v Rujukan) *string { return v.IdKunjungan })
}
