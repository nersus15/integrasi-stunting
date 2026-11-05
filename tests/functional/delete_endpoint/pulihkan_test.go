package delete_endpoint_test

import (
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const idOrangtuaBaru = "7736486d-1ddb-4fa1-8a12-9d27e17dae99"

func statusOrangtua(t *testing.T) string {
	t.Helper()
	var s string
	require.NoError(t, db.QueryRow(`SELECT concat_ws('|', id, "deletedAt" IS NULL, coalesce(deleted_by,'-'), alamat)
		FROM stunting.orangtua WHERE nik = '3175041503970004'`).Scan(&s))
	return s
}

func TestPOSTOrangtuaNikTerhapusDipulihkan(t *testing.T) {
	setup(t)
	createOrangtua(t)
	require.Equal(t, 200, request(t, "DELETE", "/orangtua/"+idOrangtua, testutil.KeyJakantro, nil).Status)

	o := newOrangtua()
	o["id"], o["alamat"] = idOrangtuaBaru, "Jl. Setelah Dipulihkan"
	res := request(t, "POST", "/orangtua", testutil.KeyJakantro, o)
	require.Equal(t, 201, res.Status, "%v", res.Body)

	assert.Equal(t, idOrangtua, res.Body["id"], "baris lama yang dipulihkan, bukan baris baru")
	assert.Equal(t, idOrangtua+"|t|-|Jl. Setelah Dipulihkan", statusOrangtua(t))
	assert.Equal(t, 0, countRows(t, `SELECT count(*) FROM stunting.orangtua WHERE id = $1`, idOrangtuaBaru))
	assert.Equal(t, 200, request(t, "GET", "/orangtua/"+idOrangtua, testutil.KeyJakantro, nil).Status)
}

func TestPOSTKunjunganBentuk1NikTerhapusDipulihkan(t *testing.T) {
	setup(t)
	createOrangtua(t)
	require.Equal(t, 200, request(t, "DELETE", "/orangtua/"+idOrangtua, testutil.KeyJakantro, nil).Status)

	o := newOrangtua()
	o["id"], o["alamat"] = idOrangtuaBaru, "Jl. Lewat Kunjungan"
	a := newAnak()
	a["id_orangtua"] = idOrangtuaBaru
	res := request(t, "POST", "/kunjungan", testutil.KeyJakantro, map[string]any{
		"orangtua": o, "anak": a,
		"kunjungan": map[string]any{"id": "f1a2b3c4-0000-4000-8000-0000000000d3", "id_anak": idAnak, "tanggal_pengukuran": "2026-09-01"},
	})
	require.Equal(t, 201, res.Status, "%v", res.Body)

	assert.Equal(t, idOrangtua+"|t|-|Jl. Lewat Kunjungan", statusOrangtua(t))
	var idOrtuAnak string
	require.NoError(t, db.QueryRow(`SELECT id_orangtua FROM stunting.anak WHERE id = $1`, idAnak).Scan(&idOrtuAnak))
	assert.Equal(t, idOrangtua, idOrtuAnak, "anak terhubung ke orangtua yang dipulihkan")
}

func TestPOSTOrangtuaNikAktifTidakDiperbarui(t *testing.T) {
	setup(t)
	createOrangtua(t)

	o := newOrangtua()
	o["id"], o["alamat"] = idOrangtuaBaru, "Jl. Tidak Boleh Menimpa"
	res := request(t, "POST", "/orangtua", testutil.KeyJakantro, o)
	require.Equal(t, 201, res.Status, "%v", res.Body)

	assert.Equal(t, idOrangtua, res.Body["id"])
	assert.Equal(t, idOrangtua+"|t|-|Jl. Kenanga No. 12", statusOrangtua(t), "pemulihan hanya untuk data terhapus")
}

func countRows(t *testing.T, q string, args ...any) int {
	t.Helper()
	var n int
	require.NoError(t, db.QueryRow(q, args...).Scan(&n))
	return n
}
