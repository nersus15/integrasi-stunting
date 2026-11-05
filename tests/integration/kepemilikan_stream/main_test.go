package kepemilikan_stream_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/require"
)

const (
	testGroup = "uji-kepemilikan-stream"
	ihsAnak   = "P-KEP-000001"
	nikAnak   = "3175049920000001"
	ihsIbu    = "RP-KEP-000001"
	nikIbu    = "3175049920000002"
	encUji    = "enc-kep-1"
	orgUji    = "100011"
)

var (
	wiring *testutil.Wiring
	app    *testutil.App
	db     *sql.DB
)

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		fmt.Fprintln(os.Stderr, "postgres tidak hidup, test kepemilikan stream dilewati")
		os.Exit(0)
	}

	var err error
	app, err = testutil.StartApp(2082)
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

func cleanup() {
	anak := `SELECT id FROM stunting.anak WHERE nik = $1 OR satusehat_id = $2`
	for _, table := range []string{"observasi", "diagnosa", "layanan", "rujukan", "kunjungan", "episode"} {
		_, _ = db.Exec(`DELETE FROM stunting.`+table+` WHERE id_anak IN (`+anak+`)`, nikAnak, ihsAnak)
	}
	_, _ = db.Exec(`DELETE FROM stunting.anak WHERE nik = $1 OR satusehat_id = $2`, nikAnak, ihsAnak)
	_, _ = db.Exec(`DELETE FROM stunting.orangtua WHERE nik = $1`, nikIbu)
	_, _ = db.Exec(`DELETE FROM kafka.transactions WHERE group_id = $1`, testGroup)
}

func setup(t *testing.T) {
	t.Helper()
	cleanup()
	t.Cleanup(cleanup)
}

type res struct {
	id string
	r  map[string]any
}

func send(t *testing.T, trx, method string, content ...res) {
	t.Helper()
	var entries, answers []any
	for _, x := range content {
		tipe := x.r["resourceType"].(string)
		url, full, status := tipe, "urn:uuid:"+x.id, "201 Created"
		if method == "PUT" {
			x.r["id"] = x.id
			url, full, status = tipe+"/"+x.id, tipe+"/"+x.id, "200 OK"
		}
		entries = append(entries, map[string]any{"fullUrl": full, "resource": x.r, "request": map[string]any{"method": method, "url": url}})
		answers = append(answers, map[string]any{"status": status, "resourceID": x.id, "resourceType": tipe})
	}
	b, err := json.Marshal(map[string]any{
		"method": method, "transactionId": trx, "patient": ihsAnak,
		"input":    map[string]any{"resourceType": "Bundle", "type": "transaction", "entry": entries},
		"response": answers,
	})
	require.NoError(t, err)
	wiring.Kafka.Run(context.Background(), b)

	var failed int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM kafka.transactions WHERE group_id = $1`, testGroup).Scan(&failed))
	require.Zero(t, failed, "pesan %s gagal diproses", trx)
}

func coding(system, code string) map[string]any {
	return map[string]any{"coding": []any{map[string]any{"system": system, "code": code, "display": code}}}
}

func ref(r string) map[string]any { return map[string]any{"reference": r} }

func pasien(name string) res {
	return res{ihsAnak, map[string]any{
		"resourceType": "Patient",
		"identifier":   []any{map[string]any{"system": "https://fhir.kemkes.go.id/id/nik", "value": nikAnak}},
		"name":         []any{map[string]any{"use": "official", "text": name}},
		"gender":       "male",
		"birthDate":    "2025-01-10",
	}}
}

func ibu(name string) res {
	return res{ihsIbu, map[string]any{
		"resourceType": "RelatedPerson",
		"patient":      ref("Patient/" + ihsAnak),
		"identifier":   []any{map[string]any{"system": "https://fhir.kemkes.go.id/id/nik", "value": nikIbu}},
		"relationship": []any{coding("http://terminology.hl7.org/CodeSystem/v3-RoleCode", "MTH")},
		"name":         []any{map[string]any{"text": name}},
		"telecom":      []any{map[string]any{"system": "phone", "value": "0811"}},
		"address":      []any{map[string]any{"text": "Jl. Uji"}},
	}}
}

func encounter() res {
	return res{encUji, map[string]any{
		"resourceType":    "Encounter",
		"status":          "finished",
		"class":           map[string]any{"system": "http://terminology.hl7.org/CodeSystem/v3-ActCode", "code": "AMB"},
		"subject":         ref("Patient/" + ihsAnak),
		"period":          map[string]any{"start": "2026-03-01T01:00:00+00:00"},
		"serviceProvider": ref("Organization/" + orgUji),
	}}
}

func tinggiStunting() res {
	return res{"obs-kep-tb", map[string]any{
		"resourceType":      "Observation",
		"status":            "final",
		"code":              coding("http://loinc.org", "8306-3"),
		"subject":           ref("Patient/" + ihsAnak),
		"encounter":         ref("Encounter/" + encUji),
		"effectiveDateTime": "2026-03-01T01:10:00+00:00",
		"valueQuantity":     map[string]any{"value": 65.0, "unit": "cm"},
		"interpretation":    []any{coding("http://terminology.kemkes.go.id/CodeSystem/clinical-term", "OI000011")},
	}}
}

func scalar(t *testing.T, q string, args ...any) string {
	t.Helper()
	var v sql.NullString
	require.NoError(t, db.QueryRow(q, args...).Scan(&v), q)
	return v.String
}
