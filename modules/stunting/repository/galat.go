package repository

import (
	"errors"
	"strings"

	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/uptrace/bun/driver/pgdriver"
)

func init() { exceptions.DaftarkanPenerjemah(galatPostgres) }

// galat Postgres -> kontrak error; nil berarti bukan galat Postgres
func galatPostgres(err error) *exceptions.Error {
	var pg pgdriver.Error
	if !errors.As(err, &pg) {
		return nil
	}

	if pg.StatementTimeout() {
		return exceptions.TidakTersedia.New(err)
	}

	switch pg.Field('C') {
	case "23505": // unique_violation, mencakup primary key
		switch pg.Field('n') {
		case "pk_orangtua":
			return exceptions.DuplikatId.WithMessage("id orangtua sudah dipakai", err)
		case "pk_anak":
			return exceptions.DuplikatId.WithMessage("id anak sudah dipakai", err)
		case "pk_kunjungan":
			return exceptions.DuplikatId.WithMessage("id kunjungan sudah dipakai", err)
		case "pk_kesehatan":
			return exceptions.DuplikatId.WithMessage("id kesehatan sudah dipakai", err)
		case "uq_orangtua_nik":
			return exceptions.DuplikatNikOrangtua.New(err)
		case "uq_orangtua_no_kk":
			return exceptions.DuplikatNoKK.New(err)
		case "uq_anak_nik":
			return exceptions.DuplikatNikAnak.New(err)
		}
		return exceptions.Duplikat.New(err)

	case "23503": // foreign_key_violation
		return exceptions.ReferensiHilang.New(err)

	case "23502": // not_null_violation
		return exceptions.KolomWajib.WithMessage("kolom wajib tidak boleh kosong: "+pg.Field('c'), err)
	}

	if pg.IntegrityViolation() || strings.HasPrefix(pg.Field('C'), "22") {
		return exceptions.Constraint.New(err)
	}

	return exceptions.Internal.New(err)
}
