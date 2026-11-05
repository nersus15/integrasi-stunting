// Paket optimasi_query mengunci jumlah statement SQL di jalur yang dioptimasi
// (summary, registrasi, sapuan turunan, cek akses faskes) beserta perilakunya.
//
// Menjalankan paket ini saja:
//
//	cd tests && ./uji.sh optimasi_query
//	cd tests && ./uji.sh optimasi_query -run Summary
//
// Golden summary ditulis ulang dengan TULIS_GOLDEN=1 (hanya bila perubahan
// bentuk respons memang disengaja).
package optimasi_query_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/repository"
	"github.com/nersus15/integrasi/tests/testutil"
)

const groupUji = "uji-optimasi-query"

var (
	app     *testutil.App
	wiring  *testutil.Wiring
	counter *testutil.QueryCounter
	db      *sql.DB
)

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		fmt.Fprintln(os.Stderr, "postgres mati, test optimasi query dilewati")
		os.Exit(0)
	}

	// jeda hapus ulang dipersingkat supaya test kebocoran tidak menunggu 12 detik
	repository.JedaLupakanUlang = 500 * time.Millisecond

	db = testutil.OpenDB()
	if err := testutil.Scenario1Seed(db); err != nil {
		fmt.Fprintf(os.Stderr, "gagal seed: %v\n", err)
		os.Exit(1)
	}

	var err error
	app, err = testutil.StartApp(2086)
	if err == nil {
		wiring, err = app.Wiring(groupUji)
	}
	if err == nil {
		counter, err = app.InstallQueryCounter()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	_ = testutil.Scenario1Cleanup(db, groupUji)
	cleanupExtras()
	app.Stop()
	db.Close()
	os.Exit(code)
}
