package types_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
)

func fillEntity(v reflect.Value, urut *int) {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() || f.Anonymous {
			continue
		}

		tag := f.Tag.Get("bun")
		if tag == "-" || strings.Contains(tag, "rel:") {
			continue
		}

		*urut++
		fillField(v.Field(i), f.Name, *urut)
	}
}

func fillField(fv reflect.Value, name string, n int) {
	if fv.Kind() == reflect.Ptr {
		if fv.Type().Elem().Kind() == reflect.Struct && fv.Type().Elem() != reflect.TypeOf(time.Time{}) {
			return
		}
		fv.Set(reflect.New(fv.Type().Elem()))
		fv = fv.Elem()
	}

	switch fv.Kind() {
	case reflect.String:
		fv.SetString(stringValue(name, n))
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		fv.SetInt(int64(n))
	case reflect.Float32, reflect.Float64:
		fv.SetFloat(float64(n))
	case reflect.Bool:
		fv.SetBool(true)
	case reflect.Struct:
		if fv.Type() == reflect.TypeOf(time.Time{}) {
			fv.Set(reflect.ValueOf(time.Date(2026, 1, 1+n%27, 0, 0, 0, 0, time.UTC)))
		}
	}
}

func stringValue(name string, n int) string {
	switch {
	case strings.Contains(name, "NIK"), strings.Contains(name, "Nik"),
		strings.Contains(name, "NoKk"), strings.Contains(name, "NoKK"):
		return "35150101010100" + pad(n)
	case name == "JenisKelamin":
		return "L"
	case strings.HasPrefix(name, "Tanggal"), name == "Onset", name == "Mulai",
		name == "Selesai", name == "TanggalLahir":
		return "2026-01-15"
	case name == "Rt", name == "Rw", name == "RT", name == "RW":
		return "00" + pad(n)[1:]
	}
	return "v" + pad(n)
}

func pad(n int) string {
	s := "00" + string(rune('0'+n%10))
	return s[len(s)-2:]
}

func compare(t *testing.T, label string, asal, result reflect.Value, abaikan map[string]bool) {
	t.Helper()
	tp := asal.Type()

	for i := 0; i < tp.NumField(); i++ {
		f := tp.Field(i)
		if !f.IsExported() || f.Anonymous || abaikan[f.Name] {
			continue
		}

		tag := f.Tag.Get("bun")
		if tag == "-" || strings.Contains(tag, "rel:") || strings.Contains(tag, "soft_delete") {
			continue
		}
		if f.Type.Kind() == reflect.Ptr && f.Type.Elem().Kind() == reflect.Struct &&
			f.Type.Elem() != reflect.TypeOf(time.Time{}) {
			continue
		}
		if f.Type.Kind() == reflect.Slice {
			continue
		}

		a, b := asal.Field(i).Interface(), result.Field(i).Interface()
		if !reflect.DeepEqual(a, b) {
			t.Errorf("%s.%s hilang atau berubah saat konversi: %v -> %v", label, f.Name, fill(a), fill(b))
		}
	}
}

func fill(v any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		return rv.Elem().Interface()
	}
	return v
}

func TestKonversiAnakBolakBalik(t *testing.T) {
	e := new(entity.Anak)
	n := 0
	fillEntity(reflect.ValueOf(e).Elem(), &n)

	var d *types.Anak
	result := d.FromEntity(e).ToPayload().ToEntity()

	compare(t, "Anak", reflect.ValueOf(e).Elem(), reflect.ValueOf(result).Elem(), nil)
}

func TestKonversiOrangtuaBolakBalik(t *testing.T) {
	e := new(entity.Orangtua)
	n := 0
	fillEntity(reflect.ValueOf(e).Elem(), &n)

	var d *types.Orangtua
	result := d.FromEntity(e).ToPayload().ToEntity()

	compare(t, "Orangtua", reflect.ValueOf(e).Elem(), reflect.ValueOf(result).Elem(), nil)
}

func TestKonversiKunjunganBolakBalik(t *testing.T) {
	e := new(entity.Kunjungan)
	n := 0
	fillEntity(reflect.ValueOf(e).Elem(), &n)

	var d *types.Kunjungan
	result := d.FromEntity(e).ToPayload().ToEntity()

	compare(t, "Kunjungan", reflect.ValueOf(e).Elem(), reflect.ValueOf(result).Elem(), nil)
}

func TestKonversiKesehatanBolakBalik(t *testing.T) {
	e := new(entity.Kesehatan)
	n := 0
	fillEntity(reflect.ValueOf(e).Elem(), &n)

	var d *types.Kesehatan
	result := d.FromEntity(e).ToPayload().ToEntity()

	compare(t, "Kesehatan", reflect.ValueOf(e).Elem(), reflect.ValueOf(result).Elem(), nil)
}

func TestKonversiDataMedisBolakBalik(t *testing.T) {
	cases := []struct {
		nama  string
		asal  any
		bolak func(any) any
	}{
		{"Diagnosa", new(entity.Diagnosa), func(e any) any {
			var d *types.Diagnosa
			return d.FromEntity(e.(*entity.Diagnosa)).ToEntity()
		}},
		{"Observasi", new(entity.Observasi), func(e any) any {
			var d *types.Observasi
			return d.FromEntity(e.(*entity.Observasi)).ToEntity()
		}},
		{"Layanan", new(entity.Layanan), func(e any) any {
			var d *types.Layanan
			return d.FromEntity(e.(*entity.Layanan)).ToEntity()
		}},
		{"Rujukan", new(entity.Rujukan), func(e any) any {
			var d *types.Rujukan
			return d.FromEntity(e.(*entity.Rujukan)).ToEntity()
		}},
		{"Episode", new(entity.Episode), func(e any) any {
			var d *types.Episode
			return d.FromEntity(e.(*entity.Episode)).ToEntity()
		}},
	}

	for _, k := range cases {
		t.Run(k.nama, func(t *testing.T) {
			n := 0
			fillEntity(reflect.ValueOf(k.asal).Elem(), &n)

			result := k.bolak(k.asal)

			compare(t, k.nama, reflect.ValueOf(k.asal).Elem(),
				reflect.ValueOf(result).Elem(), nil)
		})
	}
}
