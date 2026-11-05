package delete_endpoint_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const idAnak = "b34a5387-919d-4154-823b-05529a1e3dcd"
const idOrangtua = "7736486d-1ddb-4fa1-8a12-9d27e17daeeb"
const idKesehatan = "d5b3f169-8c47-42ae-b91d-6f0a3e28cg52"
const idKunjungan = "t5b3f169-1e24-3f4e-k82c-9f0a3e283k5h"

var (
	app *testutil.App
	db  *sql.DB
)

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		fmt.Fprintln(os.Stderr, "postgres mati, test delete endpoint dilewati")
		os.Exit(0)
	}

	db = testutil.OpenDB()
	if err := testutil.Scenario1Seed(db); err != nil {
		fmt.Fprintf(os.Stderr, "gagal seed: %v\n", err)
		os.Exit(1)
	}

	var err error
	app, err = testutil.StartApp(2090)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	cleanup()
	app.Stop()
	db.Close()
	os.Exit(code)
}

func newAnak() map[string]any {
	return map[string]any{
		"id":            idAnak,
		"id_orangtua":   idOrangtua,
		"nik":           "3175041503240004",
		"nama":          "Functional Test - Delete Endpoint Jakantro",
		"tanggal_lahir": "2024-03-15",
		"jenis_kelamin": "L",
		"anak_ke":       1,
		"imd":           1,
		"bb_lahir":      3.2,
		"tb_lahir":      49.5,
		"lk_lahir":      34.0,
		"source_data":   "jakantro",
	}
}

func newOrangtua() map[string]any {
	return map[string]any{
		"id":          idOrangtua,
		"nik":         "3175041503970004",
		"id_posyandu": "POS-JAKARTA-01",
		"no_kk":       "3175042509870001",
		"nama_ayah":   "Ahmad Suryana",
		"nama_ibu":    "Dewi Lestari",
		"telepon":     "081234567890",
		"rt":          "004",
		"rw":          "007",
		"alamat":      "Jl. Kenanga No. 12",
		"kia":         1,
	}
}

func newKesehatan() map[string]any {
	return map[string]any{

		"id":                 idKesehatan,
		"id_anak":            idAnak,
		"tanggal_pemantauan": "2026-08-14",
		"tbc_batuk":          0,
		"tbc_demam":          0,
		"tbc_bb":             0,
		"tbc_kontak":         0,
		"layanan_asi_eks":    1,
		"layanan_imunisasi":  1,
		"layanan_vit_a":      1,
		"penyuluhan_edukasi": 1,
		"source_data":        "8352cfdd-5b80-425b-b831-51efff2b22df",
	}
}

func newKunjungan() map[string]map[string]any {
	return map[string]map[string]any{
		"kunjungan": {
			"id":                 idKunjungan,
			"id_anak":            idAnak,
			"tanggal_pengukuran": "2026-09-11",
			"berat_badan":        9.8,
			"source_data":        "6dd44d76-84c0-44f7-9f64-55ea3d796dee",
		},
	}
}

func request(t *testing.T, method, path, key string, body any) testutil.Response {
	t.Helper()
	return testutil.Request(t, app.BaseURL+path, method, key, body)
}

func cleanup() {
	_, _ = db.Exec(`DELETE FROM stunting.kunjungan WHERE id = $1`, idKunjungan)
	_, _ = db.Exec(`DELETE FROM stunting.kesehatan WHERE id = $1`, idKesehatan)
	_, _ = db.Exec(`DELETE FROM stunting.anak WHERE id = $1 OR nik = '3175041503240004'`, idAnak)
	_, _ = db.Exec(`DELETE FROM stunting.orangtua WHERE id IN ($1, '7736486d-1ddb-4fa1-8a12-9d27e17dae99') OR nik = '3175041503970004' OR no_kk = '3175042509870001'`, idOrangtua)
}

func setup(t *testing.T) {
	t.Helper()
	cleanup()
	t.Cleanup(cleanup)
}

func createOrangtua(t *testing.T) {
	t.Helper()
	res := request(t, "POST", "/orangtua", testutil.KeyJakantro, newOrangtua())
	require.Equal(t, 201, res.Status, "%v", res.Body)
}

func createAnak(t *testing.T) {
	t.Helper()
	createOrangtua(t)
	res := request(t, "POST", "/anak", testutil.KeyJakantro, newAnak())
	require.Equal(t, 201, res.Status, "%v", res.Body)
}

func createKesehatan(t *testing.T) {
	t.Helper()
	createAnak(t)
	res := request(t, "POST", "/kesehatan", testutil.KeyJakantro, newKesehatan())
	require.Equal(t, 201, res.Status, "%v", res.Body)
}

func createKunjungan(t *testing.T) {
	t.Helper()
	createAnak(t)
	res := request(t, "POST", "/kunjungan", testutil.KeyJakantro, newKunjungan())
	require.Equal(t, 201, res.Status, "%v", res.Body)
}

// penghapus dicatat sebagai orgid, sama dengan updated_by, bukan API key
func deletedBy(t *testing.T, table, id string) string {
	t.Helper()
	var oleh sql.NullString
	require.NoError(t, db.QueryRow(`SELECT deleted_by FROM stunting.`+table+` WHERE id = $1`, id).Scan(&oleh))
	return oleh.String
}

func isDeleted(t *testing.T, table, id string) bool {
	t.Helper()
	var exists bool
	require.NoError(t, db.QueryRow(`SELECT "deletedAt" IS NOT NULL FROM stunting.`+table+` WHERE id = $1`, id).Scan(&exists))
	return exists
}

func TestDeleteAnakSuccess(t *testing.T) {
	setup(t)
	createAnak(t)

	res := request(t, "DELETE", "/anak/"+idAnak, testutil.KeyJakantro, nil)
	require.Equal(t, 200, res.Status, "%v", res.Body)
	assert.Contains(t, res.Body["message"], idAnak)
	assert.True(t, isDeleted(t, "anak", idAnak), "soft delete, barisnya tetap ada")
	assert.Equal(t, "jakantro", deletedBy(t, "anak", idAnak))
}

func TestDeleteAnakTidakMunculLagiDiGET(t *testing.T) {
	setup(t)
	createAnak(t)
	require.Equal(t, 200, request(t, "GET", "/anak/"+idAnak, testutil.KeyJakantro, nil).Status, "mengisi cache dulu")

	require.Equal(t, 200, request(t, "DELETE", "/anak/"+idAnak, testutil.KeyJakantro, nil).Status)

	assert.Eventually(t, func() bool {
		return request(t, "GET", "/anak/"+idAnak, testutil.KeyJakantro, nil).Status == 404
	}, 2*time.Second, 50*time.Millisecond, "cache anak ikut dibersihkan")
}

func TestDeleteAnakDuaKali404(t *testing.T) {
	setup(t)
	createAnak(t)
	require.Equal(t, 200, request(t, "DELETE", "/anak/"+idAnak, testutil.KeyJakantro, nil).Status)
	var waktu string
	require.NoError(t, db.QueryRow(`SELECT "deletedAt"::text FROM stunting.anak WHERE id = $1`, idAnak).Scan(&waktu))

	res := request(t, "DELETE", "/anak/"+idAnak, testutil.KeyJakantro, nil)
	assert.Equal(t, 404, res.Status, "%v", res.Body)
	assert.Equal(t, 2001, res.ErrorCode)

	var after string
	require.NoError(t, db.QueryRow(`SELECT "deletedAt"::text FROM stunting.anak WHERE id = $1`, idAnak).Scan(&after))
	assert.Equal(t, waktu, after, "waktu hapus pertama tidak tertimpa")
}

func TestDeleteAnakTidakAda404(t *testing.T) {
	res := request(t, "DELETE", "/anak/anak-tidak-pernah-ada", testutil.KeyJakantro, nil)
	assert.Equal(t, 404, res.Status, "%v", res.Body)
	assert.Equal(t, 2001, res.ErrorCode)
}

func TestDeleteOrangtuaSuccess(t *testing.T) {
	setup(t)
	createOrangtua(t)

	res := request(t, "DELETE", "/orangtua/"+idOrangtua, testutil.KeyJakantro, nil)
	require.Equal(t, 200, res.Status, "%v", res.Body)
	assert.Contains(t, res.Body["message"], "orang tua")
	assert.NotContains(t, res.Body["message"], "anak")
	assert.Contains(t, res.Body["message"], idOrangtua)
	assert.True(t, isDeleted(t, "orangtua", idOrangtua))
	assert.Equal(t, "jakantro", deletedBy(t, "orangtua", idOrangtua))
}

func TestDeleteOrangtuaTidakMunculLagiDiGET(t *testing.T) {
	setup(t)
	createOrangtua(t)
	require.Equal(t, 200, request(t, "GET", "/orangtua/"+idOrangtua, testutil.KeyJakantro, nil).Status, "mengisi cache dulu")

	require.Equal(t, 200, request(t, "DELETE", "/orangtua/"+idOrangtua, testutil.KeyJakantro, nil).Status)

	assert.Eventually(t, func() bool {
		return request(t, "GET", "/orangtua/"+idOrangtua, testutil.KeyJakantro, nil).Status == 404
	}, 2*time.Second, 50*time.Millisecond, "cache orangtua ikut dibersihkan")
}

func TestDeleteOrangtuaDuaKali404(t *testing.T) {
	setup(t)
	createOrangtua(t)
	require.Equal(t, 200, request(t, "DELETE", "/orangtua/"+idOrangtua, testutil.KeyJakantro, nil).Status)

	res := request(t, "DELETE", "/orangtua/"+idOrangtua, testutil.KeyJakantro, nil)
	assert.Equal(t, 404, res.Status, "%v", res.Body)
	assert.Equal(t, 2001, res.ErrorCode)
}

func TestDeleteOrangtuaTidakAda404(t *testing.T) {
	res := request(t, "DELETE", "/orangtua/ortu-tidak-pernah-ada", testutil.KeyJakantro, nil)
	assert.Equal(t, 404, res.Status, "%v", res.Body)
	assert.Equal(t, 2001, res.ErrorCode)
}

func TestDeleteHanyaUntukJakantro(t *testing.T) {
	setup(t)
	createAnak(t)

	for _, path := range []string{"/anak/" + idAnak, "/orangtua/" + idOrangtua} {
		res := request(t, "DELETE", path, testutil.KeyFaskesA, nil)
		assert.Equal(t, 403, res.Status, "%s: %v", path, res.Body)
		assert.Equal(t, 1008, res.ErrorCode, path)
	}
	assert.False(t, isDeleted(t, "anak", idAnak))
	assert.False(t, isDeleted(t, "orangtua", idOrangtua))
}

func TestDeleteKesehatanSuccess(t *testing.T) {
	setup(t)
	createKesehatan(t)

	res := request(t, "DELETE", "/kesehatan/"+idKesehatan, testutil.KeyJakantro, nil)
	require.Equal(t, 200, res.Status, "%v", res.Body)
	assert.Contains(t, res.Body["message"], idKesehatan)
	assert.True(t, isDeleted(t, "kesehatan", idKesehatan), "soft delete, barisnya tetap ada")
}
func TestDeleteKesehatanTidakMunculLagiDiGET(t *testing.T) {
	setup(t)
	createKesehatan(t)
	require.Equal(t, 200, request(t, "GET", "/kesehatan/"+idKesehatan, testutil.KeyJakantro, nil).Status, "mengisi cache dulu")

	require.Equal(t, 200, request(t, "DELETE", "/kesehatan/"+idKesehatan, testutil.KeyJakantro, nil).Status)

	assert.Eventually(t, func() bool {
		return request(t, "GET", "/kesehatan/"+idKesehatan, testutil.KeyJakantro, nil).Status == 404
	}, 2*time.Second, 50*time.Millisecond, "cache kesehatan ikut dibersihkan")
}

func TestDeleteKesehatanDuaKali404(t *testing.T) {
	setup(t)
	createKesehatan(t)
	require.Equal(t, 200, request(t, "DELETE", "/kesehatan/"+idKesehatan, testutil.KeyJakantro, nil).Status)
	var waktu string
	require.NoError(t, db.QueryRow(`SELECT "deletedAt"::text FROM stunting.kesehatan WHERE id = $1`, idKesehatan).Scan(&waktu))

	res := request(t, "DELETE", "/kesehatan/"+idKesehatan, testutil.KeyJakantro, nil)
	assert.Equal(t, 404, res.Status, "%v", res.Body)
	assert.Equal(t, 2001, res.ErrorCode)

	var after string
	require.NoError(t, db.QueryRow(`SELECT "deletedAt"::text FROM stunting.kesehatan WHERE id = $1`, idKesehatan).Scan(&after))
	assert.Equal(t, waktu, after, "waktu hapus pertama tidak tertimpa")
}

func TestDeleteKesehatanTidakAda404(t *testing.T) {
	res := request(t, "DELETE", "/kesehatan/id-tidak-pernah-ada", testutil.KeyJakantro, nil)
	assert.Equal(t, 404, res.Status, "%v", res.Body)
	assert.Equal(t, 2001, res.ErrorCode)
}

func TestDeleteKunjunganSuccess(t *testing.T) {
	setup(t)
	createKunjungan(t)

	res := request(t, "DELETE", "/kunjungan/"+idKunjungan, testutil.KeyJakantro, nil)
	require.Equal(t, 200, res.Status, "%v", res.Body)
	assert.Contains(t, res.Body["message"], idKunjungan)
	assert.True(t, isDeleted(t, "kunjungan", idKunjungan), "soft delete, barisnya tetap ada")
}
func TestDeleteKunjunganTidakMunculLagiDiGET(t *testing.T) {
	setup(t)
	createKunjungan(t)
	require.Equal(t, 200, request(t, "GET", "/kunjungan/"+idKunjungan, testutil.KeyJakantro, nil).Status, "mengisi cache dulu")

	require.Equal(t, 200, request(t, "DELETE", "/kunjungan/"+idKunjungan, testutil.KeyJakantro, nil).Status)

	assert.Eventually(t, func() bool {
		return request(t, "GET", "/kunjungan/"+idKunjungan, testutil.KeyJakantro, nil).Status == 404
	}, 2*time.Second, 50*time.Millisecond, "cache kunjungan ikut dibersihkan")
}

func TestDeleteKunjunganDuaKali404(t *testing.T) {
	setup(t)
	createKunjungan(t)
	require.Equal(t, 200, request(t, "DELETE", "/kunjungan/"+idKunjungan, testutil.KeyJakantro, nil).Status)
	var waktu string
	require.NoError(t, db.QueryRow(`SELECT "deletedAt"::text FROM stunting.kunjungan WHERE id = $1`, idKunjungan).Scan(&waktu))

	res := request(t, "DELETE", "/kunjungan/"+idKunjungan, testutil.KeyJakantro, nil)
	assert.Equal(t, 404, res.Status, "%v", res.Body)
	assert.Equal(t, 2001, res.ErrorCode)

	var after string
	require.NoError(t, db.QueryRow(`SELECT "deletedAt"::text FROM stunting.kunjungan WHERE id = $1`, idKunjungan).Scan(&after))
	assert.Equal(t, waktu, after, "waktu hapus pertama tidak tertimpa")
}

func TestDeleteKunjunganTidakAda404(t *testing.T) {
	res := request(t, "DELETE", "/kunjungan/id-tidak-pernah-ada", testutil.KeyJakantro, nil)
	assert.Equal(t, 404, res.Status, "%v", res.Body)
	assert.Equal(t, 2001, res.ErrorCode)
}
