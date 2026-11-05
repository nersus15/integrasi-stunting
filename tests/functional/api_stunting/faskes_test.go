package api_stunting_test

import "github.com/nersus15/integrasi/tests/testutil"

var (
	tidakDisimpanKode, tidakDisimpanNama = 6001, "TIDAK_DISIMPAN"
	forbiddenKode, forbiddenNama         = 1008, "FORBIDDEN"
	notFoundKode, notFoundNama           = 2001, "NOT_FOUND"
)

const (
	nikAnakFaskes = "3515019909090001"
	encFaskes     = "enc-uji-faskes-01"
	obsFaskes     = "obs-uji-faskes-01"
	condFaskes    = "cond-uji-faskes-01"
)

func pemeriksaanFaskes() peta {
	return peta{
		"anak": peta{
			"nik":           nikAnakFaskes,
			"nama":          "Anak Faskes",
			"tanggal_lahir": "2024-03-01",
			"jenis_kelamin": "P",
		},
		"kunjungan": peta{
			"id_satusehat":       encFaskes,
			"tanggal_pengukuran": "2026-06-10",
			"berat_badan":        7.1,
			"tinggi_badan":       70.0,
		},
		"observasi": []any{
			peta{"id_satusehat": obsFaskes, "system": "http://loinc.org", "kode": "29463-7",
				"display": "Body weight", "nilai_angka": 7.1, "satuan": "kg", "interpretasi": "L"},
		},
		"diagnosa": []any{
			peta{"id_satusehat": condFaskes, "jenis": "diagnosis",
				"system": "http://hl7.org/fhir/sid/icd-10", "kode": "E45",
				"display": "Nutritional stunting"},
		},
	}
}

func skenarioFaskes(tambah func(skenario)) {
	F := "U. Jalur faskes"

	tambah(skenario{bagian: F, label: "U1 kunci jakantro ditolak di endpoint faskes",
		metode: "POST", path: "/faskes/pemeriksaan", payload: pemeriksaanFaskes(),
		http: 403, errorCode: forbiddenKode, errorName: forbiddenNama,
		pesan: "hanya untuk faskes"})

	tambah(skenario{bagian: F, label: "U2 tanpa kredensial", metode: "POST",
		path: "/faskes/pemeriksaan", payload: pemeriksaanFaskes(), kunci: tanpaKunci,
		http: 401})

	tambah(skenario{bagian: F, label: "U3 pemeriksaan tersimpan", metode: "POST",
		path: "/faskes/pemeriksaan", payload: pemeriksaanFaskes(), kunci: testutil.KeyFaskesA,
		http: 201})

	tambah(skenario{bagian: F, label: "U4 kiriman ulang tetap 201", metode: "POST",
		path: "/faskes/pemeriksaan", payload: pemeriksaanFaskes(), kunci: testutil.KeyFaskesA,
		http: 201})

	tambah(skenario{bagian: F, label: "U5 anak tanpa nik dan tanpa satusehat id",
		metode: "POST", path: "/faskes/pemeriksaan", kunci: testutil.KeyFaskesA,
		payload: peta{"anak": peta{}, "kunjungan": peta{"id_satusehat": "x",
			"tanggal_pengukuran": "2026-06-10"}},
		http: 422, errorCode: valKode, errorName: valNama,
		pesan: "anak.satusehat_id atau anak.nik"})

	tambah(skenario{bagian: F, label: "U6 kunjungan tanpa id_satusehat",
		metode: "POST", path: "/faskes/pemeriksaan", kunci: testutil.KeyFaskesA,
		payload: peta{"anak": peta{"nik": nikAnakFaskes},
			"kunjungan": peta{"tanggal_pengukuran": "2026-06-10"}},
		http: 422, errorCode: valKode, errorName: valNama,
		pesan: "kunjungan.id_satusehat"})

	tambah(skenario{bagian: F, label: "U7 tanggal salah format",
		metode: "POST", path: "/faskes/pemeriksaan", kunci: testutil.KeyFaskesA,
		payload: peta{"anak": peta{"nik": nikAnakFaskes},
			"kunjungan": peta{"id_satusehat": "x", "tanggal_pengukuran": "10-06-2026"}},
		http: 422, errorCode: valKode, errorName: valNama, pesan: "YYYY-MM-DD"})

	tambah(skenario{bagian: F, label: "U8 anak lewat 5 tahun tidak disimpan",
		metode: "POST", path: "/faskes/pemeriksaan", kunci: testutil.KeyFaskesA,
		payload: peta{
			"anak":      peta{"nik": "3515019909090099", "tanggal_lahir": "1990-01-01"},
			"kunjungan": peta{"id_satusehat": "enc-uji-tua-01", "tanggal_pengukuran": "2026-06-10"}},
		http: 422, errorCode: tidakDisimpanKode, errorName: tidakDisimpanNama})

	tambah(skenario{bagian: F, label: "U9 update diagnosa", metode: "PUT",
		path: "/faskes/diagnosa", kunci: testutil.KeyFaskesA,
		payload: peta{"id_satusehat": condFaskes, "kode": "E45",
			"display": "Nutritional stunting (revisi)"},
		http: 200})

	tambah(skenario{bagian: F, label: "U10 update observasi", metode: "PUT",
		path: "/faskes/observasi", kunci: testutil.KeyFaskesA,
		payload: peta{"id_satusehat": obsFaskes, "system": "http://loinc.org",
			"kode": "29463-7", "nilai_angka": 7.9, "satuan": "kg"},
		http: 200})

	tambah(skenario{bagian: F, label: "U11 update tanpa id_satusehat ditolak",
		metode: "PUT", path: "/faskes/diagnosa", kunci: testutil.KeyFaskesA,
		payload: peta{"kode": "E45"},
		http:    422, errorCode: valKode, errorName: valNama, pesan: "id_satusehat"})

	tambah(skenario{bagian: F, label: "U12 resource tidak dikenal", metode: "PUT",
		path: "/faskes/hantu", kunci: testutil.KeyFaskesA,
		payload: peta{"id_satusehat": "x"},
		http:    404, errorCode: notFoundKode, errorName: notFoundNama})

	tambah(skenario{bagian: F, label: "U13 update resource yang belum tersimpan",
		metode: "PUT", path: "/faskes/diagnosa", kunci: testutil.KeyFaskesA,
		payload: peta{"id_satusehat": "cond-tidak-pernah-ada", "kode": "E45"},
		http:    404, errorCode: notFoundKode, errorName: notFoundNama})

	V := "V. Antrean pesan kafka gagal"

	tambah(skenario{bagian: V, label: "V1 daftar antrean", metode: "GET",
		path: "/kafka/gagal", http: 200})

	tambah(skenario{bagian: V, label: "V2 filter per patient_id", metode: "GET",
		path: "/kafka/gagal?patient_id=P-tidak-ada", http: 200})

	tambah(skenario{bagian: V, label: "V3 retry pesan yang tidak ada", metode: "POST",
		path: "/kafka/gagal/tidak-ada/retry", http: 404,
		errorCode: notFoundKode, errorName: notFoundNama})

	tambah(skenario{bagian: V, label: "V4 retry semua", metode: "POST",
		path: "/kafka/gagal/retry", http: 200})
}
