package validasi

import (
	"fmt"
	"strings"

	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
)

// JenisPayload menentukan entitas mana yang perlu dicari dan dibuat.
type JenisPayload string

const (
	JenisRegistrasiLengkap JenisPayload = "REGISTRASI_LENGKAP" // orangtua + anak + kunjungan
	JenisAnakBaru          JenisPayload = "ANAK_BARU"          // anak + kunjungan
	JenisKunjunganSaja     JenisPayload = "KUNJUNGAN_SAJA"     // kunjungan
	JenisKesehatanSaja     JenisPayload = "KESEHATAN_SAJA"     // kesehatan
)

func DeteksiJenisPayload(p *types.KunjunganNestedPayload) (JenisPayload, *types.KunjunganPayload, error) {
	if p == nil {
		return "", nil, fmt.Errorf("payload kosong")
	}

	adaNested := p.Orangtua != nil || p.Anak != nil || p.Kunjungan != nil
	adaFlat := utils.IsStrFilled(p.KunjunganPayload.Id) ||
		utils.IsStrFilled(p.KunjunganPayload.IDAnak) ||
		utils.IsStrFilled(p.KunjunganPayload.TanggalPengukuran)

	if adaFlat && adaNested {
		return "", nil, fmt.Errorf("bentuk flat hanya untuk kunjungan saja, jangan disertai key orangtua, anak, maupun kunjungan")
	}

	if adaFlat {
		kunjungan := p.KunjunganPayload
		if err := wajibKunjungan(&kunjungan); err != nil {
			return "", nil, err
		}
		return JenisKunjunganSaja, &kunjungan, nil
	}

	if p.Kunjungan == nil {
		return "", nil, fmt.Errorf("data kunjungan wajib dikirim")
	}
	if err := wajibKunjungan(p.Kunjungan); err != nil {
		return "", nil, err
	}

	jenis, err := jenisIdentitas(p.Orangtua, p.Anak, "kunjungan", p.Kunjungan.IDAnak, JenisKunjunganSaja)
	if err != nil {
		return "", nil, err
	}
	return jenis, p.Kunjungan, nil
}

// aturan sama dengan DeteksiJenisPayload, entitas terakhirnya kesehatan
func DeteksiJenisPayloadKesehatan(p *types.KesehatanNestedPayload) (JenisPayload, *types.KesehatanPayload, error) {
	if p == nil {
		return "", nil, fmt.Errorf("payload kosong")
	}

	adaNested := p.Orangtua != nil || p.Anak != nil || p.Kesehatan != nil
	adaFlat := utils.IsStrFilled(p.KesehatanPayload.Id) ||
		utils.IsStrFilled(p.KesehatanPayload.IDAnak) ||
		utils.IsStrFilled(p.KesehatanPayload.TanggalPemantauan)

	if adaFlat && adaNested {
		return "", nil, fmt.Errorf("bentuk flat hanya untuk kesehatan saja, jangan disertai key orangtua, anak, maupun kesehatan")
	}

	if adaFlat {
		kesehatan := p.KesehatanPayload
		if err := wajibKesehatan(&kesehatan); err != nil {
			return "", nil, err
		}
		return JenisKesehatanSaja, &kesehatan, nil
	}

	if p.Kesehatan == nil {
		return "", nil, fmt.Errorf("data kesehatan wajib dikirim")
	}
	if err := wajibKesehatan(p.Kesehatan); err != nil {
		return "", nil, err
	}

	jenis, err := jenisIdentitas(p.Orangtua, p.Anak, "kesehatan", p.Kesehatan.IDAnak, JenisKesehatanSaja)
	if err != nil {
		return "", nil, err
	}
	return jenis, p.Kesehatan, nil
}

// anak/orangtua hanya pelengkap dan harus saling merujuk
func jenisIdentitas(o *types.OrangtuaPayload, a *types.AnakPayload, entitas, idAnak string, saja JenisPayload) (JenisPayload, error) {
	if a != nil {
		if !utils.IsStrFilled(a.Id) {
			return "", fmt.Errorf("anak.id wajib dikirim")
		}
		if !utils.IsStrFilled(a.IDOrangtua) {
			return "", fmt.Errorf("anak.id_orangtua wajib dikirim")
		}
		if strings.TrimSpace(a.Id) != strings.TrimSpace(idAnak) {
			return "", fmt.Errorf("anak.id (%q) tidak sama dengan %s.id_anak (%q)", a.Id, entitas, idAnak)
		}
	}

	if o != nil {
		if a == nil {
			return "", fmt.Errorf("key orangtua hanya boleh dikirim bersama key anak")
		}
		if !utils.IsStrFilled(o.Id) {
			return "", fmt.Errorf("orangtua.id wajib dikirim")
		}
		if strings.TrimSpace(o.Id) != strings.TrimSpace(a.IDOrangtua) {
			return "", fmt.Errorf("orangtua.id (%q) tidak sama dengan anak.id_orangtua (%q)", o.Id, a.IDOrangtua)
		}
	}

	anakBaru := a != nil && (utils.IsFilled(a.NIK) || utils.IsStrFilled(a.Nama))
	ortuBaru := o != nil && (utils.IsStrFilled(o.Nik) || utils.IsStrFilled(o.NoKk))

	switch {
	case ortuBaru && !anakBaru:
		return "", fmt.Errorf("orangtua baru harus disertai data anak baru")
	case ortuBaru:
		return JenisRegistrasiLengkap, nil
	case anakBaru:
		return JenisAnakBaru, nil
	}
	return saja, nil
}

func wajibKesehatan(k *types.KesehatanPayload) error {
	if !utils.IsStrFilled(k.Id) {
		return fmt.Errorf("kesehatan.id wajib dikirim")
	}
	if !utils.IsStrFilled(k.IDAnak) {
		return fmt.Errorf("kesehatan.id_anak wajib dikirim")
	}
	if !utils.IsStrFilled(k.TanggalPemantauan) {
		return fmt.Errorf("kesehatan.tanggal_pemantauan wajib dikirim")
	}
	if !utils.IsValidDate(k.TanggalPemantauan) {
		return fmt.Errorf("kesehatan.tanggal_pemantauan (%q) harus berformat YYYY-MM-DD", k.TanggalPemantauan)
	}
	return nil
}

func wajibKunjungan(k *types.KunjunganPayload) error {
	if !utils.IsStrFilled(k.Id) {
		return fmt.Errorf("kunjungan.id wajib dikirim")
	}
	if !utils.IsStrFilled(k.IDAnak) {
		return fmt.Errorf("kunjungan.id_anak wajib dikirim")
	}
	if !utils.IsStrFilled(k.TanggalPengukuran) {
		return fmt.Errorf("kunjungan.tanggal_pengukuran wajib dikirim")
	}
	if !utils.IsValidDate(k.TanggalPengukuran) {
		return fmt.Errorf("kunjungan.tanggal_pengukuran (%q) harus berformat YYYY-MM-DD", k.TanggalPengukuran)
	}
	return nil
}
