package logtransaksi_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	app        *testutil.App
	db         *sql.DB
	mu         sync.Mutex
	diterima   []map[string]any
	headerLog  []http.Header
	serviceLog *httptest.Server
)

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		fmt.Fprintln(os.Stderr, "postgres mati, test log transaksi dilewati")
		os.Exit(0)
	}
	serviceLog = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var c map[string]any
		_ = json.Unmarshal(b, &c)
		mu.Lock()
		diterima = append(diterima, c)
		headerLog = append(headerLog, r.Header.Clone())
		mu.Unlock()
		w.WriteHeader(http.StatusAccepted)
	}))
	os.Setenv("LOG_TRANSAKSI_ENABLED", "true")
	os.Setenv("LOG_TRANSAKSI_ENDPOINT", serviceLog.URL)
	// sama seperti ConfigMap, termasuk spasi setelah '='
	os.Setenv("LOG_TRANSAKSI_HEADERS", "x-api-key= key-service-log,x-api-name=Service-Integrasi-Stunting-Uji")
	os.Setenv("LOG_TRANSAKSI_REDIS_KEY", "uji-log:"+uuid.NewString())

	db = testutil.OpenDB()
	if err := testutil.Scenario1Seed(db); err != nil {
		fmt.Fprintf(os.Stderr, "gagal seed: %v\n", err)
		os.Exit(1)
	}
	var err error
	app, err = testutil.StartApp(2101)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	cleanup()
	app.Stop()
	serviceLog.Close()
	db.Close()
	os.Exit(code)
}

func cleanup() {
	_, _ = db.Exec(`DELETE FROM stunting.kunjungan WHERE id LIKE 'uji-log-%'`)
	_, _ = db.Exec(`DELETE FROM stunting.anak WHERE id LIKE 'uji-log-%'`)
	_, _ = db.Exec(`DELETE FROM stunting.orangtua WHERE id LIKE 'uji-log-%'`)
}

func kirim(t *testing.T, method, path, key string, body any) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, path, r)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("X-API-Key", key)
	}
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, res.Body)
	res.Body.Close()
	return res
}

func catatanUntuk(t *testing.T, id string) map[string]any {
	t.Helper()
	var hasil map[string]any
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		for _, c := range diterima {
			if c["identifier"] == id {
				hasil, _ = c["context"].(map[string]any)
				return true
			}
		}
		return false
	}, 3*time.Second, 20*time.Millisecond, "catatan %s tidak sampai ke service log", id)
	return hasil
}

func TestRegistrasiJakantroTercatatDenganNikDisamarkan(t *testing.T) {
	cleanup()
	t.Cleanup(cleanup)
	nikAnak := "3175049977000001"

	res := kirim(t, "POST", app.BaseURL+"/kunjungan", testutil.KeyJakantro, map[string]any{
		"orangtua": map[string]any{"id": "uji-log-ortu", "id_posyandu": "POS-JAKARTA-01", "no_kk": "3175049977000002",
			"nik": "3175049977000003", "nama_ibu": "Ibu Uji Log", "telepon": "0811", "rt": "001", "rw": "001", "alamat": "Jl. Uji", "kia": 1},
		"anak": map[string]any{"id": "uji-log-anak", "id_orangtua": "uji-log-ortu", "nik": nikAnak, "nama": "Anak Uji Log",
			"tanggal_lahir": "2025-01-10", "jenis_kelamin": "L", "anak_ke": 1, "imd": 1, "bb_lahir": 3.1, "tb_lahir": 49, "lk_lahir": 33,
			"source_data": "kader-01"},
		"kunjungan": map[string]any{"id": "uji-log-kunj", "id_anak": "uji-log-anak", "tanggal_pengukuran": "2026-08-14", "berat_badan": 8.1},
	})
	require.Equal(t, 201, res.StatusCode)
	id := res.Header.Get("X-Transaction-ID")
	require.NotEmpty(t, id, "header id transaksi bawaan dari config")

	c := catatanUntuk(t, id)
	assert.Equal(t, "POST", c["method"])
	assert.Equal(t, "/api/kunjungan", c["path"])
	assert.Equal(t, float64(201), c["status"])
	assert.Equal(t, "jakantro", c["orgid"])

	b, err := json.Marshal(c)
	require.NoError(t, err)
	assert.NotContains(t, string(b), nikAnak, "nik anak tidak boleh terkirim utuh, di request maupun response")
	assert.NotContains(t, string(b), "3175049977000003", "nik orangtua")
	assert.NotContains(t, string(b), testutil.KeyJakantro)
	assert.Contains(t, string(b), "uji-log-kunj", "id kunjungan tetap utuh")
}

func TestPenolakanOtorisasiTetapTercatat(t *testing.T) {
	res := kirim(t, "POST", app.BaseURL+"/kunjungan", testutil.KeyFaskesA, map[string]any{"id": "uji-log-x"})
	require.Equal(t, 403, res.StatusCode, "key faskes di endpoint jakantro")

	c := catatanUntuk(t, res.Header.Get("X-Transaction-ID"))
	assert.Equal(t, float64(403), c["status"], "middleware log terpasang sebelum Otorisasi")
	resp, _ := c["response"].(map[string]any)
	assert.Equal(t, float64(1008), resp["errorCode"])
}

func TestKeyTidakDikenalDitolakAutentikasiTetapTercatat(t *testing.T) {
	res := kirim(t, "GET", app.BaseURL+"/anak/uji-log-x", "key-tidak-dikenal", nil)
	require.Equal(t, 401, res.StatusCode)
	id := res.Header.Get("X-Transaction-ID")
	require.NotEmpty(t, id, "log transaksi jalan sebelum autentikasi webcore")

	c := catatanUntuk(t, id)
	assert.Equal(t, float64(401), c["status"])
	assert.Equal(t, "/api/anak/uji-log-x", c["path"])
	assert.Nil(t, c["orgid"], "ditolak sebelum identitas diketahui")
	b, err := json.Marshal(c)
	require.NoError(t, err)
	assert.NotContains(t, string(b), "key-tidak-dikenal")
}

func TestHeaderServiceLogDariEnvIkutTerkirim(t *testing.T) {
	res := kirim(t, "GET", app.BaseURL+"/anak/uji-log-x", "key-tidak-dikenal", nil)
	catatanUntuk(t, res.Header.Get("X-Transaction-ID"))

	mu.Lock()
	defer mu.Unlock()
	h := headerLog[len(headerLog)-1]
	assert.Equal(t, "key-service-log", h.Get("X-Api-Key"), "spasi di sekitar nilai dibuang")
	assert.Equal(t, "Service-Integrasi-Stunting-Uji", h.Get("X-Api-Name"))
}

func TestEndpointDiLuarApiTidakDicatat(t *testing.T) {
	res := kirim(t, "GET", app.Health, "", nil)
	require.Equal(t, 200, res.StatusCode)
	assert.Empty(t, res.Header.Get("X-Transaction-ID"), "health di luar /api")
}
