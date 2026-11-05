package entity_test

import (
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/stretchr/testify/assert"
)

// --- POSITIVE CASE ---

// --- NEGATIVE CASE ---

func TestZeroValueTidakMembawaData(t *testing.T) {
	var o entity.Orangtua
	var a entity.Anak
	var k entity.Kunjungan

	assert.Empty(t, o.ID)
	assert.Empty(t, a.ID)
	assert.Empty(t, a.IDOrangtua)
	assert.Empty(t, k.ID)
	assert.Empty(t, k.IDAnak)
}
