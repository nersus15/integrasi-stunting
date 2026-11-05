package create_kunjungan_test

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var app *testutil.FunctionalApp

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		fmt.Fprintln(os.Stderr, "postgres mati, test dilewati")
		os.Exit(0)
	}

	built, err := testutil.NewFunctionalAppAuth(2089, testutil.KeyJakantro)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi untuk functional test: %v\n", err)
		os.Exit(1)
	}
	app = built

	code := m.Run()
	app.Close()
	os.Exit(code)
}

func post(t *testing.T, path string, body []byte) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, "http://localhost"+path, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

const path = testutil.APIPrefix + "/kunjungan"

// --- POSITIVE CASE ---

func TestCatatKunjungan_BentukSahMengembalikan201(t *testing.T) {
	tests := []struct {
		nama      string
		body      func(testing.TB) []byte
		prasyarat func() error
	}{
		{"registrasi lengkap", testutil.BodyRegistrasiLengkap, nil},
		{"anak baru 3 key", testutil.BodyAnakBaru3Key, testutil.SeedOrangtua},
		{"anak baru 2 key", testutil.BodyAnakBaru2Key, testutil.SeedOrangtua},
		{"kunjungan saja 3 key", testutil.BodyKunjunganSaja3Key, testutil.SeedOrangtuaAndAnak},
		{"kunjungan saja 2 key", testutil.BodyKunjunganSaja2Key, testutil.SeedOrangtuaAndAnak},
		{"kunjungan saja 1 key", testutil.BodyKunjunganSaja1Key, testutil.SeedOrangtuaAndAnak},
		{"kunjungan saja flat", testutil.BodyKunjunganFlat, testutil.SeedOrangtuaAndAnak},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			require.NoError(t, testutil.CleanFixtures())
			if tc.prasyarat != nil {
				require.NoError(t, tc.prasyarat())
			}

			resp := post(t, path, tc.body(t))
			defer resp.Body.Close()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		})
	}
}

func TestBackgroundWorker_BerjalanSetelahModulInit(t *testing.T) {
	require.NotNil(t, app.Background, "modul harus memuat library background worker")

	assert.True(t, app.Background.Stats().Running,
		"pool worker harus sudah jalan setelah Module.Init, kalau tidak job kompensasi hanya menumpuk di antrian")
}

// --- NEGATIVE CASE ---

func TestCatatKunjungan_PayloadCacatDitolak(t *testing.T) {
	tests := []struct {
		nama  string
		body  string
		harap int
	}{
		{"json rusak", `{"kunjungan":`, http.StatusBadRequest},
		{"body kosong", ``, http.StatusBadRequest},
		{"objek kosong", `{}`, http.StatusUnprocessableEntity},
		{"tanpa kunjungan", `{"anak":{"id":"A1","id_orangtua":"O1","nama":"Ahmad"}}`, http.StatusUnprocessableEntity},
		{"flat dicampur nested", `{"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16","kunjungan":{"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16"}}`, http.StatusUnprocessableEntity},
		{"anak tanpa id_orangtua", `{"anak":{"id":"A1","nama":"Ahmad"},"kunjungan":{"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16"}}`, http.StatusUnprocessableEntity},
		{"orangtua tanpa anak", `{"orangtua":{"id":"O1","nik":"3515081203800002"},"kunjungan":{"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16"}}`, http.StatusUnprocessableEntity},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			resp := post(t, path, []byte(tc.body))
			defer resp.Body.Close()

			assert.Equal(t, tc.harap, resp.StatusCode)
		})
	}
}

func TestCatatKunjungan_MethodSelainPost(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, "http://localhost"+path, nil)
			require.NoError(t, err)

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		})
	}
}

func TestCatatKunjungan_PathTanpaPrefix(t *testing.T) {
	resp := post(t, "/kunjungan", testutil.BodyKunjunganSaja1Key(t))
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
