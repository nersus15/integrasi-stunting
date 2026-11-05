package logtransaksi_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/helper/logtransaksi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultMask = []string{"nik", "no_kk", "nama", "nama_ayah", "nama_ibu", "telepon", "alamat"}

func masker(t *testing.T, cfg config.LogTransaksiConfig) *logtransaksi.Masker {
	t.Helper()
	if cfg.Mask == nil {
		cfg.Mask = defaultMask
	}
	return logtransaksi.NewMasker(cfg.Mask)
}

func prepare(t *testing.T, m *logtransaksi.Masker, r logtransaksi.Record) map[string]any {
	t.Helper()
	b, err := json.Marshal(m.Prepare(r))
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(b, &out))
	ctx, ok := out["context"].(map[string]any)
	require.True(t, ok, "detail transaksi ada di context: %v", out)
	return ctx
}

func envelope(t *testing.T, m *logtransaksi.Masker, r logtransaksi.Record) map[string]any {
	t.Helper()
	b, err := json.Marshal(m.Prepare(r))
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(b, &out))
	return out
}

// semua kecuali 3 karakter terakhir jadi '*'
func samar(s string) string {
	r := []rune(s)
	if len(r) <= 3 {
		return strings.Repeat("*", len(r))
	}
	return strings.Repeat("*", len(r)-3) + string(r[len(r)-3:])
}

func dig(t *testing.T, v any, path ...any) any {
	t.Helper()
	for _, k := range path {
		switch key := k.(type) {
		case string:
			m, ok := v.(map[string]any)
			require.True(t, ok, "bukan object di %v", k)
			v = m[key]
		case int:
			a, ok := v.([]any)
			require.True(t, ok, "bukan array di %v", k)
			v = a[key]
		}
	}
	return v
}

const registrasi = `{
  "orangtua": {"id": "o1", "nik": "3175040101900002", "no_kk": "3175042509870001", "nama_ayah": "Ahmad Suryana",
    "nama_ibu": "Dewi Lestari", "telepon": "081234567890", "alamat": "Jl. Kenanga No. 12", "id_posyandu": "POS-1"},
  "anak": {"id": "a1", "nik": "3175041503240003", "nama": "Rizky Ramadhan", "tanggal_lahir": "2024-03-15"},
  "kunjungan": {"id": "k1", "berat_badan": 9.4, "status_tbu": "Sangat Pendek"}
}`

func TestFieldPribadiDisamarkanKecualiTigaKarakterTerakhir(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{})
	out := prepare(t, m, logtransaksi.Record{Request: []byte(registrasi)})

	assert.Equal(t, "*************002", dig(t, out, "request", "orangtua", "nik"))
	assert.Equal(t, "*************001", dig(t, out, "request", "orangtua", "no_kk"))
	assert.Equal(t, "*********890", dig(t, out, "request", "orangtua", "telepon"))
	assert.Equal(t, samar("Ahmad Suryana"), dig(t, out, "request", "orangtua", "nama_ayah"))
	assert.Equal(t, samar("Dewi Lestari"), dig(t, out, "request", "orangtua", "nama_ibu"))
	assert.Equal(t, samar("Jl. Kenanga No. 12"), dig(t, out, "request", "orangtua", "alamat"))
	assert.Equal(t, "*************003", dig(t, out, "request", "anak", "nik"))
	assert.Equal(t, samar("Rizky Ramadhan"), dig(t, out, "request", "anak", "nama"))
	assert.NotContains(t, dig(t, out, "request", "anak", "nama"), "Rizky")
}

func TestFieldBukanPribadiTetapUtuh(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{})
	out := prepare(t, m, logtransaksi.Record{Request: []byte(registrasi)})

	assert.Equal(t, "o1", dig(t, out, "request", "orangtua", "id"))
	assert.Equal(t, "POS-1", dig(t, out, "request", "orangtua", "id_posyandu"))
	assert.Equal(t, "2024-03-15", dig(t, out, "request", "anak", "tanggal_lahir"))
	assert.Equal(t, 9.4, dig(t, out, "request", "kunjungan", "berat_badan"))
	assert.Equal(t, "Sangat Pendek", dig(t, out, "request", "kunjungan", "status_tbu"))
}

func TestPenyamaranMasukKeArrayDanResponse(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{})
	out := prepare(t, m, logtransaksi.Record{
		Response: []byte(`{"nik": "3175040101900002", "anak": [{"nik": "3175041503240003", "nama": "Rizky"}, {"nik": "3175041503240004"}]}`),
	})

	assert.Equal(t, "*************002", dig(t, out, "response", "nik"))
	assert.Equal(t, "*************003", dig(t, out, "response", "anak", 0, "nik"))
	assert.Equal(t, "**zky", dig(t, out, "response", "anak", 0, "nama"))
	assert.Equal(t, "*************004", dig(t, out, "response", "anak", 1, "nik"))
}

func TestNamaFieldDicocokkanTanpaMemperhatikanHurufBesar(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{})
	out := prepare(t, m, logtransaksi.Record{Request: []byte(`{"NIK": "3175040101900002", "Nama": "Ani"}`)})

	assert.Equal(t, "*************002", dig(t, out, "request", "NIK"))
	assert.Equal(t, "***", dig(t, out, "request", "Nama"), "3 karakter atau kurang disamarkan semua")
}

func TestQueryStringDisamarkan(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{})
	out := prepare(t, m, logtransaksi.Record{Query: "nik=3175040101900002&detail=1"})

	q, _ := out["query"].(string)
	assert.NotContains(t, q, "3175040101900002")
	assert.Contains(t, q, "002")
	assert.Contains(t, q, "detail=1")
}

func TestBodyBukanJSONTetapDisamarkanAngkaPanjangnya(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{})
	out := prepare(t, m, logtransaksi.Record{Request: []byte(`{"nik": "3175040101900002", rusak`)})

	body, ok := out["request"].(string)
	require.True(t, ok, "body bukan JSON dikirim sebagai teks: %v", out["request"])
	assert.NotContains(t, body, "3175040101900002")
	assert.Contains(t, body, "rusak")
}

func TestBodyBesarDikirimUtuhTanpaDipotong(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{})
	isi := strings.Repeat("x", 1<<20)
	out := prepare(t, m, logtransaksi.Record{Request: []byte(`{"catatan": "` + isi + `"}`)})

	assert.Equal(t, isi, dig(t, out, "request", "catatan"), "body 1MB tetap utuh dan tetap JSON")
	assert.NotContains(t, out, "request_terpotong")
}

func TestNikBerupaAngkaTetapDisamarkanTanpaKehilanganDigit(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{})
	out := prepare(t, m, logtransaksi.Record{Request: []byte(`{"nik": 3175041503240003, "berat_badan": 9.4, "anak_ke": 2}`)})

	assert.Equal(t, "*************003", dig(t, out, "request", "nik"))
	assert.Equal(t, 9.4, dig(t, out, "request", "berat_badan"))
	assert.Equal(t, float64(2), dig(t, out, "request", "anak_ke"))
}

func TestDaftarFieldYangDisamarkanBisaDiaturLewatConfig(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{Mask: []string{"email"}})
	out := prepare(t, m, logtransaksi.Record{Request: []byte(`{"email": "dewi@contoh.id", "nik": "3175040101900002"}`)})

	assert.Equal(t, samar("dewi@contoh.id"), dig(t, out, "request", "email"))
	assert.Equal(t, "3175040101900002", dig(t, out, "request", "nik"), "nik tidak ada di daftar config ini")
}

func TestMetadataTransaksiIkutTerkirim(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{})
	waktu := time.Date(2026, 10, 31, 10, 0, 0, 0, time.UTC)
	out := prepare(t, m, logtransaksi.Record{
		TransactionID: "trx-1", Timestamp: waktu, Duration: 1500 * time.Microsecond, Method: "POST", Path: "/api/kunjungan",
		Status: 201, Orgid: "jakantro", IP: "10.0.0.1", UserAgent: "uji",
	})

	assert.Equal(t, "trx-1", out["transaction_id"])
	assert.Equal(t, "2026-10-31T10:00:00Z", out["timestamp"])
	assert.Equal(t, 1.5, out["duration_ms"])
	assert.Equal(t, "POST", out["method"])
	assert.Equal(t, "/api/kunjungan", out["path"])
	assert.Equal(t, float64(201), out["status"])
	assert.Equal(t, "jakantro", out["orgid"])
	assert.Equal(t, "10.0.0.1", out["ip"])
	assert.Equal(t, "uji", out["user_agent"])
}

func TestPayloadDibungkusLevelMessageIdentifierContext(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{})
	out := envelope(t, m, logtransaksi.Record{TransactionID: "trx-1", Status: 422,
		Response: []byte(`{"httpCode": 422, "errorCode": 1005, "message": "id_posyandu tidak ditemukan"}`)})

	assert.Equal(t, "info", out["level"])
	assert.Equal(t, "id_posyandu tidak ditemukan", out["message"], "message dari respons")
	assert.Equal(t, "trx-1", out["identifier"], "identifier dari id transaksi")
	assert.Equal(t, "trx-1", dig(t, out, "context", "transaction_id"), "isi context tetap seperti sebelumnya")
	assert.Equal(t, float64(1005), dig(t, out, "context", "response", "errorCode"))
	assert.Len(t, out, 4, "hanya level, message, identifier, context")
}

func TestMessageTidakPernahKosong(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{})
	cases := []struct {
		status   int
		response string
		want     string
	}{
		{200, `{"id": "k1"}`, "Berhasil"},
		{201, `[{"id": "k1"}]`, "Berhasil"},
		{200, `{"message": ""}`, "Berhasil"},
		{404, `bukan json`, "Not Found"},
		{500, ``, "Internal Server Error"},
		{422, `{"message": "id anak tidak valid"}`, "id anak tidak valid"},
	}
	for _, c := range cases {
		out := envelope(t, m, logtransaksi.Record{TransactionID: "trx-1", Status: c.status, Response: []byte(c.response)})
		assert.Equal(t, c.want, out["message"], "status %d, respons %q", c.status, c.response)
	}
}

func TestMessageYangMemuatDataPribadiIkutDisamarkan(t *testing.T) {
	m := masker(t, config.LogTransaksiConfig{})
	out := envelope(t, m, logtransaksi.Record{TransactionID: "trx-1", Secrets: []string{"key-tidak-dikenal"},
		Response: []byte(`{"message": "User not found: key-tidak-dikenal"}`)})

	assert.Equal(t, "User not found: **************nal", out["message"])
}
