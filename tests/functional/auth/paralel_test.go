package auth_test

import (
	"bytes"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nersus15/integrasi/tests/testutil"
)

// Uji ujung-ke-ujung terhadap service yang sedang berjalan: dua kunci dengan
// hak berbeda ditembakkan bersamaan, berkali-kali.
//
//	kunci write     -> POST /api/kunjungan lolos auth, gagal validasi  = 422
//	kunci readonly  -> ditolak RBAC sebelum sampai handler             = 403
//
// Selisih 422 vs 403 itulah alatnya. Kalau identitas antar request tertukar,
// kunci readonly akan sesekali mendapat 422 (naik hak) atau kunci write
// mendapat 403 (turun hak). Keduanya kegagalan keras, bukan tafsiran.
//
// Perbaikannya pernah diuji pada kondisi terburuk: latensi buatan 100ms di
// pemuatan data auth dari database, 2000 request serentak, tiga putaran, nol
// penyimpangan. Instrumen latensi itu sudah dibuang setelah terkonfirmasi.
//
// Rate limiter service (bawaan 1000/menit) menjawab 429 bila PASANG dinaikkan
// terlalu tinggi. Kode selain 422/403 dicatat di sebaran, tidak dihitung
// sebagai pertukaran identitas.
//
// Aplikasi dijalankan in-process; API_BASE diisi untuk menguji service yang sudah berjalan.
//
//	API_BASE=http://localhost:2025/api PASANG=200 go test -count=1 -run Paralel ./functional/auth/
var ()

func testService(t *testing.T) (base, health string) {
	if base, health, exists := testutil.ExternalService(); exists {
		return base, health
	}
	if !testutil.DatabaseAvailable() {
		t.Skip("database tidak hidup, aplikasi in-process tidak bisa dijalankan")
	}
	app, err := testutil.StartApp(2085)
	if err != nil {
		t.Fatalf("gagal menjalankan aplikasi in-process: %v", err)
	}
	t.Cleanup(app.Stop)
	return app.BaseURL, app.Health
}

type result struct {
	kunci string
	kode  int
	harap int
}

func TestParalelIdentitasTidakTertukar(t *testing.T) {
	basis, health := testService(t)
	client := &http.Client{Timeout: 30 * time.Second}
	defer client.CloseIdleConnections() // tanpa ini Stop() aplikasi menunggu koneksi keep-alive

	if _, err := client.Get(health); err != nil {
		t.Skipf("service di %s tidak menyala: %v", health, err)
	}

	tembak := func(key string) int {
		req, err := http.NewRequest(http.MethodPost, basis+"/kunjungan",
			bytes.NewBufferString(`{"tidak":"valid"}`))
		if err != nil {
			return -1
		}
		req.Header.Set("X-API-Key", key)
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			return -1
		}
		defer res.Body.Close()
		return res.StatusCode
	}

	// 1 pasang = 1 request write + 1 request readonly, ditembakkan bersamaan.
	pasang := 200
	if v := os.Getenv("PASANG"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			pasang = n
		}
	}
	var wg sync.WaitGroup
	start := make(chan struct{})

	menyimpang := make(chan result, pasang*2)
	var naikHak, turunHak, failed int64

	// Rekam sebaran kode per kunci. Tanpa ini, kode tak terduga yang sebenarnya
	// berasal dari rate limiter (429) akan salah dibaca sebagai masalah otorisasi.
	var mu sync.Mutex
	sebaran := map[string]map[int]int{"write": {}, "readonly": {}}

	send := func(name, key string, want int) {
		defer wg.Done()
		<-start

		code := tembak(key)

		mu.Lock()
		sebaran[name][code]++
		mu.Unlock()

		if code == want {
			return
		}
		if code < 0 {
			atomic.AddInt64(&failed, 1)
			return
		}
		// Hanya 422 vs 403 yang berbicara soal identitas. Kode lain (mis. 429
		// dari rate limiter) dicatat di sebaran, tidak dihitung sebagai
		// pertukaran identitas.
		switch {
		case want == http.StatusForbidden && code == http.StatusUnprocessableEntity:
			atomic.AddInt64(&naikHak, 1)
		case want == http.StatusUnprocessableEntity && code == http.StatusForbidden:
			atomic.AddInt64(&turunHak, 1)
		default:
			return
		}
		menyimpang <- result{kunci: key, kode: code, harap: want}
	}

	mulaiWaktu := time.Now()
	for i := 0; i < pasang; i++ {
		wg.Add(2)
		go send("write", testutil.KeyJakantro, http.StatusUnprocessableEntity)
		go send("readonly", testutil.KeyReadonly, http.StatusForbidden)
	}
	close(start)
	wg.Wait()
	close(menyimpang)

	total := pasang * 2
	t.Logf("%d request serentak dalam %v", total, time.Since(mulaiWaktu).Round(time.Millisecond))

	for _, name := range []string{"write", "readonly"} {
		t.Logf("sebaran kode %-9s : %v", name, sebaran[name])
	}

	if failed > 0 {
		t.Errorf("%d request gagal terkirim", failed)
	}

	n := 0
	for h := range menyimpang {
		if n < 5 {
			t.Errorf("kunci %.12s... dapat http %d, harap %d", h.kunci, h.kode, h.harap)
		}
		n++
	}

	if naikHak > 0 {
		t.Errorf("KENAIKAN HAK: %d request readonly lolos otorisasi", naikHak)
	}
	if turunHak > 0 {
		t.Errorf("penurunan hak: %d request write ditolak", turunHak)
	}
	if n == 0 {
		t.Logf("semua %d request mendapat hasil sesuai identitasnya sendiri", total)
	}
}
