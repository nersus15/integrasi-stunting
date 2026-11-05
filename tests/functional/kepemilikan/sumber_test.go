package kepemilikan_test

import (
	"strings"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"

	"github.com/stretchr/testify/assert"
)

func TestSumberDanPembuatJalurJakantro(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)

	check := []struct{ tabel, id, sumber string }{
		{"orangtua", idOrtuJkt, "jakantro:posyandu-01"},
		{"anak", idAnakJkt, "jakantro:kader-01"},
		{"kunjungan", idKunjJkt, "jakantro"},
		{"kesehatan", idSehatJkt, "jakantro"},
		{"orangtua", idOrtuJkt2, "jakantro:rw-07"},
		{"anak", idAnakJkt2, "jakantro:kader-03"},
	}
	for _, c := range check {
		assert.Equal(t, c.sumber, column(t, c.tabel, "source_data", c.id), "%s %s", c.tabel, c.id)
		assert.Equal(t, "jakantro", column(t, c.tabel, "created_by", c.id), "%s %s", c.tabel, c.id)
	}

	res := request(t, "GET", "/anak/"+idAnakJkt, testutil.KeyJakantro, nil)
	requireStatus(t, res, 200)
	assert.Equal(t, "jakantro", res.Body["created_by"])
}

func TestSumberJakantroBerlabelUntukKunjunganDanKesehatan(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)

	requireStatus(t, request(t, "POST", "/kunjungan", testutil.KeyJakantro, map[string]any{
		"id": idKunjJkt2, "id_anak": idAnakJkt, "tanggal_pengukuran": "2026-04-01", "source_data": "kader-02",
	}), 201)
	requireStatus(t, request(t, "POST", "/kesehatan", testutil.KeyJakantro, map[string]any{
		"id": idSehatJkt2, "id_anak": idAnakJkt, "tanggal_pemantauan": "2026-04-01", "source_data": "kader-02",
	}), 201)

	assert.Equal(t, "jakantro:kader-02", column(t, "kunjungan", "source_data", idKunjJkt2))
	assert.Equal(t, "jakantro:kader-02", column(t, "kesehatan", "source_data", idSehatJkt2))
}

// source_data tersimpan maksimal 50 karakter, jadi label jakantro maksimal 50 - len("jakantro:") = 41
func TestSourceDataTersimpanMaksimal50Karakter(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)
	panjang := strings.Repeat("x", 42)

	res := request(t, "POST", "/kunjungan", testutil.KeyJakantro, map[string]any{
		"id": idKunjJkt2, "id_anak": idAnakJkt, "tanggal_pengukuran": "2026-04-01", "source_data": panjang,
	})
	assert.Equal(t, 422, res.Status, "%v", res.Body)
	assert.Equal(t, 1002, res.ErrorCode, "%v", res.Body)

	ortu := ortuJakantro("kep-o-panjang", "3175049910000099")
	ortu["source_data"] = panjang
	res = request(t, "POST", "/orangtua", testutil.KeyJakantro, ortu)
	assert.Equal(t, 422, res.Status, "%v", res.Body)
	assert.Equal(t, 1002, res.ErrorCode, "%v", res.Body)

	uuid := "8352cfdd-5b80-425b-b831-51efff2b22df"
	requireStatus(t, request(t, "POST", "/kesehatan", testutil.KeyJakantro, map[string]any{
		"id": idSehatJkt2, "id_anak": idAnakJkt, "tanggal_pemantauan": "2026-04-01", "source_data": uuid,
	}), 201)
	assert.Equal(t, "jakantro:"+uuid, column(t, "kesehatan", "source_data", idSehatJkt2))

	requireStatus(t, request(t, "POST", "/kunjungan", testutil.KeyJakantro, map[string]any{
		"id": idKunjJkt2, "id_anak": idAnakJkt, "tanggal_pengukuran": "2026-04-01", "source_data": strings.Repeat("x", 41),
	}), 201)
	assert.Len(t, column(t, "kunjungan", "source_data", idKunjJkt2), 50)
}

func TestSumberDanPembuatJalurFaskes(t *testing.T) {
	setup(t)
	idOrtu, idAnak := createKeluargaFaskes(t)
	idAnakPeriksa, idKunjungan := createPemeriksaanFaskes(t)

	check := []struct{ tabel, id string }{
		{"orangtua", idOrtu},
		{"anak", idAnak},
		{"anak", idAnakPeriksa},
		{"kunjungan", idKunjungan},
	}
	for _, c := range check {
		assert.Equal(t, "faskes:"+orgFaskes, column(t, c.tabel, "source_data", c.id), "%s %s", c.tabel, c.id)
		assert.Equal(t, orgFaskes, column(t, c.tabel, "created_by", c.id), "%s %s", c.tabel, c.id)
	}
}

func TestPembuatTidakBerubahLewatPUT(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)
	_, idAnakFaskes := createKeluargaFaskes(t)

	requireStatus(t, request(t, "PUT", "/faskes/anak", testutil.KeyFaskesA, map[string]any{"id": idAnakJkt, "nama": "Diubah Faskes", "created_by": orgFaskes}), 200)
	requireStatus(t, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"id": idAnakFaskes, "nama": "Diubah Jakantro", "created_by": "jakantro"}), 200)
	requireStatus(t, request(t, "PUT", "/kunjungan/"+idKunjJkt, testutil.KeyJakantro, map[string]any{
		"id": idKunjJkt, "id_anak": idAnakJkt, "tanggal_pengukuran": "2026-03-01", "created_by": orgLain,
	}), 200)

	assert.Equal(t, "jakantro", column(t, "anak", "created_by", idAnakJkt))
	assert.Equal(t, orgFaskes, column(t, "anak", "created_by", idAnakFaskes))
	assert.Equal(t, "jakantro", column(t, "kunjungan", "created_by", idKunjJkt))
	assert.Equal(t, "jakantro", column(t, "kunjungan", "updated_by", idKunjJkt))
	assert.Equal(t, "jakantro:kader-01", column(t, "anak", "source_data", idAnakJkt))
}
