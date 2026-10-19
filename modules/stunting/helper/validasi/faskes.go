package validasi

import (
	"fmt"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
)

func ValidateWilayahKerja(org string, data any) error {

	return nil
}

func ValidatePemeriksaanFaskes(p *types.PemeriksaanFaskes) error {
	if p == nil {
		return fmt.Errorf("payload kosong")
	}

	if !utils.IsFilled(p.Anak.IdSatusehat) && !utils.IsFilled(p.Anak.Nik) {
		return fmt.Errorf("anak.satusehat_id atau anak.nik wajib diisi salah satu")
	}
	if utils.IsFilled(p.Anak.Nik) && !utils.IsDigitsLen(*p.Anak.Nik, 16) {
		return fmt.Errorf("anak.nik harus 16 digit angka")
	}

	// faskes tidak menentukan id; pencocokan lewat satusehat id
	if err := ValidateKunjunganFaskes(*p.Kunjungan.ToPayload()); err != nil {
		return err
	}

	for i, o := range p.Observasi {
		if err := ValidateObservasiFaskes(o); err != nil {
			return fmt.Errorf("observasi[%d]: %s", i, err.Error())
		}
	}
	for i, d := range p.Diagnosa {
		if err := ValidateDiagnosaFaskes(d); err != nil {
			return fmt.Errorf("diagnosa[%d]: %s", i, err.Error())
		}
		if err := ValidateEnumDiagnosa(d.Jenis, d.Kategori, d.Kritikalitas, d.ClinicalStatus, d.VerificationStatus); err != nil {
			return fmt.Errorf("diagnosa[%d]: %s", i, err.Error())
		}
	}
	for i, l := range p.Layanan {
		if err := ValidateLayananFaskes(l); err != nil {
			return fmt.Errorf("layanan[%d]: %s", i, err.Error())
		}
		if err := ValidateStatusLayanan(l.Jenis, l.Status); err != nil {
			return fmt.Errorf("layanan[%d]: %s", i, err.Error())
		}
	}
	for i, r := range p.Rujukan {
		if err := ValidateRujukanFaskes(r); err != nil {
			return fmt.Errorf("rujukan[%d]: %s", i, err.Error())
		}
	}
	for i, e := range p.Episode {
		if err := ValidateEpisodeFaskes(e); err != nil {
			return fmt.Errorf("episode[%d]: %s", i, err.Error())
		}
	}

	return nil
}

func ValidateKunjunganFaskes(kunjungan types.KunjunganPayload) error {
	if !utils.IsFilled(kunjungan.SatusehatId) {
		return fmt.Errorf("kunjungan.id_satusehat wajib dikirim, kirim setelah data diterima SatuSehat")
	}
	if !utils.IsStrFilled(kunjungan.TanggalPengukuran) {
		return fmt.Errorf("kunjungan.tanggal_pengukuran wajib dikirim")
	}
	if !utils.IsValidDate(kunjungan.TanggalPengukuran) {
		return fmt.Errorf("kunjungan.tanggal_pengukuran (%q) harus berformat YYYY-MM-DD", kunjungan.TanggalPengukuran)
	}
	if kunjungan.TanggalSelesai != nil && !utils.IsValidDate(*kunjungan.TanggalSelesai) {
		return fmt.Errorf("kunjungan.tanggal_selesai (%q) harus berformat YYYY-MM-DD", *kunjungan.TanggalSelesai)
	}

	return nil
}

func ValidateDiagnosaFaskes(diagnosa types.Diagnosa) error {
	if !utils.IsFilled(diagnosa.IdSatusehat) {
		return fmt.Errorf("id_satusehat wajib dikirim, kirim setelah data diterima SatuSehat")
	}
	if !utils.IsStrFilled(diagnosa.Kode) {
		return fmt.Errorf("kode wajib dikirim")
	}
	if diagnosa.Jenis != "" && diagnosa.Jenis != entity.DiagnosaDiagnosis && diagnosa.Jenis != entity.DiagnosaAlergi {
		return fmt.Errorf("jenis harus %q atau %q", entity.DiagnosaDiagnosis, entity.DiagnosaAlergi)
	}

	return nil
}

func ValidateObservasiFaskes(o types.Observasi) error {
	if !utils.IsFilled(o.IdSatusehat) {
		return fmt.Errorf("id_satusehat wajib dikirim, kirim setelah data diterima SatuSehat")
	}
	if !utils.IsStrFilled(o.Kode) {
		return fmt.Errorf("kode wajib dikirim")
	}
	if !utils.IsStrFilled(o.System) {
		return fmt.Errorf("system wajib dikirim, supaya kodenya bisa dipetakan")
	}

	for i, c := range o.Component {
		if err := ValidateObservasiFaskes(c); err != nil {
			return fmt.Errorf("component[%d]: %s", i, err.Error())
		}
	}

	return nil
}

func ValidateLayananFaskes(l types.Layanan) error {
	if !utils.IsFilled(l.IdSatusehat) {
		return fmt.Errorf("id_satusehat wajib dikirim, kirim setelah data diterima SatuSehat")
	}
	if !utils.IsStrFilled(l.Jenis) {
		return fmt.Errorf("jenis wajib dikirim")
	}

	switch l.Jenis {
	case entity.LayananProcedure, entity.LayananMedicationDispense,
		entity.LayananNutritionOrder, entity.LayananImmunization, entity.LayananServiceRequest:
	default:
		return fmt.Errorf("jenis %q tidak dikenali", l.Jenis)
	}

	return nil
}

// jenis boleh kosong: arahnya disimpulkan dari faskes asal dan tujuan
func ValidateRujukanFaskes(r types.Rujukan) error {
	if !utils.IsFilled(r.IdSatusehat) {
		return fmt.Errorf("id_satusehat wajib dikirim, kirim setelah data diterima SatuSehat")
	}

	switch r.Jenis {
	case "", entity.RujukanKeluar, entity.RujukBalik, entity.RujukanInternal:
	default:
		return fmt.Errorf("jenis harus %q, %q, atau %q", entity.RujukanKeluar, entity.RujukBalik, entity.RujukanInternal)
	}
	if err := cekEnum("status", r.Status, statusPermintaan); err != nil {
		return err
	}

	return cekEnum("prioritas", r.Prioritas, prioritasRujukan)
}

func ValidateEpisodeFaskes(e types.Episode) error {
	if !utils.IsFilled(e.IdSatusehat) {
		return fmt.Errorf("id_satusehat wajib dikirim, kirim setelah data diterima SatuSehat")
	}

	return cekEnum("status", e.Status, statusEpisode)
}
