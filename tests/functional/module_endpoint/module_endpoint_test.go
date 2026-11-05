package module_endpoint_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	stunting "github.com/nersus15/integrasi/mod-stunting"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var app *testutil.FunctionalApp

func TestMain(m *testing.M) {
	built, err := testutil.NewFunctionalApp()
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

	resp, err := app.Web.Test(req, -1)
	require.NoError(t, err)
	return resp
}

// --- POSITIVE CASE ---

func TestHealth_MelaporkanModulSehat(t *testing.T) {
	resp := get(t, "/stunting/health")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestInfo_MendaftarkanSeluruhRoute(t *testing.T) {
	resp := get(t, "/stunting/info")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestModule_Identitas(t *testing.T) {
	assert.Equal(t, "stunting", app.Module.Name())
	assert.Equal(t, stunting.ModuleVersion, app.Module.Version())
	assert.Empty(t, app.Module.Dependencies())
}

func TestModule_MengeksposServiceDanRepository(t *testing.T) {
	require.NotNil(t, app.Module.Services()["stunting"])
	require.NotNil(t, app.Module.Repositories()["stunting"])
}

// --- NEGATIVE CASE ---

func TestRouteTidakDikenal(t *testing.T) {
	resp := get(t, "/tidak-ada-route-ini")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHealth_MethodTidakDiizinkan(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://localhost/stunting/health", nil)
	require.NoError(t, err)

	resp, err := app.Web.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}
