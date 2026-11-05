package kepemilikan_test

import (
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"

	"github.com/stretchr/testify/assert"
)

func TestFaskesPUTKunjunganSendiri(t *testing.T) {
	setup(t)
	_, idKunjungan := createPemeriksaanFaskes(t)

	res := request(t, "PUT", "/faskes/kunjungan", testutil.KeyFaskesA, map[string]any{
		"id_satusehat": encPeriksa, "tanggal_pengukuran": "2026-03-06", "berat_badan": 7.9,
	})
	requireStatus(t, res, 200)
	assert.NotNil(t, res.Body["updated_at"], "respons menampilkan waktu update")
	assert.Equal(t, "2026-03-06", column(t, "kunjungan", "tanggal_pengukuran", idKunjungan))
	assert.Equal(t, orgFaskes, column(t, "kunjungan", "updated_by", idKunjungan))

	res = request(t, "PUT", "/faskes/diagnosa", testutil.KeyFaskesA, map[string]any{"id_satusehat": condPeriksa, "kode": "E45", "display": "revisi"})
	requireStatus(t, res, 200)
	assert.NotNil(t, res.Body["updated_at"], "respons menampilkan waktu update")
}

func TestFaskesPUTKunjunganFaskesLainDitolak(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)
	createKunjunganFaskesLain(t, idAnakJkt)

	requireForbidden(t, request(t, "PUT", "/faskes/kunjungan", testutil.KeyFaskesA, map[string]any{
		"id_satusehat": encLain, "tanggal_pengukuran": "2026-03-08",
	}))
	assert.Equal(t, "2026-03-07", column(t, "kunjungan", "tanggal_pengukuran", idKunjLain))
}

func TestFaskesPUTTurunanKunjunganFaskesLainDitolak(t *testing.T) {
	setup(t)
	createKeluargaJakantro(t)
	createKunjunganFaskesLain(t, idAnakJkt)

	requireForbidden(t, request(t, "PUT", "/faskes/diagnosa", testutil.KeyFaskesA, map[string]any{
		"id_satusehat": condLain, "kode": "E45", "display": "diubah faskes lain",
	}))
	assert.Equal(t, "", column(t, "diagnosa", "display", idDiagLain))
}

func TestFaskesTetapBisaUbahAnakSetelahDiubahJakantro(t *testing.T) {
	setup(t)
	_, idAnak := createKeluargaFaskes(t)

	requireStatus(t, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"id": idAnak, "nama": "Diubah Jakantro"}), 200)
	assert.Equal(t, "jakantro", column(t, "anak", "updated_by", idAnak))

	requireStatus(t, request(t, "PUT", "/faskes/anak", testutil.KeyFaskesA, map[string]any{"id": idAnak, "nama": "Diubah Faskes", "nik": "3175049910000066"}), 200)
	assert.Equal(t, "Diubah Faskes", column(t, "anak", "nama", idAnak))
	assert.Equal(t, "3175049910000066", column(t, "anak", "nik", idAnak))
}

func TestFaskesPOSTAnakKeKeluargaTanpaAksesDitolak(t *testing.T) {
	setup(t)
	ortu := ortuJakantro("kep-o-luar", "3175049910000010")
	ortu["id_posyandu"] = posyanduLuar
	requireStatus(t, request(t, "POST", "/orangtua", testutil.KeyJakantro, ortu), 201)

	requireForbidden(t, request(t, "POST", "/faskes/anak", testutil.KeyFaskesA, map[string]any{
		"id": "kep-a-luar", "id_orangtua": "kep-o-luar", "nik": "3175049910000011", "nama": "Anak Luar Wilayah",
		"tanggal_lahir": "2025-02-10", "jenis_kelamin": "P", "anak_ke": 1, "bb_lahir": 3, "tb_lahir": 48, "lk_lahir": 33,
	}))

	var exists int
	assert.NoError(t, db.QueryRow(`SELECT count(*) FROM stunting.anak WHERE nik = '3175049910000011'`).Scan(&exists))
	assert.Zero(t, exists)
}

func TestFaskesPOSTIdDibangkitkanSistem(t *testing.T) {
	setup(t)

	res := request(t, "POST", "/faskes/orangtua", testutil.KeyFaskesA, map[string]any{
		"id_posyandu": posyandu, "nik": nikOrtuFks, "nama_ibu": "Ibu Tanpa Id", "telepon": "0812",
	})
	requireStatus(t, res, 201)
	idOrtu, _ := res.Body["id"].(string)
	assert.NotEmpty(t, idOrtu)

	res = request(t, "POST", "/faskes/anak", testutil.KeyFaskesA, map[string]any{
		"id": "kep-a-kiriman", "id_orangtua": idOrtu, "nik": nikAnakFks, "nama": "Anak Tanpa Id",
		"tanggal_lahir": "2025-02-10", "jenis_kelamin": "P", "anak_ke": 1, "bb_lahir": 3, "tb_lahir": 48, "lk_lahir": 33,
	})
	requireStatus(t, res, 201)
	idAnak, _ := res.Body["id"].(string)
	assert.NotEmpty(t, idAnak)
	assert.NotEqual(t, "kep-a-kiriman", idAnak, "id kiriman faskes diabaikan")
}

func TestPayloadAnakDanOrangtuaMemakaiIdSatusehat(t *testing.T) {
	setup(t)
	idOrtu, idAnak := createKeluargaFaskes(t)

	requireStatus(t, request(t, "PUT", "/faskes/anak", testutil.KeyFaskesA, map[string]any{"id": idAnak, "id_satusehat": "P-KEP-01"}), 200)
	assert.Equal(t, "P-KEP-01", column(t, "anak", "satusehat_id", idAnak))
	res := request(t, "PUT", "/faskes/anak", testutil.KeyFaskesA, map[string]any{"id_satusehat": "P-KEP-01", "nama": "Dicari Lewat Id Satusehat"})
	requireStatus(t, res, 200)
	assert.Equal(t, "P-KEP-01", res.Body["id_satusehat"])
	assert.Equal(t, "Dicari Lewat Id Satusehat", column(t, "anak", "nama", idAnak))

	requireStatus(t, request(t, "PUT", "/faskes/orangtua", testutil.KeyFaskesA, map[string]any{"id": idOrtu, "id_satusehat": "RP-KEP-01"}), 200)
	assert.Equal(t, "RP-KEP-01", column(t, "orangtua", "satusehat_id", idOrtu))
}

func TestGETByNikIkutHakAkses(t *testing.T) {
	setup(t)
	idOrtu, idAnak := createKeluargaFaskes(t)

	res := request(t, "GET", "/anak?nik="+nikAnakFks, testutil.KeyFaskesA, nil)
	requireStatus(t, res, 200)
	assert.Equal(t, idAnak, res.Body["id"])
	res = request(t, "GET", "/orangtua?nik="+nikOrtuFks, testutil.KeyFaskesA, nil)
	requireStatus(t, res, 200)
	assert.Equal(t, idOrtu, res.Body["id"])

	requireForbidden(t, request(t, "GET", "/anak?nik="+nikAnakFks, testutil.KeyFaskesB, nil))
	requireForbidden(t, request(t, "GET", "/orangtua?nik="+nikOrtuFks, testutil.KeyFaskesB, nil))
}

func TestGETAnakTanpaIdDanNik(t *testing.T) {
	res := request(t, "GET", "/anak", testutil.KeyFaskesA, nil)
	assert.Equal(t, 422, res.Status, "%v", res.Body)
	assert.Equal(t, 1002, res.ErrorCode, "%v", res.Body)

	res = request(t, "GET", "/anak?nik=3175049999999999", testutil.KeyFaskesA, nil)
	assert.Equal(t, 404, res.Status, "%v", res.Body)
}
