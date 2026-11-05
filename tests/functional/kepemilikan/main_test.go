package kepemilikan_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
)

var (
	app *testutil.App
	db  *sql.DB
)

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		fmt.Fprintln(os.Stderr, "postgres mati, test kepemilikan dilewati")
		os.Exit(0)
	}

	db = testutil.OpenDB()
	if err := testutil.Scenario1Seed(db); err != nil {
		fmt.Fprintf(os.Stderr, "gagal seed: %v\n", err)
		os.Exit(1)
	}

	var err error
	app, err = testutil.StartApp(2083)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	cleanup()
	app.Stop()
	db.Close()
	os.Exit(code)
}
