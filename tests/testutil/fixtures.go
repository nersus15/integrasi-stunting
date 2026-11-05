package testutil

import (
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/helper/types"
)

const (
	IdOrangtua  = "O1000000-0000-0000-0000-00000000001"
	IdAnak      = "A1000000-0000-0000-0000-00000000001"
	IdKunjungan = "K1000000-0000-0000-0000-00000000001"
	IdPosyandu  = "P1000000-0000-0000-0000-00000000001"

	NikOrangtua = "3515081203800002"
	NikAnak     = "3515081203230001"
	NoKK        = "3515081203000001"
	SourceData  = "20f11c5c-5a29-11f0-a136-5749ce66d036"
)

func OrangtuaPayloadValid() types.OrangtuaPayload {
	return types.OrangtuaPayload{
		Id:         IdOrangtua,
		IdPosyandu: IdPosyandu,
		NoKk:       NoKK,
		NamaAyah:   "Budi Santoso",
		NamaIbu:    "Siti Aminah",
		Nik:        NikOrangtua,
		Telepon:    "081234567890",
		Rt:         "003",
		Rw:         "002",
		Alamat:     "Jl. Mawar No. 12",
		Kia:        Ptr(int16(1)),
		SourceData: Ptr(SourceData),
	}
}

var NikAnakPtr = NikAnak

func AnakPayloadValid() types.AnakPayload {
	return types.AnakPayload{
		Id:           IdAnak,
		IDOrangtua:   IdOrangtua,
		Nama:         "Ananda Ahmad Santoso",
		NIK:          &NikAnakPtr,
		TanggalLahir: "2024-05-10",
		JenisKelamin: "L",
		AnakKe:       1,
		IMD:          Ptr(int16(1)),
		BBLahir:      3.2,
		TBLahir:      49.5,
		LKLahir:      33.0,
		SourceData:   SourceData,
	}
}

func KunjunganPayloadValid() types.KunjunganPayload {
	return types.KunjunganPayload{
		Id:                IdKunjungan,
		IDAnak:            IdAnak,
		TanggalPengukuran: "2025-02-16",
		CaraUkur:          Ptr("Telentang"),
		BeratBadan:        Ptr(8.5),
		TinggiBadan:       Ptr(68.2),
		LingkarLengan:     Ptr(13.5),
		LingkarKepala:     Ptr(42.0),
		LingkarDada:       Ptr(43.0),
		ASIBulan0:         Ptr(int16(1)),
		VitBiru:           Ptr(int16(1)),
		VitMerah:          Ptr(int16(0)),
		SourceData:        Ptr(SourceData),
	}
}

func KesehatanPayloadValid() types.KesehatanPayload {
	return types.KesehatanPayload{
		Id:                "H1000000-0000-0000-0000-00000000001",
		IDAnak:            IdAnak,
		TanggalPemantauan: "2025-02-16",
		TBCBatuk:          Ptr(int16(0)),
		LayananASIEks:     Ptr(int16(1)),
		SourceData:        Ptr(SourceData),
	}
}

// --- body JSON untuk tujuh bentuk payload yang sah ---

type BodyOpts struct {
	Orangtua  any
	Anak      any
	Kunjungan any
}

func BodyNested(t testing.TB, o BodyOpts) []byte {
	t.Helper()

	body := map[string]any{}
	if o.Orangtua != nil {
		body["orangtua"] = o.Orangtua
	}
	if o.Anak != nil {
		body["anak"] = o.Anak
	}
	if o.Kunjungan != nil {
		body["kunjungan"] = o.Kunjungan
	}
	return JSONBody(t, body)
}

// BodyRegistrasiLengkap: orangtua + anak sama-sama membawa data.
func BodyRegistrasiLengkap(t testing.TB) []byte {
	t.Helper()
	return BodyNested(t, BodyOpts{
		Orangtua:  OrangtuaPayloadValid(),
		Anak:      AnakPayloadValid(),
		Kunjungan: KunjunganPayloadValid(),
	})
}

// BodyAnakBaru3Key: orangtua hanya membawa id.
func BodyAnakBaru3Key(t testing.TB) []byte {
	t.Helper()
	return BodyNested(t, BodyOpts{
		Orangtua:  map[string]any{"id": IdOrangtua},
		Anak:      AnakPayloadValid(),
		Kunjungan: KunjunganPayloadValid(),
	})
}

// BodyAnakBaru2Key: tanpa key orangtua.
func BodyAnakBaru2Key(t testing.TB) []byte {
	t.Helper()
	return BodyNested(t, BodyOpts{
		Anak:      AnakPayloadValid(),
		Kunjungan: KunjunganPayloadValid(),
	})
}

// BodyKunjunganSaja3Key: orangtua dan anak hanya membawa id.
func BodyKunjunganSaja3Key(t testing.TB) []byte {
	t.Helper()
	return BodyNested(t, BodyOpts{
		Orangtua:  map[string]any{"id": IdOrangtua},
		Anak:      map[string]any{"id": IdAnak, "id_orangtua": IdOrangtua},
		Kunjungan: KunjunganPayloadValid(),
	})
}

func BodyKunjunganSaja2Key(t testing.TB) []byte {
	t.Helper()
	return BodyNested(t, BodyOpts{
		Anak:      map[string]any{"id": IdAnak, "id_orangtua": IdOrangtua},
		Kunjungan: KunjunganPayloadValid(),
	})
}

func BodyKunjunganSaja1Key(t testing.TB) []byte {
	t.Helper()
	return BodyNested(t, BodyOpts{Kunjungan: KunjunganPayloadValid()})
}

// BodyKunjunganFlat: field kunjungan langsung di root.
func BodyKunjunganFlat(t testing.TB) []byte {
	t.Helper()
	return JSONBody(t, KunjunganPayloadValid())
}
