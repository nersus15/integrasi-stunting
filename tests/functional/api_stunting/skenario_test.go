package api_stunting_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
)

// uid membentuk id 36 karakter yang mudah dikenali di tabel database.
func uid(awalan string) string {
	p := (awalan + "00000000")[:8]
	s := (awalan + "000000000000")[:12]
	return p + "-1111-1111-1111-" + s
}

type peta = map[string]any

var idPosyandu = "pos-01"

// ubah menyalin peta lalu mengganti satu kunci, supaya varian negatif tidak
// merusak payload dasar yang dipakai skenario lain.
func modify(m peta, key string, value any) peta {
	n := peta{}
	for k, v := range m {
		n[k] = v
	}
	n[key] = value
	return n
}

var (
	o1, o2, o3                     = uid("ortu1"), uid("ortu2"), uid("ortu3")
	a1, a2, a3, a4, a5, a6, a7     = uid("anakA"), uid("anakB"), uid("anakC"), uid("anakD"), uid("anakE"), uid("anakF"), uid("anakG")
	nikO1, nokkO1                  = "3515010101010001", "3515010101010099"
	nikA1                          = "3515010101010002"
	valKode, valNama               = 1002, "VALIDATION_ERROR"
	bentukKode, bentukNama         = 1003, "PAYLOAD_SHAPE_INVALID"
	bodyKode, bodyNama             = 1001, "BODY_INVALID"
	referensiKode, referensiNama   = 1005, "REFERENCE_NOT_FOUND"
	duplikatIdKode, duplikatIdNama = 3001, "DUPLICATE_ID"
	noKKKode, noKKNama             = 3003, "DUPLICATE_NO_KK"
)

func ortu(id, nik, nokk string) peta {
	return peta{
		"id": id, "id_posyandu": idPosyandu, "no_kk": nokk, "nama_ayah": "Budi",
		"nama_ibu": "Siti", "nik": nik, "telepon": "08123", "rt": "001",
		"rw": "002", "alamat": "Jl. Mawar", "kia": 1,
	}
}

func anak(id, idOrtu, nik string, ke int) peta {
	name := "Anak"
	if len(id) >= 5 {
		name = "Anak " + id[:5]
	}
	m := peta{
		"id": id, "id_orangtua": idOrtu, "nama": name,
		"tanggal_lahir": "2024-03-15", "jenis_kelamin": "L", "anak_ke": ke,
		"imd": 1, "bb_lahir": 3.2, "tb_lahir": 49.5, "lk_lahir": 34.0,
		"source_data": "jakantro",
	}
	if nik != "-" { // "-" berarti key nik sengaja dihilangkan
		m["nik"] = nik
	}
	return m
}

func kunj(id, idAnak, tgl string) peta {
	return peta{"id": id, "id_anak": idAnak, "tanggal_pengukuran": tgl, "berat_badan": 9.4}
}

func kes(id, idAnak, tgl string) peta {
	return peta{"id": id, "id_anak": idAnak, "tanggal_pemantauan": tgl,
		"tbc_batuk": 0, "layanan_asi_eks": 1}
}

const K = "/kunjungan"

func allScenarios() []skenario {
	var s []skenario
	tambah := func(x skenario) { s = append(s, x) }

	sah := func(label string, p any) {
		tambah(skenario{bagian: "A. Bentuk payload sah", label: label,
			metode: "POST", path: K, payload: p, http: 201})
	}
	sah("A1 REGISTRASI_LENGKAP, 3 key orangtua bawa data",
		peta{"orangtua": ortu(o1, nikO1, nokkO1), "anak": anak(a1, o1, nikA1, 1), "kunjungan": kunj(uid("k01"), a1, "2026-08-01")})
	sah("A2 ANAK_BARU, 3 key orangtua {id} saja",
		peta{"orangtua": peta{"id": o1}, "anak": anak(a2, o1, "3515010101010003", 2), "kunjungan": kunj(uid("k02"), a2, "2026-08-02")})
	sah("A3 ANAK_BARU, 2 key",
		peta{"anak": anak(a3, o1, "3515010101010004", 3), "kunjungan": kunj(uid("k03"), a3, "2026-08-03")})
	sah("A4 KUNJUNGAN_SAJA, 3 key id saja",
		peta{"orangtua": peta{"id": o1}, "anak": peta{"id": a1, "id_orangtua": o1}, "kunjungan": kunj(uid("k04"), a1, "2026-08-04")})
	sah("A5 KUNJUNGAN_SAJA, 2 key",
		peta{"anak": peta{"id": a1, "id_orangtua": o1}, "kunjungan": kunj(uid("k05"), a1, "2026-08-05")})
	sah("A6 KUNJUNGAN_SAJA, 1 key nested",
		peta{"kunjungan": kunj(uid("k06"), a1, "2026-08-06")})
	sah("A7 KUNJUNGAN_SAJA, flat di root", kunj(uid("k07"), a1, "2026-08-07"))

	opsional := func(label string, p any) {
		tambah(skenario{bagian: "B. NIK anak opsional", label: label,
			metode: "POST", path: K, payload: p, http: 201})
	}
	opsional("B1 nik dikirim string kosong",
		peta{"anak": anak(a4, o1, "", 4), "kunjungan": kunj(uid("k08"), a4, "2026-08-08")})
	opsional("B2 anak kedua, nik string kosong",
		peta{"anak": anak(a5, o1, "", 5), "kunjungan": kunj(uid("k09"), a5, "2026-08-09")})
	opsional("B3 key nik dihilangkan",
		peta{"anak": anak(a6, o1, "-", 6), "kunjungan": kunj(uid("k10"), a6, "2026-08-10")})
	opsional("B4 nik berisi spasi",
		peta{"anak": anak(a7, o1, "   ", 7), "kunjungan": kunj(uid("k11"), a7, "2026-08-11")})

	bentuk := func(label string, p any, message string) {
		tambah(skenario{bagian: "C. Bentuk payload ditolak", label: label,
			metode: "POST", path: K, payload: p, http: 422,
			errorCode: bentukKode, errorName: bentukNama, pesan: message})
	}
	flatPlus := kunj(uid("c01"), a1, "2026-08-12")
	flatPlus["anak"] = anak(uid("ca1"), o1, nikA1, 11)
	bentuk("C1 flat disertai key nested", flatPlus, "bentuk flat hanya untuk kunjungan saja")
	bentuk("C2 nested tanpa key kunjungan",
		peta{"anak": anak(uid("ca2"), o1, "3515010101010011", 12)}, "data kunjungan wajib dikirim")
	bentuk("C3 kunjungan.id kosong",
		peta{"kunjungan": peta{"id": "", "id_anak": a1, "tanggal_pengukuran": "2026-08-12"}}, "kunjungan.id wajib dikirim")
	bentuk("C4 kunjungan.id_anak kosong",
		peta{"kunjungan": peta{"id": uid("c04"), "id_anak": "", "tanggal_pengukuran": "2026-08-12"}}, "kunjungan.id_anak wajib dikirim")
	bentuk("C5 kunjungan.tanggal_pengukuran kosong",
		peta{"kunjungan": peta{"id": uid("c05"), "id_anak": a1, "tanggal_pengukuran": ""}}, "tanggal_pengukuran wajib dikirim")
	bentuk("C6 tanggal_pengukuran format salah",
		peta{"kunjungan": kunj(uid("c06"), a1, "12-08-2026")}, "berformat YYYY-MM-DD")
	bentuk("C7 anak.id kosong",
		peta{"anak": anak("", o1, "3515010101010012", 13), "kunjungan": kunj(uid("c07"), a1, "2026-08-13")}, "anak.id wajib dikirim")
	bentuk("C8 anak.id_orangtua kosong",
		peta{"anak": anak(uid("ca8"), "", "3515010101010013", 14), "kunjungan": kunj(uid("c08"), uid("ca8"), "2026-08-14")}, "anak.id_orangtua wajib dikirim")
	bentuk("C9 anak.id != kunjungan.id_anak",
		peta{"anak": anak(uid("ca9"), o1, "3515010101010014", 15), "kunjungan": kunj(uid("c09"), a1, "2026-08-15")}, "tidak sama dengan kunjungan.id_anak")
	bentuk("C10 key orangtua tanpa key anak",
		peta{"orangtua": peta{"id": o1}, "kunjungan": kunj(uid("c10"), a1, "2026-08-16")}, "bersama key anak")
	bentuk("C11 orangtua.id kosong",
		peta{"orangtua": peta{"id": ""}, "anak": anak(uid("cb1"), o1, "3515010101010015", 16), "kunjungan": kunj(uid("c11"), uid("cb1"), "2026-08-17")}, "orangtua.id wajib dikirim")
	bentuk("C12 orangtua.id != anak.id_orangtua",
		peta{"orangtua": peta{"id": uid("beda")}, "anak": anak(uid("cb2"), o1, "3515010101010016", 17), "kunjungan": kunj(uid("c12"), uid("cb2"), "2026-08-18")}, "tidak sama dengan anak.id_orangtua")
	bentuk("C13 orangtua baru tanpa data anak baru",
		peta{"orangtua": ortu(uid("cb3"), "3515010101010017", "3515010101010018"),
			"anak":      peta{"id": uid("cb4"), "id_orangtua": uid("cb3")},
			"kunjungan": kunj(uid("c13"), uid("cb4"), "2026-08-19")}, "orangtua baru harus disertai data anak baru")

	// Validasi orangtua: setiap varian memakai nik dan no_kk unik supaya yang
	// gagal benar-benar field yang diuji, bukan bentrok data.
	nOrtu := 0
	ortuD := func(label string, ubahan peta, sk skenario) {
		nOrtu++
		t := fmt.Sprintf("%02d", nOrtu)
		id, idA := uid("d"+t), uid("da"+t)
		o := ortu(id, "35150102"+t+"000000", "35150103"+t+"000000")
		for k, v := range ubahan {
			o = modify(o, k, v)
		}
		sk.bagian, sk.label, sk.metode, sk.path = "D. Validasi orangtua", fmt.Sprintf("D%d %s", nOrtu, label), "POST", K
		sk.payload = peta{
			"orangtua":  o,
			"anak":      anak(idA, id, "35150104"+t+"000000", 1),
			"kunjungan": kunj(uid("dk"+t), idA, "2026-09-01")}
		tambah(sk)
	}
	valOrtu := func(label, key string, value any, message string) {
		ortuD(label, peta{key: value}, skenario{http: 422, errorCode: valKode, errorName: valNama, pesan: message})
	}
	valOrtu("id_posyandu kosong", "id_posyandu", "", "id_posyandu tidak boleh kosong")
	ortuD("no_kk kosong diterima", peta{"no_kk": ""}, skenario{http: 201})
	valOrtu("no_kk bukan 16 digit", "no_kk", "123", "no_kk harus 16 digit angka")
	ortuD("hanya nama_ibu diterima", peta{"nama_ayah": ""}, skenario{http: 201})
	ortuD("hanya nama_ayah diterima", peta{"nama_ibu": ""}, skenario{http: 201})
	ortuD("nama_ayah dan nama_ibu kosong", peta{"nama_ayah": "", "nama_ibu": ""},
		skenario{http: 422, errorCode: valKode, errorName: valNama, pesan: "nama_ayah atau nama_ibu harus diisi"})
	valOrtu("nik kosong", "nik", "", "nik tidak boleh kosong")
	valOrtu("nik bukan 16 digit", "nik", "abc", "nik harus 16 digit angka")
	valOrtu("telepon kosong", "telepon", "", "telepon tidak boleh kosong")
	valOrtu("rt kosong", "rt", "", "rt tidak boleh kosong")
	valOrtu("rw kosong", "rw", "", "rw tidak boleh kosong")
	valOrtu("alamat kosong", "alamat", "", "alamat tidak boleh kosong")
	valOrtu("kia bukan 0/1", "kia", 5, "kia harus bernilai 0 atau 1")
	valOrtu("usia_hamil negatif", "usia_hamil", -3, "usia_hamil tidak boleh bernilai negatif")
	valOrtu("kia_bayi_kecil bukan 0/1", "kia_bayi_kecil", 9, "kia_bayi_kecil harus bernilai 0 atau 1")

	nAnak := 0
	valAnak := func(label, key string, value any, message string) {
		nAnak++
		t := fmt.Sprintf("%02d", nAnak)
		id := uid("e" + t)
		tambah(skenario{bagian: "E. Validasi anak",
			label:  fmt.Sprintf("E%d %s", nAnak, label),
			metode: "POST", path: K,
			payload: peta{
				"anak":      modify(anak(id, o1, "35150105"+t+"000000", 20), key, value),
				"kunjungan": kunj(uid("ek"+t), id, "2026-09-02")},
			http: 422, errorCode: valKode, errorName: valNama, pesan: message})
	}
	valAnak("nama kosong", "nama", "", "nama tidak boleh kosong")
	valAnak("nik bukan 16 digit", "nik", "123", "nik harus 16 digit angka")
	valAnak("tanggal_lahir kosong", "tanggal_lahir", "", "tanggal_lahir tidak boleh kosong")
	valAnak("tanggal_lahir format salah", "tanggal_lahir", "15-03-2024", "berformat YYYY-MM-DD")
	valAnak("jenis_kelamin salah", "jenis_kelamin", "X", "jenis_kelamin harus")
	valAnak("anak_ke 0", "anak_ke", 0, "anak_ke harus lebih dari 0")
	valAnak("anak_ke negatif", "anak_ke", -2, "anak_ke harus lebih dari 0")
	valAnak("imd bukan 0/1", "imd", 7, "imd harus bernilai 0 atau 1")
	valAnak("bb_lahir 0", "bb_lahir", 0, "bb_lahir harus lebih dari 0")
	valAnak("tb_lahir negatif", "tb_lahir", -1, "tb_lahir harus lebih dari 0")
	valAnak("lk_lahir 0", "lk_lahir", 0, "lk_lahir harus lebih dari 0")
	valAnak("source_data kosong", "source_data", "", "source_data tidak boleh kosong")

	nKunj := 0
	valKunj := func(label, key string, value any, message string) {
		nKunj++
		t := fmt.Sprintf("%02d", nKunj)
		tambah(skenario{bagian: "F. Validasi kunjungan",
			label:  fmt.Sprintf("F%d %s", nKunj, label),
			metode: "POST", path: K,
			payload: modify(kunj(uid("f"+t), a1, "2026-09-03"), key, value),
			http:    422, errorCode: valKode, errorName: valNama, pesan: message})
	}
	valKunj("berat_badan 0", "berat_badan", 0, "berat_badan harus lebih dari 0")
	valKunj("tinggi_badan negatif", "tinggi_badan", -5, "tinggi_badan harus lebih dari 0")
	valKunj("lingkar_kepala 0", "lingkar_kepala", 0, "lingkar_kepala harus lebih dari 0")
	valKunj("lingkar_lengan negatif", "lingkar_lengan", -1, "lingkar_lengan harus lebih dari 0")
	valKunj("lingkar_dada 0", "lingkar_dada", 0, "lingkar_dada harus lebih dari 0")
	valKunj("asi_bulan_0 bukan 0/1", "asi_bulan_0", 5, "asi_bulan_0 harus bernilai 0 atau 1")
	valKunj("asi_bulan_6 bukan 0/1", "asi_bulan_6", 3, "asi_bulan_6 harus bernilai 0 atau 1")
	valKunj("vit_biru bukan 0/1", "vit_biru", 2, "vit_biru harus bernilai 0 atau 1")
	valKunj("vit_merah bukan 0/1", "vit_merah", 4, "vit_merah harus bernilai 0 atau 1")
	valKunj("pitting_edema bukan 0/1", "pitting_edema", 8, "pitting_edema harus bernilai 0 atau 1")
	valKunj("kelas_ibu_balita bukan 0/1", "kelas_ibu_balita", 6, "kelas_ibu_balita harus bernilai 0 atau 1")

	G := "G. Aturan NIK anak"
	tambah(skenario{bagian: G, label: "G1 nik terdaftar, id_orangtua SAMA", metode: "POST", path: K,
		payload: peta{"anak": modify(anak(uid("g01"), o1, nikA1, 30), "nama", "Nama Kiriman"),
			"kunjungan": kunj(uid("gk1"), uid("g01"), "2026-09-05")}, http: 201})
	tambah(skenario{bagian: G, label: "G2 siapkan orangtua kedua", metode: "POST", path: K,
		payload: peta{"orangtua": ortu(o2, "3515010101010777", "3515010101010888"),
			"anak":      anak(uid("g02"), o2, "3515010101010999", 1),
			"kunjungan": kunj(uid("gk2"), uid("g02"), "2026-09-06")}, http: 201})
	tambah(skenario{bagian: G, label: "G3 nik terdaftar, id_orangtua BEDA", metode: "POST", path: K,
		payload: peta{"anak": anak(uid("g03"), o2, nikA1, 31),
			"kunjungan": kunj(uid("gk3"), uid("g03"), "2026-09-07")},
		http: 409, errorCode: 3006, errorName: "ANAK_TERDAFTAR_DI_ORANGTUA_LAIN", pesan: "orangtua yang berbeda"})

	H := "H. Bentrok database"
	tambah(skenario{bagian: H, label: "H1 id kunjungan dikirim ulang", metode: "POST", path: K,
		payload: kunj(uid("k07"), a1, "2026-09-08"),
		http:    409, errorCode: duplikatIdKode, errorName: duplikatIdNama, pesan: "id kunjungan sudah dipakai"})
	tambah(skenario{bagian: H, label: "H2 no_kk sama, nik orangtua beda", metode: "POST", path: K,
		payload: peta{"orangtua": ortu(uid("h02"), "3515010101011111", nokkO1),
			"anak":      anak(uid("ha2"), uid("h02"), "3515010101012222", 1),
			"kunjungan": kunj(uid("hk2"), uid("ha2"), "2026-09-09")},
		http: 409, errorCode: noKKKode, errorName: noKKNama, pesan: "no_kk sudah terdaftar"})
	tambah(skenario{bagian: H, label: "H3 id_anak tidak ada di database", metode: "POST", path: K,
		payload: kunj(uid("h03"), uid("hantu"), "2026-09-10"),
		http:    422, errorCode: referensiKode, errorName: referensiNama})
	tambah(skenario{bagian: H, label: "H4 nik orangtua terdaftar, id payload beda", metode: "POST", path: K,
		payload: peta{"orangtua": ortu(uid("h04"), nikO1, nokkO1),
			"anak":      anak(uid("ha4"), uid("h04"), "3515010101013333", 40),
			"kunjungan": kunj(uid("hk4"), uid("ha4"), "2026-09-11")}, http: 201})

	I := "I. Body rusak"
	rusak := func(label string, body any, code int, name, message string) {
		http := 422
		if code == bodyKode {
			http = 400
		}
		tambah(skenario{bagian: I, label: label, metode: "POST", path: K,
			payload: body, http: http, errorCode: code, errorName: name, pesan: message})
	}
	rusak("I1 bukan JSON", "bukan json", bodyKode, bodyNama, "")
	rusak("I2 array JSON", "[]", bodyKode, bodyNama, "")
	rusak("I3 objek kosong", "{}", bentukKode, bentukNama, "data kunjungan wajib dikirim")
	rusak("I4 JSON terpotong", `{"kunjungan":{`, bodyKode, bodyNama, "")
	rusak("I5 body benar-benar kosong", "", bodyKode, bodyNama, "")

	J := "J. POST /api/orangtua"
	tambah(skenario{bagian: J, label: "J1 orangtua baru", metode: "POST", path: "/orangtua",
		payload: ortu(o3, "3515010101014444", "3515010101015555"), http: 201})
	tambah(skenario{bagian: J, label: "J2 kirim ulang, dipakai yang sudah ada", metode: "POST", path: "/orangtua",
		payload: ortu(o3, "3515010101014444", "3515010101015555"), http: 201})
	tambah(skenario{bagian: J, label: "J3 nik kosong", metode: "POST", path: "/orangtua",
		payload: ortu(uid("j03"), "", "3515010101016000"), http: 422,
		errorCode: valKode, errorName: valNama, pesan: "nik tidak boleh kosong"})
	tambah(skenario{bagian: J, label: "J4 no_kk kosong diterima", metode: "POST", path: "/orangtua",
		payload: ortu(uid("j04"), "3515010101016666", ""), http: 201})
	tambah(skenario{bagian: J, label: "J5 id kosong", metode: "POST", path: "/orangtua",
		payload: ortu("", "3515010101017777", "3515010101018888"), http: 422,
		errorCode: valKode, errorName: valNama, pesan: "id tidak boleh kosong"})
	tambah(skenario{bagian: J, label: "J6 body bukan JSON", metode: "POST", path: "/orangtua",
		payload: "bukan json", http: 400, errorCode: bodyKode, errorName: bodyNama})
	tambah(skenario{bagian: J, label: "J7 no_kk sudah dipakai keluarga lain", metode: "POST", path: "/orangtua",
		payload: ortu(uid("j07"), "3515010101019999", "3515010101015555"), http: 409,
		errorCode: noKKKode, errorName: noKKNama, pesan: "no_kk sudah terdaftar"})

	L := "K. POST /api/anak"
	tambah(skenario{bagian: L, label: "K1 anak baru", metode: "POST", path: "/anak",
		payload: anak(uid("k1a"), o1, "3515010101020001", 50), http: 201})
	tambah(skenario{bagian: L, label: "K2 id_orangtua kosong", metode: "POST", path: "/anak",
		payload: anak(uid("k2a"), "", "3515010101020002", 51), http: 422,
		errorCode: valKode, errorName: valNama, pesan: "id_orangtua tidak boleh kosong"})
	tambah(skenario{bagian: L, label: "K3 nik bukan 16 digit", metode: "POST", path: "/anak",
		payload: anak(uid("k3a"), o1, "123", 52), http: 422,
		errorCode: valKode, errorName: valNama, pesan: "nik harus 16 digit angka"})
	tambah(skenario{bagian: L, label: "K4 tanpa nik", metode: "POST", path: "/anak",
		payload: anak(uid("k4a"), o1, "", 53), http: 201})
	tambah(skenario{bagian: L, label: "K5 body bukan JSON", metode: "POST", path: "/anak",
		payload: "bukan json", http: 400, errorCode: bodyKode, errorName: bodyNama})
	tambah(skenario{bagian: L, label: "K6 id_orangtua tidak ada di database", metode: "POST", path: "/anak",
		payload: anak(uid("k6a"), uid("hantu"), "3515010101020003", 54), http: 422,
		errorCode: referensiKode, errorName: referensiNama})
	tambah(skenario{bagian: L, label: "K7 id anak kosong", metode: "POST", path: "/anak",
		payload: anak("", o1, "3515010101020004", 55), http: 422,
		errorCode: valKode, errorName: valNama, pesan: "id tidak boleh kosong"})
	tambah(skenario{bagian: L, label: "K8 nama kosong", metode: "POST", path: "/anak",
		payload: modify(anak(uid("k8a"), o1, "3515010101020005", 56), "nama", ""), http: 422,
		errorCode: valKode, errorName: valNama, pesan: "nama tidak boleh kosong"})

	M := "L. GET /api/orangtua"
	tambah(skenario{bagian: M, label: "L1 cari lewat nik yang ada", metode: "GET",
		path: "/orangtua?nik=" + nikO1, http: 200})
	tambah(skenario{bagian: M, label: "L2 cari lewat id yang ada", metode: "GET",
		path: "/orangtua/" + o1, http: 200})
	tambah(skenario{bagian: M, label: "L3 nik tidak ada", metode: "GET",
		path: "/orangtua?nik=9999999999999999", http: 404,
		errorCode: 2001, errorName: "NOT_FOUND", pesan: "tidak ditemukan"})
	tambah(skenario{bagian: M, label: "L4 id tidak ada", metode: "GET",
		path: "/orangtua/" + uid("hantu"), http: 404,
		errorCode: 2001, errorName: "NOT_FOUND"})
	tambah(skenario{bagian: M, label: "L5 tanpa kata kunci", metode: "GET",
		path: "/orangtua", http: 422, errorCode: valKode, errorName: valNama,
		pesan: "minimal satu kata kunci"})
	tambah(skenario{bagian: M, label: "L6 nik hanya spasi", metode: "GET",
		path: "/orangtua?nik=%20%20", http: 422, errorCode: valKode, errorName: valNama,
		pesan: "minimal satu kata kunci"})

	N := "M. GET /api/anak/:id"
	tambah(skenario{bagian: N, label: "M1 id anak yang ada", metode: "GET",
		path: "/anak/" + a1, http: 200})
	tambah(skenario{bagian: N, label: "M2 anak tanpa nik", metode: "GET",
		path: "/anak/" + a4, http: 200})
	tambah(skenario{bagian: N, label: "M3 id anak tidak ada", metode: "GET",
		path: "/anak/" + uid("hantu"), http: 404,
		errorCode: 2001, errorName: "NOT_FOUND", pesan: "tidak ditemukan"})
	tambah(skenario{bagian: N, label: "M4 tanpa id dan nik", metode: "GET",
		path: "/anak", http: 422, errorCode: valKode, errorName: valNama,
		pesan: "sebutkan minimal satu kata kunci (id | nik)"})
	tambah(skenario{bagian: N, label: "M5 cari lewat nik", metode: "GET",
		path: "/anak?nik=" + nikA1, http: 200})

	O := "N. GET /api/orangtua/:id/anak"
	tambah(skenario{bagian: O, label: "N1 orangtua yang punya anak", metode: "GET",
		path: "/orangtua/" + o1 + "/anak", http: 200})
	tambah(skenario{bagian: O, label: "N2 orangtua tidak ada", metode: "GET",
		path: "/orangtua/" + uid("hantu") + "/anak", http: 404,
		errorCode: 2001, errorName: "NOT_FOUND"})

	P := "O. GET /api/anak/:id/kunjungan"
	tambah(skenario{bagian: P, label: "O1 anak yang punya kunjungan", metode: "GET",
		path: "/anak/" + a1 + "/kunjungan", http: 200})
	tambah(skenario{bagian: P, label: "O2 anak tidak ada", metode: "GET",
		path: "/anak/" + uid("hantu") + "/kunjungan", http: 404,
		errorCode: 2001, errorName: "NOT_FOUND"})

	// ── R. POST /api/kesehatan ──────────────────────────────────────────
	// Aturannya sama persis dengan kunjungan, jadi yang diuji di sini adalah
	// bahwa kesamaan itu benar-benar berlaku -- bukan cuma diniatkan.
	const KS = "/kesehatan"
	oKS, aKS := uid("ksort"), uid("ksana")

	sahKS := func(label string, p any) {
		tambah(skenario{bagian: "R. POST /api/kesehatan", label: label,
			metode: "POST", path: KS, payload: p, http: 201})
	}
	sahKS("R1 REGISTRASI_LENGKAP",
		peta{"orangtua": ortu(oKS, "3599010101010001", "3599010101010002"),
			"anak":      anak(aKS, oKS, "3599010101010003", 1),
			"kesehatan": kes(uid("ks01"), aKS, "2026-09-01")})
	sahKS("R2 KESEHATAN_SAJA nested",
		peta{"kesehatan": kes(uid("ks02"), aKS, "2026-09-02")})
	sahKS("R3 KESEHATAN_SAJA flat di root", kes(uid("ks03"), aKS, "2026-09-03"))
	sahKS("R4 ANAK_BARU",
		peta{"anak": anak(uid("ksan2"), oKS, "3599010101010004", 2),
			"kesehatan": kes(uid("ks04"), uid("ksan2"), "2026-09-04")})

	bentukKS := func(label string, p any, message string) {
		tambah(skenario{bagian: "R. POST /api/kesehatan", label: label,
			metode: "POST", path: KS, payload: p, http: 422,
			errorCode: bentukKode, errorName: bentukNama, pesan: message})
	}
	flatKS := kes(uid("ks05"), aKS, "2026-09-05")
	flatKS["anak"] = anak(uid("ksan3"), oKS, "3599010101010005", 3)
	bentukKS("R5 flat disertai key nested", flatKS, "bentuk flat hanya untuk kesehatan saja")
	bentukKS("R6 nested tanpa key kesehatan",
		peta{"anak": anak(uid("ksan4"), oKS, "3599010101010006", 4)}, "data kesehatan wajib dikirim")
	bentukKS("R7 kesehatan.id kosong",
		peta{"kesehatan": peta{"id": "", "id_anak": aKS, "tanggal_pemantauan": "2026-09-06"}}, "kesehatan.id wajib dikirim")
	bentukKS("R8 tanggal_pemantauan format salah",
		peta{"kesehatan": kes(uid("ks06"), aKS, "06-09-2026")}, "berformat YYYY-MM-DD")
	bentukKS("R9 anak.id != kesehatan.id_anak",
		peta{"anak": anak(uid("ksan5"), oKS, "3599010101010007", 5),
			"kesehatan": kes(uid("ks07"), aKS, "2026-09-07")}, "tidak sama dengan kesehatan.id_anak")
	bentukKS("R10 key orangtua tanpa key anak",
		peta{"orangtua": peta{"id": oKS}, "kesehatan": kes(uid("ks08"), aKS, "2026-09-08")}, "bersama key anak")

	tambah(skenario{bagian: "R. POST /api/kesehatan", label: "R11 id_anak tidak ada",
		metode: "POST", path: KS, payload: kes(uid("ks09"), uid("hantu"), "2026-09-09"),
		http: 422, errorCode: referensiKode, errorName: referensiNama})
	tambah(skenario{bagian: "R. POST /api/kesehatan", label: "R12 id kesehatan dipakai ulang",
		metode: "POST", path: KS, payload: kes(uid("ks01"), aKS, "2026-09-10"),
		http: 409, errorCode: duplikatIdKode, errorName: duplikatIdNama})

	// ── S. GET kesehatan, kunjungan by id, dan summary ──────────────────
	S := "S. GET kesehatan & summary"
	tambah(skenario{bagian: S, label: "S1 riwayat kesehatan satu anak", metode: "GET",
		path: "/anak/" + aKS + "/kesehatan", http: 200})
	tambah(skenario{bagian: S, label: "S2 anak tidak ada", metode: "GET",
		path: "/anak/" + uid("hantu") + "/kesehatan", http: 404,
		errorCode: 2001, errorName: "NOT_FOUND"})
	tambah(skenario{bagian: S, label: "S3 satu kesehatan by id", metode: "GET",
		path: "/kesehatan/" + uid("ks02"), http: 200})
	tambah(skenario{bagian: S, label: "S4 kesehatan tidak ada", metode: "GET",
		path: "/kesehatan/" + uid("hantu"), http: 404,
		errorCode: 2001, errorName: "NOT_FOUND"})
	tambah(skenario{bagian: S, label: "S5 satu kunjungan by id", metode: "GET",
		path: "/kunjungan/" + uid("k01"), http: 200})
	tambah(skenario{bagian: S, label: "S6 kunjungan tidak ada", metode: "GET",
		path: "/kunjungan/" + uid("hantu"), http: 404,
		errorCode: 2001, errorName: "NOT_FOUND"})
	tambah(skenario{bagian: S, label: "S7 summary anak berisi dua riwayat", metode: "GET",
		path: "/anak/" + a1 + "/summary", http: 200})
	tambah(skenario{bagian: S, label: "S8 summary anak tidak ada", metode: "GET",
		path: "/anak/" + uid("hantu") + "/summary", http: 404,
		errorCode: 2001, errorName: "NOT_FOUND"})

	// ── Q. Autentikasi ──────────────────────────────────────────────────
	// Sejak auth aktif, setiap request wajib membawa kredensial. Skenario di
	// bawah menguji ketiga jawaban yang berbeda: kredensial bermasalah (401),
	// kredensial sah tapi hak kurang (403), dan kredensial sah dengan hak
	// cukup (lolos ke handler).
	Q := "Q. Autentikasi"
	authTolak := func(label, method, path string, payload any, key string, http, code int, name string) {
		tambah(skenario{bagian: Q, label: label, metode: method, path: path,
			payload: payload, kunci: key, http: http, errorCode: code, errorName: name})
	}

	authTolak("Q1 GET tanpa kredensial", "GET", "/orangtua/"+o1, nil,
		tanpaKunci, 401, 4001, "UNAUTHORIZED")
	authTolak("Q2 POST tanpa kredensial", "POST", K, kunj(uid("q02"), a1, "2026-10-01"),
		tanpaKunci, 401, 4001, "UNAUTHORIZED")
	authTolak("Q3 kunci tidak terdaftar", "GET", "/orangtua/"+o1, nil,
		"kunci-yang-tidak-pernah-ada", 401, 4001, "UNAUTHORIZED")
	authTolak("Q4 kunci kosong", "GET", "/orangtua/"+o1, nil,
		" ", 401, 4001, "UNAUTHORIZED")

	// Kunci readonly punya peran {read}; aturan di access.auth_resources
	// menuntut {write} untuk POST /api/kunjungan.
	authTolak("Q5 readonly menulis kunjungan", "POST", K, kunj(uid("q05"), a1, "2026-10-02"),
		testutil.KeyReadonly, 403, 4002, "FORBIDDEN")

	// Yang tidak dibatasi aturan tetap boleh dibaca readonly -- membuktikan
	// 403 di Q5 memang soal hak, bukan kunci readonly ditolak di mana-mana.
	tambah(skenario{bagian: Q, label: "Q6 readonly membaca orangtua", metode: "GET",
		path: "/orangtua/" + o1, kunci: testutil.KeyReadonly, http: 200})
	tambah(skenario{bagian: Q, label: "Q7 readonly membaca anak", metode: "GET",
		path: "/anak/" + a1, kunci: testutil.KeyReadonly, http: 200})

	// Health berada di luar PathPrefix, jadi tetap terbuka untuk probe.
	tambah(skenario{bagian: Q, label: "Q8 health tetap terbuka tanpa kredensial",
		metode: "GET", path: healthURL, kunci: tanpaKunci, http: 200})

	// ── T. Dokumentasi ──────────────────────────────────────────────────
	// Berada di luar prefix /api, jadi memakai URL penuh dan tanpa kredensial.
	T := "T. Dokumentasi"
	docsURL := strings.TrimSuffix(healthURL, "/health") + "/docs"

	tambah(skenario{bagian: T, label: "T1 index dirender", metode: "GET",
		path: docsURL, kunci: tanpaKunci, http: 200})
	tambah(skenario{bagian: T, label: "T2 katalog error dirender", metode: "GET",
		path: docsURL + "?doc=errors", kunci: tanpaKunci, http: 200})
	tambah(skenario{bagian: T, label: "T3 stylesheet tersaji", metode: "GET",
		path: docsURL + "?doc=gaya", kunci: tanpaKunci, http: 200})
	tambah(skenario{bagian: T, label: "T4 markdown mentah", metode: "GET",
		path: docsURL + "?raw=1", kunci: tanpaKunci, http: 200})
	tambah(skenario{bagian: T, label: "T5 dokumen tidak dikenal", metode: "GET",
		path: docsURL + "?doc=hantu", kunci: tanpaKunci, http: 404,
		errorCode: 2001, errorName: "NOT_FOUND"})
	tambah(skenario{bagian: T, label: "T6 penjelajahan path ditolak", metode: "GET",
		path: docsURL + "?doc=../../config", kunci: tanpaKunci, http: 404,
		errorCode: 2001, errorName: "NOT_FOUND"})

	tambah(skenario{bagian: T, label: "T7 laporan pengujian tidak disajikan", metode: "GET",
		path: docsURL + "?doc=laporan", kunci: tanpaKunci, http: 404,
		errorCode: 2001, errorName: "NOT_FOUND"})

	skenarioFaskes(tambah)

	return s
}

// TestApiStunting menjalankan seluruh skenario berurutan. Urutannya penting:
// skenario awal membuat data yang dirujuk skenario berikutnya.
func TestSkenarioEndToEnd(t *testing.T) {
	if !databaseAvailable() {
		t.Skipf("database tidak terjangkau di PG_DSN, uji dilewati")
	}
	if !serviceAvailable() {
		t.Skipf("service tidak menyala di %s, uji dilewati", healthURL)
	}

	truncateDB(t)
	idPosyandu = fetchIdPosyandu(t)
	rekam = nil

	for _, s := range allScenarios() {
		t.Run(s.label, func(t *testing.T) { run(t, s) })
	}

	failed := 0
	for _, r := range rekam {
		if r.Salah != "" {
			failed++
		}
	}
	t.Logf("%d skenario, %d sesuai harapan, %d bermasalah", len(rekam), len(rekam)-failed, failed)

	writeReport(t)
}
