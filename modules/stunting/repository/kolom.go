package repository

import (
	"encoding/json"
	"reflect"
	"strings"
)

func StructToMap(thestruct any) map[string]any {
	rv := reflect.ValueOf(thestruct)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return structToMapJSON(thestruct)
	}

	hasil := make(map[string]any)
	if !kumpulkanKolomBun(rv, hasil) {
		return structToMapJSON(thestruct)
	}

	return hasil
}

func kumpulkanKolomBun(rv reflect.Value, hasil map[string]any) bool {
	rt := rv.Type()
	adaTagBun := false

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		tag, punyaTag := field.Tag.Lookup("bun")
		if punyaTag {
			adaTagBun = true
		}

		if field.Anonymous && !punyaTag {
			if field.Type.Kind() == reflect.Struct && kumpulkanKolomBun(rv.Field(i), hasil) {
				adaTagBun = true
			}
			continue
		}

		nama, lewati := namaKolomBun(field, tag, punyaTag)
		if lewati || nama == "" {
			continue
		}

		hasil[nama] = rv.Field(i).Interface()
	}

	return adaTagBun
}

func namaKolomBun(field reflect.StructField, tag string, punyaTag bool) (nama string, lewati bool) {
	if !punyaTag {
		nama = strings.Split(field.Tag.Get("json"), ",")[0]
		if nama == "-" {
			return "", true
		}
		if nama == "" {
			nama = field.Name
		}
		return nama, false
	}

	if tag == "-" {
		return "", true
	}

	bagian := strings.Split(tag, ",")
	nama = strings.TrimSpace(bagian[0])

	if penandaStruktur(nama) {
		return "", true
	}

	for _, opsi := range bagian[1:] {
		opsi = strings.TrimSpace(opsi)
		switch {
		case opsi == "pk", opsi == "soft_delete", opsi == "scanonly", opsi == "skipupdate", opsi == "extend":
			return "", true
		case penandaStruktur(opsi):
			return "", true
		}
	}

	if nama == "" {
		nama = field.Name
	}

	return nama, false
}

func penandaStruktur(s string) bool {
	for _, awalan := range []string{"rel:", "table:", "alias:", "embed:", "select:", "polymorphic:"} {
		if strings.HasPrefix(s, awalan) {
			return true
		}
	}
	return false
}

func structToMapJSON(thestruct any) map[string]any {
	data, err := json.Marshal(thestruct)
	if err != nil {
		return nil
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err == nil {
		return result
	}

	return nil
}
