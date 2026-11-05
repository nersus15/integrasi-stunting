package update_endpoint_test

import (
	"fmt"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ditemukan lewat benchmark: PUT jakantro menulis satusehat_id = '' untuk baris yang belum punya
// satusehat_id, sehingga baris kedua bentrok di unique constraint.

func ortuTanpaSatusehat(n int) map[string]any {
	return map[string]any{
		"id": fmt.Sprintf("uji-ss-kosong-ortu-%d", n), "id_posyandu": "POS-JAKARTA-01",
		"nik": fmt.Sprintf("31750499000001%02d", n), "nama_ayah": "Ayah", "nama_ibu": "Ibu",
		"telepon": "081200000000", "rt": "001", "rw": "002", "alamat": "Jl. Uji", "kia": 1,
	}
}

func anakTanpaSatusehat(n int) map[string]any {
	return map[string]any{
		"id": fmt.Sprintf("uji-ss-kosong-anak-%d", n), "id_orangtua": fmt.Sprintf("uji-ss-kosong-ortu-%d", n),
		"nik": fmt.Sprintf("31750499000002%02d", n), "nama": "Anak", "tanggal_lahir": "2025-01-01",
		"jenis_kelamin": "L", "anak_ke": 1, "imd": 1, "bb_lahir": 3.0, "tb_lahir": 48.0, "lk_lahir": 33.5,
		"source_data": "jakantro",
	}
}

func setupTanpaSatusehat(t *testing.T) {
	t.Helper()
	cleanup := func() {
		_, _ = db.Exec(`DELETE FROM stunting.anak WHERE id LIKE 'uji-ss-kosong-%'`)
		_, _ = db.Exec(`DELETE FROM stunting.orangtua WHERE id LIKE 'uji-ss-kosong-%'`)
	}
	cleanup()
	t.Cleanup(cleanup)
	for n := 1; n <= 2; n++ {
		j := request(t, "POST", "/orangtua", testutil.KeyJakantro, ortuTanpaSatusehat(n))
		require.Equal(t, 201, j.Kode, "%v", j.Isi)
		j = request(t, "POST", "/anak", testutil.KeyJakantro, anakTanpaSatusehat(n))
		require.Equal(t, 201, j.Kode, "%v", j.Isi)
	}
}

func TestPUTOrangtuaJakantroTanpaSatusehatTetapNull(t *testing.T) {
	setupTanpaSatusehat(t)

	for n := 1; n <= 2; n++ {
		o := ortuTanpaSatusehat(n)
		j := request(t, "PUT", "/orangtua", testutil.KeyJakantro, map[string]any{
			"id": o["id"], "nik": o["nik"], "nama_ibu": "Ibu Ubah", "telepon": "081299990000",
		})
		require.Equal(t, 200, j.Kode, "orangtua ke-%d: %v", n, j.Isi)
	}
	assert.Equal(t, 2, scalar[int](t, `SELECT count(*) FROM stunting.orangtua WHERE id LIKE 'uji-ss-kosong-%' AND satusehat_id IS NULL`))
}

func TestPUTAnakJakantroTanpaSatusehatTetapNull(t *testing.T) {
	setupTanpaSatusehat(t)

	for n := 1; n <= 2; n++ {
		a := anakTanpaSatusehat(n)
		j := request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{
			"id": a["id"], "nik": a["nik"], "nama": "Anak Ubah", "tanggal_lahir": "2025-01-01", "jenis_kelamin": "L",
		})
		require.Equal(t, 200, j.Kode, "anak ke-%d: %v", n, j.Isi)
	}
	assert.Equal(t, 2, scalar[int](t, `SELECT count(*) FROM stunting.anak WHERE id LIKE 'uji-ss-kosong-%' AND satusehat_id IS NULL`))
}
