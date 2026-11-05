package delete_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webcore-go/webcore/port"
)

const (
	idOrtu      = "ortu-hapus"
	nikOrtu     = "3175049700000001"
	ihsOrtu     = "RP-HAPUS"
	idAnak      = "anak-hapus"
	nikAnak     = "3175049700000002"
	ihsAnak     = "P-HAPUS"
	penghapus   = "jakantro"
	tabelOrtu   = "stunting.orangtua"
	tabelAnak   = "stunting.anak"
	kodeHilang  = 2001
	tungguCache = 2 * time.Second
)

func TestMain(m *testing.M) {
	testutil.SetupLogger()
	os.Exit(m.Run())
}

func errorCode(err error) int {
	var e *exceptions.Error
	if errors.As(err, &e) {
		return e.ErrorCode
	}
	return 0
}

func TestDeleteSoftDeleteLewatUpdate(t *testing.T) {
	for name, remove := range map[string]func(s *testutil.Stack) error{
		tabelAnak: func(s *testutil.Stack) error { return s.Repository.DeleteAnakByID(idAnak, penghapus) },
		tabelOrtu: func(s *testutil.Stack) error { return s.Repository.DeleteOrangtuaByID(idOrtu, penghapus) },
	} {
		t.Run(name, func(t *testing.T) {
			db := testutil.NewFakeDatabase()
			db.UpdateOneAffected = 1
			require.NoError(t, remove(testutil.NewStack(t, db)))

			require.Len(t, db.UpdateOneCalls, 1)
			c := db.UpdateOneCalls[0]
			assert.Equal(t, name, c.Table)
			data, ok := c.Data.(map[string]any)
			require.True(t, ok, "data update berupa map kolom")
			assert.IsType(t, time.Time{}, data["deletedAt"])
			assert.Equal(t, penghapus, data["deleted_by"])
			assert.Len(t, data, 2, "hanya deletedAt dan deleted_by yang diubah")
			assert.Empty(t, db.DeleteCalls, "soft delete, bukan DELETE")
		})
	}
}

func TestDeleteTidakAdaBarisJadi404(t *testing.T) {
	db := testutil.NewFakeDatabase()
	s := testutil.NewStack(t, db)

	assert.Equal(t, kodeHilang, errorCode(s.Repository.DeleteAnakByID(idAnak, penghapus)))
	assert.Equal(t, kodeHilang, errorCode(s.Repository.DeleteOrangtuaByID(idOrtu, penghapus)))
}

func TestDeleteMeneruskanErrorDatabase(t *testing.T) {
	dbErr := errors.New("koneksi terputus")
	db := testutil.NewFakeDatabase()
	db.UpdateOneErr = dbErr
	s := testutil.NewStack(t, db)

	assert.ErrorIs(t, s.Repository.DeleteAnakByID(idAnak, penghapus), dbErr)
	assert.ErrorIs(t, s.Repository.DeleteOrangtuaByID(idOrtu, penghapus), dbErr)
}

// setelah soft delete, FindOne (bun dengan tag soft_delete) tidak lagi menemukan barisnya;
// hanya FindOneDeleted yang masih bisa
func simulateDeletedRow(db *testutil.FakeDatabase, content func(any)) {
	db.FindOneFunc = func(ctx context.Context, result any, table string, filter []port.DbExpression) error {
		return sql.ErrNoRows
	}
	db.FindOneDeletedFunc = func(ctx context.Context, result any, table string, filter []port.DbExpression) error {
		content(result)
		return nil
	}
}

func fillOrangtua(result any) {
	if o, ok := result.(*entity.Orangtua); ok {
		o.ID, o.NIK, o.SatusehatId, o.CreatedBy = idOrtu, nikOrtu, ihsOrtu, testutil.Ptr(penghapus)
	}
}

func fillAnak(result any) {
	if a, ok := result.(*entity.Anak); ok {
		a.ID, a.NIK, a.SatusehatId, a.IDOrangtua, a.AnakKe, a.CreatedBy = idAnak, testutil.Ptr(nikAnak), ihsAnak, idOrtu, 1, testutil.Ptr(penghapus)
	}
}

func TestDeleteOrangtuaMelupakanCache(t *testing.T) {
	db := testutil.NewFakeDatabase()
	db.UpdateOneAffected = 1
	db.FindOneFunc = func(ctx context.Context, result any, table string, filter []port.DbExpression) error {
		fillOrangtua(result)
		return nil
	}
	s := testutil.NewStack(t, db)

	_, err := s.Repository.FindOrangTuaById(idOrtu, true)
	require.NoError(t, err)
	_, err = s.Repository.FindOrangTuaByNik(nikOrtu, true)
	require.NoError(t, err)
	_, err = s.Repository.FindOrangtuaBySatusehatId(ihsOrtu, true)
	require.NoError(t, err)

	simulateDeletedRow(db, fillOrangtua)
	require.NoError(t, s.Service.DeleteOrangtuaByID(idOrtu, penghapus))

	find := map[string]func() error{
		"id":           func() error { _, err := s.Repository.FindOrangTuaById(idOrtu, true); return err },
		"nik":          func() error { _, err := s.Repository.FindOrangTuaByNik(nikOrtu, true); return err },
		"satusehat_id": func() error { _, err := s.Repository.FindOrangtuaBySatusehatId(ihsOrtu, true); return err },
	}
	for key, fn := range find {
		assert.Eventually(t, func() bool { return fn() != nil }, tungguCache, 20*time.Millisecond,
			"cache orangtua by %s masih menyimpan orangtua yang sudah dihapus", key)
	}
}

func TestDeleteAnakMelupakanCache(t *testing.T) {
	db := testutil.NewFakeDatabase()
	db.UpdateOneAffected = 1
	db.FindOneFunc = func(ctx context.Context, result any, table string, filter []port.DbExpression) error {
		fillAnak(result)
		return nil
	}
	s := testutil.NewStack(t, db)

	_, err := s.Repository.FindAnak(testutil.Ptr(idAnak), nil, nil, nil, true)
	require.NoError(t, err)
	_, err = s.Repository.FindAnak(nil, testutil.Ptr(nikAnak), nil, nil, true)
	require.NoError(t, err)
	_, err = s.Repository.FindAnakBySatusehatId(ihsAnak, true)
	require.NoError(t, err)

	simulateDeletedRow(db, fillAnak)
	require.NoError(t, s.Service.DeleteAnakByID(idAnak, penghapus))

	find := map[string]func() error{
		"id":           func() error { _, err := s.Repository.FindAnak(testutil.Ptr(idAnak), nil, nil, nil, true); return err },
		"nik":          func() error { _, err := s.Repository.FindAnak(nil, testutil.Ptr(nikAnak), nil, nil, true); return err },
		"satusehat_id": func() error { _, err := s.Repository.FindAnakBySatusehatId(ihsAnak, true); return err },
	}
	for key, fn := range find {
		assert.Eventually(t, func() bool { return fn() != nil }, tungguCache, 20*time.Millisecond,
			"cache anak by %s masih menyimpan anak yang sudah dihapus", key)
	}
}
