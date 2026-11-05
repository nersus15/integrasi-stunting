package stream_kafka_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"

	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func failedRecord(t *testing.T, key string) (attempt int, pesanError string, patient string, resolved bool) {
	t.Helper()
	err := db.QueryRow(`SELECT attempt, coalesce(error::text,''), coalesce(patient_id,''), resolved_at IS NOT NULL
		FROM kafka.transactions WHERE transaction_id = $1 AND group_id = $2`, key, testGroup).
		Scan(&attempt, &pesanError, &patient, &resolved)
	require.NoError(t, err, "catatan %s tidak ada", key)
	return
}

func brokenEnvelope(t *testing.T) []byte {
	a := load(t, "04-kafka-puskesmas.json")
	a["transactionId"] = "uji-overflow"
	a.entry(2)["resource"].(map[string]any)["valueQuantity"].(map[string]any)["value"] = 1e12
	return a.bytes(t)
}

func TestPesanBukanJSONDicatatDenganSidik(t *testing.T) {
	cleanup()
	t.Cleanup(cleanup)

	wiring.Kafka.Run(context.Background(), []byte(`{bukan json`))

	var key string
	require.NoError(t, db.QueryRow(`SELECT transaction_id FROM kafka.transactions WHERE group_id = $1`, testGroup).Scan(&key))
	assert.True(t, strings.HasPrefix(key, "sha256-"), key)

	_, e, patient, _ := failedRecord(t, key)
	assert.Contains(t, e, "Json Invalid")
	assert.Empty(t, patient)
}

func TestAmplopTidakLengkapDicatat(t *testing.T) {
	cases := map[string]func(a envelope){
		"tanpa-input":    func(a envelope) { delete(a, "input") },
		"tanpa-response": func(a envelope) { delete(a, "response") },
		"entry-beda":     func(a envelope) { a["response"] = a["response"].([]any)[:1] },
		"tanpa-trx":      func(a envelope) { delete(a, "transactionId"); delete(a, "input") },
	}
	message := map[string]string{
		"tanpa-input":    "request body kosong",
		"tanpa-response": "response body kosong",
		"entry-beda":     "tidak sama",
		"tanpa-trx":      "request body kosong",
	}

	for name, modify := range cases {
		t.Run(name, func(t *testing.T) {
			cleanup()
			t.Cleanup(cleanup)

			a := load(t, "04-kafka-puskesmas.json")
			a["transactionId"] = "uji-" + name
			modify(a)
			wiring.Kafka.Run(context.Background(), a.bytes(t))

			var key, e, patient string
			require.NoError(t, db.QueryRow(`SELECT transaction_id, error::text, coalesce(patient_id,'')
				FROM kafka.transactions WHERE group_id = $1`, testGroup).Scan(&key, &e, &patient))
			assert.Contains(t, e, message[name])
			assert.Equal(t, ihsAnak, patient)
			if name == "tanpa-trx" {
				assert.True(t, strings.HasPrefix(key, "sha256-"), key)
			} else {
				assert.Equal(t, "uji-"+name, key)
			}
		})
	}
}

func TestGagalSimpanDicatatDanAttemptBertambah(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json")

	wiring.Kafka.Run(context.Background(), brokenEnvelope(t))
	attempt, e, patient, resolved := failedRecord(t, "uji-overflow")
	assert.Equal(t, 1, attempt)
	assert.NotEmpty(t, e)
	assert.Equal(t, ihsAnak, patient)
	assert.False(t, resolved)
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`),
		"transaksi data medis harus batal utuh")

	wiring.Kafka.Run(context.Background(), brokenEnvelope(t))
	attempt, _, _, _ = failedRecord(t, "uji-overflow")
	assert.Equal(t, 2, attempt)
}

func TestRetryBerhasilMenandaiSelesai(t *testing.T) {
	setupJakantro(t)

	require.NoError(t, wiring.Stream.SaveKafkaTransaction("sim1-patient", fixture(t, "02-kafka-patient.json"), errors.New("bug lama")))
	require.NoError(t, wiring.Stream.RetryKafkaTransaction("sim1-patient"))

	_, e, _, resolved := failedRecord(t, "sim1-patient")
	assert.True(t, resolved)
	assert.Empty(t, e)
	assert.Equal(t, ihsAnak, scalar[string](t, `SELECT satusehat_id FROM stunting.anak WHERE nik = $1`, nikAnak))

	err := wiring.Stream.RetryKafkaTransaction("sim1-patient")
	assert.Equal(t, 404, exceptions.Classify(err).HttpCode, "yang sudah selesai tidak diproses lagi")
}

func TestRetryGagalMenaikkanAttempt(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json")
	wiring.Kafka.Run(context.Background(), brokenEnvelope(t))

	assert.Error(t, wiring.Stream.RetryKafkaTransaction("uji-overflow"))
	attempt, _, _, resolved := failedRecord(t, "uji-overflow")
	assert.Equal(t, 2, attempt)
	assert.False(t, resolved)
}

func TestRetryPesanKosongAtauRusak(t *testing.T) {
	cleanup()
	t.Cleanup(cleanup)

	require.NoError(t, wiring.Stream.SaveKafkaTransaction("uji-kosong", nil, nil))
	err := wiring.Stream.RetryKafkaTransaction("uji-kosong")
	assert.Equal(t, exceptions.TidakDitemukan.ErrorCode, exceptions.Classify(err).ErrorCode)
	_, e, _, _ := failedRecord(t, "uji-kosong")
	assert.NotEmpty(t, e)

	require.NoError(t, wiring.Stream.SaveKafkaTransaction("uji-rusak", []byte(`{rusak`), errors.New("x")))
	err = wiring.Stream.RetryKafkaTransaction("uji-rusak")
	assert.Equal(t, exceptions.BodyRusak.ErrorCode, exceptions.Classify(err).ErrorCode)

	a := load(t, "04-kafka-puskesmas.json")
	delete(a, "input")
	require.NoError(t, wiring.Stream.SaveKafkaTransaction("uji-tanpa-input", a.bytes(t), errors.New("x")))
	err = wiring.Stream.RetryKafkaTransaction("uji-tanpa-input")
	assert.Contains(t, err.Error(), "request body kosong")
}

func TestRetrySemuaDanDaftarPerPatient(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json")

	require.NoError(t, wiring.Stream.SaveKafkaTransaction("sim1-ortu", fixture(t, "03-kafka-relatedperson.json"), errors.New("bug")))
	wiring.Kafka.Run(context.Background(), brokenEnvelope(t))
	require.NoError(t, wiring.Stream.SaveKafkaTransaction("uji-lain", []byte(`{"patient":"P-LAIN"}`), errors.New("bug")))

	all, err := wiring.Stream.DaftarKafkaTransactionGagal(nil)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	patient := ihsAnak
	punyaAnak, err := wiring.Stream.DaftarKafkaTransactionGagal(&patient)
	require.NoError(t, err)
	assert.Len(t, punyaAnak, 2)
	for _, tr := range punyaAnak {
		assert.Equal(t, ihsAnak, *tr.PatientId)
		assert.NotEmpty(t, tr.Created)
		assert.Nil(t, tr.Resolved)
	}

	berhasil, failed, err := wiring.Stream.RetryAllKafkaTransactions()
	require.NoError(t, err)
	assert.Equal(t, 1, berhasil)
	assert.Equal(t, 2, failed)

	remaining, _ := wiring.Stream.DaftarKafkaTransactionGagal(nil)
	assert.Len(t, remaining, 2)
	assert.NotEmpty(t, scalar[string](t, `SELECT coalesce(satusehat_id,'') FROM stunting.orangtua WHERE nik = $1`, nikIbu))
}

func TestEndpointKafkaGagalDaftarDanRetry(t *testing.T) {
	key := testutil.KeyJakantro

	module, err := app.Wiring("")
	require.NoError(t, err)
	group := module.Config.Kafka.GroupID
	remove := func() {
		_, _ = db.Exec(`DELETE FROM kafka.transactions WHERE group_id = $1 AND transaction_id LIKE 'uji-http-%'`, group)
		cleanup()
	}
	remove()
	t.Cleanup(remove)

	var initial int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM kafka.transactions WHERE group_id = $1 AND resolved_at IS NULL`, group).Scan(&initial))
	if initial > 0 {
		t.Skipf("antrean %s berisi %d pesan milik orang lain; retry semua akan ikut memprosesnya", group, initial)
	}

	call := func(method, path string) (int, map[string]any) {
		req, _ := http.NewRequest(method, app.BaseURL+path, nil)
		req.Header.Set("X-API-Key", key)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		content, _ := io.ReadAll(res.Body)
		var m map[string]any
		_ = json.Unmarshal(content, &m)
		return res.StatusCode, m
	}

	setupJakantro(t)
	require.NoError(t, module.Stream.SaveKafkaTransaction("uji-http-patient", fixture(t, "02-kafka-patient.json"), errors.New("bug")))
	require.NoError(t, module.Stream.SaveKafkaTransaction("uji-http-rusak", []byte(`{"patient":"P-HTTP"}`), errors.New("bug")))

	status, content := call("GET", "/kafka/gagal?patient_id="+ihsAnak)
	assert.Equal(t, 200, status)
	assert.EqualValues(t, 1, content["jumlah"])

	status, content = call("GET", "/kafka/gagal")
	assert.Equal(t, 200, status)
	assert.EqualValues(t, 2, content["jumlah"])

	status, content = call("POST", "/kafka/gagal/uji-http-patient/retry")
	assert.Equal(t, 200, status)
	assert.Equal(t, "beres", content["status"])

	status, _ = call("POST", "/kafka/gagal/uji-http-patient/retry")
	assert.Equal(t, 404, status)

	status, content = call("POST", "/kafka/gagal/retry")
	assert.Equal(t, 200, status)
	assert.EqualValues(t, 0, content["berhasil"])
	assert.EqualValues(t, 1, content["gagal"])

}
