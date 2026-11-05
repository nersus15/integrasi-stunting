package service_create_kunjungan_test

import (
	"errors"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tabelDisisipkan mengembalikan urutan tabel yang menerima INSERT.
func insertedTables(db *testutil.FakeDatabase) []string {
	table := make([]string, 0, db.InsertOneCallCount())
	for _, c := range db.InsertOneCalls {
		table = append(table, c.Table)
	}
	return table
}

// --- POSITIVE CASE ---

func TestCreateKunjungan_KunjunganSajaHanyaMenyimpanKunjungan(t *testing.T) {
	db := testutil.NewFakeDatabase()
	stack := testutil.NewStack(t, db)

	res, err := stack.CreateKunjungan(t, testutil.BodyKunjunganSaja1Key(t))

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, []string{entity.Kunjungan{}.TableName()}, insertedTables(db),
		"kunjungan saja tidak boleh menyentuh tabel orangtua maupun anak")
	require.NotNil(t, res.IdAnak)
	assert.Equal(t, testutil.IdAnak, *res.IdAnak)
	assert.Nil(t, res.Orangtua)
	assert.Nil(t, res.Anak)
}

func TestCreateKunjungan_BentukKunjunganSajaLain(t *testing.T) {
	tests := []struct {
		nama string
		body func(testing.TB) []byte
	}{
		{"3 key", testutil.BodyKunjunganSaja3Key},
		{"2 key", testutil.BodyKunjunganSaja2Key},
		{"1 key", testutil.BodyKunjunganSaja1Key},
		{"flat", testutil.BodyKunjunganFlat},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			db := testutil.NewFakeDatabase()
			stack := testutil.NewStack(t, db)

			res, err := stack.CreateKunjungan(t, tc.body(t))

			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, 1, db.InsertOneCallCount(), "hanya kunjungan yang disimpan")
		})
	}
}

// --- NEGATIVE CASE: payload ---

func TestCreateKunjungan_BodyTidakValid(t *testing.T) {
	tests := []struct {
		nama string
		body string
	}{
		{"json rusak", `{"kunjungan":`},
		{"bukan objek", `"cuma string"`},
		{"body kosong", ``},
		{"objek kosong", `{}`},
		{"tanpa kunjungan", `{"anak":{"id":"A1","id_orangtua":"O1","nama":"Ahmad"}}`},
		{"flat dicampur nested", `{"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16","kunjungan":{"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16"}}`},
		{"anak tanpa id_orangtua", `{"anak":{"id":"A1","nama":"Ahmad"},"kunjungan":{"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16"}}`},
		{"anak.id beda dengan kunjungan.id_anak", `{"anak":{"id":"A9","id_orangtua":"O1"},"kunjungan":{"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16"}}`},
		{"orangtua tanpa anak", `{"orangtua":{"id":"O1","nik":"3515081203800002"},"kunjungan":{"id":"K1","id_anak":"A1","tanggal_pengukuran":"2025-02-16"}}`},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			db := testutil.NewFakeDatabase()
			stack := testutil.NewStack(t, db)

			res, err := stack.CreateKunjungan(t, []byte(tc.body))

			require.Error(t, err)
			assert.Nil(t, res)
			assert.Equal(t, 0, db.InsertOneCallCount(),
				"payload tidak valid tidak boleh menyentuh database")
		})
	}
}

func TestCreateKunjungan_DataAnakTidakLolosValidasi(t *testing.T) {
	db := testutil.NewFakeDatabase()
	stack := testutil.NewStack(t, db)

	anak := testutil.AnakPayloadValid()
	anak.TanggalLahir = "" // wajib menurut ValidateAnak

	body := testutil.BodyNested(t, testutil.BodyOpts{
		Anak:      anak,
		Kunjungan: testutil.KunjunganPayloadValid(),
	})

	res, err := stack.CreateKunjungan(t, body)

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, 0, db.InsertOneCallCount(), "anak yang cacat tidak boleh disimpan")
}

// --- NEGATIVE CASE: kompensasi ---

func TestCreateKunjungan_GagalTransaksiTidakMenulisSebagian(t *testing.T) {
	db := testutil.NewFakeDatabase()
	stack := testutil.NewStack(t, db)

	_, err := stack.CreateKunjungan(t, testutil.BodyRegistrasiLengkap(t))
	require.Error(t, err, "tanpa dukungan transaksi, registrasi lengkap harus gagal")

	stack.DrainBackground(t)

	assert.Zero(t, db.InsertOneCallCount(),
		"tidak boleh ada baris yang tersimpan ketika transaksi tidak bisa dijalankan")
	assert.Empty(t, db.DeletedTables(),
		"tanpa tulisan, tidak ada yang perlu dihapus")
}

func TestCreateKunjungan_KunjunganSajaGagalTidakMenghapusApaPun(t *testing.T) {
	db := testutil.NewFailingFakeDatabase(errors.New("database bermasalah"))
	stack := testutil.NewStack(t, db)

	_, err := stack.CreateKunjungan(t, testutil.BodyKunjunganSaja1Key(t))
	require.Error(t, err)

	stack.DrainBackground(t)
	assert.Empty(t, db.DeletedTables(),
		"tidak ada entitas yang dibuat request ini, jadi tidak ada yang boleh dihapus")
}

func TestCreateKunjungan_ErrorDatabaseDiteruskan(t *testing.T) {
	dbErr := errors.New("koneksi database terputus")
	db := testutil.NewFailingFakeDatabase(dbErr)
	stack := testutil.NewStack(t, db)

	res, err := stack.CreateKunjungan(t, testutil.BodyKunjunganSaja1Key(t))

	require.Error(t, err)
	assert.ErrorIs(t, err, dbErr)
	assert.Nil(t, res)
}

func TestCreateKunjungan_TidakPanic(t *testing.T) {
	bentuk := map[string]func(testing.TB) []byte{
		"registrasi lengkap": testutil.BodyRegistrasiLengkap,
		"anak baru 3 key":    testutil.BodyAnakBaru3Key,
		"anak baru 2 key":    testutil.BodyAnakBaru2Key,
		"kunjungan saja":     testutil.BodyKunjunganSaja1Key,
		"kunjungan flat":     testutil.BodyKunjunganFlat,
	}

	for name, body := range bentuk {
		t.Run(name, func(t *testing.T) {
			db := testutil.NewFakeDatabase()
			stack := testutil.NewStack(t, db)
			raw := body(t)

			value := testutil.CatchPanic(func() {
				_, _ = stack.Service.CreateKunjungan(raw)
			})

			assert.Nil(t, value, "payload sah tidak boleh membuat service panic")
		})
	}
}
