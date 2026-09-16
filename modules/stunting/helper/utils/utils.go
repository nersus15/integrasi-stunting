package utils

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/helper/types"
)

var ErrTidakDitemukan = errors.New("data tidak ditemukan")

var dumpAktif atomic.Bool

func SetDumpAktif(level string) {
	dumpAktif.Store(strings.EqualFold(strings.TrimSpace(level), "debug"))
}

func DumpAktif() bool { return dumpAktif.Load() }

func IsFilled(s *string) bool {
	return s != nil && strings.TrimSpace(*s) != ""
}

func IsStrFilled(s string) bool {
	return strings.TrimSpace(s) != ""
}

func IsDigitsLen(s string, n int) bool {
	s = strings.TrimSpace(s)
	if len(s) != n {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func IsValidDate(d string) bool {
	_, err := time.Parse("2006-01-02", d)
	return err == nil
}

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
	adaFlat := IsStrFilled(p.KunjunganPayload.Id) ||
		IsStrFilled(p.KunjunganPayload.IDAnak) ||
		IsStrFilled(p.KunjunganPayload.TanggalPengukuran)

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

	// Key anak wajib membawa id dan id_orangtua, dan harus rujuk-silang dengan kunjungan.
	if p.Anak != nil {
		if !IsStrFilled(p.Anak.Id) {
			return "", nil, fmt.Errorf("anak.id wajib dikirim")
		}
		if !IsStrFilled(p.Anak.IDOrangtua) {
			return "", nil, fmt.Errorf("anak.id_orangtua wajib dikirim")
		}
		if strings.TrimSpace(p.Anak.Id) != strings.TrimSpace(p.Kunjungan.IDAnak) {
			return "", nil, fmt.Errorf("anak.id (%q) tidak sama dengan kunjungan.id_anak (%q)",
				p.Anak.Id, p.Kunjungan.IDAnak)
		}
	}

	// Key orangtua tidak berdiri sendiri: hanya melengkapi key anak.
	if p.Orangtua != nil {
		if p.Anak == nil {
			return "", nil, fmt.Errorf("key orangtua hanya boleh dikirim bersama key anak")
		}
		if !IsStrFilled(p.Orangtua.Id) {
			return "", nil, fmt.Errorf("orangtua.id wajib dikirim")
		}
		if strings.TrimSpace(p.Orangtua.Id) != strings.TrimSpace(p.Anak.IDOrangtua) {
			return "", nil, fmt.Errorf("orangtua.id (%q) tidak sama dengan anak.id_orangtua (%q)",
				p.Orangtua.Id, p.Anak.IDOrangtua)
		}
	}

	anakBaru := p.Anak != nil && (IsFilled(p.Anak.NIK) || IsStrFilled(p.Anak.Nama))
	ortuBaru := p.Orangtua != nil && (IsStrFilled(p.Orangtua.Nik) || IsStrFilled(p.Orangtua.NoKk))

	switch {
	case ortuBaru && !anakBaru:
		return "", nil, fmt.Errorf("orangtua baru harus disertai data anak baru")
	case ortuBaru:
		return JenisRegistrasiLengkap, p.Kunjungan, nil
	case anakBaru:
		return JenisAnakBaru, p.Kunjungan, nil
	}

	return JenisKunjunganSaja, p.Kunjungan, nil
}

// aturan sama dengan DeteksiJenisPayload, entitas terakhirnya kesehatan
func DeteksiJenisPayloadKesehatan(p *types.KesehatanNestedPayload) (JenisPayload, *types.KesehatanPayload, error) {
	if p == nil {
		return "", nil, fmt.Errorf("payload kosong")
	}

	adaNested := p.Orangtua != nil || p.Anak != nil || p.Kesehatan != nil
	adaFlat := IsStrFilled(p.KesehatanPayload.Id) ||
		IsStrFilled(p.KesehatanPayload.IDAnak) ||
		IsStrFilled(p.KesehatanPayload.TanggalPemantauan)

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

	// Key anak wajib membawa id dan id_orangtua, dan harus rujuk-silang dengan kesehatan.
	if p.Anak != nil {
		if !IsStrFilled(p.Anak.Id) {
			return "", nil, fmt.Errorf("anak.id wajib dikirim")
		}
		if !IsStrFilled(p.Anak.IDOrangtua) {
			return "", nil, fmt.Errorf("anak.id_orangtua wajib dikirim")
		}
		if strings.TrimSpace(p.Anak.Id) != strings.TrimSpace(p.Kesehatan.IDAnak) {
			return "", nil, fmt.Errorf("anak.id (%q) tidak sama dengan kesehatan.id_anak (%q)",
				p.Anak.Id, p.Kesehatan.IDAnak)
		}
	}

	// Key orangtua tidak berdiri sendiri: hanya melengkapi key anak.
	if p.Orangtua != nil {
		if p.Anak == nil {
			return "", nil, fmt.Errorf("key orangtua hanya boleh dikirim bersama key anak")
		}
		if !IsStrFilled(p.Orangtua.Id) {
			return "", nil, fmt.Errorf("orangtua.id wajib dikirim")
		}
		if strings.TrimSpace(p.Orangtua.Id) != strings.TrimSpace(p.Anak.IDOrangtua) {
			return "", nil, fmt.Errorf("orangtua.id (%q) tidak sama dengan anak.id_orangtua (%q)",
				p.Orangtua.Id, p.Anak.IDOrangtua)
		}
	}

	anakBaru := p.Anak != nil && (IsFilled(p.Anak.NIK) || IsStrFilled(p.Anak.Nama))
	ortuBaru := p.Orangtua != nil && (IsStrFilled(p.Orangtua.Nik) || IsStrFilled(p.Orangtua.NoKk))

	switch {
	case ortuBaru && !anakBaru:
		return "", nil, fmt.Errorf("orangtua baru harus disertai data anak baru")
	case ortuBaru:
		return JenisRegistrasiLengkap, p.Kesehatan, nil
	case anakBaru:
		return JenisAnakBaru, p.Kesehatan, nil
	}

	return JenisKesehatanSaja, p.Kesehatan, nil
}

func wajibKesehatan(k *types.KesehatanPayload) error {
	if !IsStrFilled(k.Id) {
		return fmt.Errorf("kesehatan.id wajib dikirim")
	}
	if !IsStrFilled(k.IDAnak) {
		return fmt.Errorf("kesehatan.id_anak wajib dikirim")
	}
	if !IsStrFilled(k.TanggalPemantauan) {
		return fmt.Errorf("kesehatan.tanggal_pemantauan wajib dikirim")
	}
	if !IsValidDate(k.TanggalPemantauan) {
		return fmt.Errorf("kesehatan.tanggal_pemantauan (%q) harus berformat YYYY-MM-DD", k.TanggalPemantauan)
	}
	return nil
}

func wajibKunjungan(k *types.KunjunganPayload) error {
	if !IsStrFilled(k.Id) {
		return fmt.Errorf("kunjungan.id wajib dikirim")
	}
	if !IsStrFilled(k.IDAnak) {
		return fmt.Errorf("kunjungan.id_anak wajib dikirim")
	}
	if !IsStrFilled(k.TanggalPengukuran) {
		return fmt.Errorf("kunjungan.tanggal_pengukuran wajib dikirim")
	}
	if !IsValidDate(k.TanggalPengukuran) {
		return fmt.Errorf("kunjungan.tanggal_pengukuran (%q) harus berformat YYYY-MM-DD", k.TanggalPengukuran)
	}
	return nil
}

// Nilai membaca isi pointer string dengan aman untuk keperluan log.
func Nilai(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func Substr(s string, start, length int) string {
	runes := []rune(s)

	if start < 0 || start >= len(runes) {
		return ""
	}

	end := start + length
	if end > len(runes) {
		end = len(runes)
	}

	return string(runes[start:end])
}

// level wilayah => [0 => Nasional, 1 => Provinsi, 2 => Kab/Kota, 3 => Kecamatan, 4 => Kelurahan, -1 => Invalid]
func LevelWilayah(kodeWilayah *string) int {
	if !IsFilled(kodeWilayah) {
		return -1
	}
	arr := strings.Split(*kodeWilayah, ".")
	level := 0

	for _, s := range arr {
		if s == "00" || s == "0000" {
			continue
		}
		level += 1
	}

	return level
}

// null berarti statusnya belum ditentukan, bukan bukan-stunting
func SmallintToBool(num *int) bool {
	return num != nil && *num == 1
}

func BoolToSmallint(v bool) int {
	if v {
		return 1
	}

	return 0
}

func BoolToSmallintPtr(v bool) *int {
	n := BoolToSmallint(v)
	return &n
}
