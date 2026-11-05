package kepemilikan_test

import (
	"database/sql"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/require"
)

const (
	posyandu     = "POS-JAKARTA-01"
	posyanduLuar = "003eea17-3fb4-f85e-a8fc-90f39c5cd577"
	orgFaskes    = "100011"
	orgLain      = "200022"
	idFaskesLain = "FAS-RS-HARAPAN"

	idOrtuJkt   = "kep-o-jkt"
	nikOrtuJkt  = "3175049910000001"
	idOrtuJkt2  = "kep-o-jkt2"
	nikOrtuJkt2 = "3175049910000002"
	idAnakJkt   = "kep-a-jkt"
	nikAnakJkt  = "3175049910000003"
	idAnakJkt2  = "kep-a-jkt2"
	nikAnakJkt2 = "3175049910000007"
	idKunjJkt   = "kep-k-jkt"
	idKunjJkt2  = "kep-k-jkt2"
	idSehatJkt  = "kep-h-jkt"
	idSehatJkt2 = "kep-h-jkt2"

	idOrtuFks  = "kep-o-fks"
	nikOrtuFks = "3175049910000004"
	idAnakFks  = "kep-a-fks"
	nikAnakFks = "3175049910000005"

	nikAnakPeriksa = "3175049910000006"
	encPeriksa     = "kep-enc-1"
	condPeriksa    = "kep-cond-1"

	idKunjLain = "kep-k-lain"
	encLain    = "kep-enc-lain"
	idDiagLain = "kep-d-lain"
	condLain   = "kep-cond-lain"
)

func request(t *testing.T, method, path, key string, body any) testutil.Response {
	t.Helper()
	return testutil.Request(t, app.BaseURL+path, method, key, body)
}

func cleanup() {
	anak := `SELECT id FROM stunting.anak WHERE id LIKE 'kep-%' OR nik LIKE '317504991000%'`
	for _, table := range []string{"observasi", "diagnosa", "layanan", "rujukan", "kunjungan", "kesehatan", "episode"} {
		_, _ = db.Exec(`DELETE FROM stunting.` + table + ` WHERE id_anak IN (` + anak + `)`)
	}
	_, _ = db.Exec(`DELETE FROM stunting.anak WHERE id LIKE 'kep-%' OR nik LIKE '317504991000%'`)
	_, _ = db.Exec(`DELETE FROM stunting.orangtua WHERE id LIKE 'kep-%' OR nik LIKE '317504991000%'`)
}

func setup(t *testing.T) {
	t.Helper()
	cleanup()
	t.Cleanup(cleanup)
}

func column(t *testing.T, table, column, id string) string {
	t.Helper()
	var v sql.NullString
	require.NoError(t, db.QueryRow(`SELECT trim(`+column+`::text) FROM stunting.`+table+` WHERE id = $1`, id).Scan(&v), "%s.%s %s", table, column, id)
	return v.String
}

func isDeleted(t *testing.T, table, id string) bool {
	t.Helper()
	var exists bool
	require.NoError(t, db.QueryRow(`SELECT "deletedAt" IS NOT NULL FROM stunting.`+table+` WHERE id = $1`, id).Scan(&exists))
	return exists
}

func requireStatus(t *testing.T, res testutil.Response, code int) {
	t.Helper()
	require.Equal(t, code, res.Status, "%v", res.Body)
}

func requireForbidden(t *testing.T, res testutil.Response) {
	t.Helper()
	require.Equal(t, 403, res.Status, "%v", res.Body)
	require.Equal(t, 1008, res.ErrorCode, "%v", res.Body)
}

func ortuJakantro(id, nik string) map[string]any {
	return map[string]any{
		"id": id, "id_posyandu": posyandu, "nik": nik, "nama_ibu": "Ibu Jakantro",
		"telepon": "0811", "rt": "001", "rw": "002", "alamat": "Jl. Posyandu",
	}
}

func anakJakantro(id, idOrtu, nik, label string) map[string]any {
	return map[string]any{
		"id": id, "id_orangtua": idOrtu, "nik": nik, "nama": "Anak Jakantro",
		"tanggal_lahir": "2025-01-10", "jenis_kelamin": "L", "anak_ke": 1,
		"bb_lahir": 3.1, "tb_lahir": 49, "lk_lahir": 33, "source_data": label,
	}
}

// keluarga jakantro lewat POST /kunjungan bentuk 1, ditambah kesehatan, orangtua kedua, dan anak kedua
func createKeluargaJakantro(t *testing.T) {
	t.Helper()
	ortu := ortuJakantro(idOrtuJkt, nikOrtuJkt)
	ortu["source_data"] = "posyandu-01"
	ortu["created_by"] = "palsu"
	requireStatus(t, request(t, "POST", "/kunjungan", testutil.KeyJakantro, map[string]any{
		"orangtua": ortu,
		"anak":     anakJakantro(idAnakJkt, idOrtuJkt, nikAnakJkt, "kader-01"),
		"kunjungan": map[string]any{
			"id": idKunjJkt, "id_anak": idAnakJkt, "tanggal_pengukuran": "2026-03-01",
			"berat_badan": 8.1, "created_by": "palsu",
		},
	}), 201)
	requireStatus(t, request(t, "POST", "/kesehatan", testutil.KeyJakantro, map[string]any{
		"id": idSehatJkt, "id_anak": idAnakJkt, "tanggal_pemantauan": "2026-03-01", "tbc_batuk": 0,
	}), 201)

	ortu2 := ortuJakantro(idOrtuJkt2, nikOrtuJkt2)
	ortu2["source_data"] = "rw-07"
	requireStatus(t, request(t, "POST", "/orangtua", testutil.KeyJakantro, ortu2), 201)
	requireStatus(t, request(t, "POST", "/anak", testutil.KeyJakantro, anakJakantro(idAnakJkt2, idOrtuJkt2, nikAnakJkt2, "kader-03")), 201)
}

// orangtua dan anak didaftarkan faskes; id dibaca dari respons, bukan dari payload
func createKeluargaFaskes(t *testing.T) (idOrtu, idAnak string) {
	t.Helper()
	res := request(t, "POST", "/faskes/orangtua", testutil.KeyFaskesA, map[string]any{
		"id": idOrtuFks, "id_posyandu": posyandu, "nik": nikOrtuFks, "nama_ibu": "Ibu Faskes",
		"telepon": "0812", "source_data": "palsu", "created_by": "palsu",
	})
	requireStatus(t, res, 201)
	idOrtu = res.Body["id"].(string)

	res = request(t, "POST", "/faskes/anak", testutil.KeyFaskesA, map[string]any{
		"id": idAnakFks, "id_orangtua": idOrtu, "nik": nikAnakFks, "nama": "Anak Faskes",
		"tanggal_lahir": "2025-02-10", "jenis_kelamin": "P", "anak_ke": 1,
		"bb_lahir": 3, "tb_lahir": 48, "lk_lahir": 33, "source_data": "palsu", "created_by": "palsu",
	})
	requireStatus(t, res, 201)
	return idOrtu, res.Body["id"].(string)
}

func createPemeriksaanFaskes(t *testing.T) (idAnak, idKunjungan string) {
	t.Helper()
	res := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, map[string]any{
		"anak": map[string]any{
			"nik": nikAnakPeriksa, "nama": "Anak Pemeriksaan", "tanggal_lahir": "2025-03-10", "jenis_kelamin": "L",
		},
		"kunjungan": map[string]any{"id_satusehat": encPeriksa, "tanggal_pengukuran": "2026-03-05"},
		"diagnosa": []any{map[string]any{
			"id_satusehat": condPeriksa, "system": "http://hl7.org/fhir/sid/icd-10", "kode": "E45",
			"display": "Nutritional stunting", "tanggal_catat": "2026-03-05",
		}},
	})
	requireStatus(t, res, 201)
	return res.Body["id_anak"].(string), res.Body["id_kunjungan"].(string)
}

// kunjungan dan diagnosa milik faskes lain, langsung ke database
func createKunjunganFaskesLain(t *testing.T, idAnak string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO stunting.kunjungan (id, id_anak, id_faskes, satusehat_id, tanggal_pengukuran, source_data, created_by)
		VALUES ($1, $2, $3, $4, '2026-03-07', 'faskes:'||$5, $5)`, idKunjLain, idAnak, idFaskesLain, encLain, orgLain)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO stunting.diagnosa (id, id_anak, id_kunjungan, satusehat_id, jenis, system, kode, tanggal_catat)
		VALUES ($1, $2, $3, $4, 'diagnosis', 'http://hl7.org/fhir/sid/icd-10', 'E45', '2026-03-07')`, idDiagLain, idAnak, idKunjLain, condLain)
	require.NoError(t, err)
}
