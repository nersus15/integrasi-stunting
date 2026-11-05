package api_stunting_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
)

func TestMain(m *testing.M) {
	if base, health, exists := testutil.ExternalService(); exists {
		baseURL, healthURL = base, health
		os.Exit(m.Run())
	}
	if !databaseAvailable() {
		os.Exit(m.Run())
	}

	app, err := testutil.StartApp(2098)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi in-process: %v\n", err)
		os.Exit(1)
	}

	baseURL = app.BaseURL
	healthURL = app.Health

	code := m.Run()
	app.Stop()
	os.Exit(code)
}
