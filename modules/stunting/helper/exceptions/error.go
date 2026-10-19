package exceptions

import (
	"context"
	databasesql "database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/webcore-go/webcore/infra/logger"
)

var ErrTidakDitemukan = errors.New("data tidak ditemukan")

type Kind struct {
	HttpCode  int
	ErrorCode int
	Name      string
	Message   string
}

// Error() hanya berisi Message, sebab asli tidak ikut ke klien
type Error struct {
	HttpCode  int
	ErrorCode int
	Name      string
	Message   string

	cause error
}

func (e *Error) Error() string { return e.Message }

func (e *Error) Unwrap() error { return e.cause }

func (e *Error) Cause() error { return e.cause }

func (k Kind) New(cause error) *Error {
	return &Error{k.HttpCode, k.ErrorCode, k.Name, k.Message, cause}
}

func (k Kind) WithMessage(message string, cause error) *Error {
	if strings.TrimSpace(message) == "" {
		message = k.Message
	}
	return &Error{k.HttpCode, k.ErrorCode, k.Name, message, cause}
}

func (k Kind) Messagef(format string, args ...any) *Error {
	return &Error{k.HttpCode, k.ErrorCode, k.Name, fmt.Sprintf(format, args...), nil}
}

// 1xxx — payload perlu diperbaiki
var (
	BodyRusak         = Kind{400, 1001, "BODY_INVALID", "Body tidak bisa dibaca sebagai JSON"}
	Validasi          = Kind{422, 1002, "VALIDATION_ERROR", "Data tidak memenuhi aturan"}
	BentukPayload     = Kind{422, 1003, "PAYLOAD_SHAPE_INVALID", "Bentuk payload tidak dikenali"}
	KolomWajib        = Kind{422, 1004, "REQUIRED_FIELD_MISSING", "Kolom wajib tidak boleh kosong"}
	ReferensiHilang   = Kind{422, 1005, "REFERENCE_NOT_FOUND", "Referensi tidak ditemukan"}
	Constraint        = Kind{422, 1006, "CONSTRAINT_VIOLATION", "Data tidak memenuhi aturan database"}
	ParameterRequired = Kind{422, 1007, "PARAM_REQUIRED", "Parameter harus dikirim"}
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

// penerjemah galat dari adapter, mis. driver database
var penerjemah []func(error) *Error

func DaftarkanPenerjemah(f func(error) *Error) { penerjemah = append(penerjemah, f) }

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
	if errors.Is(err, ErrTidakDitemukan) || errors.Is(err, databasesql.ErrNoRows) {
		return TidakDitemukan.New(err)
	}

	for _, terjemah := range penerjemah {
		if e := terjemah(err); e != nil {
			return e
		}
	}

	return Internal.New(err)
}

func AbaikanJikaBelumTersimpan(err error) error {
	if err == nil {
		return nil
	}

	var siap *Error
	if errors.As(err, &siap) && siap.ErrorCode == TidakDitemukan.ErrorCode {
		logger.Info("ProsesTransaksiFHIR:Update => resource tidak ada di database, dilewati")
		return nil
	}

	return err
}
