package optimasi_query_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/require"
)

type jawaban struct {
	Kode   int
	Isi    map[string]any
	Mentah []byte
}

func request(t *testing.T, method, path, key string, body any) jawaban {
	t.Helper()
	var r io.Reader
	if body != nil {
		content, err := json.Marshal(body)
		require.NoError(t, err)
		r = bytes.NewReader(content)
	}
	req, err := http.NewRequest(method, app.BaseURL+path, r)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", key)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	j := jawaban{Kode: res.StatusCode}
	j.Mentah, _ = io.ReadAll(res.Body)
	_ = json.Unmarshal(j.Mentah, &j.Isi)
	return j
}

func scalar[T any](t *testing.T, q string, args ...any) T {
	t.Helper()
	var v T
	require.NoError(t, db.QueryRow(q, args...).Scan(&v))
	return v
}

func exec(t *testing.T, q string, args ...any) {
	t.Helper()
	_, err := db.Exec(q, args...)
	require.NoError(t, err, q)
}

// skenario 1 lengkap: posyandu -> puskesmas -> RS -> rujuk balik -> lanjutan
func setupRiwayat(t *testing.T) {
	t.Helper()
	setupScenario1(t, testutil.Scenario1KafkaFiles...)
}

// belum pernah dirujuk: akses faskes A murni dari posyandu
func setupBelumDirujuk(t *testing.T) {
	t.Helper()
	setupScenario1(t, "02-kafka-patient.json", "03-kafka-relatedperson.json")
}

func setupScenario1(t *testing.T, kafka ...string) {
	t.Helper()
	cleanup := func() {
		require.NoError(t, testutil.Scenario1Cleanup(db, groupUji))
		cleanupExtras()
	}
	cleanup()
	t.Cleanup(cleanup)
	require.NoError(t, wiring.Scenario1Jakantro())
	require.NoError(t, wiring.Scenario1Send(kafka...))
	forgetCache()
}

// baris di luar skenario 1 yang dibuat paket ini; turunan anak ikut terhapus lewat cascade
func cleanupExtras() {
	_, _ = db.Exec(`DELETE FROM stunting.anak WHERE id LIKE 'uji-opt-%' OR nik LIKE '31750477%'`)
	_, _ = db.Exec(`DELETE FROM stunting.orangtua WHERE id LIKE 'uji-opt-%' OR nik LIKE '31750477%'`)
	_, _ = db.Exec(`DELETE FROM stunting.posyandu WHERE id LIKE 'UJI-OPT-%'`)
	_, _ = db.Exec(`DELETE FROM stunting.faskes WHERE id LIKE 'UJI-OPT-%'`)
}

func forgetCache() {
	wiring.Repository.LupakanRiwayatAnak(testutil.Scenario1IdAnak)
	wiring.Repository.LupakanRujukanTerakhir(testutil.Scenario1IdAnak)
}

// harapQuery: jumlah statement yang cocok dengan pola harus tepat; bila beda, hanya statement itu yang dicetak
func expectQuery(t *testing.T, q testutil.Statement, pattern string, want int, apa string) {
	t.Helper()
	if c := q.Match(pattern); len(c) != want {
		t.Errorf("%s: harap %d statement, ada %d\n%s", apa, want, len(c), stringsJoin(c))
	}
}

// harapQueryMaks: jumlah statement yang cocok tidak boleh melebihi maks
func expectQueryMax(t *testing.T, q testutil.Statement, pattern string, maks int, apa string) {
	t.Helper()
	if c := q.Match(pattern); len(c) > maks {
		t.Errorf("%s: maksimal %d statement, ada %d\n%s", apa, maks, len(c), stringsJoin(c))
	}
}
