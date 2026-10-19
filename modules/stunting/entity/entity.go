package entity

import (
	"reflect"
)

type Entity interface {
	TableName() string
	GetID() string
	GetPkName() string
}

func timpaField(tujuan any, new any, skipFields map[string]bool, force bool) {
	valExisting := reflect.ValueOf(tujuan).Elem()
	valNew := reflect.ValueOf(new)

	for i := 0; i < valNew.NumField(); i++ {
		fieldName := valNew.Type().Field(i).Name
		if skipFields[fieldName] {
			continue
		}

		fieldNew := valNew.Field(i)
		fieldExisting := valExisting.FieldByName(fieldName)

		if !fieldExisting.IsValid() || !fieldExisting.CanSet() {
			continue
		}

		if (!fieldNew.IsZero() && fieldNew.Interface() != fieldExisting.Interface()) || force {
			fieldExisting.Set(fieldNew)
		}
	}
}
