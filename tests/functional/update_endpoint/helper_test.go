package update_endpoint_test

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
	Kode  int
	Isi   map[string]any
	Galat int
}

func request(t *testing.T, method, path, key string, body any) jawaban {
	t.Helper()
	var r io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		r = bytes.NewBufferString(b)
	default:
		content, err := json.Marshal(b)
		require.NoError(t, err)
		r = bytes.NewReader(content)
	}

	req, err := http.NewRequest(method, app.BaseURL+path, r)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("X-API-Key", key)
	}
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	j := jawaban{Kode: res.StatusCode}
	content, _ := io.ReadAll(res.Body)
	_ = json.Unmarshal(content, &j.Isi)
	if k, ok := j.Isi["errorCode"].(float64); ok {
		j.Galat = int(k)
	}
	return j
}

func setupRiwayat(t *testing.T) {
	t.Helper()
	cleanup := func() { require.NoError(t, testutil.Scenario1Cleanup(db, testGroup)) }
	cleanup()
	t.Cleanup(cleanup)
	forgetPembuat()
	require.NoError(t, wiring.Scenario1Jakantro())
	require.NoError(t, wiring.Scenario1Send(testutil.Scenario1KafkaFiles...))
}

func scalar[T any](t *testing.T, q string, args ...any) T {
	t.Helper()
	var v T
	require.NoError(t, db.QueryRow(q, args...).Scan(&v), q)
	return v
}

var pemilikTercache = map[string][]string{
	"kunjungan": {"enc-pkm-01", "enc-rs-01", "enc-langsung-01"},
	"observasi": {"obs-pkm-bb-01", "obs-rs-bb-01", "obs-langsung-bb"},
	"diagnosa":  {"cond-pkm-01", "cond-rs-01", "cond-langsung-01"},
	"layanan":   {"proc-pkm-02", "proc-rs-01", "proc-langsung-01"},
	"rujukan":   {"sr-rujuk-01", "sr-balik-01"},
	"episode":   {"ep-stunting-sim1"},
}

func forgetPembuat() {
	for jenis, list := range pemilikTercache {
		for _, id := range list {
			_ = wiring.Memory.Delete(jenis + "::" + id)
		}
	}
}
