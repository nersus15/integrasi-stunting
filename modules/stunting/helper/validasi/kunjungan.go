package validasi

import (
	"fmt"
	"strings"

	"github.com/nersus15/integrasi/mod-stunting/helper/satusehat"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
)

func ValidateKesehatan(p types.KesehatanPayload) error {
	if !utils.IsStrFilled(p.Id) {
		return fmt.Errorf("id tidak boleh kosong")
	}
	if !utils.IsStrFilled(p.IDAnak) {
		return fmt.Errorf("id_anak tidak boleh kosong")
	}

	if !utils.IsStrFilled(p.TanggalPemantauan) {
		return fmt.Errorf("tanggal_pemantauan tidak boleh kosong")
	}
	if !utils.IsValidDate(p.TanggalPemantauan) {
		return fmt.Errorf("tanggal_pemantauan harus berformat YYYY-MM-DD")
	}

	// nullable, hanya divalidasi jika dikirim
	flags := map[string]*int16{
		"tbc_batuk":           p.TBCBatuk,
		"tbc_demam":           p.TBCDemam,
		"tbc_bb":              p.TBCBB,
		"tbc_kontak":          p.TBCKontak,
		"layanan_asi_eks":     p.LayananASIEks,
		"layanan_mpasi":       p.LayananMPASI,
		"layanan_imunisasi":   p.LayananImunisasi,
		"layanan_vit_a":       p.LayananVitA,
		"layanan_obat_cacing": p.LayananObatCacing,
		"layanan_mt_pangan":   p.LayananMTPangan,
		"penyuluhan_edukasi":  p.PenyuluhanEdukasi,
		"penyuluhan_rujukan":  p.PenyuluhanRujukan,
	}

	for fieldName, val := range flags {
		if val != nil && *val != 0 && *val != 1 {
			return fmt.Errorf("%s harus bernilai 0 atau 1", fieldName)
		}
	}

	return nil
}

func ValidateKunjungan(p types.KunjunganPayload) error {
	if !utils.IsStrFilled(p.Id) {
		return fmt.Errorf("id tidak boleh kosong")
	}
	if !utils.IsStrFilled(p.IDAnak) {
		return fmt.Errorf("id_anak tidak boleh kosong")
	}

	if !utils.IsStrFilled(p.TanggalPengukuran) {
		return fmt.Errorf("tanggal_pengukuran tidak boleh kosong")
	}
	if !utils.IsValidDate(p.TanggalPengukuran) {
		return fmt.Errorf("tanggal_pengukuran harus berformat YYYY-MM-DD")
	}

	// nullable, harus > 0 jika dikirim
	if p.BeratBadan != nil && *p.BeratBadan <= 0 {
		return fmt.Errorf("berat_badan harus lebih dari 0")
	}
	if p.TinggiBadan != nil && *p.TinggiBadan <= 0 {
		return fmt.Errorf("tinggi_badan harus lebih dari 0")
	}
	if p.LingkarKepala != nil && *p.LingkarKepala <= 0 {
		return fmt.Errorf("lingkar_kepala harus lebih dari 0")
	}
	if p.LingkarLengan != nil && *p.LingkarLengan <= 0 {
		return fmt.Errorf("lingkar_lengan harus lebih dari 0")
	}
	if p.LingkarDada != nil && *p.LingkarDada <= 0 {
		return fmt.Errorf("lingkar_dada harus lebih dari 0")
	}

	asiFlags := map[string]*int16{
		"asi_bulan_0":      p.ASIBulan0,
		"asi_bulan_1":      p.ASIBulan1,
		"asi_bulan_2":      p.ASIBulan2,
		"asi_bulan_3":      p.ASIBulan3,
		"asi_bulan_4":      p.ASIBulan4,
		"asi_bulan_5":      p.ASIBulan5,
		"asi_bulan_6":      p.ASIBulan6,
		"vit_biru":         p.VitBiru,
		"vit_merah":        p.VitMerah,
		"pitting_edema":    p.PittingEdema,
		"kelas_ibu_balita": p.KelasIbuBalita,
	}

	for fieldName, val := range asiFlags {
		if val != nil && *val != 0 && *val != 1 {
			return fmt.Errorf("%s harus bernilai 0 atau 1", fieldName)
		}
	}

	return nil
}

// z-score satu indeks dari kunjungan atau observasi, tidak keduanya
func ValidateZscoreSatuSumber(p *types.PemeriksaanFaskes) error {
	if p == nil {
		return nil
	}
	for i, o := range p.Observasi {
		if kolom := satusehat.KolomUkuran(o.System, o.Kode); zscoreDiKunjungan(p.Kunjungan, kolom) {
			indeks := strings.TrimPrefix(kolom, "zscore_")
			return fmt.Errorf("observasi[%d] (%s) sudah membawa %s; kunjungan.zscore_%s dan kunjungan.status_%s jangan dikirim bersamaan",
				i, o.Kode, kolom, indeks, indeks)
		}
	}
	return nil
}

func zscoreDiKunjungan(k types.Kunjungan, kolom string) bool {
	switch kolom {
	case "zscore_bbu":
		return k.ZscoreBbu != nil || utils.IsFilled(k.StatusBbu)
	case "zscore_tbu":
		return k.ZscoreTbu != nil || utils.IsFilled(k.StatusTbu)
	case "zscore_bbtb":
		return k.ZscoreBbtb != nil || utils.IsFilled(k.StatusBbtb)
	}
	return false
}
