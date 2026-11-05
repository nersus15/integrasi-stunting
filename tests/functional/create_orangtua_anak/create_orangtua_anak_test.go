package create_orangtua_anak_test

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

	built, err := testutil.NewFunctionalAppAuth(2088, testutil.KeyJakantro)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi: %v\n", err)
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

const (
	pathOrangtua = testutil.APIPrefix + "/orangtua"
	pathAnak     = testutil.APIPrefix + "/anak"
)

// --- POSITIVE CASE ---

func TestDaftarOrangtua_Berhasil(t *testing.T) {
	resp := post(t, pathOrangtua, testutil.JSONBody(t, testutil.OrangtuaPayloadValid()))
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestDaftarAnak_Berhasil(t *testing.T) {
	resp := post(t, pathAnak, testutil.JSONBody(t, testutil.AnakPayloadValid()))
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode,
		"endpoint anak harus menyimpan dan mengembalikan 201, bukan respons kosong")
}

// --- NEGATIVE CASE ---

func TestDaftarOrangtua_PayloadCacat(t *testing.T) {
	tests := []struct {
		nama  string
		body  string
		harap int
	}{
		{"json rusak", `{"nik":`, http.StatusBadRequest},
		{"body kosong", ``, http.StatusBadRequest},
		{"objek kosong", `{}`, http.StatusUnprocessableEntity},
		{"tanpa id", `{"id_posyandu":"P1","no_kk":"3515081203000001","nama_ayah":"Budi","nama_ibu":"Siti","nik":"3515081203800002","telepon":"08","rt":"003","rw":"002","alamat":"Jl"}`, http.StatusUnprocessableEntity},
		{"nik bukan 16 digit", `{"id":"O1","id_posyandu":"P1","no_kk":"3515081203000001","nama_ayah":"Budi","nama_ibu":"Siti","nik":"351508","telepon":"08","rt":"003","rw":"002","alamat":"Jl"}`, http.StatusUnprocessableEntity},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			resp := post(t, pathOrangtua, []byte(tc.body))
			defer resp.Body.Close()

			assert.Equal(t, tc.harap, resp.StatusCode)
		})
	}
}

func TestDaftarAnak_PayloadCacat(t *testing.T) {
	tests := []struct {
		nama  string
		body  string
		harap int
	}{
		{"json rusak", `{"nik":`, http.StatusBadRequest},
		{"objek kosong", `{}`, http.StatusUnprocessableEntity},
		{"tanpa id_orangtua", `{"id":"A1","nama":"Ahmad","nik":"3515081203230001","tanggal_lahir":"2024-05-10","jenis_kelamin":"L","anak_ke":1}`, http.StatusUnprocessableEntity},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			resp := post(t, pathAnak, []byte(tc.body))
			defer resp.Body.Close()

			assert.Equal(t, tc.harap, resp.StatusCode,
				"payload cacat harus ditolak, bukan diterima diam-diam")
		})
	}
}

func TestMethodSelainPost(t *testing.T) {
	t.Run("PUT "+pathOrangtua+" terdaftar", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, "http://localhost"+pathOrangtua, nil)
		require.NoError(t, err)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode,
			"PUT /orangtua sudah punya handler sendiri")
	})

	t.Run("PUT "+pathAnak+" terdaftar", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, "http://localhost"+pathAnak, nil)
		require.NoError(t, err)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode,
			"PUT /anak sudah punya handler sendiri")
	})
}
