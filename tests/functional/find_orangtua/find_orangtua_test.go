package find_orangtua_test

import (
	"fmt"
	"net/http"
	"net/url"
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

	built, err := testutil.NewFunctionalAppAuth(2087, testutil.KeyJakantro)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi: %v\n", err)
		os.Exit(1)
	}
	app = built

	code := m.Run()
	app.Close()
	os.Exit(code)
}

func get(t *testing.T, path string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "http://localhost"+path, nil)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

var muatanInjeksi = []string{
	"' OR '1'='1",
	"1' --",
	"' UNION SELECT * FROM jakantro.orangtua --",
	"'; DROP TABLE jakantro.orangtua; --",
	"1; DELETE FROM jakantro.anak",
	`" OR ""="`,
	`\' OR 1=1`,
}

// --- NEGATIVE CASE: SQL injection lewat HTTP ---

func TestFindOrangtua_InjeksiLewatPathId(t *testing.T) {
	for _, m := range muatanInjeksi {
		t.Run(m, func(t *testing.T) {
			resp := get(t, testutil.APIPrefix+"/orangtua/"+url.PathEscape(m))
			defer resp.Body.Close()

			assert.NotEqual(t, http.StatusOK, resp.StatusCode,
				"muatan injeksi tidak boleh menghasilkan data")
		})
	}
}

func TestFindOrangtua_InjeksiLewatQueryNik(t *testing.T) {
	for _, m := range muatanInjeksi {
		t.Run(m, func(t *testing.T) {
			resp := get(t, testutil.APIPrefix+"/orangtua/"+testutil.IdOrangtua+
				"?nik="+url.QueryEscape(m))
			defer resp.Body.Close()

			assert.NotEqual(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestFindOrangtua_TidakAdaKataKunci(t *testing.T) {
	resp := get(t, testutil.APIPrefix+"/orangtua")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode,
		"tanpa id maupun query apa pun harus ditolak 422")
}

func TestFindOrangtua_IdHanyaSpasiTetap404(t *testing.T) {
	// Spasi bukan kata kunci, tapi path-nya tetap cocok dengan route :id?
	// sehingga hasilnya "tidak ditemukan", bukan "permintaan salah".
	resp := get(t, testutil.APIPrefix+"/orangtua/%20")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestFindOrangtua_TidakDitemukanBukanErrorServer(t *testing.T) {
	// Pencarian yang sah tapi tidak menemukan baris bukan kegagalan server.
	// Status 5xx membuatnya tercampur dengan gangguan nyata di monitoring.
	resp := get(t, testutil.APIPrefix+"/orangtua/"+testutil.IdOrangtua)
	defer resp.Body.Close()

	assert.Less(t, resp.StatusCode, 500,
		"data tidak ditemukan seharusnya 404, bukan 500")
}

// --- Perilaku endpoint ---

func TestFindOrangtua_CariLewatNikTanpaId(t *testing.T) {
	resp := get(t, testutil.APIPrefix+"/orangtua?nik="+testutil.NikOrangtua)
	defer resp.Body.Close()

	assert.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode,
		"id opsional, pencarian lewat query nik harus terlayani")
	assert.NotEqual(t, http.StatusBadRequest, resp.StatusCode,
		"nik adalah kata kunci yang sah, tidak boleh dianggap tanpa kata kunci")
}
