package update_endpoint_test

import (
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const idKesehatan = "a1b2c3d4-4444-4aaa-8bbb-000000000004"

func newKunjungan(bb float64) map[string]any {
	return map[string]any{
		"id":                 testutil.Scenario1IdKunjungan,
		"id_anak":            testutil.Scenario1IdAnak,
		"tanggal_pengukuran": "2026-01-15",
		"berat_badan":        bb,
	}
}

func setupKesehatan(t *testing.T) {
	t.Helper()
	j := request(t, "POST", "/kesehatan", testutil.KeyJakantro, map[string]any{
		"id": idKesehatan, "id_anak": testutil.Scenario1IdAnak, "tanggal_pemantauan": "2026-01-15",
		"tbc_batuk": 0, "layanan_asi_eks": 1,
	})
	require.Equal(t, 201, j.Kode, "%v", j.Isi)
}

func TestPUTKunjunganJakantro(t *testing.T) {
	setupRiwayat(t)

	j := request(t, "PUT", "/kunjungan/"+testutil.Scenario1IdKunjungan, testutil.KeyJakantro, newKunjungan(7.3))
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.EqualValues(t, 7.3, j.Isi["berat_badan"])
	assert.EqualValues(t, 43.8, j.Isi["lingkar_kepala"], "respons dari baris gabungan")
	assert.Equal(t, "7.3/43.8/Sangat Pendek", scalar[string](t,
		`SELECT berat_badan || '/' || lingkar_kepala || '/' || status_tbu FROM stunting.kunjungan WHERE id = $1`, testutil.Scenario1IdKunjungan))

	j = request(t, "PUT", "/kunjungan/"+testutil.Scenario1IdKunjungan, testutil.KeyFaskesA, newKunjungan(7.1))
	assert.Equal(t, 403, j.Kode)

	j = request(t, "PUT", "/kunjungan/tidak-ada", testutil.KeyJakantro, newKunjungan(7.1))
	assert.Equal(t, 404, j.Kode, "%v", j.Isi)

	j = request(t, "PUT", "/kunjungan/"+testutil.Scenario1IdKunjungan, testutil.KeyJakantro, `{rusak`)
	assert.Equal(t, 400, j.Kode)

	tanpaTanggal := newKunjungan(7.1)
	delete(tanpaTanggal, "tanggal_pengukuran")
	j = request(t, "PUT", "/kunjungan/"+testutil.Scenario1IdKunjungan, testutil.KeyJakantro, tanpaTanggal)
	assert.Equal(t, 422, j.Kode)
}

func TestPUTKesehatanJakantro(t *testing.T) {
	setupRiwayat(t)
	setupKesehatan(t)

	body := map[string]any{"id": idKesehatan, "id_anak": testutil.Scenario1IdAnak, "tanggal_pemantauan": "2026-01-15", "tbc_batuk": 1}
	j := request(t, "PUT", "/kesehatan/"+idKesehatan, testutil.KeyJakantro, body)
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.EqualValues(t, 1, j.Isi["tbc_batuk"])
	assert.EqualValues(t, 1, j.Isi["layanan_asi_eks"], "field yang tidak dikirim dipertahankan")
	assert.Equal(t, "1/1", scalar[string](t, `SELECT tbc_batuk || '/' || layanan_asi_eks FROM stunting.kesehatan WHERE id = $1`, idKesehatan))

	assert.Equal(t, 403, request(t, "PUT", "/kesehatan/"+idKesehatan, testutil.KeyFaskesA, body).Kode)
	assert.Equal(t, 404, request(t, "PUT", "/kesehatan/tidak-ada", testutil.KeyJakantro, body).Kode)
	assert.Equal(t, 400, request(t, "PUT", "/kesehatan/"+idKesehatan, testutil.KeyJakantro, `{rusak`).Kode)

	delete(body, "tanggal_pemantauan")
	assert.Equal(t, 422, request(t, "PUT", "/kesehatan/"+idKesehatan, testutil.KeyJakantro, body).Kode)
}

func newAnak(name string) map[string]any {
	return map[string]any{
		"id": testutil.Scenario1IdAnak, "id_orangtua": testutil.Scenario1IdOrtu, "nik": testutil.Scenario1NikAnak,
		"nama": name, "tanggal_lahir": "2024-05-20", "jenis_kelamin": "P", "anak_ke": 1, "imd": 1,
		"bb_lahir": 2.9, "tb_lahir": 48.0, "lk_lahir": 33.5, "source_data": "uji",
	}
}

func TestPUTAnak(t *testing.T) {
	setupRiwayat(t)

	j := request(t, "PUT", "/anak", testutil.KeyJakantro, newAnak("Siti Aminah J"))
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, "Siti Aminah J", scalar[string](t, `SELECT nama FROM stunting.anak WHERE id = $1`, testutil.Scenario1IdAnak))
	assert.Equal(t, testutil.Scenario1IhsAnak, scalar[string](t, `SELECT satusehat_id FROM stunting.anak WHERE id = $1`, testutil.Scenario1IdAnak),
		"satusehat_id dari stream tidak terhapus")

	j = request(t, "PUT", "/faskes/anak", testutil.KeyFaskesA, newAnak("Siti Aminah F"))
	assert.Equal(t, 200, j.Kode, "puskesmas pembina posyandu boleh: %v", j.Isi)

	j = request(t, "PUT", "/faskes/anak", keyRS, newAnak("Siti Aminah RS"))
	assert.Equal(t, 403, j.Kode, "RS yang sudah merujuk balik tidak boleh")

	assert.Equal(t, 400, request(t, "PUT", "/anak", testutil.KeyJakantro, `{rusak`).Kode)
	assert.Equal(t, 422, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"nama": "tanpa id"}).Kode)
	assert.Equal(t, 422, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"id": testutil.Scenario1IdAnak, "nik": "123"}).Kode)
	assert.Equal(t, 404, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"id": "tidak-ada"}).Kode)

	ortuHantu := newAnak("x")
	ortuHantu["id_orangtua"] = "00000000-0000-0000-0000-000000000000"
	assert.Equal(t, 1005, request(t, "PUT", "/anak", testutil.KeyJakantro, ortuHantu).Galat, "jakantro boleh memindah anak, tapi orangtuanya harus ada")

	assert.Equal(t, 403, request(t, "PUT", "/anak", keyWithoutGroup, newAnak("x")).Kode)
}
