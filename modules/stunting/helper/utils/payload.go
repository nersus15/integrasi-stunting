package utils

import "reflect"

var fieldSistem = map[string]bool{
	"CreatedAt": true,
	"UpdatedAt": true,
	"DeletedAt": true,
	"UpdatedBy": true,
	"DeletedBy": true,
}

// kolom audit diisi sistem; nilai dari pengirim dibuang, termasuk di struct bersarang
func AbaikanFieldSistem(v any) {
	abaikanFieldSistem(reflect.ValueOf(v))
}

func abaikanFieldSistem(v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			abaikanFieldSistem(v.Elem())
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			abaikanFieldSistem(v.Index(i))
		}
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if !f.CanSet() {
				continue
			}
			if fieldSistem[t.Field(i).Name] {
				f.SetZero()
				continue
			}
			abaikanFieldSistem(f)
		}
	}
}
