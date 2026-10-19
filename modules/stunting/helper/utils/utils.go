package utils

import (
	"strings"
	"sync/atomic"
	"time"
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
