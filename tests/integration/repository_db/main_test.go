package repository_db_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
)

const groupUji = "uji-repository-db"

var (
	wiring *testutil.Wiring
	db     *sql.DB
)

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		os.Exit(0)
	}
	db = testutil.OpenDB()
	if err := testutil.Scenario1Seed(db); err != nil {
		fmt.Fprintf(os.Stderr, "gagal seed: %v\n", err)
		os.Exit(1)
	}

	app, err := testutil.StartApp(2095)
	if err == nil {
		wiring, err = app.Wiring(groupUji)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	_ = testutil.Scenario1Cleanup(db, groupUji)
	app.Stop()
	db.Close()
	os.Exit(code)
}
