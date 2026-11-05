package entity_test

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

func fill(v reflect.Value, dasar int) {
	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		if !f.IsExported() || f.Anonymous || strings.Contains(f.Tag.Get("bun"), "rel:") {
			continue
		}
		fv := v.Field(i)
		if fv.Kind() == reflect.Ptr {
			if fv.Type().Elem().Kind() == reflect.Struct && fv.Type().Elem() != reflect.TypeOf(time.Time{}) {
				continue
			}
			fv.Set(reflect.New(fv.Type().Elem()))
			fv = fv.Elem()
		}
		n := dasar + i
		switch fv.Kind() {
		case reflect.String:
			fv.SetString(f.Name + string(rune('a'+n%26)))
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
			fv.SetInt(int64(n))
		case reflect.Float32, reflect.Float64:
			fv.SetFloat(float64(n) + 0.5)
		case reflect.Bool:
			fv.SetBool(true)
		case reflect.Struct:
			if fv.Type() == reflect.TypeOf(time.Time{}) {
				fv.Set(reflect.ValueOf(time.Date(2026, 1, 1+n%27, 0, 0, 0, 0, time.UTC)))
			}
		}
	}
}

type kasusOverride struct {
	nama       string
	baru       func() any
	timpa      func(old, baru any, force bool)
	dilindungi []string
}

func allOverrides() []kasusOverride {
	return []kasusOverride{
		{"Kunjungan", func() any { return new(entity.Kunjungan) },
			func(l, b any, f bool) { l.(*entity.Kunjungan).Override(*b.(*entity.Kunjungan), f) },
			[]string{"ID", "SatusehatId", "IDAnak", "IDFaskes", "IDEpisode", "IDRujukan", "RefEpisode", "RefRujukan", "Stunting", "CreatedAt", "UpdatedAt", "DeletedAt", "DeletedBy", "CreatedBy", "SourceData"}},
		{"Kesehatan", func() any { return new(entity.Kesehatan) },
			func(l, b any, f bool) { l.(*entity.Kesehatan).Override(*b.(*entity.Kesehatan), f) },
			[]string{"ID", "IDAnak", "CreatedAt", "DeletedAt", "DeletedBy", "CreatedBy", "SourceData"}},
		{"Anak", func() any { return new(entity.Anak) },
			func(l, b any, f bool) { l.(*entity.Anak).Override(*b.(*entity.Anak), f) },
			[]string{"ID", "CreatedAt", "DeletedAt", "DeletedBy", "CreatedBy", "SourceData"}},
		{"Orangtua", func() any { return new(entity.Orangtua) },
			func(l, b any, f bool) { l.(*entity.Orangtua).Override(*b.(*entity.Orangtua), f) },
			[]string{"ID", "CreatedAt", "DeletedAt", "DeletedBy", "CreatedBy", "SourceData"}},
		{"Diagnosa", func() any { return new(entity.Diagnosa) },
			func(l, b any, f bool) { l.(*entity.Diagnosa).Override(*b.(*entity.Diagnosa), f) },
			[]string{"ID", "SatusehatId", "IDAnak", "IDKunjungan", "RefEncounter", "CreatedAt", "UpdatedAt", "DeletedAt", "Jenis", "TanggalCatat"}},
		{"Observasi", func() any { return new(entity.Observasi) },
			func(l, b any, f bool) { l.(*entity.Observasi).Override(*b.(*entity.Observasi), f) },
			[]string{"ID", "SatusehatId", "IDAnak", "IDKunjungan", "IDInduk", "RefEncounter", "CreatedAt", "UpdatedAt", "DeletedAt"}},
		{"Layanan", func() any { return new(entity.Layanan) },
			func(l, b any, f bool) { l.(*entity.Layanan).Override(*b.(*entity.Layanan), f) },
			[]string{"ID", "SatusehatId", "IDAnak", "IDKunjungan", "RefEncounter", "Jenis", "CreatedAt", "UpdatedAt", "DeletedAt"}},
		{"Rujukan", func() any { return new(entity.Rujukan) },
			func(l, b any, f bool) { l.(*entity.Rujukan).Override(*b.(*entity.Rujukan), f) },
			[]string{"ID", "SatusehatId", "IDAnak", "IDKunjungan", "IDFaskesAsal", "IDFaskesTujuan", "RefEncounter", "RefFaskesAsal", "RefFaskesTujuan", "Jenis", "Tanggal", "CreatedAt", "UpdatedAt", "DeletedAt"}},
		{"Episode", func() any { return new(entity.Episode) },
			func(l, b any, f bool) { l.(*entity.Episode).Override(*b.(*entity.Episode), f) },
			[]string{"ID", "SatusehatId", "IDAnak", "IDFaskes", "CreatedAt", "UpdatedAt", "DeletedAt"}},
	}
}

func skipped(f reflect.StructField) bool {
	if !f.IsExported() || f.Anonymous || strings.Contains(f.Tag.Get("bun"), "rel:") {
		return true
	}
	return f.Type.Kind() == reflect.Ptr && f.Type.Elem().Kind() == reflect.Struct && f.Type.Elem() != reflect.TypeOf(time.Time{})
}

func TestOverrideMenimpaFieldTerisiDanMelindungiIdentitas(t *testing.T) {
	for _, k := range allOverrides() {
		t.Run(k.nama, func(t *testing.T) {
			old, baru := k.baru(), k.baru()
			fill(reflect.ValueOf(old).Elem(), 1)
			fill(reflect.ValueOf(baru).Elem(), 7)
			asli := reflect.ValueOf(old).Elem().Interface()

			k.timpa(old, baru, false)

			vl, vb, va := reflect.ValueOf(old).Elem(), reflect.ValueOf(baru).Elem(), reflect.ValueOf(asli)
			lindung := map[string]bool{}
			for _, n := range k.dilindungi {
				lindung[n] = true
				_, exists := vl.Type().FieldByName(n)
				assert.True(t, exists, "field %s tidak ada di %s", n, k.nama)
			}
			for i := 0; i < vl.NumField(); i++ {
				f := vl.Type().Field(i)
				if skipped(f) {
					continue
				}
				if lindung[f.Name] {
					assert.Equal(t, va.Field(i).Interface(), vl.Field(i).Interface(), "%s.%s tidak boleh tertimpa", k.nama, f.Name)
				} else {
					assert.Equal(t, vb.Field(i).Interface(), vl.Field(i).Interface(), "%s.%s harus tertimpa", k.nama, f.Name)
				}
			}
		})
	}
}

func TestOverrideNilaiKosongTidakMenghapus(t *testing.T) {
	for _, k := range allOverrides() {
		t.Run(k.nama, func(t *testing.T) {
			old := k.baru()
			fill(reflect.ValueOf(old).Elem(), 1)
			asli := reflect.ValueOf(old).Elem().Interface()

			k.timpa(old, k.baru(), false)
			assert.Equal(t, asli, reflect.ValueOf(old).Elem().Interface())
		})
	}
}

func TestOverrideForceMengosongkanKecualiIdentitas(t *testing.T) {
	for _, k := range allOverrides() {
		t.Run(k.nama, func(t *testing.T) {
			old := k.baru()
			fill(reflect.ValueOf(old).Elem(), 1)
			asli := reflect.ValueOf(reflect.ValueOf(old).Elem().Interface())

			k.timpa(old, k.baru(), true)

			vl := reflect.ValueOf(old).Elem()
			lindung := map[string]bool{}
			for _, n := range k.dilindungi {
				lindung[n] = true
			}
			for i := 0; i < vl.NumField(); i++ {
				f := vl.Type().Field(i)
				if skipped(f) {
					continue
				}
				if lindung[f.Name] {
					assert.Equal(t, asli.Field(i).Interface(), vl.Field(i).Interface(), "%s.%s", k.nama, f.Name)
				} else {
					assert.True(t, vl.Field(i).IsZero(), "%s.%s harus kosong saat force", k.nama, f.Name)
				}
			}
		})
	}
}

func TestBeforeAppendModelHanyaSaatUpdate(t *testing.T) {
	type pengait interface {
		BeforeAppendModel(context.Context, bun.Query) error
	}
	list := map[string]func() (pengait, func() *time.Time){
		"Orangtua": func() (pengait, func() *time.Time) {
			e := new(entity.Orangtua)
			return e, func() *time.Time { return e.UpdatedAt }
		},
		"Anak": func() (pengait, func() *time.Time) {
			e := new(entity.Anak)
			return e, func() *time.Time { return e.UpdatedAt }
		},
		"Kesehatan": func() (pengait, func() *time.Time) {
			e := new(entity.Kesehatan)
			return e, func() *time.Time { return e.UpdatedAt }
		},
		"Kunjungan": func() (pengait, func() *time.Time) {
			e := new(entity.Kunjungan)
			return e, func() *time.Time { return e.UpdatedAt }
		},
		"Diagnosa": func() (pengait, func() *time.Time) {
			e := new(entity.Diagnosa)
			return e, func() *time.Time { return e.UpdatedAt }
		},
		"Observasi": func() (pengait, func() *time.Time) {
			e := new(entity.Observasi)
			return e, func() *time.Time { return e.UpdatedAt }
		},
		"Layanan": func() (pengait, func() *time.Time) {
			e := new(entity.Layanan)
			return e, func() *time.Time { return e.UpdatedAt }
		},
	}
	for name, create := range list {
		t.Run(name, func(t *testing.T) {
			e, updated := create()
			assert.NoError(t, e.BeforeAppendModel(context.Background(), (*bun.InsertQuery)(nil)))
			assert.Nil(t, updated(), "insert tidak mengisi updatedAt")

			assert.NoError(t, e.BeforeAppendModel(context.Background(), (*bun.UpdateQuery)(nil)))
			if assert.NotNil(t, updated()) {
				assert.WithinDuration(t, time.Now(), *updated(), time.Second)
			}
		})
	}
}

func TestNamaTabelDanAliasPkUnik(t *testing.T) {
	table := map[string][2]string{
		"Orangtua":  {entity.Orangtua{}.TableName(), entity.Orangtua{}.GetPkName()},
		"Anak":      {entity.Anak{}.TableName(), entity.Anak{}.GetPkName()},
		"Kunjungan": {entity.Kunjungan{}.TableName(), entity.Kunjungan{}.GetPkName()},
		"Kesehatan": {entity.Kesehatan{}.TableName(), entity.Kesehatan{}.GetPkName()},
		"Diagnosa":  {entity.Diagnosa{}.TableName(), entity.Diagnosa{}.GetPkName()},
		"Observasi": {entity.Observasi{}.TableName(), entity.Observasi{}.GetPkName()},
		"Layanan":   {entity.Layanan{}.TableName(), entity.Layanan{}.GetPkName()},
		"Rujukan":   {entity.Rujukan{}.TableName(), entity.Rujukan{}.GetPkName()},
		"Episode":   {entity.Episode{}.TableName(), entity.Episode{}.GetPkName()},
		"Faskes":    {entity.Faskes{}.TableName(), entity.Faskes{}.GetPkName()},
		"Posyandu":  {entity.Posyandu{}.TableName(), entity.Posyandu{}.GetPkName()},
	}
	alias := map[string]string{}
	for name, tp := range table {
		assert.Equal(t, "stunting."+strings.ToLower(name), tp[0])
		assert.True(t, strings.HasSuffix(tp[1], ".id"), "%s pk %q", name, tp[1])
		a := strings.TrimSuffix(tp[1], ".id")
		assert.Empty(t, alias[a], "alias %q dipakai %s dan %s", a, alias[a], name)
		alias[a] = name
	}
	assert.Equal(t, "kafka.transactions", entity.FailedTransactions{}.TableName())
}
