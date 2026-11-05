package delete_endpoint_test

import (
	"testing"
	"time"

	"github.com/nersus15/integrasi/tests/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func kunjunganJakantro(t *testing.T) {
	t.Helper()
	res := request(t, "POST", "/kunjungan", testutil.KeyJakantro, map[string]any{
		"id": "f1a2b3c4-0000-4000-8000-0000000000d1", "id_anak": idAnak, "tanggal_pengukuran": "2026-09-01", "berat_badan": 9.1,
	})
	require.Equal(t, 201, res.Status, "%v", res.Body)
	res = request(t, "POST", "/kesehatan", testutil.KeyJakantro, map[string]any{
		"id": "f1a2b3c4-0000-4000-8000-0000000000d2", "id_anak": idAnak, "tanggal_pemantauan": "2026-09-01", "tbc_batuk": 0,
	})
	require.Equal(t, 201, res.Status, "%v", res.Body)
}

func TestGETAnakTerhapus404(t *testing.T) {
	setup(t)
	createAnak(t)
	kunjunganJakantro(t)

	paths := []string{"/anak/" + idAnak, "/anak/" + idAnak + "/summary", "/anak/" + idAnak + "/kunjungan", "/anak/" + idAnak + "/kesehatan"}
	for _, p := range paths {
		require.Equal(t, 200, request(t, "GET", p, testutil.KeyJakantro, nil).Status, "%s sebelum dihapus", p)
	}

	require.Equal(t, 200, request(t, "DELETE", "/anak/"+idAnak, testutil.KeyJakantro, nil).Status)

	for _, p := range paths {
		assert.Eventually(t, func() bool { return request(t, "GET", p, testutil.KeyJakantro, nil).Status == 404 },
			2*time.Second, 50*time.Millisecond, "%s setelah anak dihapus", p)
	}
}

func TestGETDaftarAnakTanpaAnakTerhapus(t *testing.T) {
	setup(t)
	createAnak(t)
	require.Equal(t, 200, request(t, "GET", "/orangtua/"+idOrangtua+"/anak", testutil.KeyJakantro, nil).Status)

	require.Equal(t, 200, request(t, "DELETE", "/anak/"+idAnak, testutil.KeyJakantro, nil).Status)

	assert.Eventually(t, func() bool {
		res := request(t, "GET", "/orangtua/"+idOrangtua+"/anak", testutil.KeyJakantro, nil)
		if res.Status == 404 {
			return true
		}
		list, _ := res.Body["anak"].([]any)
		for _, a := range list {
			if m, ok := a.(map[string]any); ok && m["id"] == idAnak {
				return false
			}
		}
		return res.Status == 200
	}, 2*time.Second, 50*time.Millisecond, "anak terhapus tidak ikut di daftar anak orangtua")
}

func TestGETOrangtuaTerhapus404(t *testing.T) {
	setup(t)
	createOrangtua(t)

	paths := []string{"/orangtua/" + idOrangtua, "/orangtua?nik=3175041503970004", "/orangtua/" + idOrangtua + "/anak"}
	require.Equal(t, 200, request(t, "GET", paths[0], testutil.KeyJakantro, nil).Status)
	require.Equal(t, 200, request(t, "GET", paths[1], testutil.KeyJakantro, nil).Status)

	require.Equal(t, 200, request(t, "DELETE", "/orangtua/"+idOrangtua, testutil.KeyJakantro, nil).Status)

	for _, p := range paths {
		assert.Eventually(t, func() bool { return request(t, "GET", p, testutil.KeyJakantro, nil).Status == 404 },
			2*time.Second, 50*time.Millisecond, "%s setelah orangtua dihapus", p)
	}
}
