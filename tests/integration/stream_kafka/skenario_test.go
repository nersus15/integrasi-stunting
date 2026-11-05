package stream_kafka_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkenarioPosyanduPuskesmasRSRujukBalik(t *testing.T) {
	setupJakantro(t)

	send(t,
		"02-kafka-patient.json",
		"03-kafka-relatedperson.json",
		"04-kafka-puskesmas.json",
		"05-kafka-rs.json",
		"06-kafka-rujuk-balik.json",
		"07-kafka-puskesmas-lanjutan.json",
	)

	assert.Equal(t, 0, count(t, `SELECT count(*) FROM kafka.transactions WHERE group_id = $1`, testGroup),
		"tidak ada pesan yang boleh gagal")

	assert.Equal(t, ihsAnak, scalar[string](t, `SELECT satusehat_id FROM stunting.anak WHERE nik = $1`, nikAnak))
	assert.Equal(t, idAnakJkt, scalar[string](t, `SELECT id FROM stunting.anak WHERE nik = $1`, nikAnak),
		"anak jakantro dilengkapi, bukan dibuat ulang")
	assert.NotEmpty(t, scalar[string](t, `SELECT coalesce(satusehat_id,'') FROM stunting.orangtua WHERE nik = $1`, nikIbu))

	row, err := db.Query(`
		SELECT to_char(k.tanggal_pengukuran,'YYYY-MM-DD'), coalesce(f.satusehat_id,'posyandu'),
		       k.berat_badan::text, k.tinggi_badan::text, coalesce(k.stunting::text,'NULL'), coalesce(ep.satusehat_id,'')
		FROM stunting.kunjungan k
		JOIN stunting.anak a ON a.id = k.id_anak
		LEFT JOIN stunting.faskes f ON f.id = k.id_faskes
		LEFT JOIN stunting.episode ep ON ep.id = k.id_episode
		WHERE a.nik = $1 ORDER BY k.tanggal_pengukuran`, nikAnak)
	require.NoError(t, err)
	defer row.Close()

	var result [][6]string
	for row.Next() {
		var r [6]string
		require.NoError(t, row.Scan(&r[0], &r[1], &r[2], &r[3], &r[4], &r[5]))
		result = append(result, r)
	}

	assert.Equal(t, [][6]string{
		{"2026-01-15", "posyandu", "7.2", "72", "NULL", ""},
		{"2026-02-05", "100011", "7.4", "72.5", "1", "ep-stunting-sim1"},
		{"2026-02-20", "200022", "7.5", "72.8", "1", "ep-stunting-sim1"},
		{"2026-03-15", "100011", "8.1", "74", "0", "ep-stunting-sim1"},
	}, result)

	assert.Equal(t, 1, count(t, `SELECT count(*) FROM stunting.episode WHERE id_anak = $1`, idAnakJkt))

	rujuk, err := db.Query(`
		SELECT r.jenis, fa.satusehat_id, ft.satusehat_id FROM stunting.rujukan r
		LEFT JOIN stunting.faskes fa ON fa.id = r.id_faskes_asal
		LEFT JOIN stunting.faskes ft ON ft.id = r.id_faskes_tujuan
		WHERE r.id_anak = $1 ORDER BY r.tanggal`, idAnakJkt)
	require.NoError(t, err)
	defer rujuk.Close()
	var alur [][3]string
	for rujuk.Next() {
		var r [3]string
		require.NoError(t, rujuk.Scan(&r[0], &r[1], &r[2]))
		alur = append(alur, r)
	}
	assert.Equal(t, [][3]string{{"rujukan", "100011", "200022"}, {"rujuk_balik", "200022", "100011"}}, alur)

	for _, table := range []string{"observasi", "diagnosa", "layanan", "rujukan"} {
		assert.Zero(t, count(t, `SELECT count(*) FROM stunting.`+table+` WHERE id_anak = $1 AND id_kunjungan IS NULL`, idAnakJkt),
			"%s menggantung", table)
	}
}

func TestPesanGandaTidakMenggandakanData(t *testing.T) {
	setupJakantro(t)
	send(t, "02-kafka-patient.json", "04-kafka-puskesmas.json")

	counter := func() [4]int {
		return [4]int{
			count(t, `SELECT count(*) FROM stunting.kunjungan WHERE id_anak = $1`, idAnakJkt),
			count(t, `SELECT count(*) FROM stunting.observasi WHERE id_anak = $1`, idAnakJkt),
			count(t, `SELECT count(*) FROM stunting.diagnosa WHERE id_anak = $1`, idAnakJkt),
			count(t, `SELECT count(*) FROM stunting.rujukan WHERE id_anak = $1`, idAnakJkt),
		}
	}
	initial := counter()
	send(t, "04-kafka-puskesmas.json")

	assert.Equal(t, initial, counter())
	assert.Zero(t, count(t, `SELECT count(*) FROM kafka.transactions WHERE group_id = $1`, testGroup))
}
