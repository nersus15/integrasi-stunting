package update_endpoint_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
)

const testGroup = "uji-update-endpoint"

const (
	keyRS           = "uji-cakupan-rs-200022-000000000000000000000000000000000000"
	keyWithoutGroup = "uji-cakupan-tanpa-grup-0000000000000000000000000000000000"
	kunciGrupLain   = "uji-cakupan-grup-lain-00000000000000000000000000000000000"
)

var (
	app    *testutil.App
	wiring *testutil.Wiring
	db     *sql.DB
)

func setupKeys() error {
	pengguna := []struct {
		nama, kunci, grup string
	}{
		{"uji-cakupan-rs", keyRS, "{orgid:200022}"},
		{"uji-cakupan-tanpa-grup", keyWithoutGroup, "{}"},
		{"uji-cakupan-grup-lain", kunciGrupLain, "{service}"},
	}
	for _, p := range pengguna {
		_, err := db.Exec(`WITH u AS (
			INSERT INTO access.auth_users (username, password_hash, groups, roles)
			VALUES ($1, 'x', $3::text[], '{read,write}')
			ON CONFLICT (username) DO UPDATE SET groups = EXCLUDED.groups
			RETURNING id)
			INSERT INTO access.auth_apikey (api_key, user_id) SELECT $2, id FROM u
			ON CONFLICT (api_key) DO NOTHING`, p.nama, p.kunci, p.grup)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteKeys() {
	_, _ = db.Exec(`DELETE FROM access.auth_users WHERE username LIKE 'uji-cakupan-%'`)
}

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		fmt.Fprintln(os.Stderr, "postgres mati, test update endpoint dilewati")
		os.Exit(0)
	}

	db = testutil.OpenDB()
	if err := testutil.Scenario1Seed(db); err != nil {
		fmt.Fprintf(os.Stderr, "gagal seed: %v\n", err)
		os.Exit(1)
	}
	if err := setupKeys(); err != nil {
		fmt.Fprintf(os.Stderr, "gagal menyiapkan api key uji: %v\n", err)
		os.Exit(1)
	}

	var err error
	app, err = testutil.StartApp(2097)
	if err == nil {
		wiring, err = app.Wiring(testGroup)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi: %v\n", err)
		deleteKeys()
		os.Exit(1)
	}

	code := m.Run()
	_ = testutil.Scenario1Cleanup(db, testGroup)
	deleteKeys()
	app.Stop()
	db.Close()
	os.Exit(code)
}
