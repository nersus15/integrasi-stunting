package repository_db_test

import (
	"testing"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func errorCode(err error) int {
	if err == nil {
		return 0
	}
	return exceptions.Classify(err).ErrorCode
}

func setupRiwayat(t *testing.T) {
	t.Helper()
	cleanup := func() { require.NoError(t, testutil.Scenario1Cleanup(db, groupUji)) }
	cleanup()
	t.Cleanup(cleanup)
	require.NoError(t, wiring.Scenario1Jakantro())
	require.NoError(t, wiring.Scenario1Send(testutil.Scenario1KafkaFiles...))
}

func scalar[T any](t *testing.T, q string, args ...any) T {
	t.Helper()
	var v T
	require.NoError(t, db.QueryRow(q, args...).Scan(&v), q)
	return v
}

func TestFindRujukanDanKunjunganByIdAnak(t *testing.T) {
	setupRiwayat(t)
	repo := wiring.Repository

	r, err := repo.FindRujukanByIdAnak(testutil.Scenario1IdAnak)
	require.NoError(t, err)
	assert.Equal(t, testutil.Scenario1IdAnak, r.IDAnak)
	assert.NotNil(t, r.FaskesTujuan)

	_, err = repo.FindRujukanByIdAnak("")
	assert.Equal(t, exceptions.Validasi.ErrorCode, errorCode(err))
	_, err = repo.FindRujukanByIdAnak("tidak-ada")
	assert.Equal(t, exceptions.TidakDitemukan.ErrorCode, errorCode(err))

	k, err := repo.FindKunjunganByIdAnak(testutil.Scenario1IdAnak)
	require.NoError(t, err)
	assert.Equal(t, testutil.Scenario1IdAnak, k.IDAnak)
	_, err = repo.FindKunjunganByIdAnak("tidak-ada")
	assert.Equal(t, exceptions.TidakDitemukan.ErrorCode, errorCode(err))
}

func TestCreateDataMedisSatuPerSatuDanHubungkan(t *testing.T) {
	setupRiwayat(t)
	repo := wiring.Repository
	sekarang := time.Now()
	ref := "enc-uji-repo"
	idAnak := testutil.Scenario1IdAnak

	_, err := repo.CreateObservasi(&entity.Observasi{ID: "uji-repo-obs", IDAnak: idAnak, RefEncounter: &ref, System: "s", Kode: "k"}, nil)
	require.NoError(t, err)
	_, err = repo.CreateDiagnosa(&entity.Diagnosa{ID: "uji-repo-dg", IDAnak: idAnak, RefEncounter: &ref, Jenis: "diagnosis", System: "s", Kode: "E45"}, nil)
	require.NoError(t, err)
	_, err = repo.CreateLayanan(&entity.Layanan{ID: "uji-repo-ly", IDAnak: idAnak, RefEncounter: &ref, Jenis: "procedure"}, nil)
	require.NoError(t, err)
	_, err = repo.CreateRujukan(&entity.Rujukan{ID: "uji-repo-rj", IDAnak: idAnak, RefEncounter: &ref, Jenis: "internal", Tanggal: &sekarang}, nil)
	require.NoError(t, err)
	_, err = repo.CreateEpisode(&entity.Episode{ID: "uji-repo-ep", IDAnak: idAnak}, nil)
	require.NoError(t, err)

	idKunjungan := scalar[string](t, `SELECT id FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`)
	for _, table := range []string{"stunting.observasi", "stunting.diagnosa", "stunting.layanan", "stunting.rujukan"} {
		n, err := repo.HubungkanKeKunjungan(table, idKunjungan, ref, nil)
		require.NoError(t, err, table)
		assert.EqualValues(t, 1, n, table)

		n, err = repo.HubungkanKeKunjungan(table, idKunjungan, ref, nil)
		require.NoError(t, err)
		assert.EqualValues(t, 0, n, "%s: yang sudah terhubung tidak disentuh lagi", table)
	}

	_, err = repo.HubungkanKeKunjungan("stunting.observasi", "", ref, nil)
	assert.Equal(t, exceptions.Validasi.ErrorCode, errorCode(err))

	_, err = repo.CreateObservasi(nil, nil)
	assert.Equal(t, exceptions.BodyRusak.ErrorCode, errorCode(err))
	_, err = repo.CreateDiagnosa(nil, nil)
	assert.Equal(t, exceptions.BodyRusak.ErrorCode, errorCode(err))
	_, err = repo.CreateLayanan(nil, nil)
	assert.Equal(t, exceptions.BodyRusak.ErrorCode, errorCode(err))
	_, err = repo.CreateRujukan(nil, nil)
	assert.Equal(t, exceptions.BodyRusak.ErrorCode, errorCode(err))
	_, err = repo.CreateEpisode(nil, nil)
	assert.Equal(t, exceptions.BodyRusak.ErrorCode, errorCode(err))

	_, err = repo.CreateObservasi(&entity.Observasi{ID: "uji-repo-obs", IDAnak: idAnak, System: "s", Kode: "k"}, nil)
	assert.Error(t, err)
	_, err = repo.CreateDiagnosa(&entity.Diagnosa{ID: "uji-repo-dg", IDAnak: idAnak, Jenis: "diagnosis", Kode: "E45"}, nil)
	assert.Error(t, err)
	_, err = repo.CreateLayanan(&entity.Layanan{ID: "uji-repo-ly", IDAnak: idAnak, Jenis: "procedure"}, nil)
	assert.Error(t, err)
	_, err = repo.CreateRujukan(&entity.Rujukan{ID: "uji-repo-rj", IDAnak: idAnak, Jenis: "internal"}, nil)
	assert.Error(t, err)
	_, err = repo.CreateEpisode(&entity.Episode{ID: "uji-repo-ep", IDAnak: idAnak}, nil)
	assert.Error(t, err)
}

func TestUpdateOrangTuaById(t *testing.T) {
	setupRiwayat(t)
	svc := wiring.Service

	o, err := svc.UpdateOrangTuaById(&entity.Orangtua{ID: testutil.Scenario1IdOrtu, Telepon: "0899"}, testutil.Ptr("jakantro"))
	require.NoError(t, err)
	assert.Equal(t, "0899", o.Telepon)
	assert.Equal(t, "0899/Rina Marlina/jakantro", scalar[string](t,
		`SELECT telepon || '/' || nama_ibu || '/' || updated_by FROM stunting.orangtua WHERE id = $1`, testutil.Scenario1IdOrtu))

	_, err = svc.UpdateOrangTuaById(&entity.Orangtua{ID: "tidak-ada"}, nil)
	assert.Equal(t, exceptions.TidakDitemukan.ErrorCode, errorCode(err))

	repo := wiring.Repository
	_, err = repo.UpdateOrangTuaById(nil, nil, true)
	assert.Equal(t, exceptions.BodyRusak.ErrorCode, errorCode(err))
	_, err = repo.UpdateOrangTuaById(&entity.Orangtua{}, nil, true)
	assert.Equal(t, exceptions.Validasi.ErrorCode, errorCode(err))
	_, err = repo.UpdateOrangTuaById(&entity.Orangtua{ID: "tidak-ada"}, nil, true)
	assert.Equal(t, exceptions.TidakDitemukan.ErrorCode, errorCode(err))
}
