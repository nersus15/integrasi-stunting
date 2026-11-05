package kontrak_field_test

import (
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func auditRow(t *testing.T, table, id string) string {
	t.Helper()
	return scalar[string](t, `SELECT concat_ws('|', extract(year FROM "createdAt"), "deletedAt" IS NULL,
		coalesce(to_jsonb(x)->>'updated_by', '-'), coalesce(to_jsonb(x)->>'deleted_by', '-'))
		FROM stunting.`+table+` x WHERE id = $1`, id)
}

func TestFieldSistemDiabaikanSaatPOSTJakantro(t *testing.T) {
	setup(t)

	require.Equal(t, 201, request(t, "POST", "/orangtua", testutil.KeyJakantro, withSystemFields(orangtua())).Kode)
	require.Equal(t, 201, request(t, "POST", "/anak", testutil.KeyJakantro, withSystemFields(anak())).Kode)
	require.Equal(t, 201, request(t, "POST", "/kunjungan", testutil.KeyJakantro, withSystemFields(kunjungan())).Kode)
	require.Equal(t, 201, request(t, "POST", "/kesehatan", testutil.KeyJakantro, withSystemFields(kesehatan())).Kode)

	for table, id := range map[string]string{"orangtua": idOrtu, "anak": idAnak, "kunjungan": idKunjungan, "kesehatan": idKesehatan} {
		assert.Equal(t, "2026|t|-|-", auditRow(t, table, id), table)
	}
}

func TestFieldSistemDiabaikanSaatPUTJakantro(t *testing.T) {
	setup(t)
	setupKeluarga(t)
	require.Equal(t, 201, request(t, "POST", "/kunjungan", testutil.KeyJakantro, kunjungan()).Kode)
	require.Equal(t, 201, request(t, "POST", "/kesehatan", testutil.KeyJakantro, kesehatan()).Kode)

	for _, r := range []struct{ path, tabel, id, harap string }{
		{"/orangtua", "orangtua", idOrtu, "2026|t|jakantro|-"},
		{"/anak", "anak", idAnak, "2026|t|jakantro|-"},
		{"/kunjungan/" + idKunjungan, "kunjungan", idKunjungan, "2026|t|jakantro|-"},
		{"/kesehatan/" + idKesehatan, "kesehatan", idKesehatan, "2026|t|jakantro|-"},
	} {
		body := withSystemFields(map[string]any{"id": r.id})
		if r.tabel == "orangtua" {
			body = withSystemFields(ortuPUT(map[string]any{}))
		}
		switch r.tabel {
		case "kunjungan":
			body = withSystemFields(kunjungan())
		case "kesehatan":
			body = withSystemFields(kesehatan())
		}
		j := request(t, "PUT", r.path, testutil.KeyJakantro, body)
		require.Equal(t, 200, j.Kode, "%s: %v", r.path, j.Isi)
		assert.Equal(t, r.harap, auditRow(t, r.tabel, r.id), r.path)
	}
}

func TestKunjunganJakantroTidakMengisiFieldJalurFaskes(t *testing.T) {
	setup(t)
	setupKeluarga(t)

	k := kunjungan()
	k["id_faskes"], k["id_satusehat"], k["stunting"] = "FAS-PKM-KENANGA", "enc-dari-jakantro", 1
	k["ref_episode"], k["ref_rujukan"], k["id_episode"], k["id_rujukan"] = "ep-x", "sr-x", "ep-x", "sr-x"
	j := request(t, "POST", "/kunjungan", testutil.KeyJakantro, k)
	require.Equal(t, 201, j.Kode, "%v", j.Isi)

	jalurFaskes := `SELECT count(*) FROM stunting.kunjungan WHERE id = $1 AND (id_faskes IS NOT NULL OR satusehat_id IS NOT NULL
		OR stunting IS NOT NULL OR ref_episode IS NOT NULL OR ref_rujukan IS NOT NULL OR id_episode IS NOT NULL OR id_rujukan IS NOT NULL)`
	assert.Zero(t, count(t, jalurFaskes, idKunjungan))

	j = request(t, "PUT", "/kunjungan/"+idKunjungan, testutil.KeyJakantro, kunjungan())
	assert.Equal(t, 200, j.Kode, "kunjungan tetap milik jakantro: %v", j.Isi)

	j = request(t, "PUT", "/kunjungan/"+idKunjungan, testutil.KeyJakantro, k)
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Zero(t, count(t, jalurFaskes, idKunjungan), "PUT juga mengabaikan field jalur faskes, sama dengan POST")
}

func TestSatusehatIdDariJakantroDiabaikan(t *testing.T) {
	setup(t)

	o, a := orangtua(), anak()
	o["id_satusehat"], a["id_satusehat"] = "SS-ORTU-JKT", "SS-ANAK-JKT"
	require.Equal(t, 201, request(t, "POST", "/orangtua", testutil.KeyJakantro, o).Kode)
	require.Equal(t, 201, request(t, "POST", "/anak", testutil.KeyJakantro, a).Kode)
	assert.Equal(t, "-|-", scalar[string](t, `SELECT coalesce((SELECT satusehat_id FROM stunting.orangtua WHERE id = $1),'-') || '|' ||
		coalesce((SELECT satusehat_id FROM stunting.anak WHERE id = $2),'-')`, idOrtu, idAnak))

	_, err := db.Exec(`UPDATE stunting.orangtua SET satusehat_id = 'RP-DARI-SATUSEHAT' WHERE id = $1`, idOrtu)
	require.NoError(t, err)
	_, err = db.Exec(`UPDATE stunting.anak SET satusehat_id = 'P-DARI-SATUSEHAT' WHERE id = $1`, idAnak)
	require.NoError(t, err)

	require.Equal(t, 200, request(t, "PUT", "/orangtua", testutil.KeyJakantro, ortuPUT(map[string]any{"id_satusehat": "SS-ORTU-JKT"})).Kode)
	require.Equal(t, 200, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"id": idAnak, "id_satusehat": "SS-ANAK-JKT"}).Kode)
	assert.Equal(t, "RP-DARI-SATUSEHAT|P-DARI-SATUSEHAT", scalar[string](t, `SELECT (SELECT satusehat_id FROM stunting.orangtua WHERE id = $1) || '|' ||
		(SELECT satusehat_id FROM stunting.anak WHERE id = $2)`, idOrtu, idAnak))
}

func TestFlagBisaDiubahKeNolLewatPUT(t *testing.T) {
	setup(t)
	setupKeluarga(t)
	flag := func() string {
		return scalar[string](t, `SELECT (SELECT kia FROM stunting.orangtua WHERE id = $1) || '|' || (SELECT imd FROM stunting.anak WHERE id = $2)`, idOrtu, idAnak)
	}

	require.Equal(t, 200, request(t, "PUT", "/orangtua", testutil.KeyJakantro, ortuPUT(map[string]any{"alamat": "Jl Baru"})).Kode)
	require.Equal(t, 200, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"id": idAnak, "nama": "Nama Baru"}).Kode)
	assert.Equal(t, "1|1", flag(), "tidak dikirim berarti dipertahankan")

	require.Equal(t, 200, request(t, "PUT", "/orangtua", testutil.KeyJakantro, ortuPUT(map[string]any{"kia": 0})).Kode)
	require.Equal(t, 200, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"id": idAnak, "imd": 0}).Kode)
	assert.Equal(t, "0|0", flag())

	require.Equal(t, 200, request(t, "PUT", "/orangtua", testutil.KeyJakantro, ortuPUT(map[string]any{"alamat": "Jl Lain"})).Kode)
	require.Equal(t, 200, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"id": idAnak, "nama": "Nama Lain"}).Kode)
	assert.Equal(t, "0|0", flag(), "nol yang tersimpan juga dipertahankan")
}

func TestSourceDataTidakBerubahLewatPUT(t *testing.T) {
	setup(t)
	setupKeluarga(t)
	source := func() string {
		return scalar[string](t, `SELECT trim((SELECT source_data FROM stunting.orangtua WHERE id = $1)) || '|' ||
			trim((SELECT source_data FROM stunting.anak WHERE id = $2))`, idOrtu, idAnak)
	}
	initial := source()

	require.Equal(t, 200, request(t, "PUT", "/orangtua", testutil.KeyJakantro, ortuPUT(map[string]any{"source_data": "lain"})).Kode)
	require.Equal(t, 200, request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"id": idAnak, "source_data": "lain"}).Kode)
	assert.Equal(t, initial, source())
}

// PUT orangtua selalu membawa nik dan salah satu nama
func ortuPUT(content map[string]any) map[string]any {
	b := map[string]any{"id": idOrtu, "nik": "3175049800000101", "nama_ayah": "Ayah Kontrak"}
	for k, v := range content {
		b[k] = v
	}
	return b
}
