package update_endpoint_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const idOrtuLain = "uji-update-ortu-lain"

func setupUntilRS(t *testing.T) {
	t.Helper()
	cleanup := func() {
		require.NoError(t, testutil.Scenario1Cleanup(db, testGroup))
		_, _ = db.Exec(`DELETE FROM stunting.orangtua WHERE id = $1`, idOrtuLain)
	}
	cleanup()
	t.Cleanup(cleanup)
	forgetPembuat()
	require.NoError(t, wiring.Scenario1Jakantro())
	require.NoError(t, wiring.Scenario1Send("02-kafka-patient.json", "03-kafka-relatedperson.json", "04-kafka-puskesmas.json", "05-kafka-rs.json"))
}

func ortuUpdate(modify map[string]any) map[string]any {
	b := map[string]any{"id": testutil.Scenario1IdOrtu, "nik": testutil.Scenario1NikIbu, "nama_ibu": "Rina Marlina"}
	for k, v := range modify {
		b[k] = v
	}
	return b
}

func TestAksesRSSelamaRujukanAktif(t *testing.T) {
	setupUntilRS(t)

	j := request(t, "PUT", "/faskes/anak", keyRS, map[string]any{"id": testutil.Scenario1IdAnak, "nik": "3175042003240099"})
	require.Equal(t, 200, j.Kode, "RS tujuan rujukan boleh mengubah nik: %v", j.Isi)
	assert.Equal(t, "3175042003240099/Siti Aminah", scalar[string](t, `SELECT nik || '/' || nama FROM stunting.anak WHERE id = $1`, testutil.Scenario1IdAnak),
		"field yang tidak dikirim dipertahankan")

	j = request(t, "PUT", "/faskes/orangtua", keyRS, ortuUpdate(map[string]any{"no_kk": "3175042509870099", "id_posyandu": "POS-LAIN"}))
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, "3175042509870099/POS-JAKARTA-01", scalar[string](t,
		`SELECT no_kk || '/' || id_posyandu FROM stunting.orangtua WHERE id = $1`, testutil.Scenario1IdOrtu), "id_posyandu tidak boleh diubah faskes")

	assert.Equal(t, 200, request(t, "GET", "/orangtua/"+testutil.Scenario1IdOrtu, keyRS, nil).Kode)
	assert.Equal(t, 200, request(t, "GET", "/anak/"+testutil.Scenario1IdAnak, keyRS, nil).Kode)
	assert.Equal(t, 403, request(t, "PUT", "/faskes/anak", testutil.KeyFaskesA, map[string]any{"id": testutil.Scenario1IdAnak, "nama": "Siti A"}).Kode,
		"puskesmas yang merujuk kehilangan akses selama pasien di RS")
	assert.Equal(t, 403, request(t, "GET", "/anak/"+testutil.Scenario1IdAnak, testutil.KeyFaskesA, nil).Kode)

	require.NoError(t, wiring.Scenario1Send("06-kafka-rujuk-balik.json"))
	assert.Equal(t, 403, request(t, "PUT", "/faskes/orangtua", keyRS, ortuUpdate(map[string]any{"telepon": "0800"})).Kode, "setelah rujuk balik")
	assert.Equal(t, 403, request(t, "GET", "/anak/"+testutil.Scenario1IdAnak, keyRS, nil).Kode)
	assert.Equal(t, 200, request(t, "GET", "/anak/"+testutil.Scenario1IdAnak, testutil.KeyFaskesA, nil).Kode, "akses kembali ke tujuan rujuk balik")
}

func TestFaskesTidakBisaMemindahAnak(t *testing.T) {
	setupUntilRS(t)

	j := request(t, "POST", "/orangtua", testutil.KeyJakantro, map[string]any{
		"id": idOrtuLain, "id_posyandu": "POS-JAKARTA-01", "no_kk": "3175042509870077", "nik": "3175040202900077",
		"nama_ayah": "A", "nama_ibu": "I", "telepon": "08", "rt": "001", "rw": "001", "alamat": "Jl", "kia": 1,
	})
	require.Equal(t, 201, j.Kode, "%v", j.Isi)

	j = request(t, "PUT", "/faskes/anak", keyRS, map[string]any{"id": testutil.Scenario1IdAnak, "id_orangtua": idOrtuLain, "nama": "Siti F"})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, testutil.Scenario1IdOrtu+"/Siti F", scalar[string](t, `SELECT id_orangtua || '/' || nama FROM stunting.anak WHERE id = $1`, testutil.Scenario1IdAnak))

	j = request(t, "PUT", "/anak", testutil.KeyJakantro, map[string]any{"id": testutil.Scenario1IdAnak, "id_orangtua": idOrtuLain})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, idOrtuLain, scalar[string](t, `SELECT id_orangtua FROM stunting.anak WHERE id = $1`, testutil.Scenario1IdAnak), "jakantro boleh")
}

func TestPUTOrangtua(t *testing.T) {
	setupRiwayat(t)

	j := request(t, "PUT", "/orangtua", testutil.KeyJakantro, ortuUpdate(map[string]any{"telepon": "0811111"}))
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, "0811111", j.Isi["telepon"])
	assert.Equal(t, "0811111/Rina Marlina/jakantro", scalar[string](t,
		`SELECT telepon || '/' || nama_ibu || '/' || updated_by FROM stunting.orangtua WHERE id = $1`, testutil.Scenario1IdOrtu))

	assert.Equal(t, 200, request(t, "PUT", "/faskes/orangtua", testutil.KeyFaskesA, ortuUpdate(map[string]any{"alamat": "Jl. Baru"})).Kode, "puskesmas pembina")
	assert.Equal(t, 400, request(t, "PUT", "/orangtua", testutil.KeyJakantro, `{rusak`).Kode)
	assert.Equal(t, 422, request(t, "PUT", "/orangtua", testutil.KeyJakantro, map[string]any{"telepon": "08"}).Kode)
	assert.Equal(t, 422, request(t, "PUT", "/orangtua", testutil.KeyJakantro, ortuUpdate(map[string]any{"nik": "123"})).Kode)
	assert.Equal(t, 404, request(t, "PUT", "/orangtua", testutil.KeyJakantro, map[string]any{"id": "tidak-ada", "nik": "3175040202900999", "nama_ibu": "X"}).Kode)

	j = request(t, "PUT", "/orangtua", testutil.KeyJakantro, map[string]any{"id": testutil.Scenario1IdOrtu, "telepon": "0822"})
	require.Equal(t, 200, j.Kode, "PUT cukup id: %v", j.Isi)
	j = request(t, "PUT", "/faskes/orangtua", testutil.KeyFaskesA, map[string]any{"id": testutil.Scenario1IdOrtu, "alamat": "Jl. Cukup Id"})
	require.Equal(t, 200, j.Kode, "PUT faskes cukup id: %v", j.Isi)
	assert.Equal(t, testutil.Scenario1NikIbu+"/Rina Marlina/0822/Jl. Cukup Id", scalar[string](t,
		`SELECT nik || '/' || nama_ibu || '/' || telepon || '/' || alamat FROM stunting.orangtua WHERE id = $1`, testutil.Scenario1IdOrtu),
		"nik dan nama yang tidak dikirim dipertahankan")
	assert.Equal(t, 403, request(t, "PUT", "/orangtua", keyWithoutGroup, ortuUpdate(nil)).Kode)
}

func TestPemeriksaanAnakLamaButuhAkses(t *testing.T) {
	content, err := os.ReadFile(filepath.Join(testutil.RepoRoot(), "simulasi", "faskes-langsung", "01-pemeriksaan.json"))
	require.NoError(t, err)

	setupUntilRS(t)
	j := request(t, "POST", "/faskes/pemeriksaan", keyRS, string(content))
	assert.Equal(t, 201, j.Kode, "RS tujuan rujukan aktif: %v", j.Isi)

	require.NoError(t, wiring.Scenario1Send("06-kafka-rujuk-balik.json"))
	j = request(t, "POST", "/faskes/pemeriksaan", keyRS, string(content))
	assert.Equal(t, 403, j.Kode, "setelah rujuk balik: %v", j.Isi)
	assert.Equal(t, 1008, j.Galat)
}

func pemeriksaan(t *testing.T, modify func(p map[string]any)) map[string]any {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(testutil.RepoRoot(), "simulasi", "faskes-langsung", "01-pemeriksaan.json"))
	require.NoError(t, err)
	var p map[string]any
	require.NoError(t, json.Unmarshal(content, &p))
	modify(p)
	return p
}

func TestPemeriksaanTidakBolehMenempelKeEncounterFaskesLain(t *testing.T) {
	setupRiwayat(t)

	cases := map[string]func(p map[string]any){
		"kunjungan milik RS": func(p map[string]any) {
			p["kunjungan"].(map[string]any)["id_satusehat"] = "enc-rs-01"
		},
		"observasi menunjuk encounter RS": func(p map[string]any) {
			p["observasi"].([]any)[0].(map[string]any)["ref_encounter"] = "enc-rs-01"
		},
		"diagnosa menunjuk encounter yang belum ada": func(p map[string]any) {
			p["diagnosa"].([]any)[0].(map[string]any)["ref_encounter"] = "enc-hantu"
		},
		"layanan menunjuk encounter RS": func(p map[string]any) {
			p["layanan"].([]any)[0].(map[string]any)["ref_encounter"] = "enc-rs-01"
		},
	}
	for name, modify := range cases {
		j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, pemeriksaan(t, modify))
		assert.Equal(t, 403, j.Kode, "%s: %v", name, j.Isi)
		assert.Equal(t, 1008, j.Galat, name)
	}
	assert.Zero(t, scalar[int](t, `SELECT count(*) FROM stunting.observasi WHERE satusehat_id LIKE 'obs-langsung-%'`),
		"tidak ada yang tersimpan dari kiriman yang ditolak")
	assert.Equal(t, 2, scalar[int](t, `SELECT count(*) FROM stunting.observasi o JOIN stunting.kunjungan k ON k.id = o.id_kunjungan WHERE k.satusehat_id = 'enc-rs-01'`),
		"observasi kunjungan RS tetap milik RS")

	sah := pemeriksaan(t, func(p map[string]any) {
		p["observasi"].([]any)[0].(map[string]any)["ref_encounter"] = "enc-pkm-01"
		p["diagnosa"].([]any)[0].(map[string]any)["ref_encounter"] = "enc-langsung-01"
	})
	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, sah)
	require.Equal(t, 201, j.Kode, "encounter milik sendiri: %v", j.Isi)
	j = request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, sah)
	assert.Equal(t, 201, j.Kode, "kiriman ulang ke kunjungan sendiri: %v", j.Isi)
}

func TestPUTFaskesFieldTakBisaDiubah(t *testing.T) {
	setupRiwayat(t)

	for resource, body := range map[string]map[string]any{
		"observasi": {"id_satusehat": "obs-pkm-bb-01", "system": "http://loinc.org", "kode": "29463-7"},
		"diagnosa":  {"id_satusehat": "cond-pkm-01", "kode": "E45"},
		"layanan":   {"id_satusehat": "proc-pkm-02", "jenis": "procedure"},
		"rujukan":   {"id_satusehat": "sr-rujuk-01"},
	} {
		body["ref_encounter"] = "enc-rs-01"
		j := request(t, "PUT", "/faskes/"+resource, testutil.KeyFaskesA, body)
		require.Equal(t, 200, j.Kode, "%s: %v", resource, j.Isi)
		assert.NotEqual(t, "enc-rs-01", scalar[string](t, `SELECT coalesce(ref_encounter,'') FROM stunting.`+resource+` WHERE satusehat_id = $1`, body["id_satusehat"]),
			"%s: ref_encounter tidak ikut berubah", resource)
	}

	idKunjunganRS := scalar[string](t, `SELECT id FROM stunting.kunjungan WHERE satusehat_id = 'enc-rs-01'`)
	j := request(t, "PUT", "/faskes/diagnosa", testutil.KeyFaskesA, map[string]any{
		"id_satusehat": "cond-pkm-01", "kode": "E45", "id_anak": "anak-lain", "id_kunjungan": idKunjunganRS, "jenis": "alergi",
	})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, testutil.Scenario1IdAnak+"/enc-pkm-01/diagnosis", scalar[string](t, `SELECT d.id_anak || '/' || k.satusehat_id || '/' || d.jenis
		FROM stunting.diagnosa d JOIN stunting.kunjungan k ON k.id = d.id_kunjungan WHERE d.satusehat_id = 'cond-pkm-01'`))

	j = request(t, "PUT", "/faskes/layanan", testutil.KeyFaskesA, map[string]any{"id_satusehat": "proc-pkm-02", "jenis": "immunization"})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, "procedure", scalar[string](t, `SELECT jenis FROM stunting.layanan WHERE satusehat_id = 'proc-pkm-02'`))

	j = request(t, "PUT", "/faskes/kunjungan", testutil.KeyFaskesA, map[string]any{
		"id_satusehat": "enc-pkm-01", "tanggal_pengukuran": "2026-02-05", "stunting": 0, "ref_episode": "ep-lain", "ref_rujukan": "sr-lain",
	})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, "1/ep-stunting-sim1/", scalar[string](t, `SELECT stunting || '/' || coalesce(ref_episode,'') || '/' || coalesce(ref_rujukan,'')
		FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`))
}

func TestRSTidakBisaMembukaAksesLewatRujukBalik(t *testing.T) {
	setupRiwayat(t)
	require.Equal(t, 403, request(t, "GET", "/anak/"+testutil.Scenario1IdAnak, keyRS, nil).Kode, "sudah dirujuk balik")

	j := request(t, "PUT", "/faskes/rujukan", keyRS, map[string]any{
		"id_satusehat": "sr-balik-01", "jenis": "internal", "tanggal": "2020-01-01T00:00:00Z",
		"ref_faskes_asal": "100011", "ref_faskes_tujuan": "200022", "status": "completed",
	})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, "rujuk_balik/200022/100011/completed", scalar[string](t, `SELECT r.jenis || '/' || r.ref_faskes_asal || '/' || r.ref_faskes_tujuan || '/' || r.status
		FROM stunting.rujukan r WHERE r.satusehat_id = 'sr-balik-01'`))
	assert.NotEqual(t, "2020-01-01", scalar[string](t, `SELECT to_char(tanggal,'YYYY-MM-DD') FROM stunting.rujukan WHERE satusehat_id = 'sr-balik-01'`))

	assert.Equal(t, 403, request(t, "GET", "/anak/"+testutil.Scenario1IdAnak, keyRS, nil).Kode, "akses tetap tertutup")
}

// skenario 1: kunjungan baru hanya oleh pemegang akses saat ini
func TestPemeriksaanHanyaOlehTujuanRujukan(t *testing.T) {
	setupUntilRS(t)
	content, err := os.ReadFile(filepath.Join(testutil.RepoRoot(), "simulasi", "faskes-langsung", "01-pemeriksaan.json"))
	require.NoError(t, err)
	pemeriksaan := func(enc string) map[string]any {
		var p map[string]any
		require.NoError(t, json.Unmarshal(content, &p))
		p["kunjungan"].(map[string]any)["id_satusehat"] = enc
		return p
	}

	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, pemeriksaan("enc-uji-update-a1"))
	assert.Equal(t, 403, j.Kode, "pasien sedang dirujuk ke RS: %v", j.Isi)

	require.NoError(t, wiring.Scenario1Send("06-kafka-rujuk-balik.json"))
	j = request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, pemeriksaan("enc-uji-update-a2"))
	assert.Equal(t, 201, j.Kode, "tujuan rujuk balik: %v", j.Isi)
	j = request(t, "POST", "/faskes/pemeriksaan", keyRS, pemeriksaan("enc-uji-update-rs"))
	assert.Equal(t, 403, j.Kode, "RS sudah merujuk balik: %v", j.Isi)
}

// skenario 1: by nik ikut tujuan rujukan, sama dengan by id
func TestGETByNikIkutTujuanRujukan(t *testing.T) {
	setupUntilRS(t)
	anak, ortu := "/anak?nik="+testutil.Scenario1NikAnak, "/orangtua?nik="+testutil.Scenario1NikIbu

	assert.Equal(t, 200, request(t, "GET", anak, keyRS, nil).Kode, "RS tujuan rujukan")
	assert.Equal(t, 200, request(t, "GET", ortu, keyRS, nil).Kode, "RS tujuan rujukan")
	assert.Equal(t, 403, request(t, "GET", anak, testutil.KeyFaskesA, nil).Kode, "puskesmas sudah merujuk ke RS")
	assert.Equal(t, 403, request(t, "GET", ortu, testutil.KeyFaskesA, nil).Kode, "puskesmas sudah merujuk ke RS")

	require.NoError(t, wiring.Scenario1Send("06-kafka-rujuk-balik.json"))
	assert.Equal(t, 403, request(t, "GET", anak, keyRS, nil).Kode, "RS sudah merujuk balik")
	assert.Equal(t, 200, request(t, "GET", anak, testutil.KeyFaskesA, nil).Kode, "tujuan rujuk balik")
}

