package klasifikasi_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var db *sql.DB

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		os.Exit(0)
	}
	testutil.SetupLogger()
	db = testutil.OpenDB()
	code := m.Run()
	db.Close()
	os.Exit(code)
}

const (
	ortu = `INSERT INTO stunting.orangtua (id, no_kk, nik, nama_ayah, nama_ibu, telepon, rt, rw, alamat, kia)
		VALUES ('%s', '%s', '%s', 'A', 'I', '08', '001', '001', 'Jl', 1)`
	anak = `INSERT INTO stunting.anak (id, id_orangtua, nama, nik, tanggal_lahir, jenis_kelamin, anak_ke, imd, bb_lahir, tb_lahir, lk_lahir, satusehat_id)
		VALUES ('%s', 'klas-o1', 'N', '%s', '2025-01-01', 'L', 1, 1, 3, 50, 33, '%s')`
)

func dbError(t *testing.T, persiapan []string, failed string) error {
	t.Helper()
	tx, err := db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()
	for _, q := range persiapan {
		_, err := tx.Exec(q)
		require.NoError(t, err, q)
	}
	_, err = tx.Exec(failed)
	require.Error(t, err, failed)
	return err
}

func TestClassifyGalatPostgres(t *testing.T) {
	dasar := []string{
		fmt.Sprintf(ortu, "klas-o1", "3100000000000001", "3100000000000002"),
		fmt.Sprintf(anak, "klas-a1", "3100000000000003", "klas-ihs-1"),
	}

	cases := []struct {
		nama  string
		sql   string
		kind  exceptions.Kind
		pesan string
	}{
		{"pk orangtua", fmt.Sprintf(ortu, "klas-o1", "3100000000000011", "3100000000000012"), exceptions.DuplikatId, "id orangtua"},
		{"nik orangtua", fmt.Sprintf(ortu, "klas-o2", "3100000000000021", "3100000000000002"), exceptions.DuplikatNikOrangtua, ""},
		{"no kk", fmt.Sprintf(ortu, "klas-o3", "3100000000000001", "3100000000000032"), exceptions.DuplikatNoKK, ""},
		{"pk anak", fmt.Sprintf(anak, "klas-a1", "3100000000000043", "klas-ihs-4"), exceptions.DuplikatId, "id anak"},
		{"nik anak", fmt.Sprintf(anak, "klas-a5", "3100000000000003", "klas-ihs-5"), exceptions.DuplikatNikAnak, ""},
		{"unik lain", fmt.Sprintf(anak, "klas-a6", "3100000000000063", "klas-ihs-1"), exceptions.Duplikat, ""},
		{"pk kunjungan", `INSERT INTO stunting.kunjungan (id, id_anak, tanggal_pengukuran) VALUES ('klas-k1', 'klas-a1', now()), ('klas-k1', 'klas-a1', now())`, exceptions.DuplikatId, "id kunjungan"},
		{"pk kesehatan", `INSERT INTO stunting.kesehatan (id, id_anak, tanggal_pemantauan) VALUES ('klas-s1', 'klas-a1', now()), ('klas-s1', 'klas-a1', now())`, exceptions.DuplikatId, "id kesehatan"},
		{"foreign key", fmt.Sprintf(anak, "klas-a7", "3100000000000073", "klas-ihs-7") + "; ", exceptions.ReferensiHilang, ""},
		{"not null", `INSERT INTO stunting.orangtua (id, nik, nama_ayah, nama_ibu, rt, rw, alamat, kia) VALUES ('klas-o9', '3100000000000092', 'A', 'I', '001', '001', 'Jl', 1)`, exceptions.KolomWajib, "telepon"},
		{"check", `INSERT INTO stunting.faskes (id, nama, jenis, status) VALUES ('klas-f1', 'F', 'PUSTU', 7)`, exceptions.Constraint, ""},
		{"angka melampaui kolom", `INSERT INTO stunting.observasi (id, id_anak, system, kode, nilai_angka) VALUES ('klas-ob1', 'klas-a1', 's', 'k', 1e12)`, exceptions.Constraint, ""},
		{"teks terlalu panjang", `INSERT INTO stunting.faskes (id, nama, jenis, wilayah) VALUES ('klas-f2', 'F', 'PUSTU', '31.74.02.10010000')`, exceptions.Constraint, ""},
		{"sintaks", `SELEKT 1`, exceptions.Internal, ""},
	}

	for _, k := range cases {
		t.Run(k.nama, func(t *testing.T) {
			persiapan := dasar
			if k.nama == "foreign key" {
				persiapan = nil
			}
			result := exceptions.Classify(dbError(t, persiapan, k.sql))
			assert.Equal(t, k.kind.ErrorCode, result.ErrorCode, result.Message)
			assert.Equal(t, k.kind.HttpCode, result.HttpCode)
			if k.pesan != "" {
				assert.Contains(t, result.Message, k.pesan)
			}
		})
	}
}

func TestClassifyStatementTimeout(t *testing.T) {
	err := dbError(t, []string{`SET LOCAL statement_timeout = 20`}, `SELECT pg_sleep(1)`)
	assert.Equal(t, exceptions.TidakTersedia.ErrorCode, exceptions.Classify(err).ErrorCode)
}

func TestClassifyTanpaDatabase(t *testing.T) {
	assert.Nil(t, exceptions.Classify(nil))

	siap := exceptions.Validasi.Messagef("x %d", 1)
	assert.Same(t, siap, exceptions.Classify(fmt.Errorf("bungkus: %w", siap)))

	cases := map[error]exceptions.Kind{
		context.DeadlineExceeded:     exceptions.TidakTersedia,
		context.Canceled:             exceptions.TidakTersedia,
		exceptions.ErrTidakDitemukan: exceptions.TidakDitemukan,
		sql.ErrNoRows:                exceptions.TidakDitemukan,
		errors.New("tak dikenal"):    exceptions.Internal,
	}
	for err, kind := range cases {
		assert.Equal(t, kind.ErrorCode, exceptions.Classify(err).ErrorCode, err.Error())
	}
}

func TestKindDanError(t *testing.T) {
	cause := errors.New("sebab")

	e := exceptions.Validasi.WithMessage("", cause)
	assert.Equal(t, exceptions.Validasi.Message, e.Error())
	assert.Same(t, cause, e.Cause())
	assert.ErrorIs(t, e, cause)

	e = exceptions.Validasi.WithMessage("khusus", nil)
	assert.Equal(t, "khusus", e.Error())
	assert.Nil(t, e.Unwrap())

	assert.Equal(t, exceptions.Internal.Message, exceptions.Internal.New(nil).Message)
}

func TestAbaikanJikaBelumTersimpan(t *testing.T) {
	assert.NoError(t, exceptions.AbaikanJikaBelumTersimpan(nil))
	assert.NoError(t, exceptions.AbaikanJikaBelumTersimpan(exceptions.TidakDitemukan.New(nil)))

	other := exceptions.Internal.New(nil)
	assert.Same(t, other, exceptions.AbaikanJikaBelumTersimpan(other))
	polos := errors.New("polos")
	assert.Same(t, polos, exceptions.AbaikanJikaBelumTersimpan(polos))
}
