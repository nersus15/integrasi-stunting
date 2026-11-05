package utils_test

import (
	"math"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/repository"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
)

func TestLevelWilayah(t *testing.T) {
	cases := map[string]int{
		"00.00.00.0000": 0,
		"31.00.00.0000": 1,
		"31.74.00.0000": 2,
		"31.74.02.0000": 3,
		"31.74.02.1001": 4,
	}
	for code, level := range cases {
		assert.Equal(t, level, entity.LevelWilayah(testutil.Ptr(code)), code)
	}
	assert.Equal(t, -1, entity.LevelWilayah(nil))
	assert.Equal(t, -1, entity.LevelWilayah(testutil.Ptr("")))
}

func TestSubstr(t *testing.T) {
	assert.Equal(t, "31.74.02.", utils.Substr("31.74.02.1001", 0, 9))
	assert.Equal(t, "1001", utils.Substr("31.74.02.1001", 9, 10), "dipotong di ujung")
	assert.Equal(t, "", utils.Substr("abc", -1, 2))
	assert.Equal(t, "", utils.Substr("abc", 3, 1))
	assert.Equal(t, "é", utils.Substr("aéb", 1, 1), "per rune, bukan byte")
}

type berTag struct {
	ID          string  `bun:"id,pk"`
	Nama        string  `bun:"nama"`
	Tanpa       string  `bun:",nullzero"`
	Abaikan     string  `bun:"-"`
	Hapus       *string `bun:"deletedAt,soft_delete"`
	Relasi      *berTag `bun:"rel:belongs-to"`
	Pindai      int     `bun:"pindai,scanonly"`
	Lewat       int     `bun:"lewat,skipupdate"`
	Opsi        int     `bun:"opsi,rel:has-one"`
	JSONSaja    string  `json:"json_saja"`
	JSONBuang   string  `json:"-"`
	Polos       string
	tersembunyi string
	Sematan
}

type Sematan struct {
	Dalam string `bun:"dalam"`
}

type tanpaBun struct {
	A string `json:"a"`
	B int
}

func TestStructToMapMembacaTagBun(t *testing.T) {
	m := repository.StructToMap(&berTag{ID: "1", Nama: "n", Tanpa: "t", JSONSaja: "j", Polos: "p", Sematan: Sematan{Dalam: "d"}})
	assert.Equal(t, map[string]any{
		"nama": "n", "Tanpa": "t", "json_saja": "j", "Polos": "p", "dalam": "d",
	}, m)

	var nilPtr *berTag
	assert.Nil(t, repository.StructToMap(nilPtr))
}

func TestStructToMapJatuhKeJSON(t *testing.T) {
	assert.Equal(t, map[string]any{"a": "x", "B": float64(2)}, repository.StructToMap(tanpaBun{A: "x", B: 2}))
	assert.Equal(t, map[string]any{"k": "v"}, repository.StructToMap(map[string]string{"k": "v"}))
	assert.Nil(t, repository.StructToMap([]int{1}), "bukan objek JSON")
	assert.Nil(t, repository.StructToMap(math.Inf(1)), "gagal marshal")
}

func TestKonversiSmallint(t *testing.T) {
	assert.False(t, utils.SmallintToBool(nil))
	assert.False(t, utils.SmallintToBool(testutil.Ptr(0)))
	assert.True(t, utils.SmallintToBool(testutil.Ptr(1)))
	assert.Equal(t, 1, *utils.BoolToSmallintPtr(true))
	assert.Equal(t, 0, utils.BoolToSmallint(false))
}
