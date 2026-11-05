package update_endpoint_test

import (
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouteFaskesAnakDanOrangtuaTidakTertangkapResourceName(t *testing.T) {
	setupRiwayat(t)

	j := request(t, "PUT", "/faskes/anak", testutil.KeyFaskesA, map[string]any{"id": testutil.Scenario1IdAnak, "nama": "Siti Lewat Rute Faskes"})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, testutil.Scenario1IdAnak, j.Isi["id"], "dijawab handler anak, bukan handler resource")
	assert.Equal(t, "Siti Lewat Rute Faskes", scalar[string](t, `SELECT nama FROM stunting.anak WHERE id = $1`, testutil.Scenario1IdAnak))

	j = request(t, "PUT", "/faskes/orangtua", testutil.KeyFaskesA, ortuUpdate(map[string]any{"alamat": "Jl. Rute Faskes"}))
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, testutil.Scenario1IdOrtu, j.Isi["id"], "dijawab handler orangtua, bukan handler resource")
	assert.Equal(t, "Jl. Rute Faskes", scalar[string](t, `SELECT alamat FROM stunting.orangtua WHERE id = $1`, testutil.Scenario1IdOrtu))
}

func TestRouteFaskesResourceNameTetapBerjalan(t *testing.T) {
	setupRiwayat(t)

	j := request(t, "PUT", "/faskes/diagnosa", testutil.KeyFaskesA, map[string]any{"id_satusehat": "cond-pkm-01", "kode": "E45", "display": "lewat resourceName"})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, "lewat resourceName", scalar[string](t, `SELECT display FROM stunting.diagnosa WHERE satusehat_id = 'cond-pkm-01'`))

	for _, r := range []string{"tidak-dikenal", "kesehatan", "pemeriksaan"} {
		j := request(t, "PUT", "/faskes/"+r, testutil.KeyFaskesA, map[string]any{"id_satusehat": "x"})
		assert.Equal(t, 404, j.Kode, "%s: %v", r, j.Isi)
		assert.Equal(t, 2001, j.Galat, r)
	}
}

func TestRouteAnakOrangtuaTerpisahPerJalur(t *testing.T) {
	setupRiwayat(t)

	for _, k := range []struct{ path, kunci string }{
		{"/faskes/anak", testutil.KeyJakantro},
		{"/faskes/orangtua", testutil.KeyJakantro},
		{"/anak", testutil.KeyFaskesA},
		{"/orangtua", testutil.KeyFaskesA},
	} {
		j := request(t, "PUT", k.path, k.kunci, map[string]any{"id": testutil.Scenario1IdAnak})
		assert.Equal(t, 403, j.Kode, "%s: %v", k.path, j.Isi)
		assert.Equal(t, 1008, j.Galat, k.path)
	}
}
