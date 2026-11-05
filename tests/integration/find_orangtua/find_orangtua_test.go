package find_orangtua_test

import (
	"context"
	"strings"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webcore-go/webcore/port"
)

// muatan berbahaya yang lazim dipakai untuk menguji SQL injection
var muatanInjeksi = []struct {
	nama  string
	nilai string
}{
	{"tutup kutip lalu OR", "' OR '1'='1"},
	{"komentar sql", "1' --"},
	{"union select", "' UNION SELECT * FROM jakantro.orangtua --"},
	{"drop table", "'; DROP TABLE jakantro.orangtua; --"},
	{"titik koma ganda", "1; DELETE FROM jakantro.anak"},
	{"kutip ganda", `" OR ""="`},
	{"backslash", `\' OR 1=1`},
	{"null byte", "abc\x00def"},
}

// --- POSITIVE CASE ---

func TestFindOrangTuaById_MemakaiFilterTerparameter(t *testing.T) {
	db := testutil.NewFakeDatabase()
	db.FindOneFunc = func(_ context.Context, result any, _ string, _ []port.DbExpression) error {
		dest := result.(*entity.Orangtua)
		*dest = entity.Orangtua{ID: testutil.IdOrangtua}
		return nil
	}
	stack := testutil.NewStack(t, db)

	res, err := stack.Repository.FindOrangTuaById(testutil.IdOrangtua, true)

	require.NoError(t, err)
	require.NotNil(t, res)

	require.Len(t, db.FindCalls, 1)
	f := db.FindCalls[0]
	assert.Equal(t, "[me]", f.Table, "lewat model supaya tag soft_delete berlaku")
	require.Len(t, f.Filter, 1)
	assertSafeExpression(t, f.Expr(), "id")
	assert.Equal(t, []any{testutil.IdOrangtua}, f.Filter[0].Args)
}

func TestFindOrangTuaByNik_MemakaiFilterTerparameter(t *testing.T) {
	db := testutil.NewFakeDatabase()
	stack := testutil.NewStack(t, db)

	_, _ = stack.Repository.FindOrangTuaByNik(testutil.NikOrangtua, true)

	require.Len(t, db.FindCalls, 1)
	assertSafeExpression(t, db.FindCalls[0].Expr(), "nik")
}

// --- NEGATIVE CASE: SQL injection ---

func TestFindOrangTuaById_MuatanInjeksiTidakMasukKeSQL(t *testing.T) {
	for _, m := range muatanInjeksi {
		t.Run(m.nama, func(t *testing.T) {
			db := testutil.NewFakeDatabase()
			stack := testutil.NewStack(t, db)

			value := testutil.CatchPanic(func() {
				_, _ = stack.Repository.FindOrangTuaById(m.nilai, true)
			})
			require.Nil(t, value, "muatan injeksi tidak boleh membuat repository panic")

			require.Len(t, db.FindCalls, 1)
			f := db.FindCalls[0]

			assertSafeExpression(t, f.Expr(), "id")
			assert.NotContains(t, f.Expr(), m.nilai,
				"muatan menyusup ke ekspresi SQL, bukan hanya jadi nilai parameter")

			// Muatannya harus utuh sebagai argumen, bukan dipotong atau di-escape
			require.Len(t, f.Filter[0].Args, 1)
			assert.Equal(t, m.nilai, f.Filter[0].Args[0],
				"nilai harus sampai apa adanya sebagai parameter terikat")
		})
	}
}

func TestFindOrangTuaByNik_MuatanInjeksiTidakMasukKeSQL(t *testing.T) {
	for _, m := range muatanInjeksi {
		t.Run(m.nama, func(t *testing.T) {
			db := testutil.NewFakeDatabase()
			stack := testutil.NewStack(t, db)

			value := testutil.CatchPanic(func() {
				_, _ = stack.Repository.FindOrangTuaByNik(m.nilai, true)
			})
			require.Nil(t, value)

			require.Len(t, db.FindCalls, 1)
			f := db.FindCalls[0]

			assertSafeExpression(t, f.Expr(), "nik")
			assert.NotContains(t, f.Expr(), m.nilai)
			require.Len(t, f.Filter[0].Args, 1)
			assert.Equal(t, m.nilai, f.Filter[0].Args[0])
		})
	}
}

// pastikanEkspresiAman memeriksa bagian filter yang benar-benar disambung ke
// string SQL oleh lib-sql. Dua bentuk sama-sama sah: nama kolom polos ("nik")
// dan ekspresi berplaceholder ("nik = ?"). Yang tidak boleh adalah kutip,
// titik koma, komentar, atau kata kunci SQL - itu tanda ada nilai yang bocor
// ke dalam SQL alih-alih diikat sebagai parameter.
func assertSafeExpression(t *testing.T, expr, kolom string) {
	t.Helper()

	assert.Contains(t, expr, kolom, "ekspresi harus menyebut kolom yang dimaksud")
	assert.LessOrEqual(t, strings.Count(expr, "?"), 1, "hanya boleh satu placeholder")

	for _, terlarang := range []string{"'", `"`, ";", "--", "/*"} {
		assert.NotContains(t, expr, terlarang,
			"ekspresi SQL tidak boleh mengandung %q", terlarang)
	}
	for _, kata := range []string{"SELECT", "DROP", "DELETE", "UNION", "INSERT", "UPDATE"} {
		assert.NotContains(t, strings.ToUpper(expr), kata,
			"ekspresi SQL tidak boleh mengandung kata kunci %s", kata)
	}
}
