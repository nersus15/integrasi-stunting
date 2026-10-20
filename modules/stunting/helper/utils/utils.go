package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"
)

var dumpAktif atomic.Bool

func SetDumpAktif(level string) {
	dumpAktif.Store(strings.EqualFold(strings.TrimSpace(level), "debug"))
}

func DumpAktif() bool { return dumpAktif.Load() }

func IsFilled(s *string) bool {
	return s != nil && strings.TrimSpace(*s) != ""
}

func IsStrFilled(s string) bool {
	return strings.TrimSpace(s) != ""
}

func IsDigitsLen(s string, n int) bool {
	s = strings.TrimSpace(s)
	if len(s) != n {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func IsValidDate(d string) bool {
	_, err := time.Parse("2006-01-02", d)
	return err == nil
}

// Nilai membaca isi pointer string dengan aman untuk keperluan log.
func Nilai(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func StrPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func PtrStr(s string, emptyToNil bool) *string {
	if s == "" && emptyToNil {
		return nil
	} else {
		return &s
	}
}

func Substr(s string, start, length int) string {
	runes := []rune(s)

	if start < 0 || start >= len(runes) {
		return ""
	}

	end := start + length
	if end > len(runes) {
		end = len(runes)
	}

	return string(runes[start:end])
}

// null berarti statusnya belum ditentukan, bukan bukan-stunting
func SmallintToBool(num *int) bool {
	return num != nil && *num == 1
}

func BoolToSmallint(v bool) int {
	if v {
		return 1
	}

	return 0
}

func BoolToSmallintPtr(v bool) *int {
	n := BoolToSmallint(v)
	return &n
}
func Ucfirst(s string) string {
	if s == "" {
		return s
	}

	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size == 1 {
		return s
	}

	upper := unicode.ToUpper(r)
	if upper == r {
		return s
	}

	return string(upper) + s[size:]
}

func FromAnyToInt(v any) (int, error) {
	return FromAny[int](v)
}

func FromAnyToI8(v any) (int8, error) {
	return FromAny[int8](v)
}

func FromAnyToI16(v any) (int16, error) {
	return FromAny[int16](v)
}

func FromAnyToI32(v any) (int32, error) {
	return FromAny[int32](v)
}

func FromAnyToI64(v any) (int64, error) {
	return FromAny[int64](v)
}

func FromAnyToUint(v any) (uint, error) {
	return FromAny[uint](v)
}

func FromAnyToU8(v any) (uint8, error) {
	return FromAny[uint8](v)
}

func FromAnyToU16(v any) (uint16, error) {
	return FromAny[uint16](v)
}

func FromAnyToU32(v any) (uint32, error) {
	return FromAny[uint32](v)
}

func FromAnyToU64(v any) (uint64, error) {
	return FromAny[uint64](v)
}

func FromAnyToUintptr(v any) (uintptr, error) {
	return FromAny[uintptr](v)
}

func FromAnyToF32(v any) (float32, error) {
	return FromAny[float32](v)
}

func FromAnyToF64(v any) (float64, error) {
	return FromAny[float64](v)
}

func FromAnyToComplex64(v any) (complex64, error) {
	return FromAny[complex64](v)
}

func FromAnyToComplex128(v any) (complex128, error) {
	return FromAny[complex128](v)
}

func FromAnyToBool(v any) (bool, error) {
	return FromAny[bool](v)
}

func FromAnyToString(v any) (string, error) {
	return FromAny[string](v)
}

func FromAny[T any](src any) (T, error) {
	var zero T

	if src == nil {
		return zero, fmt.Errorf("cannot convert nil")
	}

	dst := reflect.TypeFor[T]()
	v := reflect.ValueOf(src)

	// Nilai sudah cocok dengan tipe tujuan.
	if v.Type().AssignableTo(dst) {
		return v.Interface().(T), nil
	}

	// Konversi string ke tipe tujuan.
	if s, ok := src.(string); ok {
		out := reflect.New(dst).Elem()

		switch dst.Kind() {
		case reflect.String:
			out.SetString(s)
			return out.Interface().(T), nil

		case reflect.Bool:
			n, err := strconv.ParseBool(s)
			if err != nil {
				return zero, err
			}
			out.SetBool(n)
			return out.Interface().(T), nil

		case reflect.Int, reflect.Int8, reflect.Int16,
			reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(s, 0, dst.Bits())
			if err != nil {
				// Coba parsing desimal, misalnya "08".
				n, err = strconv.ParseInt(s, 10, dst.Bits())
			}
			if err != nil {
				return zero, err
			}
			out.SetInt(n)
			return out.Interface().(T), nil

		case reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			n, err := strconv.ParseUint(s, 10, dst.Bits())
			if err != nil {
				return zero, err
			}
			out.SetUint(n)
			return out.Interface().(T), nil

		case reflect.Float32, reflect.Float64:
			n, err := strconv.ParseFloat(s, dst.Bits())
			if err != nil {
				return zero, err
			}
			out.SetFloat(n)
			return out.Interface().(T), nil
		}
	}

	// Konversi numerik dan konversi lain yang diizinkan Go.
	if v.Type().ConvertibleTo(dst) {
		out := v.Convert(dst)
		return out.Interface().(T), nil
	}

	return zero, fmt.Errorf(
		"cannot convert %T to %v", src, dst,
	)
}
