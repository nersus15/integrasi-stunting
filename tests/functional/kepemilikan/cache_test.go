package kepemilikan_test

import (
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"

	"github.com/stretchr/testify/assert"
)

func idAnakDiDaftar(t *testing.T, idOrtu string) []string {
	t.Helper()
	res := request(t, "GET", "/orangtua/"+idOrtu+"/anak", testutil.KeyJakantro, nil)
	requireStatus(t, res, 200)
	var id []string
	list, _ := res.Body["anak"].([]any)
	for _, x := range list {
		if m, ok := x.(map[string]any); ok {
			id = append(id, m["id"].(string))
		}
	}
	return id
}

func TestCacheNikLamaDibuangSaatNikAnakDiubah(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)
	nikBaru := "3175049910000077"
	requireStatus(t, request(t, "GET", "/anak?nik="+nikAnakJkt, testutil.KeyJakantro, nil), 200)

	requireStatus(t, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"id": idAnakJkt, "nik": nikBaru}), 200)

	assert.Equal(t, 404, request(t, "GET", "/anak?nik="+nikAnakJkt, testutil.KeyJakantro, nil).Status, "nik lama tidak boleh tersisa di cache")
	res := request(t, "GET", "/anak?nik="+nikBaru, testutil.KeyJakantro, nil)
	requireStatus(t, res, 200)
	assert.Equal(t, idAnakJkt, res.Body["id"])
}

func TestCacheDaftarAnakOrangtuaLamaDibuangSaatAnakPindah(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)
	assert.Contains(t, idAnakDiDaftar(t, idOrtuJkt), idAnakJkt)
	assert.NotContains(t, idAnakDiDaftar(t, idOrtuJkt2), idAnakJkt)

	requireStatus(t, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"id": idAnakJkt, "id_orangtua": idOrtuJkt2}), 200)

	assert.NotContains(t, idAnakDiDaftar(t, idOrtuJkt), idAnakJkt, "daftar orangtua lama")
	assert.Contains(t, idAnakDiDaftar(t, idOrtuJkt2), idAnakJkt, "daftar orangtua baru")
}

func TestCacheDaftarAnakDibuangSeketikaSaatAnakDihapus(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)
	assert.Contains(t, idAnakDiDaftar(t, idOrtuJkt2), idAnakJkt2)

	requireStatus(t, request(t, "DELETE", "/anak/"+idAnakJkt2, testutil.KeyJakantro, nil), 200)

	assert.NotContains(t, idAnakDiDaftar(t, idOrtuJkt2), idAnakJkt2, "tanpa menunggu job latar")
}
