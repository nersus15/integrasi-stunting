package optimasi_query_test

import (
	"testing"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/repository"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Cek akses faskes membaca orangtua dan anak dari cache. Test di sini memastikan cache
// tidak pernah membuat akses bertahan setelah data yang menentukannya berubah.
//
// "Pembaca lambat" ditiru dengan menulis data lama ke cache tepat setelah perubahan:
// itulah yang terjadi bila sebuah GET membaca DB sebelum commit lalu menyimpan hasilnya
// setelah invalidasi.

// sama dengan keyOrangtuaId / keyAnakId di repository/cache.go
func ortuCacheKey(id string) string { return "OrangtuaByID::" + id }

func setupPosyanduLuar(t *testing.T) {
	t.Helper()
	exec(t, `INSERT INTO stunting.faskes (id, satusehat_id, nama, jenis, wilayah, status)
		VALUES ('UJI-OPT-PKM', 'uji-opt-900', 'Puskesmas Uji Luar', 'puskesmas', '91.71.01.0000', 1)`)
	exec(t, `INSERT INTO stunting.posyandu (id, id_puskesmas, nama, rt, rw) VALUES ('UJI-OPT-POS', 'UJI-OPT-PKM', 'Posyandu Uji Luar', '001', '001')`)
}

func revokeWindow() time.Duration { return repository.JedaLupakanUlang + 2*time.Second }

func TestAksesTidakBertahanWalauDataLamaTersimpanLagiSetelahPindahPosyandu(t *testing.T) {
	setupBelumDirujuk(t)
	setupPosyanduLuar(t)
	path := "/anak/" + testutil.Scenario1IdAnak
	require.Equal(t, 200, request(t, "GET", path, testutil.KeyFaskesA, nil).Kode)
	old, err := wiring.Repository.FindOrangTuaById(testutil.Scenario1IdOrtu, false)
	require.NoError(t, err)

	j := request(t, "PUT", "/orangtua", testutil.KeyJakantro, map[string]any{
		"id": testutil.Scenario1IdOrtu, "nik": testutil.Scenario1NikIbu, "nama_ibu": "Rina Marlina", "id_posyandu": "UJI-OPT-POS",
	})
	require.Equal(t, 200, j.Kode, "%s", j.Mentah)
	require.NoError(t, wiring.Memory.Set(ortuCacheKey(testutil.Scenario1IdOrtu), old, 24*time.Hour))

	assert.Eventually(t, func() bool { return request(t, "GET", path, testutil.KeyFaskesA, nil).Kode == 403 },
		revokeWindow(), 100*time.Millisecond, "akses wilayah lama harus tercabut paling lambat setelah JedaLupakanUlang")
}

func TestAksesTidakBertahanWalauDataLamaTersimpanLagiSetelahOrangtuaDihapus(t *testing.T) {
	setupBelumDirujuk(t)
	path := "/anak/" + testutil.Scenario1IdAnak
	require.Equal(t, 200, request(t, "GET", path, testutil.KeyFaskesA, nil).Kode)
	old, err := wiring.Repository.FindOrangTuaById(testutil.Scenario1IdOrtu, false)
	require.NoError(t, err)

	require.Equal(t, 200, request(t, "DELETE", "/orangtua/"+testutil.Scenario1IdOrtu, testutil.KeyJakantro, nil).Kode)
	require.NoError(t, wiring.Memory.Set(ortuCacheKey(testutil.Scenario1IdOrtu), old, 24*time.Hour))

	assert.Eventually(t, func() bool { return request(t, "GET", path, testutil.KeyFaskesA, nil).Kode != 200 },
		revokeWindow(), 100*time.Millisecond, "orangtua terhapus tidak boleh lagi memberi akses wilayah")
}

func TestAksesOrangtuaTerhapusTertutupSeketika(t *testing.T) {
	setupRiwayat(t)
	path := "/anak/" + testutil.Scenario1IdAnak
	require.Equal(t, 200, request(t, "GET", path, testutil.KeyFaskesA, nil).Kode)

	require.Equal(t, 200, request(t, "DELETE", "/orangtua/"+testutil.Scenario1IdOrtu, testutil.KeyJakantro, nil).Kode)
	_, masihAda := cacheExists(ortuCacheKey(testutil.Scenario1IdOrtu))
	assert.False(t, masihAda, "kunci id orangtua harus terhapus sebelum DELETE selesai, bukan menunggu job latar")
}

// pemulihan lewat registrasi harus menghapus cache baris yang dipulihkan, bukan id di payload
func TestRegistrasiMemulihkanOrangtuaMenghapusCacheBarisAsli(t *testing.T) {
	cleanupRegistrasi(t)
	setupPosyanduLuar(t)
	require.Equal(t, 201, request(t, "POST", "/kunjungan", testutil.KeyJakantro, registrasi(1, "3175047710000004")).Kode)
	require.Equal(t, 200, request(t, "GET", "/anak/uji-opt-anak-1", testutil.KeyFaskesA, nil).Kode, "mengisi cache orangtua")

	// sisa cache dari baris yang kemudian terhapus (mis. akibat race saat penghapusan)
	exec(t, `UPDATE stunting.orangtua SET "deletedAt" = now() WHERE id = 'uji-opt-ortu-1'`)
	_, masihAda := cacheExists(ortuCacheKey("uji-opt-ortu-1"))
	require.True(t, masihAda, "prasyarat: cache baris lama masih ada")

	kedua := registrasi(2, "3175047710000004")
	kedua["orangtua"].(map[string]any)["id_posyandu"] = "UJI-OPT-POS"
	j := request(t, "POST", "/kunjungan", testutil.KeyJakantro, kedua)
	require.Equal(t, 201, j.Kode, "%s", j.Mentah)
	require.Equal(t, "uji-opt-ortu-1", j.Isi["id_orangtua"], "baris lama yang dipulihkan")

	assert.Equal(t, 403, request(t, "GET", "/anak/uji-opt-anak-2", testutil.KeyFaskesA, nil).Kode,
		"orangtua dipulihkan ke posyandu di luar wilayah; cache lama tidak boleh memberi akses")
}

func cacheExists(key string) (map[string]any, bool) {
	content := map[string]any{}
	return content, wiring.Memory.Get(key, &content)
}
