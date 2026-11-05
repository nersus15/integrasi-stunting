package kepemilikan_test

import (
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"

	"github.com/stretchr/testify/assert"
)

func TestJakantroPUTAnakBuatanFaskesMengunciNikDanOrangtua(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)
	idOrtu, idAnak := createKeluargaFaskes(t)

	res := request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{
		"id": idAnak, "nama": "Nama Dari Jakantro", "nik": "3175049910000088", "id_orangtua": idOrtuJkt, "anak_ke": 2,
	})
	requireStatus(t, res, 200)

	assert.Equal(t, "Nama Dari Jakantro", column(t, "anak", "nama", idAnak))
	assert.Equal(t, "2", column(t, "anak", "anak_ke", idAnak))
	assert.Equal(t, nikAnakFks, column(t, "anak", "nik", idAnak), "nik anak buatan faskes tidak boleh diubah jakantro")
	assert.Equal(t, idOrtu, column(t, "anak", "id_orangtua", idAnak), "anak buatan faskes tidak boleh dipindah jakantro")
}

func TestJakantroPUTOrangtuaBuatanFaskesMengunciNik(t *testing.T) {
	setup(t)
	idOrtu, _ := createKeluargaFaskes(t)

	res := request(t, "PUT", "/orangtua", testutil.KeyJakantro, map[string]any{
		"id": idOrtu, "nik": "3175049910000089", "nama_ibu": "Ibu Diubah Jakantro", "telepon": "0899",
	})
	requireStatus(t, res, 200)

	assert.Equal(t, "Ibu Diubah Jakantro", column(t, "orangtua", "nama_ibu", idOrtu))
	assert.Equal(t, "0899", column(t, "orangtua", "telepon", idOrtu))
	assert.Equal(t, nikOrtuFks, column(t, "orangtua", "nik", idOrtu), "nik orangtua buatan faskes tidak boleh diubah jakantro")
}

func TestJakantroPUTAnakSendiriBolehUbahNikDanOrangtua(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)

	requireStatus(t, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{
		"id": idAnakJkt, "nik": "3175049910000077", "id_orangtua": idOrtuJkt2, "anak_ke": 2,
	}), 200)

	assert.Equal(t, "3175049910000077", column(t, "anak", "nik", idAnakJkt))
	assert.Equal(t, idOrtuJkt2, column(t, "anak", "id_orangtua", idAnakJkt))
}

func TestJakantroDELETEAnakBuatanFaskesDitolak(t *testing.T) {
	setup(t)
	_, idAnak := createKeluargaFaskes(t)

	requireForbidden(t, request(t, "DELETE", "/anak/"+idAnak, testutil.KeyJakantro, nil))
	assert.False(t, isDeleted(t, "anak", idAnak))
}

func TestJakantroDELETEOrangtuaBuatanFaskesDitolak(t *testing.T) {
	setup(t)
	idOrtu, _ := createKeluargaFaskes(t)

	requireForbidden(t, request(t, "DELETE", "/orangtua/"+idOrtu, testutil.KeyJakantro, nil))
	assert.False(t, isDeleted(t, "orangtua", idOrtu))
}

func TestJakantroPUTKunjunganBuatanFaskesDitolak(t *testing.T) {
	setup(t)
	idAnak, idKunjungan := createPemeriksaanFaskes(t)

	requireForbidden(t, request(t, "PUT", "/kunjungan/"+idKunjungan, testutil.KeyJakantro, map[string]any{
		"id": idKunjungan, "id_anak": idAnak, "tanggal_pengukuran": "2026-03-09", "berat_badan": 7.7,
	}))
	assert.Equal(t, "2026-03-05", column(t, "kunjungan", "tanggal_pengukuran", idKunjungan))
}

func TestJakantroDELETEKunjunganBuatanFaskesDitolak(t *testing.T) {
	setup(t)
	_, idKunjungan := createPemeriksaanFaskes(t)

	requireForbidden(t, request(t, "DELETE", "/kunjungan/"+idKunjungan, testutil.KeyJakantro, nil))
	assert.False(t, isDeleted(t, "kunjungan", idKunjungan))
}

func TestJakantroPUTKesehatanMengisiUpdatedBy(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)

	requireStatus(t, request(t, "PUT", "/kesehatan/"+idSehatJkt, testutil.KeyJakantro, map[string]any{
		"id": idSehatJkt, "id_anak": idAnakJkt, "tanggal_pemantauan": "2026-03-01", "tbc_demam": 1,
	}), 200)
	assert.Equal(t, "1", column(t, "kesehatan", "tbc_demam", idSehatJkt))
	assert.Equal(t, "jakantro", column(t, "kesehatan", "updated_by", idSehatJkt))
}

func TestKunjunganDanKesehatanTerhapusHilangDariRiwayat(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)

	riwayat := func(path string) map[string]any {
		res := request(t, "GET", "/anak/"+idAnakJkt+path, testutil.KeyJakantro, nil)
		requireStatus(t, res, 200)
		return res.Body
	}
	exists := func(content map[string]any, key, id string) bool {
		list, _ := content[key].([]any)
		for _, x := range list {
			if m, ok := x.(map[string]any); ok && m["id"] == id {
				return true
			}
		}
		return false
	}

	assert.True(t, exists(riwayat("/kunjungan"), "kunjungan", idKunjJkt))
	assert.True(t, exists(riwayat("/kesehatan"), "kesehatan", idSehatJkt))
	ringkas := riwayat("/summary")
	assert.True(t, exists(ringkas, "kunjungan", idKunjJkt))
	assert.True(t, exists(ringkas, "kesehatan", idSehatJkt))

	requireStatus(t, request(t, "DELETE", "/kunjungan/"+idKunjJkt, testutil.KeyJakantro, nil), 200)
	requireStatus(t, request(t, "DELETE", "/kesehatan/"+idSehatJkt, testutil.KeyJakantro, nil), 200)

	assert.False(t, exists(riwayat("/kunjungan"), "kunjungan", idKunjJkt), "cache riwayat kunjungan harus ikut dibuang")
	assert.False(t, exists(riwayat("/kesehatan"), "kesehatan", idSehatJkt), "cache riwayat kesehatan harus ikut dibuang")
	ringkas = riwayat("/summary")
	assert.False(t, exists(ringkas, "kunjungan", idKunjJkt), "cache summary harus ikut dibuang")
	assert.False(t, exists(ringkas, "kesehatan", idSehatJkt), "cache summary harus ikut dibuang")
}
