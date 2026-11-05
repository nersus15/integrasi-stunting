package kepemilikan_test

import (
	"encoding/json"
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOverrideTidakMengubahCreatedBy(t *testing.T) {
	for _, force := range []bool{false, true} {
		asal, other := "jakantro", "200022"

		o := entity.Orangtua{CreatedBy: &asal}
		o.Override(entity.Orangtua{CreatedBy: &other}, force)
		assert.Equal(t, "jakantro", *o.CreatedBy, "orangtua force=%v", force)

		a := entity.Anak{CreatedBy: &asal}
		a.Override(entity.Anak{CreatedBy: &other}, force)
		assert.Equal(t, "jakantro", *a.CreatedBy, "anak force=%v", force)

		k := entity.Kunjungan{CreatedBy: &asal}
		k.Override(entity.Kunjungan{CreatedBy: &other}, force)
		assert.Equal(t, "jakantro", *k.CreatedBy, "kunjungan force=%v", force)

		h := entity.Kesehatan{CreatedBy: &asal}
		h.Override(entity.Kesehatan{CreatedBy: &other}, force)
		assert.Equal(t, "jakantro", *h.CreatedBy, "kesehatan force=%v", force)
	}
}

func TestOverrideKesehatanTidakMengubahDeletedBy(t *testing.T) {
	for _, force := range []bool{false, true} {
		asal, other := "jakantro", "lain"
		h := entity.Kesehatan{DeletedBy: &asal}
		h.Override(entity.Kesehatan{DeletedBy: &other}, force)
		assert.Equal(t, "jakantro", *h.DeletedBy, "force=%v", force)
	}
}

func jsonKeys(t *testing.T, v any) map[string]any {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	m := map[string]any{}
	require.NoError(t, json.Unmarshal(b, &m))
	return m
}

func TestResponMemuatKolomAudit(t *testing.T) {
	for name, v := range map[string]any{
		"orangtua":  types.Orangtua{},
		"anak":      types.Anak{},
		"kunjungan": types.Kunjungan{},
		"kesehatan": types.Kesehatan{},
	} {
		m := jsonKeys(t, v)
		for _, k := range []string{"created_by", "updated_by", "deleted_by"} {
			assert.Contains(t, m, k, "%s harus memuat %s", name, k)
		}
	}
}

func TestDTOMembawaCreatedByDariEntity(t *testing.T) {
	oleh := "100011"
	assert.Equal(t, &oleh, (*types.Orangtua)(nil).FromEntity(&entity.Orangtua{CreatedBy: &oleh}).CreatedBy)
	assert.Equal(t, &oleh, (*types.Anak)(nil).FromEntity(&entity.Anak{CreatedBy: &oleh}).CreatedBy)
	assert.Equal(t, &oleh, (*types.Kunjungan)(nil).FromEntity(&entity.Kunjungan{CreatedBy: &oleh}).CreatedBy)
	assert.Equal(t, &oleh, (*types.Kesehatan)(nil).FromEntity(&entity.Kesehatan{CreatedBy: &oleh}).CreatedBy)
}
