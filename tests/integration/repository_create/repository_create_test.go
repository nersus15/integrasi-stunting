package repository_create_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webcore-go/webcore/port"
)

func orangtuaEntity() *entity.Orangtua {
	p := testutil.OrangtuaPayloadValid()
	e := p.ToEntity()
	e.ID = testutil.IdOrangtua
	return e
}

func anakEntity() *entity.Anak {
	p := testutil.AnakPayloadValid()
	e := p.ToEntity()
	e.ID = testutil.IdAnak
	return e
}

func kunjunganEntity() *entity.Kunjungan {
	p := testutil.KunjunganPayloadValid()
	e := p.ToEntity()
	e.ID = testutil.IdKunjungan
	return e
}

// --- POSITIVE CASE ---

func TestCreateOrangTua_MengirimEntityKeTabelYangBenar(t *testing.T) {
	db := testutil.NewFakeDatabase()
	stack := testutil.NewStack(t, db)
	e := orangtuaEntity()

	var res any
	var err error
	value := testutil.CatchPanic(func() { res, err = stack.Repository.CreateOrangTua(e, nil) })

	require.Nil(t, value, "repository tidak boleh panic")
	require.NoError(t, err)
	require.NotNil(t, res)

	call, ok := db.LastInsertOne()
	require.True(t, ok)
	assert.Equal(t, entity.Orangtua{}.TableName(), call.Table)
}

func TestCreateAnak_MengirimEntityKeTabelYangBenar(t *testing.T) {
	db := testutil.NewFakeDatabase()
	stack := testutil.NewStack(t, db)

	var err error
	value := testutil.CatchPanic(func() { _, err = stack.Repository.CreateAnak(anakEntity(), nil) })

	require.Nil(t, value, "repository tidak boleh panic")
	require.NoError(t, err)

	call, ok := db.LastInsertOne()
	require.True(t, ok)
	assert.Equal(t, entity.Anak{}.TableName(), call.Table)
}

func TestCreateKunjungan_MengirimEntityKeTabelYangBenar(t *testing.T) {
	db := testutil.NewFakeDatabase()
	stack := testutil.NewStack(t, db)

	var err error
	value := testutil.CatchPanic(func() { _, err = stack.Repository.CreateKunjungan(kunjunganEntity(), nil) })

	require.Nil(t, value, "repository tidak boleh panic")
	require.NoError(t, err)

	call, ok := db.LastInsertOne()
	require.True(t, ok)
	assert.Equal(t, entity.Kunjungan{}.TableName(), call.Table)
}

func TestCreateOrangTua_MencariNikSebelumMenyimpan(t *testing.T) {
	db := testutil.NewFakeDatabase()
	stack := testutil.NewStack(t, db)

	_ = testutil.CatchPanic(func() { _, _ = stack.Repository.CreateOrangTua(orangtuaEntity(), nil) })

	require.NotEmpty(t, db.FindCalls, "orangtua harus dicari dulu berdasarkan NIK sebelum disimpan")
	assert.Equal(t, entity.Orangtua{}.TableName(), db.FindCalls[0].Table)
}

func TestDeleteOrangTua_MenghapusDariTabelOrangtua(t *testing.T) {
	db := testutil.NewFakeDatabase()
	stack := testutil.NewStack(t, db)

	require.NoError(t, stack.Repository.DeleteOrangTua(testutil.Ptr(testutil.IdOrangtua)))

	require.Len(t, db.DeleteCalls, 1)
	assert.Equal(t, entity.Orangtua{}.TableName(), db.DeleteCalls[0].Table)
}

func TestDeleteAnak_MenghapusDariTabelAnak(t *testing.T) {
	db := testutil.NewFakeDatabase()
	stack := testutil.NewStack(t, db)

	require.NoError(t, stack.Repository.DeleteAnak(testutil.IdAnak))

	require.Len(t, db.DeleteCalls, 1)
	assert.Equal(t, entity.Anak{}.TableName(), db.DeleteCalls[0].Table)
}

// --- NEGATIVE CASE ---

func TestCreate_MeneruskanErrorDatabase(t *testing.T) {
	dbErr := errors.New("koneksi database terputus")

	tests := []struct {
		nama    string
		panggil func(*testutil.Stack) error
	}{
		{"orangtua", func(s *testutil.Stack) error {
			_, err := s.Repository.CreateOrangTua(orangtuaEntity(), nil)
			return err
		}},
		{"anak", func(s *testutil.Stack) error { _, err := s.Repository.CreateAnak(anakEntity(), nil); return err }},
		{"kunjungan", func(s *testutil.Stack) error {
			_, err := s.Repository.CreateKunjungan(kunjunganEntity(), nil)
			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			db := testutil.NewFailingFakeDatabase(dbErr)
			stack := testutil.NewStack(t, db)

			var err error
			value := testutil.CatchPanic(func() { err = tc.panggil(stack) })

			require.Nil(t, value, "error database harus dikembalikan, bukan panic")
			require.Error(t, err)
			assert.ErrorIs(t, err, dbErr)
		})
	}
}

func TestDelete_MeneruskanErrorDatabase(t *testing.T) {
	dbErr := errors.New("gagal menghapus")
	db := testutil.NewFakeDatabase()
	db.DeleteErr = dbErr
	stack := testutil.NewStack(t, db)

	assert.ErrorIs(t, stack.Repository.DeleteOrangTua(testutil.Ptr(testutil.IdOrangtua)), dbErr)
	assert.ErrorIs(t, stack.Repository.DeleteAnak(testutil.IdAnak), dbErr)
}

func TestFindAnak_TanpaKunciPencarian(t *testing.T) {
	db := testutil.NewFakeDatabase()
	stack := testutil.NewStack(t, db)

	var err error
	value := testutil.CatchPanic(func() { _, err = stack.Repository.FindAnak(nil, nil, nil, nil, true) })

	require.Nil(t, value, "argumen nil harus ditolak dengan error, bukan panic")
	require.Error(t, err)
}

// --- NEGATIVE CASE: pencarian sebelum simpan ---

func TestFindOrangTuaByNik_MengembalikanSatuBaris(t *testing.T) {
	db := testutil.NewFakeDatabase()
	db.FindOneFunc = func(_ context.Context, result any, _ string, _ []port.DbExpression) error {
		dest, ok := result.(*entity.Orangtua)
		if !ok {
			return fmt.Errorf("FindOne harus menerima pointer ke entity supaya hasilnya bisa diisi, dapat %T", result)
		}
		*dest = entity.Orangtua{ID: testutil.IdOrangtua, NIK: testutil.NikOrangtua}
		return nil
	}
	stack := testutil.NewStack(t, db)

	res, err := stack.Repository.FindOrangTuaByNik(testutil.NikOrangtua, true)

	require.NoError(t, err)
	require.NotNil(t, res, "baris dari database harus sampai ke pemanggil")
	assert.Equal(t, testutil.IdOrangtua, res.ID)
}

func TestFindOrangTuaByNik_TidakDitemukan(t *testing.T) {
	db := testutil.NewFakeDatabase()
	stack := testutil.NewStack(t, db)

	res, err := stack.Repository.FindOrangTuaByNik(testutil.NikOrangtua, true)

	require.Error(t, err, "tidak ditemukan datang sebagai sql.ErrNoRows dari bun")
	assert.Nil(t, res)
}

func TestCreateOrangTua_MemakaiDataLamaBilaNikSudahAda(t *testing.T) {
	db := testutil.NewFakeDatabase()
	db.FindOneFunc = func(_ context.Context, result any, _ string, _ []port.DbExpression) error {
		dest, ok := result.(*entity.Orangtua)
		if !ok {
			return fmt.Errorf("FindOne harus menerima pointer ke entity, dapat %T", result)
		}
		*dest = entity.Orangtua{ID: "ORANGTUA-LAMA", NIK: testutil.NikOrangtua}
		return nil
	}
	stack := testutil.NewStack(t, db)

	res, err := stack.Repository.CreateOrangTua(orangtuaEntity(), nil)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "ORANGTUA-LAMA", res.ID,
		"orangtua dengan NIK yang sama harus dipakai ulang, bukan dibuat lagi")
	assert.Equal(t, 0, db.InsertOneCallCount(),
		"tidak boleh menyisipkan baris baru ketika NIK-nya sudah terdaftar")
}

func TestFindAnak_MengembalikanBarisDariDatabase(t *testing.T) {
	db := testutil.NewFakeDatabase()
	db.FindOneFunc = func(_ context.Context, result any, _ string, _ []port.DbExpression) error {
		dest, ok := result.(*entity.Anak)
		if !ok {
			return fmt.Errorf("FindOne harus menerima pointer ke entity, dapat %T", result)
		}
		*dest = entity.Anak{ID: testutil.IdAnak, NIK: testutil.Ptr(testutil.NikAnak)}
		return nil
	}
	stack := testutil.NewStack(t, db)

	res, err := stack.Repository.FindAnak(nil, testutil.Ptr(testutil.NikAnak), nil, nil, true)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, testutil.IdAnak, res.ID)
}

func TestFindAnak_ErrorSaatOrangtuaDanUrutanNil(t *testing.T) {
	db := testutil.NewFakeDatabase()
	db.FindOneFunc = func(_ context.Context, _ any, _ string, _ []port.DbExpression) error {
		return errors.New("query gagal")
	}
	stack := testutil.NewStack(t, db)

	var err error
	value := testutil.CatchPanic(func() {
		_, err = stack.Repository.FindAnak(nil, testutil.Ptr(testutil.NikAnak), nil, nil, true)
	})

	require.Nil(t, value,
		"jalur error tidak boleh men-dereference id_orangtua dan urutan yang memang boleh nil")
	require.Error(t, err)
}
