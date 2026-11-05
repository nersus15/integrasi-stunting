package optimasi_query_test

import (
	"fmt"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NIK unik di seluruh baris orangtua, termasuk yang terhapus, jadi satu SELECT cukup
// untuk memutuskan: insert baru, pakai yang ada, atau pulihkan yang terhapus.

const selectOrangtua = `^\s*SELECT .* FROM "stunting"\."orangtua"`

func registrasi(n int, nikOrtu string) map[string]any {
	ortu := fmt.Sprintf("uji-opt-ortu-%d", n)
	anak := fmt.Sprintf("uji-opt-anak-%d", n)
	return map[string]any{
		"orangtua": map[string]any{
			"id": ortu, "id_posyandu": "POS-JAKARTA-01", "nik": nikOrtu,
			"nama_ayah": "Ayah Uji", "nama_ibu": "Ibu Uji", "telepon": "081200000000",
			"rt": "001", "rw": "002", "alamat": "Jl. Uji", "kia": 1,
		},
		"anak": map[string]any{
			"id": anak, "id_orangtua": ortu, "nik": fmt.Sprintf("317504770000%04d", n), "nama": "Anak Uji",
			"tanggal_lahir": "2025-01-01", "jenis_kelamin": "L", "anak_ke": n, "imd": 1,
			"bb_lahir": 3.0, "tb_lahir": 48.0, "lk_lahir": 33.5, "source_data": "jakantro",
		},
		"kunjungan": map[string]any{
			"id": fmt.Sprintf("uji-opt-kunj-%d", n), "id_anak": anak,
			"tanggal_pengukuran": "2026-03-01", "berat_badan": 7.1, "tinggi_badan": 66.0,
		},
	}
}

func cleanupRegistrasi(t *testing.T) {
	t.Helper()
	cleanupExtras()
	t.Cleanup(cleanupExtras)
}

func TestRegistrasiOrangtuaBaruSatuSelectOrangtua(t *testing.T) {
	cleanupRegistrasi(t)

	var j jawaban
	q := counter.Measure(func() {
		j = request(t, "POST", "/kunjungan", testutil.KeyJakantro, registrasi(1, "3175047710000001"))
	})
	require.Equal(t, 201, j.Kode, "%s", j.Mentah)
	expectQuery(t, q, selectOrangtua, 1, "SELECT orangtua untuk orangtua baru")
	assert.Equal(t, 1, scalar[int](t, `SELECT count(*) FROM stunting.orangtua WHERE nik = '3175047710000001'`))
}

func TestRegistrasiOrangtuaSudahAdaDipakaiUlang(t *testing.T) {
	cleanupRegistrasi(t)
	require.Equal(t, 201, request(t, "POST", "/kunjungan", testutil.KeyJakantro, registrasi(1, "3175047710000002")).Kode)

	var j jawaban
	q := counter.Measure(func() {
		j = request(t, "POST", "/kunjungan", testutil.KeyJakantro, registrasi(2, "3175047710000002"))
	})
	require.Equal(t, 201, j.Kode, "%s", j.Mentah)
	expectQuery(t, q, selectOrangtua, 1, "SELECT orangtua untuk orangtua yang sudah ada")
	assert.Equal(t, "uji-opt-ortu-1", j.Isi["id_orangtua"], "orangtua lama dipakai, bukan dibuat baru")
	assert.Equal(t, 1, scalar[int](t, `SELECT count(*) FROM stunting.orangtua WHERE nik = '3175047710000002'`))
}

func TestRegistrasiOrangtuaTerhapusDipulihkanSatuSelect(t *testing.T) {
	cleanupRegistrasi(t)
	require.Equal(t, 201, request(t, "POST", "/kunjungan", testutil.KeyJakantro, registrasi(1, "3175047710000003")).Kode)
	exec(t, `UPDATE stunting.orangtua SET "deletedAt" = now(), deleted_by = 'uji' WHERE nik = '3175047710000003'`)

	var j jawaban
	q := counter.Measure(func() {
		j = request(t, "POST", "/kunjungan", testutil.KeyJakantro, registrasi(2, "3175047710000003"))
	})
	require.Equal(t, 201, j.Kode, "%s", j.Mentah)
	expectQuery(t, q, selectOrangtua, 1, "SELECT orangtua untuk orangtua terhapus")
	assert.Equal(t, "t|1", scalar[string](t, `SELECT concat_ws('|', bool_and("deletedAt" IS NULL), count(*))
		FROM stunting.orangtua WHERE nik = '3175047710000003'`), "dipulihkan, bukan baris baru")
}
