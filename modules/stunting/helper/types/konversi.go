package types

import (
	"time"
)

// kebalikan waktuDariString, supaya bentuk JSON-nya sama dengan yang masuk
func teksWaktu(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func teksTanggal(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

// waktuDariString menerima dateTime FHIR maupun tanggal saja.
func waktuDariString(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, *s); err == nil {
			return &t
		}
	}
	return nil
}

func nilaiInt16(v *int16) int16 {
	if v == nil {
		return 0
	}
	return *v
}

// daftar mengubah slice entity menjadi slice DTO; hasil nil dilewati
func daftar[E any, T any](list []*E, ubah func(*E) *T) []T {
	out := make([]T, 0, len(list))
	for _, e := range list {
		if t := ubah(e); t != nil {
			out = append(out, *t)
		}
	}
	return out
}

// belumBerkunjung menyaring turunan yang belum tertempel ke kunjungan
func belumBerkunjung[T any](list []T, idKunjungan func(T) *string) []T {
	out := make([]T, 0)
	for _, v := range list {
		if idKunjungan(v) == nil {
			out = append(out, v)
		}
	}
	return out
}
