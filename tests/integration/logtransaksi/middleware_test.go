package logtransaksi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/handler"
	"github.com/nersus15/integrasi/mod-stunting/helper/logtransaksi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pengganti service log: hanya kiriman yang dibalas 2xx dianggap tersimpan
type receiver struct {
	srv *httptest.Server
	mu  sync.Mutex
	// records: isi context; envelopes: kiriman utuh
	records   []map[string]any
	envelopes []map[string]any
	headers   []http.Header
	status    int
	delay     time.Duration
}

func newReceiver(t *testing.T) *receiver {
	t.Helper()
	rc := &receiver{status: http.StatusOK}
	rc.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc.mu.Lock()
		delay, status := rc.delay, rc.status
		rc.mu.Unlock()
		select {
		case <-time.After(delay):
		case <-r.Context().Done():
			return // pengirim membatalkan, belum ada balasan berarti belum tersimpan
		}
		b, _ := io.ReadAll(r.Body)
		if status < 300 {
			var m map[string]any
			_ = json.Unmarshal(b, &m)
			ctx, _ := m["context"].(map[string]any)
			rc.mu.Lock()
			rc.records = append(rc.records, ctx)
			rc.envelopes = append(rc.envelopes, m)
			rc.headers = append(rc.headers, r.Header.Clone())
			rc.mu.Unlock()
		}
		w.WriteHeader(status)
	}))
	t.Cleanup(rc.srv.Close)
	return rc
}

func (rc *receiver) set(status int, delay time.Duration) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.status, rc.delay = status, delay
}

func (rc *receiver) waitFor(t *testing.T, n int) []map[string]any {
	t.Helper()
	require.Eventually(t, func() bool {
		rc.mu.Lock()
		defer rc.mu.Unlock()
		return len(rc.records) >= n
	}, 5*time.Second, 10*time.Millisecond, "catatan log tidak sampai ke service log")
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return append([]map[string]any(nil), rc.records...)
}

func (rc *receiver) count() int {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return len(rc.records)
}

// setiap id transaksi harus diterima tepat sekali
func (rc *receiver) exactlyOnce(t *testing.T, ids []string) {
	t.Helper()
	got := map[string]int{}
	for _, c := range rc.waitFor(t, len(ids)) {
		id, _ := c["transaction_id"].(string)
		got[id]++
	}
	for _, id := range ids {
		assert.Equal(t, 1, got[id], "id %s", id)
	}
	assert.Len(t, got, len(ids), "tidak ada kiriman ganda atau asing")
}

func testKey(t *testing.T) string {
	t.Helper()
	k := "uji-log:" + uuid.NewString()
	t.Cleanup(func() { rdb.Del(context.Background(), k) })
	return k
}

func queued(t *testing.T, key string) []string {
	t.Helper()
	items, err := rdb.LRange(context.Background(), key, 0, -1).Result()
	require.NoError(t, err)
	return items
}

func newSender(t *testing.T, cfg config.LogTransaksiConfig) *logtransaksi.Sender {
	t.Helper()
	if cfg.Mask == nil {
		cfg.Mask = []string{"nik", "nama"}
	}
	if cfg.RedisKey == "" {
		cfg.RedisKey = testKey(t)
	}
	if cfg.RetryInterval == 0 {
		cfg.RetryInterval = 50 * time.Millisecond
	}
	s, err := logtransaksi.New(cfg, rdb)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close(context.Background()) })
	return s
}

type testApp struct {
	web    *fiber.App
	sender *logtransaksi.Sender
}

func newApp(t *testing.T, cfg config.LogTransaksiConfig) *testApp {
	t.Helper()
	s := newSender(t, cfg)

	web := fiber.New()
	web.Use(fiberrecover.New())
	web.Use("/api", handler.LogTransaksi(cfg, "X-API-Key", s.Push))
	// meniru autentikasi webcore (401 menyebut key) dan Otorisasi (orgid dari key, key salah ditolak)
	web.Use("/api", func(c *fiber.Ctx) error {
		if k := c.Get("X-API-Key"); strings.HasPrefix(k, "tak-dikenal") {
			return c.Status(401).JSON(fiber.Map{"httpCode": 401, "errorCode": 4001, "message": "User not found: " + k})
		}
		if a := c.Get("Authorization"); a != "" {
			return c.Status(401).JSON(fiber.Map{"httpCode": 401, "message": "token ditolak: " + strings.TrimPrefix(a, "Bearer ")})
		}
		switch c.Get("X-API-Key") {
		case "key-jakantro":
			orgid := "jakantro"
			c.Locals("orgid", &orgid)
			return c.Next()
		default:
			return c.Status(403).JSON(fiber.Map{"httpCode": 403, "errorCode": 4002})
		}
	})
	web.Post("/api/kunjungan", func(c *fiber.Ctx) error {
		var body map[string]any
		_ = json.Unmarshal(c.Body(), &body)
		return c.Status(201).JSON(fiber.Map{"id": "k1", "anak": body["anak"]})
	})
	web.Get("/api/anak/:id", func(c *fiber.Ctx) error {
		return c.Status(422).JSON(fiber.Map{"httpCode": 422, "errorCode": 1002, "message": "id anak tidak valid"})
	})
	web.Get("/api/kafka/gagal", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"total": 0}) })
	web.Get("/api/panik", func(c *fiber.Ctx) error { panic("uji panic") })
	return &testApp{web: web, sender: s}
}

func (a *testApp) call(t *testing.T, method, path, key, body string, header ...string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("X-API-Key", key)
	}
	for i := 0; i+1 < len(header); i += 2 {
		req.Header.Set(header[i], header[i+1])
	}
	res, err := a.web.Test(req, -1)
	require.NoError(t, err)
	return res
}

func idOf(res *http.Response) string { return res.Header.Get("X-Transaction-ID") }

const payload = `{"anak": {"id": "a1", "nik": "3175041503240003", "nama": "Rizky Ramadhan"}}`

func TestTransaksiSuksesDicatatLengkap(t *testing.T) {
	rc := newReceiver(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL})

	res := a.call(t, "POST", "/api/kunjungan?debug=1", "key-jakantro", payload)
	require.Equal(t, 201, res.StatusCode)
	id := idOf(res)
	require.NotEmpty(t, id, "header id transaksi bawaan")

	c := rc.waitFor(t, 1)[0]
	assert.Equal(t, id, c["transaction_id"], "id di header sama dengan di log")
	assert.Equal(t, "POST", c["method"])
	assert.Equal(t, "/api/kunjungan", c["path"])
	assert.Equal(t, "debug=1", c["query"])
	assert.Equal(t, float64(201), c["status"])
	assert.Equal(t, "jakantro", c["orgid"], "orgid diisi Otorisasi, terbaca setelah handler")
	assert.NotEmpty(t, c["timestamp"])
	assert.NotNil(t, c["duration_ms"])

	req, _ := c["request"].(map[string]any)
	anakReq, _ := req["anak"].(map[string]any)
	assert.Equal(t, "a1", anakReq["id"])
	assert.Equal(t, "*************003", anakReq["nik"], "nik request disamarkan")

	resp, _ := c["response"].(map[string]any)
	assert.Equal(t, "k1", resp["id"], "body response ikut tercatat")
	anakRes, _ := resp["anak"].(map[string]any)
	assert.Equal(t, "*************003", anakRes["nik"], "nik response disamarkan")
}

func TestKirimanDibungkusLevelMessageIdentifier(t *testing.T) {
	rc := newReceiver(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL})

	id := idOf(a.call(t, "GET", "/api/anak/x", "key-jakantro", ""))
	rc.waitFor(t, 1)
	rc.mu.Lock()
	defer rc.mu.Unlock()
	e := rc.envelopes[0]
	assert.Equal(t, "info", e["level"])
	assert.Equal(t, "id anak tidak valid", e["message"], "message dari respons API")
	assert.Equal(t, id, e["identifier"], "identifier sama dengan header id transaksi")
	ctx, _ := e["context"].(map[string]any)
	assert.Equal(t, id, ctx["transaction_id"])
	assert.Equal(t, float64(422), ctx["status"])
}

func TestResponseErrorDanPenolakanAksesIkutTercatat(t *testing.T) {
	rc := newReceiver(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL})

	require.Equal(t, 422, a.call(t, "GET", "/api/anak/x", "key-jakantro", "").StatusCode)
	require.Equal(t, 403, a.call(t, "GET", "/api/anak/x", "key-salah", "").StatusCode)

	byStatus := map[float64]map[string]any{}
	for _, c := range rc.waitFor(t, 2) {
		byStatus[c["status"].(float64)] = c
	}
	require.Contains(t, byStatus, float64(422))
	require.Contains(t, byStatus, float64(403))
	assert.Equal(t, "id anak tidak valid", byStatus[422]["response"].(map[string]any)["message"])
	assert.Equal(t, float64(4002), byStatus[403]["response"].(map[string]any)["errorCode"])
	assert.Nil(t, byStatus[403]["orgid"], "key ditolak sebelum orgid diketahui")
}

func TestApiKeyTidakPernahTercatat(t *testing.T) {
	rc := newReceiver(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL})
	a.call(t, "POST", "/api/kunjungan", "key-jakantro", payload)

	b, err := json.Marshal(rc.waitFor(t, 1)[0])
	require.NoError(t, err)
	assert.NotContains(t, string(b), "key-jakantro")
}

func TestKredensialYangMunculDiResponseDisamarkan(t *testing.T) {
	rc := newReceiver(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL})

	require.Equal(t, 401, a.call(t, "GET", "/api/anak/x", "tak-dikenal-123456", "").StatusCode)
	require.Equal(t, 401, a.call(t, "GET", "/api/anak/x", "", "", "Authorization", "Bearer token-rahasia-9876").StatusCode)

	b, err := json.Marshal(rc.waitFor(t, 2))
	require.NoError(t, err)
	assert.NotContains(t, string(b), "tak-dikenal-123456")
	assert.NotContains(t, string(b), "token-rahasia-9876")
	assert.Contains(t, string(b), "User not found: ***************456", "pesan tetap terbaca, hanya kuncinya yang disamarkan")
}

func TestIdTransaksiDariKlienDipakaiUlangKalauSah(t *testing.T) {
	rc := newReceiver(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL})

	res := a.call(t, "POST", "/api/kunjungan", "key-jakantro", payload, "X-Transaction-ID", "klien-123")
	assert.Equal(t, "klien-123", idOf(res))

	res = a.call(t, "POST", "/api/kunjungan", "key-jakantro", payload, "X-Transaction-ID", "ada spasi <script>")
	replaced := idOf(res)
	assert.NotEqual(t, "ada spasi <script>", replaced, "id yang tidak sah diganti sistem")
	assert.NotEmpty(t, replaced)

	rc.exactlyOnce(t, []string{"klien-123", replaced})
}

func TestIdTransaksiTidakDisamarkan(t *testing.T) {
	rc := newReceiver(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL})
	// angka panjang seperti NIK, tapi id transaksi harus tetap utuh
	id := "20261102000123"
	a.call(t, "GET", "/api/anak/x", "key-jakantro", "", "X-Transaction-ID", id)

	rc.waitFor(t, 1)
	rc.mu.Lock()
	defer rc.mu.Unlock()
	assert.Equal(t, id, rc.envelopes[0]["identifier"])
	assert.Equal(t, id, rc.records[0]["transaction_id"])
}

func TestNamaHeaderIdTransaksiBisaDiaturLewatConfig(t *testing.T) {
	rc := newReceiver(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL, Header: "X-Jejak"})

	res := a.call(t, "POST", "/api/kunjungan", "key-jakantro", payload)
	assert.NotEmpty(t, res.Header.Get("X-Jejak"))
	assert.Empty(t, idOf(res))
	assert.Equal(t, res.Header.Get("X-Jejak"), rc.waitFor(t, 1)[0]["transaction_id"])
}

func TestEndpointDiBlacklistTidakDicatat(t *testing.T) {
	rc := newReceiver(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL,
		Blacklist: []string{"/api/kafka/*", "GET /api/anak/*"}})

	a.call(t, "GET", "/api/kafka/gagal", "key-jakantro", "")
	res := a.call(t, "GET", "/api/anak/x", "key-jakantro", "")
	assert.NotEmpty(t, idOf(res), "header tetap dipasang walau tidak dicatat")
	a.call(t, "POST", "/api/kunjungan", "key-jakantro", payload)

	records := rc.waitFor(t, 1)
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, 1, rc.count(), "hanya POST /api/kunjungan yang dicatat")
	assert.Equal(t, "/api/kunjungan", records[0]["path"])
}

func TestHeaderTambahanDikirimKeServiceLog(t *testing.T) {
	rc := newReceiver(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL,
		Headers: map[string]string{"Authorization": "Bearer rahasia-log"}})
	a.call(t, "POST", "/api/kunjungan", "key-jakantro", payload)

	rc.waitFor(t, 1)
	rc.mu.Lock()
	defer rc.mu.Unlock()
	assert.Equal(t, "Bearer rahasia-log", rc.headers[0].Get("Authorization"))
	assert.Equal(t, "application/json", rc.headers[0].Get("Content-Type"))
}

func TestHeaderDariEnvDigabungDenganHeaderConfig(t *testing.T) {
	rc := newReceiver(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL,
		Headers:    map[string]string{"x-api-name": "dari-yaml", "x-sumber": "yaml"},
		RawHeaders: " x-api-key= rahasia-log , x-api-name=dari-env "})
	a.call(t, "POST", "/api/kunjungan", "key-jakantro", payload)

	rc.waitFor(t, 1)
	rc.mu.Lock()
	defer rc.mu.Unlock()
	h := rc.headers[0]
	assert.Equal(t, "rahasia-log", h.Get("X-Api-Key"), "spasi di sekitar key dan nilai dibuang")
	assert.Equal(t, "dari-env", h.Get("X-Api-Name"), "env menang atas config.yaml")
	assert.Equal(t, "yaml", h.Get("X-Sumber"), "header yaml lain tetap terkirim")
}

func TestFormatHeaderEnvSalahMenggagalkanStart(t *testing.T) {
	for _, raw := range []string{"x-api-key", "x-api-key=abc,=tanpa-nama"} {
		_, err := logtransaksi.New(config.LogTransaksiConfig{Enabled: true, Endpoint: "http://localhost:1",
			RedisKey: testKey(t), RawHeaders: raw}, rdb)
		assert.Error(t, err, "format %q", raw)
	}
}

func TestPanicTercatatSebagai500DanTetapDitanganiRecover(t *testing.T) {
	rc := newReceiver(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL})

	res := a.call(t, "GET", "/api/panik", "key-jakantro", "")
	assert.Equal(t, 500, res.StatusCode, "recover tetap menulis 500")

	c := rc.waitFor(t, 1)[0]
	assert.Equal(t, "/api/panik", c["path"])
	assert.Equal(t, float64(500), c["status"])
}

func TestServiceLogMatiTidakMemperlambatRequestDanCatatanTetapDiRedis(t *testing.T) {
	rc := newReceiver(t)
	rc.srv.Close()
	key := testKey(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL, Timeout: time.Second,
		RedisKey: key, RetryInterval: time.Hour})

	start := time.Now()
	for range 5 {
		require.Equal(t, 201, a.call(t, "POST", "/api/kunjungan", "key-jakantro", payload).StatusCode)
	}
	assert.Less(t, time.Since(start), 500*time.Millisecond, "pengiriman ke service log di background")
	assert.Eventually(t, func() bool { return a.sender.Failed() > 0 && len(queued(t, key)) == 5 },
		3*time.Second, 10*time.Millisecond, "kiriman gagal dikembalikan, semua catatan tetap di Redis")
	assert.Zero(t, a.sender.Lost())
	assert.NotContains(t, strings.Join(queued(t, key), ""), "3175041503240003", "isi Redis sudah disamarkan")
}

func TestCatatanSaatServiceLogGagalDikirimUlangSetelahPulih(t *testing.T) {
	rc := newReceiver(t)
	rc.set(http.StatusServiceUnavailable, 0)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL})

	var ids []string
	for range 5 {
		ids = append(ids, idOf(a.call(t, "POST", "/api/kunjungan", "key-jakantro", payload)))
	}
	require.Eventually(t, func() bool { return a.sender.Failed() > 0 }, 3*time.Second, 10*time.Millisecond)
	time.Sleep(200 * time.Millisecond)
	assert.Zero(t, rc.count(), "service log masih menolak")

	rc.set(http.StatusOK, 0)
	rc.exactlyOnce(t, ids)
	assert.Zero(t, a.sender.Lost())
}

func TestBanyakTransaksiSekaligusTercatatSemua(t *testing.T) {
	rc := newReceiver(t)
	rc.set(http.StatusOK, 20*time.Millisecond)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL})

	var mu sync.Mutex
	var ids []string
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := idOf(a.call(t, "POST", "/api/kunjungan", "key-jakantro", payload))
			mu.Lock()
			ids = append(ids, id)
			mu.Unlock()
		}()
	}
	wg.Wait()
	rc.exactlyOnce(t, ids)
	assert.Zero(t, a.sender.Lost())
}

func TestCatatanYangTerkirimDihapusDariRedis(t *testing.T) {
	rc := newReceiver(t)
	key := testKey(t)
	a := newApp(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL, RedisKey: key})

	var ids []string
	for range 10 {
		ids = append(ids, idOf(a.call(t, "POST", "/api/kunjungan", "key-jakantro", payload)))
	}
	rc.exactlyOnce(t, ids)
	assert.Empty(t, queued(t, key), "tidak ada sisa di antrean setelah terkirim")
	assert.Zero(t, rdb.Exists(context.Background(), key).Val())
}

func TestWorkerMengirimParalelSesuaiConfig(t *testing.T) {
	rc := newReceiver(t)
	rc.set(http.StatusOK, 200*time.Millisecond)
	key := testKey(t)
	s := newSender(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL, RedisKey: key, Workers: 20})

	var ids []string
	for i := range 40 {
		id := fmt.Sprintf("trx-%d", i)
		ids = append(ids, id)
		s.Push(logtransaksi.Record{TransactionID: id})
	}
	start := time.Now()
	rc.exactlyOnce(t, ids)
	// berurutan butuh 40 x 200ms = 8 detik; 20 worker cukup 2 putaran
	assert.Less(t, time.Since(start), 2*time.Second)
	assert.Empty(t, queued(t, key))
}

func TestCatatanDariPodLainIkutDikirim(t *testing.T) {
	down := newReceiver(t)
	down.srv.Close()
	rc := newReceiver(t)
	key := testKey(t)

	podA := newSender(t, config.LogTransaksiConfig{Enabled: true, Endpoint: down.srv.URL, RedisKey: key, RetryInterval: time.Hour})
	var ids []string
	for i := range 4 {
		id := fmt.Sprintf("trx-%d", i)
		ids = append(ids, id)
		podA.Push(logtransaksi.Record{TransactionID: id})
	}
	require.Eventually(t, func() bool { return podA.Failed() > 0 && len(queued(t, key)) == 4 }, 3*time.Second, 10*time.Millisecond)

	newSender(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL, RedisKey: key})
	rc.exactlyOnce(t, ids)
}

func TestCloseSaatServiceLambatCatatanKembaliKeRedis(t *testing.T) {
	rc := newReceiver(t)
	rc.set(http.StatusOK, 2*time.Second)
	key := testKey(t)
	cfg := config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL, RedisKey: key, RetryInterval: 50 * time.Millisecond}
	s, err := logtransaksi.New(cfg, rdb)
	require.NoError(t, err)

	var ids []string
	for i := range 6 {
		id := fmt.Sprintf("trx-%d", i)
		ids = append(ids, id)
		s.Push(logtransaksi.Record{TransactionID: id})
	}
	time.Sleep(100 * time.Millisecond) // worker sedang mengirim
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	s.Close(ctx)
	assert.Less(t, time.Since(start), time.Second, "shutdown tidak menunggu service log yang lambat")
	assert.Len(t, queued(t, key), 6, "kiriman yang dibatalkan dikembalikan ke Redis")

	rc.set(http.StatusOK, 0)
	newSender(t, cfg)
	rc.exactlyOnce(t, ids)
}

func TestCatatanRusakDiRedisDilewatiTanpaMenahanSisanya(t *testing.T) {
	rc := newReceiver(t)
	key := testKey(t)
	require.NoError(t, rdb.RPush(context.Background(), key,
		`{"identifier":"trx-1","context":{"transaction_id":"trx-1"}}`, `{"identifier": rusak`,
		`{"identifier":"trx-2","context":{"transaction_id":"trx-2"}}`).Err())

	s := newSender(t, config.LogTransaksiConfig{Enabled: true, Endpoint: rc.srv.URL, RedisKey: key})
	rc.exactlyOnce(t, []string{"trx-1", "trx-2"})
	assert.Equal(t, int64(1), s.Lost(), "catatan rusak dihitung, tidak diam-diam hilang")
	assert.Empty(t, queued(t, key))
}

func TestRedisTidakTerjangkauMenggagalkanStart(t *testing.T) {
	down := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	defer down.Close()
	_, err := logtransaksi.New(config.LogTransaksiConfig{Enabled: true, Endpoint: "http://localhost:1", Timeout: time.Second}, down)
	assert.Error(t, err)
}
