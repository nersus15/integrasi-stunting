package akses_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Cek akses orangtua membaca daftar anaknya dari cache. Test di sini menjaga dua hal:
// daftar itu selalu diperbarui saat anak dibuat/pindah/dihapus, dan saat cache kosong
// yang dibaca dari DB hanya id anak (usulan "daftar id anak", lihat TestCekAksesOrangtuaHanyaMembacaIdAnak).

const (
	ortuDua   = "uji-akses-ortu-dua"
	kakak     = "uji-akses-kakak"
	adik      = "uji-akses-adik"
	ortuRujuk = "uji-akses-ortu-rujuk"
	sulung    = "uji-akses-sulung"
	bungsu    = "uji-akses-bungsu"
)

func insertOrangtua(t *testing.T, id, nik string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO stunting.orangtua (id, id_posyandu, no_kk, nik, nama_ayah, nama_ibu, telepon, rt, rw, alamat, kia, created_by)
		VALUES ($1, 'UJI-POS-A', $2, $3, 'Ayah', 'Ibu', '0811', '001', '001', 'Jl. Uji', 1, 'jakantro')`, id, nik+"1", nik+"2")
	require.NoError(t, err)
}

func insertAnak(t *testing.T, id, idOrtu, nik string, urutan int) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO stunting.anak (id, id_orangtua, nama, nik, tanggal_lahir, jenis_kelamin, anak_ke, imd, bb_lahir, tb_lahir, lk_lahir, created_by)
		VALUES ($1, $2, 'Anak', $3, '2025-01-01', 'L', $4, 1, 3, 50, 33, 'jakantro')`, id, idOrtu, nik, urutan)
	require.NoError(t, err)
}

// baris ditulis lewat SQL, jadi cache-nya dibuang manual; LupakanOrangtua juga harus membuang daftar anaknya
func forgetKeluarga(idOrtu string, anak ...string) {
	wiring.Repository.LupakanOrangtua(&entity.Orangtua{ID: idOrtu})
	for _, id := range anak {
		wiring.Repository.LupakanAnak(&entity.Anak{ID: id, IDOrangtua: idOrtu})
		wiring.Repository.LupakanRujukanTerakhir(id)
	}
}

func cleanupKeluarga(t *testing.T, idOrtu string, anak ...string) {
	t.Helper()
	cleanup := func() {
		for _, id := range anak {
			_, _ = db.Exec(`DELETE FROM stunting.kunjungan WHERE id_anak = $1`, id)
			_, _ = db.Exec(`DELETE FROM stunting.rujukan WHERE id_anak = $1`, id)
			_, _ = db.Exec(`DELETE FROM stunting.anak WHERE id = $1`, id)
		}
		_, _ = db.Exec(`DELETE FROM stunting.orangtua WHERE id = $1`, idOrtu)
		forgetKeluarga(idOrtu, anak...)
	}
	cleanup()
	t.Cleanup(cleanup)
}

// kakak (anak ke-1) belum dirujuk, adik (anak ke-2) dirujuk ke RS luar
func setupKeluargaDuaAnak(t *testing.T) {
	t.Helper()
	removeRujukan(t)
	cleanupKeluarga(t, ortuDua, kakak, adik)
	insertOrangtua(t, ortuDua, "317402000000094")
	insertAnak(t, kakak, ortuDua, "3174020000000943", 1)
	insertAnak(t, adik, ortuDua, "3174020000000944", 2)
	insertRujukan(t, "uji-akses-rj-adik", adik, "rujukan", "UJI-AKSES-PKM-A", "UJI-AKSES-RS-LUAR", "2026-01-01")
}

// satu-satunya anak dirujuk ke RS luar: puskesmas posyandu tidak memegang orangtuanya
func setupKeluargaSemuaDirujuk(t *testing.T) {
	t.Helper()
	removeRujukan(t)
	cleanupKeluarga(t, ortuRujuk, sulung, bungsu)
	insertOrangtua(t, ortuRujuk, "317402000000095")
	insertAnak(t, sulung, ortuRujuk, "3174020000000953", 1)
	insertRujukan(t, "uji-akses-rj-sulung", sulung, "rujukan", "UJI-AKSES-PKM-A", "UJI-AKSES-RS-LUAR", "2026-01-01")
}

// anak pertama belum dirujuk, anak kedua dirujuk ke RS: RS tetap memegang orangtuanya
func TestAksesOrangtuaLewatAnakYangBukanPertama(t *testing.T) {
	setupKeluargaDuaAnak(t)

	assert.NoError(t, accessOrangtua("uji-rs-luar", ortuDua), "RS memegang adik, walau kakak (anak pertama) belum dirujuk")
	assert.NoError(t, accessOrangtua("uji-pkm-a", ortuDua), "puskesmas posyandu memegang kakak")
	assert.Error(t, accessOrangtua("uji-pustu-b", ortuDua))
}

// --- daftar anak diperbarui saat anak dibuat ---

func TestDaftarAnakDiperbaruiSaatAnakDibuatLewatPOSTAnak(t *testing.T) {
	setupKeluargaSemuaDirujuk(t)
	require.Error(t, accessOrangtua("uji-pkm-a", ortuRujuk), "prasyarat: satu-satunya anak sudah dirujuk keluar")

	nik := "3174020000000954"
	_, err := wiring.Service.CreateAnak(&entity.Anak{
		ID: bungsu, IDOrangtua: ortuRujuk, Nama: "Bungsu", NIK: &nik, TanggalLahir: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		JenisKelamin: "P", AnakKe: 2, IMD: new(int16(1)), BBLahir: 3, TBLahir: 49, LKLahir: 33,
	}, "jakantro")
	require.NoError(t, err)

	assert.NoError(t, accessOrangtua("uji-pkm-a", ortuRujuk), "anak baru belum dirujuk, dipegang puskesmas posyandu")
}

func TestDaftarAnakDiperbaruiSaatAnakDibuatLewatRegistrasi(t *testing.T) {
	setupKeluargaSemuaDirujuk(t)
	require.Error(t, accessOrangtua("uji-pkm-a", ortuRujuk), "prasyarat: satu-satunya anak sudah dirujuk keluar")

	body := fmt.Sprintf(`{
		"anak": {"id": %q, "id_orangtua": %q, "nik": "3174020000000955", "nama": "Bungsu", "tanggal_lahir": "2025-06-01",
			"jenis_kelamin": "P", "anak_ke": 2, "imd": 1, "bb_lahir": 3, "tb_lahir": 49, "lk_lahir": 33, "source_data": "kader-01"},
		"kunjungan": {"id": "uji-akses-kunj-bungsu", "id_anak": %q, "tanggal_pengukuran": "2026-08-14", "berat_badan": 7.1}
	}`, bungsu, ortuRujuk, bungsu)
	_, err := wiring.Service.CreateKunjungan([]byte(body))
	require.NoError(t, err)

	assert.NoError(t, accessOrangtua("uji-pkm-a", ortuRujuk), "anak baru belum dirujuk, dipegang puskesmas posyandu")
}

// --- usulan daftar id anak: saat cache kosong hanya id anak yang dibaca ---

// Gagal selama daftar anak masih memakai ListAnak (orangtua + seluruh kolom anak, 2 query).
// Lulus setelah cek akses orangtua memakai cache daftar id anak sendiri (1 query kecil).
func TestCekAksesOrangtuaHanyaMembacaIdAnak(t *testing.T) {
	setupKeluargaDuaAnak(t)
	cases := []struct {
		nama, orgid string
		statement   int
		rincian     string
	}{
		{"puskesmas posyandu", "uji-pkm-a", 3, "orangtua, id anak, rujukan terakhir kakak"},
		{"RS tujuan rujukan adik", "uji-rs-luar", 3, "orangtua, id anak, rujukan terakhir semua anak dalam satu query"},
	}
	for _, k := range cases {
		t.Run(k.nama, func(t *testing.T) {
			require.NoError(t, accessOrangtua(k.orgid, ortuDua), "memanaskan cache master faskes/posyandu")
			forgetKeluarga(ortuDua, kakak, adik)

			q := counter.Measure(func() { require.NoError(t, accessOrangtua(k.orgid, ortuDua)) })
			anak := q.Match(`stunting"?\."?anak"?\b`)
			if assert.Len(t, anak, 1, "satu query ke tabel anak\n%v", q) {
				assert.NotContains(t, anak[0], "nama", "yang dibaca hanya id anak, bukan seluruh kolomnya")
				assert.NotRegexp(t, `(?i)FROM\s+"?stunting"?\."?orangtua"?.*JOIN|"o"\."id_posyandu".*"a"\.`, anak[0])
			}
			assert.Len(t, q, k.statement, "%s\n%v", k.rincian, q)

			assert.Empty(t, counter.Measure(func() { require.NoError(t, accessOrangtua(k.orgid, ortuDua)) }),
				"pemanggilan kedua sepenuhnya dari cache")
		})
	}
}

// rujukan semua anak diambil sekaligus: jumlah query tidak bertambah dengan jumlah anak
func TestCekAksesOrangtuaTanpaQueryPerAnak(t *testing.T) {
	setupKeluargaDuaAnak(t)
	cleanupKeluarga(t, ortuDua, kakak, adik, "uji-akses-tengah")
	insertOrangtua(t, ortuDua, "317402000000094")
	insertAnak(t, kakak, ortuDua, "3174020000000943", 1)
	insertAnak(t, "uji-akses-tengah", ortuDua, "3174020000000945", 2)
	insertAnak(t, adik, ortuDua, "3174020000000944", 3)
	insertRujukan(t, "uji-akses-rj-tengah", "uji-akses-tengah", "rujukan", "UJI-AKSES-PKM-A", "UJI-AKSES-PUSTU-B", "2026-01-01")
	insertRujukan(t, "uji-akses-rj-adik", adik, "rujukan", "UJI-AKSES-PKM-A", "UJI-AKSES-RS-LUAR", "2026-01-01")

	require.NoError(t, accessOrangtua("uji-rs-luar", ortuDua), "memanaskan cache master faskes/posyandu")
	forgetKeluarga(ortuDua, kakak, adik, "uji-akses-tengah")

	q := counter.Measure(func() { require.NoError(t, accessOrangtua("uji-rs-luar", ortuDua), "RS memegang anak terakhir") })
	assert.Len(t, q, 3, "orangtua, id anak, rujukan terakhir ketiga anak dalam satu query\n%v", q)
	assert.Len(t, q.Match(`stunting"?\."?rujukan"?\b`), 1, "satu query rujukan untuk semua anak")

	assert.Empty(t, counter.Measure(func() { require.NoError(t, accessOrangtua("uji-rs-luar", ortuDua)) }),
		"pemanggilan kedua sepenuhnya dari cache, termasuk rujukan terakhir tiap anak")
	assert.NoError(t, accessOrangtua("uji-pustu-b", ortuDua), "PUSTU B memegang anak tengah")
}

// rujukan milik keluarga lain tidak boleh ikut terbaca saat mengecek keluarga ini
func TestRujukanKeluargaLainTidakMemberiAkses(t *testing.T) {
	setupKeluargaDuaAnak(t) // adik keluarga lain dirujuk ke RS luar
	forgetKeluarga("uji-akses-ortu", "uji-akses-anak")

	assert.Error(t, accessOrangtua("uji-rs-luar", "uji-akses-ortu"), "keluarga ini tidak punya anak yang dirujuk ke RS luar")
	assert.NoError(t, accessOrangtua("uji-rs-luar", ortuDua))
}

// rujukan ke faskes yang belum terdaftar (ref terisi, id kosong) tetap memindah akses ke ref itu
func TestRujukanKeFaskesBelumTerdaftarMemindahAkses(t *testing.T) {
	setupKeluargaSemuaDirujuk(t)
	_, err := db.Exec(`INSERT INTO stunting.rujukan (id, id_anak, jenis, id_faskes_asal, ref_faskes_asal, ref_faskes_tujuan, tanggal)
		VALUES ('uji-akses-rj-luar-sistem', $1, 'rujukan', 'UJI-AKSES-PKM-A', 'uji-pkm-a', 'org-belum-terdaftar', '2026-02-01')`, sulung)
	require.NoError(t, err)
	forgetKeluarga(ortuRujuk, sulung)

	assert.Error(t, accessOrangtua("uji-rs-luar", ortuRujuk), "rujukan terakhir sekarang ke faskes lain")
	assert.Error(t, accessAnak("uji-rs-luar", sulung))
}
