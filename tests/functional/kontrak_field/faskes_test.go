package kontrak_field_test

import (
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRiwayatScenario1(t *testing.T) {
	t.Helper()
	setup(t)
	require.NoError(t, wiring.Scenario1Jakantro())
	require.NoError(t, wiring.Scenario1Send(testutil.Scenario1KafkaFiles...))
}

func add(p map[string]any, key string, content map[string]any) {
	list, _ := p[key].([]any)
	p[key] = append(list, content)
}

func TestFieldSistemDiabaikanSaatPemeriksaan(t *testing.T) {
	setup(t)

	p := newPemeriksaan()
	withSystemFields(p["anak"].(map[string]any))
	withSystemFields(p["kunjungan"].(map[string]any))
	withSystemFields(p["diagnosa"].([]any)[0].(map[string]any))
	add(p, "observasi", withSystemFields(map[string]any{"id_satusehat": "obs-kontrak-01", "system": "http://loinc.org", "kode": "29463-7", "nilai_angka": 7.9}))
	add(p, "layanan", withSystemFields(map[string]any{"id_satusehat": "lay-kontrak-01", "jenis": "procedure", "status": "completed"}))
	add(p, "rujukan", withSystemFields(map[string]any{"id_satusehat": "sr-kontrak-01", "jenis": "internal"}))
	add(p, "episode", withSystemFields(map[string]any{"id_satusehat": "ep-kontrak-01", "status": "active"}))
	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, p)
	require.Equal(t, 201, j.Kode, "%v", j.Isi)

	for _, r := range [][2]string{
		{"anak", ihsAnakBaru}, {"kunjungan", "enc-kontrak-01"}, {"observasi", "obs-kontrak-01"},
		{"diagnosa", "cond-kontrak-01"}, {"layanan", "lay-kontrak-01"}, {"rujukan", "sr-kontrak-01"}, {"episode", "ep-kontrak-01"},
	} {
		assert.Equal(t, "2026|t|-|-", scalar[string](t, `SELECT concat_ws('|', extract(year FROM "createdAt"), "deletedAt" IS NULL,
			coalesce(to_jsonb(x)->>'updated_by', '-'), coalesce(to_jsonb(x)->>'deleted_by', '-'))
			FROM stunting.`+r[0]+` x WHERE satusehat_id = $1`, r[1]), r[0])
	}
}

func TestFieldSistemDiabaikanSaatPUTFaskes(t *testing.T) {
	setupRiwayatScenario1(t)

	for resource, id := range map[string]string{
		"kunjungan": "enc-pkm-01", "observasi": "obs-pkm-bb-01", "diagnosa": "cond-pkm-01", "rujukan": "sr-rujuk-01", "episode": "ep-stunting-sim1",
	} {
		before := scalar[string](t, `SELECT extract(year FROM "createdAt")::text FROM stunting.`+resource+` WHERE satusehat_id = $1`, id)
		body := withSystemFields(map[string]any{"id_satusehat": id})
		if resource == "kunjungan" {
			body["tanggal_pengukuran"] = "2026-02-05"
		}
		if resource == "observasi" || resource == "diagnosa" {
			body["system"], body["kode"] = "http://loinc.org", "29463-7"
			if resource == "diagnosa" {
				body["system"], body["kode"] = "http://hl7.org/fhir/sid/icd-10", "E45"
			}
		}
		j := request(t, "PUT", "/faskes/"+resource, testutil.KeyFaskesA, body)
		require.Equal(t, 200, j.Kode, "%s: %v", resource, j.Isi)
		assert.Equal(t, before+"|t|-", scalar[string](t, `SELECT concat_ws('|', extract(year FROM "createdAt"), "deletedAt" IS NULL,
			CASE WHEN coalesce(to_jsonb(x)->>'updated_by','') = $2 OR coalesce(to_jsonb(x)->>'deleted_by','') = $2 THEN 'dari-pengirim' ELSE '-' END)
			FROM stunting.`+resource+` x WHERE satusehat_id = $1`, id, isianPalsu), resource)
	}
}

func TestIdAnakDanOrangtuaDariFaskesDiabaikan(t *testing.T) {
	setupRiwayatScenario1(t)

	p := newPemeriksaan()
	a := p["anak"].(map[string]any)
	a["id"], a["id_orangtua"] = "anak-dari-faskes", testutil.Scenario1IdOrtu
	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, p)
	require.Equal(t, 201, j.Kode, "%v", j.Isi)

	assert.NotEqual(t, "anak-dari-faskes", j.Isi["id_anak"])
	assert.Equal(t, "-", scalar[string](t, `SELECT coalesce(id_orangtua, '-') FROM stunting.anak WHERE satusehat_id = $1`, ihsAnakBaru),
		"faskes tidak menentukan keluarga anak")
}

func TestPemeriksaanDitolakTidakMeninggalkanAnak(t *testing.T) {
	setupRiwayatScenario1(t)

	p := newPemeriksaan()
	p["kunjungan"].(map[string]any)["id_satusehat"] = "enc-rs-01"
	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, p)
	require.Equal(t, 403, j.Kode, "%v", j.Isi)

	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.anak WHERE satusehat_id = $1 OR nik = $2`, ihsAnakBaru, nikAnakBaru))
}

func TestPemeriksaanMenyimpanIsianPosyandu(t *testing.T) {
	setup(t)

	p := newPemeriksaan()
	k := p["kunjungan"].(map[string]any)
	for _, f := range []string{"asi_bulan_0", "asi_bulan_1", "asi_bulan_2", "asi_bulan_3", "asi_bulan_4", "asi_bulan_5", "asi_bulan_6",
		"vit_biru", "vit_merah", "pitting_edema", "kelas_ibu_balita"} {
		k[f] = 1
	}
	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, p)
	require.Equal(t, 201, j.Kode, "%v", j.Isi)

	assert.Equal(t, 11, scalar[int](t, `SELECT asi_bulan_0 + asi_bulan_1 + asi_bulan_2 + asi_bulan_3 + asi_bulan_4 + asi_bulan_5 + asi_bulan_6
		+ vit_biru + vit_merah + pitting_edema + kelas_ibu_balita FROM stunting.kunjungan WHERE satusehat_id = 'enc-kontrak-01'`))
}

func TestRujukanKeFaskesBelumTerdaftarDitolak(t *testing.T) {
	setup(t)

	p := newPemeriksaan()
	add(p, "rujukan", map[string]any{"id_satusehat": "sr-kontrak-01", "jenis": "rujukan", "ref_faskes_tujuan": "ORG-BELUM-TERDAFTAR"})
	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, p)
	assert.Equal(t, 422, j.Kode, "%v", j.Isi)
	assert.Equal(t, 1005, j.Galat)
	assert.Contains(t, j.Isi["message"], "ORG-BELUM-TERDAFTAR")
	assert.Zero(t, count(t, `SELECT count(*) FROM stunting.anak WHERE satusehat_id = $1`, ihsAnakBaru))

	p = newPemeriksaan()
	add(p, "rujukan", map[string]any{"id_satusehat": "sr-kontrak-01", "jenis": "rujukan", "ref_faskes_tujuan": "200022"})
	add(p, "rujukan", map[string]any{"id_satusehat": "sr-kontrak-02", "jenis": "internal"})
	j = request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, p)
	assert.Equal(t, 201, j.Kode, "tujuan terdaftar atau tanpa tujuan: %v", j.Isi)
}

func TestKomponenObservasiMewarisiTanggalInduk(t *testing.T) {
	setup(t)

	p := newPemeriksaan()
	add(p, "observasi", map[string]any{
		"id_satusehat": "obs-kontrak-panel", "system": "http://snomed.info/sct", "kode": "1156892006", "tanggal": "2026-09-01T02:30:00Z",
		"component": []any{map[string]any{"id_satusehat": "obs-kontrak-c1", "system": "http://snomed.info/sct", "kode": "709261005", "nilai_angka": 1}},
	})
	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, p)
	require.Equal(t, 201, j.Kode, "%v", j.Isi)

	assert.Equal(t, "2026-09-01", scalar[string](t, `SELECT to_char(tanggal AT TIME ZONE 'UTC','YYYY-MM-DD') FROM stunting.observasi WHERE satusehat_id = 'obs-kontrak-c1'`))
}

func TestIdSistemDiPemeriksaanDiabaikan(t *testing.T) {
	setup(t)

	p := newPemeriksaan()
	k := p["kunjungan"].(map[string]any)
	k["id_episode"], k["id_rujukan"], k["id_faskes"] = "ep-x", "sr-x", "FAS-RS-HARAPAN"
	add(p, "observasi", map[string]any{"id_satusehat": "obs-kontrak-01", "system": "http://loinc.org", "kode": "29463-7", "nilai_angka": 7.9, "id_induk": "obs-x"})
	add(p, "episode", map[string]any{"id_satusehat": "ep-kontrak-01", "id_faskes": "FAS-RS-HARAPAN"})
	add(p, "rujukan", map[string]any{"id_satusehat": "sr-kontrak-01", "jenis": "rujukan", "ref_faskes_tujuan": "200022",
		"id_faskes_asal": "FAS-RS-HARAPAN", "id_faskes_tujuan": "FAS-PKM-KENANGA"})
	j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, p)
	require.Equal(t, 201, j.Kode, "%v", j.Isi)

	assert.Equal(t, "-|-|FAS-PKM-KENANGA", scalar[string](t, `SELECT coalesce(id_episode,'-') || '|' || coalesce(id_rujukan,'-') || '|' || id_faskes
		FROM stunting.kunjungan WHERE satusehat_id = 'enc-kontrak-01'`))
	assert.Equal(t, "-", scalar[string](t, `SELECT coalesce(id_induk,'-') FROM stunting.observasi WHERE satusehat_id = 'obs-kontrak-01'`))
	assert.Equal(t, "-", scalar[string](t, `SELECT coalesce(id_faskes,'-') FROM stunting.episode WHERE satusehat_id = 'ep-kontrak-01'`),
		"faskes episode hanya dari ref_faskes")
	assert.Equal(t, "FAS-PKM-KENANGA|FAS-RS-HARAPAN", scalar[string](t, `SELECT id_faskes_asal || '|' || id_faskes_tujuan FROM stunting.rujukan WHERE satusehat_id = 'sr-kontrak-01'`),
		"faskes rujukan dari ref (asal: pengirim), bukan dari id kiriman")
}

func faskesRujukan(t *testing.T) string {
	t.Helper()
	return scalar[string](t, `SELECT coalesce(id_faskes_asal,'-') || '|' || coalesce(id_faskes_tujuan,'-') || '|' || coalesce(ref_faskes_tujuan,'-')
		FROM stunting.rujukan WHERE satusehat_id = 'sr-rujuk-01'`)
}

// faskes rujukan hanya ditentukan saat dibuat, dari ref; PUT tidak mengubahnya
func TestPUTRujukanTidakMengubahFaskes(t *testing.T) {
	setupRiwayatScenario1(t)
	awal := faskesRujukan(t)
	require.Equal(t, "FAS-PKM-KENANGA|FAS-RS-HARAPAN|200022", awal)

	j := request(t, "PUT", "/faskes/rujukan", testutil.KeyFaskesA, map[string]any{
		"id_satusehat": "sr-rujuk-01", "id_faskes_asal": "FAS-RS-HARAPAN", "id_faskes_tujuan": "FAS-PKM-KENANGA", "ref_faskes_tujuan": "100011",
	})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, awal, faskesRujukan(t))
}

func TestPUTFaskesMengabaikanIdFaskes(t *testing.T) {
	setupRiwayatScenario1(t)

	j := request(t, "PUT", "/faskes/kunjungan", testutil.KeyFaskesA, map[string]any{
		"id_satusehat": "enc-pkm-01", "tanggal_pengukuran": "2026-02-05", "id_faskes": "FAS-RS-HARAPAN",
	})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, "FAS-PKM-KENANGA", scalar[string](t, `SELECT id_faskes FROM stunting.kunjungan WHERE satusehat_id = 'enc-pkm-01'`))

	j = request(t, "PUT", "/faskes/episode", testutil.KeyFaskesA, map[string]any{"id_satusehat": "ep-stunting-sim1", "id_faskes": "FAS-RS-HARAPAN"})
	require.Equal(t, 200, j.Kode, "%v", j.Isi)
	assert.Equal(t, "FAS-PKM-KENANGA", scalar[string](t, `SELECT id_faskes FROM stunting.episode WHERE satusehat_id = 'ep-stunting-sim1'`))
}

func TestAnakBaruTanpaDataWajibDitolak422(t *testing.T) {
	setup(t)

	for _, f := range []string{"nama", "tanggal_lahir", "jenis_kelamin"} {
		p := newPemeriksaan()
		delete(p["anak"].(map[string]any), f)
		j := request(t, "POST", "/faskes/pemeriksaan", testutil.KeyFaskesA, p)
		assert.Equal(t, 422, j.Kode, "%s: %v", f, j.Isi["message"])
		assert.Equal(t, 1002, j.Galat, f)
		assert.Contains(t, j.Isi["message"], f)
	}
}
