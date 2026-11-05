package stream_kafka_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
)

const testGroup = "uji-stream-kafka"

var (
	wiring *testutil.Wiring
	app    *testutil.App
	db     *sql.DB
)

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		fmt.Fprintln(os.Stderr, "postgres tidak hidup, test stream kafka dilewati")
		os.Exit(0)
	}

	var err error
	app, err = testutil.StartApp(2099)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi: %v\n", err)
		os.Exit(1)
	}

	wiring, err = app.Wiring(testGroup)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal merangkai service: %v\n", err)
		os.Exit(1)
	}

	db = testutil.OpenDB()
	if err := testutil.Scenario1Seed(db); err != nil {
		fmt.Fprintf(os.Stderr, "gagal seed faskes: %v\n", err)
		os.Exit(1)
	}

	exitCode := m.Run()
	cleanup()
	db.Close()
	app.Stop()
	os.Exit(exitCode)
}
