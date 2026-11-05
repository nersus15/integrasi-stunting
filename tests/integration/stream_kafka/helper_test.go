package stream_kafka_test

import (
	"encoding/json"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
)

const (
	ihsAnak   = testutil.Scenario1IhsAnak
	nikAnak   = testutil.Scenario1NikAnak
	nikIbu    = testutil.Scenario1NikIbu
	idAnakJkt = testutil.Scenario1IdAnak
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	content, err := testutil.Scenario1File(name)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func cleanup() {
	if err := testutil.Scenario1Cleanup(db, testGroup); err != nil {
		panic(err)
	}
}

func setupJakantro(t *testing.T) {
	t.Helper()
	cleanup()
	t.Cleanup(cleanup)
	if err := wiring.Scenario1Jakantro(); err != nil {
		t.Fatalf("kunjungan jakantro gagal: %v", err)
	}
}

func send(t *testing.T, file ...string) {
	t.Helper()
	if err := wiring.Scenario1Send(file...); err != nil {
		t.Fatal(err)
	}
}

func scalar[T any](t *testing.T, q string, args ...any) T {
	t.Helper()
	var v T
	if err := db.QueryRow(q, args...).Scan(&v); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return v
}

func count(t *testing.T, q string, args ...any) int {
	t.Helper()
	return scalar[int](t, q, args...)
}

type envelope map[string]any

func load(t *testing.T, file string) envelope {
	t.Helper()
	var a envelope
	if err := json.Unmarshal(fixture(t, file), &a); err != nil {
		t.Fatal(err)
	}
	return a
}

func (a envelope) entry(i int) map[string]any {
	return a["input"].(map[string]any)["entry"].([]any)[i].(map[string]any)
}

func (a envelope) bytes(t *testing.T) []byte {
	t.Helper()
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func putEnvelope(t *testing.T, trx string, resource map[string]any, id string) []byte {
	t.Helper()
	tipe := resource["resourceType"].(string)
	resource["id"] = id
	a := envelope{
		"method":        "PUT",
		"transactionId": trx,
		"patient":       ihsAnak,
		"input": map[string]any{
			"resourceType": "Bundle",
			"type":         "transaction",
			"entry": []any{map[string]any{
				"fullUrl":  tipe + "/" + id,
				"resource": resource,
				"request":  map[string]any{"method": "PUT", "url": tipe + "/" + id},
			}},
		},
		"response": []any{map[string]any{"status": "200 OK", "resourceID": id, "resourceType": tipe}},
	}
	return a.bytes(t)
}

func copyResource(t *testing.T, file string, i int) map[string]any {
	t.Helper()
	r := load(t, file).entry(i)["resource"].(map[string]any)
	return r
}
