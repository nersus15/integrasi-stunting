package exceptions

import (
	"context"
	databasesql "database/sql"
	"errors"
	"fmt"

	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Kind adalah cetakan satu macam kesalahan.
type Kind struct {
	HttpCode  int
	ErrorCode int
	Name      string
	Message   string
}

// Error adalah kesalahan siap kirim. Sebab aslinya disimpan tapi tidak ikut ke
// klien: Error() hanya mengembalikan Message.
type Error struct {
	HttpCode  int
	ErrorCode int
	Name      string
	Message   string

	cause error
}

func (e *Error) Error() string { return e.Message }

// Unwrap menjaga errors.Is dan errors.As tetap bekerja terhadap sebab aslinya.
func (e *Error) Unwrap() error { return e.cause }

// Cause mengembalikan error asli untuk keperluan log.
func (e *Error) Cause() error { return e.cause }

// New memakai pesan bawaan Kind.
func (k Kind) New(cause error) *Error {
	return &Error{k.HttpCode, k.ErrorCode, k.Name, k.Message, cause}
}

// WithMessage mengganti pesan bawaan, misalnya untuk menyebut field yang gagal.
func (k Kind) WithMessage(message string, cause error) *Error {
	if !utils.IsStrFilled(message) {
		message = k.Message
	}
	return &Error{k.HttpCode, k.ErrorCode, k.Name, message, cause}
}

// Messagef seperti WithMessage tapi memformat pesannya, tanpa sebab.
func (k Kind) Messagef(format string, args ...any) *Error {
	return &Error{k.HttpCode, k.ErrorCode, k.Name, fmt.Sprintf(format, args...), nil}
}

// 1xxx — payload perlu diperbaiki
var (
	BodyRusak         = Kind{400, 1001, "BODY_INVALID", "Body tidak bisa dibaca sebagai JSON"}
	Validasi          = Kind{400, 1002, "VALIDATION_ERROR", "Data tidak memenuhi aturan"}
	BentukPayload     = Kind{400, 1003, "PAYLOAD_SHAPE_INVALID", "Bentuk payload tidak dikenali"}
	KolomWajib        = Kind{400, 1004, "REQUIRED_FIELD_MISSING", "Kolom wajib tidak boleh kosong"}
	ReferensiHilang   = Kind{400, 1005, "REFERENCE_NOT_FOUND", "Referensi tidak ditemukan"}
	Constraint        = Kind{400, 1006, "CONSTRAINT_VIOLATION", "Data tidak memenuhi aturan database"}
	ParameterRequired = Kind{400, 1007, "PARAM_REQUIRED", "Parameter harus dikirim"}
	Forbidden         = Kind{403, 1008, "FORBIDDEN", "User tidak memiliki akses"}
	Unauthorized      = Kind{401, 1009, "UNAUTHORIZED", "Authorization required"}
)

// 2xxx — tidak ditemukan
var (
	TidakDitemukan = Kind{404, 2001, "NOT_FOUND", "Data tidak ditemukan"}
)

// 3xxx — bentrok dengan data tersimpan
var (
	DuplikatId          = Kind{409, 3001, "DUPLICATE_ID", "Id sudah dipakai"}
	DuplikatNikOrangtua = Kind{409, 3002, "DUPLICATE_NIK_ORANGTUA", "nik orangtua sudah terdaftar"}
	DuplikatNoKK        = Kind{409, 3003, "DUPLICATE_NO_KK", "no_kk sudah terdaftar"}
	DuplikatNikAnak     = Kind{409, 3004, "DUPLICATE_NIK_ANAK", "nik anak sudah terdaftar"}
	Duplikat            = Kind{409, 3005, "DUPLICATE_DATA", "Data sudah ada"}

	AnakMilikOrangtuaLain = Kind{409, 3006, "ANAK_TERDAFTAR_DI_ORANGTUA_LAIN",
		"nik anak sudah terdaftar pada orangtua yang berbeda"}
)

// 6xxx — payload sah tapi sengaja tidak disimpan
var (
	TidakDisimpan = Kind{422, 6001, "TIDAK_DISIMPAN", "Data tidak memenuhi kriteria pemantauan stunting"}
)

// 5xxx — masalah di sisi layanan
var (
	Internal      = Kind{500, 5001, "INTERNAL_ERROR", "Terjadi kesalahan"}
	TidakTersedia = Kind{503, 5002, "SERVICE_UNAVAILABLE", "Layanan sedang sibuk, silakan coba lagi"}
)

// Classify menerjemahkan error apa pun menjadi *Error siap kirim.
// Error yang sudah bertipe *Error dikembalikan apa adanya.
func Classify(err error) *Error {
	if err == nil {
		return nil
	}

	var siap *Error
	if errors.As(err, &siap) {
		return siap
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return TidakTersedia.New(err)
	}

	// jaring pengaman supaya ErrNoRows tetap 404, bukan 500
	if errors.Is(err, utils.ErrTidakDitemukan) || errors.Is(err, databasesql.ErrNoRows) {
		return TidakDitemukan.New(err)
	}

	var pg pgdriver.Error
	if !errors.As(err, &pg) {
		return Internal.New(err)
	}

	if pg.StatementTimeout() {
		return TidakTersedia.New(err)
	}

	switch pg.Field('C') {
	case "23505": // unique_violation, mencakup primary key
		switch pg.Field('n') {
		case "pk_orangtua":
			return DuplikatId.WithMessage("id orangtua sudah dipakai", err)
		case "pk_anak":
			return DuplikatId.WithMessage("id anak sudah dipakai", err)
		case "pk_kunjungan":
			return DuplikatId.WithMessage("id kunjungan sudah dipakai", err)
		case "pk_kesehatan":
			return DuplikatId.WithMessage("id kesehatan sudah dipakai", err)
		case "uq_orangtua_nik":
			return DuplikatNikOrangtua.New(err)
		case "uq_orangtua_no_kk":
			return DuplikatNoKK.New(err)
		case "uq_anak_nik":
			return DuplikatNikAnak.New(err)
		}
		return Duplikat.New(err)

	case "23503": // foreign_key_violation
		return ReferensiHilang.New(err)

	case "23502": // not_null_violation
		return KolomWajib.WithMessage("kolom wajib tidak boleh kosong: "+pg.Field('c'), err)
	}

	if pg.IntegrityViolation() {
		return Constraint.New(err)
	}

	return Internal.New(err)
}
