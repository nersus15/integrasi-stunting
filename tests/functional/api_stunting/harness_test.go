// Package api_stunting_test menjalankan skenario end-to-end lewat HTTP, lalu menulis
// laporan HTML berisi payload, kode HTTP, dan response penuh setiap skenario.
//
//	go test ./functional/api_stunting/
//
// Aplikasi dijalankan in-process; API_BASE diisi untuk menguji service yang sudah berjalan.
// DSN dan lokasi laporan lewat PG_DSN/PG_*, LAPORAN_HTML.
package api_stunting_test

import (
	"bytes"
	"context"
	databasesql "database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var orgidFaskes = testutil.Env("ORGID_FASKES", "100011")

var (
	baseURL    string
	healthURL  string
	dsn        = testutil.DSN()
	berkasHTML = testutil.Env("LAPORAN_HTML",
		filepath.Join(testutil.RepoRoot(), "tests", "functional", "api_stunting", "laporan.html"))

	// Kunci bawaan seluruh skenario. Sejak auth aktif, request tanpa kredensial
	// dijawab 401 sebelum sampai ke handler.
	//
	// Nilainya diambil dari env, tidak ditanam di source: kunci yang tertulis di
	// repo ikut terbawa ke mana-mana -- termasuk ke laporan HTML yang diterbitkan.
)

// tanpaKunci dipakai di field kunci untuk menandai request sengaja dikirim
// tanpa header kredensial. String kosong berarti "pakai kunci bawaan", jadi
// perlu penanda tersendiri.
const tanpaKunci = "-"

// skenario adalah satu kasus uji. Field pemeriksaan yang bernilai nol tidak
// diperiksa, sehingga skenario sukses cukup menyebut http saja.
type skenario struct {
	bagian, label string
	metode, path  string
	payload       any    // nil: tanpa body. string: body mentah apa adanya.
	kunci         string // kosong: kunci bawaan. tanpaKunci: tanpa header.
	http          int
	errorCode     int
	errorName     string
	pesan         string
}

// rekaman menyimpan hasil satu skenario untuk dituangkan ke laporan.
type rekaman struct {
	Bagian   string         `json:"bagian"`
	Label    string         `json:"label"`
	Metode   string         `json:"metode"`
	Path     string         `json:"path"`
	Http     int            `json:"http"`
	Payload  any            `json:"payload"`
	Kunci    string         `json:"kunci"`
	Response map[string]any `json:"response"`
	Mentah   string         `json:"-"`
	Salah    string         `json:"salah"`
}

var rekam []rekaman

var client = &http.Client{Timeout: 20 * time.Second}

func send(method, path string, payload any, key string) (int, map[string]any, string, error) {
	var reqBody io.Reader
	if payload != nil {
		switch v := payload.(type) {
		case string:
			reqBody = bytes.NewBufferString(v)
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return 0, nil, "", err
			}
			reqBody = bytes.NewBuffer(b)
		}
	}

	// Path yang sudah berupa URL penuh dipakai apa adanya, untuk menjangkau
	// endpoint di luar prefix API seperti /stunting/health. Jangan memakai
	// "/../" -- Go mengirimnya mentah (curl menormalkannya), sehingga request
	// tetap dianggap berada di bawah prefix API.
	url := baseURL + path
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		url = path
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return 0, nil, "", err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	switch key {
	case tanpaKunci: // sengaja tanpa kredensial
	case "":
		req.Header.Set("X-API-Key", testutil.KeyJakantro)
	default:
		req.Header.Set("X-API-Key", key)
	}

	res, err := client.Do(req)
	if err != nil {
		return 0, nil, "", err
	}
	defer res.Body.Close()

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, nil, "", err
	}

	var terurai map[string]any
	_ = json.Unmarshal(content, &terurai)
	return res.StatusCode, terurai, string(content), nil
}

// jalankan mengeksekusi satu skenario dan mencatat hasilnya.
func run(t *testing.T, s skenario) {
	t.Helper()

	code, resp, mentah, err := send(s.metode, s.path, s.payload, s.kunci)
	if err != nil {
		t.Fatalf("request gagal: %v", err)
	}

	var salah []string
	if code != s.http {
		salah = append(salah, fmt.Sprintf("http %d, harap %d", code, s.http))
	}
	if s.errorCode != 0 {
		if got := number(resp["errorCode"]); got != s.errorCode {
			salah = append(salah, fmt.Sprintf("errorCode %d, harap %d", got, s.errorCode))
		}
	}
	if s.errorName != "" {
		if got, _ := resp["errorName"].(string); got != s.errorName {
			salah = append(salah, fmt.Sprintf("errorName %q, harap %q", got, s.errorName))
		}
	}
	if s.pesan != "" {
		got, _ := resp["message"].(string)
		if !contains(got, s.pesan) {
			salah = append(salah, fmt.Sprintf("pesan %q tidak memuat %q", got, s.pesan))
		}
	}

	rekam = append(rekam, rekaman{
		Bagian: s.bagian, Label: s.label, Metode: s.metode, Path: s.path,
		Http: code, Payload: s.payload, Response: resp, Mentah: mentah,
		Kunci: keyLabel(s.kunci),
		Salah: join(salah),
	})

	for _, m := range salah {
		t.Error(m)
	}
}

func number(v any) int {
	f, _ := v.(float64)
	return int(f)
}

func contains(teks, find string) bool {
	return len(find) == 0 || bytes.Contains(
		bytes.ToLower([]byte(teks)), bytes.ToLower([]byte(find)))
}

func join(s []string) string {
	if len(s) == 0 {
		return ""
	}
	out := s[0]
	for _, x := range s[1:] {
		out += "; " + x
	}
	return out
}

// ---------------------------------------------------------------- database

func openDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb := databasesql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())
	t.Cleanup(func() { sqldb.Close() })
	return db
}

func fetchIdPosyandu(t *testing.T) string {
	t.Helper()
	db := openDB(t)

	var id string
	err := db.QueryRowContext(context.Background(),
		`SELECT p.id FROM stunting.posyandu p
		 JOIN stunting.faskes f ON f.id = p.id_puskesmas
		 WHERE f.satusehat_id = ?
		 ORDER BY p.id LIMIT 1`, orgidFaskes).Scan(&id)

	if err != nil {
		err = db.QueryRowContext(context.Background(),
			`SELECT id FROM stunting.posyandu ORDER BY id LIMIT 1`).Scan(&id)
	}
	if err != nil {
		t.Fatalf("gagal mengambil id posyandu: %v", err)
	}
	return id
}

func truncateDB(t *testing.T) {
	t.Helper()
	db := openDB(t)
	_, err := db.ExecContext(context.Background(),
		`TRUNCATE stunting.kunjungan, stunting.kesehatan, stunting.anak, stunting.orangtua CASCADE`)
	if err != nil {
		t.Fatalf("gagal mengosongkan database: %v", err)
	}
}

// ambilTabel menjalankan query dan mengembalikan barisnya sebagai teks.
func fetchTable(t *testing.T, db *bun.DB, sql string, kolom int) [][]string {
	t.Helper()
	row, err := db.QueryContext(context.Background(), sql)
	if err != nil {
		t.Fatalf("query gagal: %v", err)
	}
	defer row.Close()

	var out [][]string
	for row.Next() {
		sel := make([]any, kolom)
		ptr := make([]any, kolom)
		for i := range sel {
			ptr[i] = &sel[i]
		}
		if err := row.Scan(ptr...); err != nil {
			t.Fatalf("scan gagal: %v", err)
		}
		r := make([]string, kolom)
		for i, v := range sel {
			switch x := v.(type) {
			case nil:
				r[i] = "<NULL>"
			case []byte:
				r[i] = string(x)
			case time.Time:
				r[i] = x.Format("2006-01-02")
			default:
				r[i] = fmt.Sprint(x)
			}
		}
		out = append(out, r)
	}
	return out
}

// ---------------------------------------------------------------- kesiapan

func serviceAvailable() bool {
	c := &http.Client{Timeout: 3 * time.Second}
	res, err := c.Get(healthURL)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	return res.StatusCode == http.StatusOK
}

func databaseAvailable() bool {
	sqldb := databasesql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	defer sqldb.Close()
	ctx, batal := context.WithTimeout(context.Background(), 3*time.Second)
	defer batal()
	return sqldb.PingContext(ctx) == nil
}

func writeFile(name, content string) error {
	return os.WriteFile(name, []byte(content), 0o644)
}

// labelKunci menerjemahkan field kunci jadi keterangan singkat untuk laporan,
// tanpa menampilkan nilai kuncinya.
func keyLabel(k string) string {
	switch k {
	case tanpaKunci:
		return "tanpa kredensial"
	case "":
		return "kunci write"
	case testutil.KeyReadonly:
		return "kunci readonly"
	default:
		return "kunci lain"
	}
}
