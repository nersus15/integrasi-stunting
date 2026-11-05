package kontrak_field_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/require"
)

const (
	idOrtu      = "bbbbbbbb-0000-4000-8000-000000000001"
	idAnak      = "bbbbbbbb-0000-4000-8000-000000000002"
	idKunjungan = "bbbbbbbb-0000-4000-8000-000000000003"
	idKesehatan = "bbbbbbbb-0000-4000-8000-000000000004"

	ihsAnakBaru = "P-KONTRAK-01"
	nikAnakBaru = "3175049800000201"

	waktuPalsu = "2020-01-01T00:00:00Z"
	isianPalsu = "dari-pengirim"
)

type jawaban struct {
	Kode  int
	Isi   map[string]any
	Galat int
}

func request(t *testing.T, method, path, key string, body any) jawaban {
	t.Helper()
	content, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequest(method, app.BaseURL+path, bytes.NewReader(content))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", key)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	j := jawaban{Kode: res.StatusCode}
	b, _ := io.ReadAll(res.Body)
	_ = json.Unmarshal(b, &j.Isi)
	if k, ok := j.Isi["errorCode"].(float64); ok {
		j.Galat = int(k)
	}
	return j
}

func scalar[T any](t *testing.T, q string, args ...any) T {
	t.Helper()
	var v T
	require.NoError(t, db.QueryRow(q, args...).Scan(&v), q)
	return v
}

func count(t *testing.T, q string, args ...any) int {
	t.Helper()
	return scalar[int](t, q, args...)
}

func cleanup() {
	_, _ = db.Exec(`DELETE FROM stunting.anak WHERE id = $1 OR nik LIKE '317504980000%' OR satusehat_id LIKE 'P-KONTRAK%'`, idAnak)
	_, _ = db.Exec(`DELETE FROM stunting.orangtua WHERE id = $1 OR nik LIKE '317504980000%'`, idOrtu)
	_, _ = db.Exec(`DELETE FROM stunting.episode WHERE satusehat_id LIKE 'ep-kontrak%'`)
	_ = testutil.Scenario1Cleanup(db, groupUji)
}

func setup(t *testing.T) {
	t.Helper()
	cleanup()
	t.Cleanup(cleanup)
}

// field yang selalu diisi sistem; pengirim tidak boleh menentukannya
func withSystemFields(m map[string]any) map[string]any {
	m["created_at"], m["updated_at"], m["deleted_at"] = waktuPalsu, waktuPalsu, waktuPalsu
	m["updated_by"], m["deleted_by"] = isianPalsu, isianPalsu
	return m
}

func orangtua() map[string]any {
	return map[string]any{
		"id": idOrtu, "id_posyandu": "POS-JAKARTA-01", "no_kk": "3175049800000102", "nik": "3175049800000101",
		"nama_ayah": "Ayah Kontrak", "nama_ibu": "Ibu Kontrak", "telepon": "0811", "rt": "001", "rw": "002",
		"alamat": "Jl Kontrak", "kia": 1,
	}
}

func anak() map[string]any {
	return map[string]any{
		"id": idAnak, "id_orangtua": idOrtu, "nik": "3175049800000103", "nama": "Anak Kontrak",
		"tanggal_lahir": "2025-01-01", "jenis_kelamin": "L", "anak_ke": 1, "imd": 1,
		"bb_lahir": 3.1, "tb_lahir": 49, "lk_lahir": 34, "source_data": "jakantro",
	}
}

func kunjungan() map[string]any {
	return map[string]any{"id": idKunjungan, "id_anak": idAnak, "tanggal_pengukuran": "2026-08-14", "berat_badan": 8.1}
}

func kesehatan() map[string]any {
	return map[string]any{"id": idKesehatan, "id_anak": idAnak, "tanggal_pemantauan": "2026-08-14", "tbc_batuk": 0}
}

func setupKeluarga(t *testing.T) {
	t.Helper()
	require.Equal(t, 201, request(t, "POST", "/orangtua", testutil.KeyJakantro, orangtua()).Kode)
	require.Equal(t, 201, request(t, "POST", "/anak", testutil.KeyJakantro, anak()).Kode)
}

// anak yang belum terdaftar; diagnosa E45 supaya lolos kriteria pemantauan
func newPemeriksaan() map[string]any {
	return map[string]any{
		"anak": map[string]any{
			"id_satusehat": ihsAnakBaru, "nik": nikAnakBaru, "nama": "Anak Faskes Kontrak",
			"tanggal_lahir": "2025-03-01", "jenis_kelamin": "L",
		},
		"kunjungan": map[string]any{"id_satusehat": "enc-kontrak-01", "tanggal_pengukuran": "2026-09-01", "berat_badan": 7.9},
		"diagnosa": []any{map[string]any{
			"id_satusehat": "cond-kontrak-01", "system": "http://hl7.org/fhir/sid/icd-10", "kode": "E45",
		}},
	}
}
