package optimasi_query_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// absolut: JalankanAplikasi memindah working directory ke root repo
var berkasGolden = filepath.Join(testutil.RepoRoot(), "tests", "functional", "optimasi_query", "testdata", "summary.golden.json")

// kasus tepi yang harus tetap benar setelah turunan diambil sekali per anak
func setupSummary(t *testing.T) {
	t.Helper()
	setupRiwayat(t)
	a := testutil.Scenario1IdAnak

	// observasi menggantung; komponen di-insert terbalik dari urutan kode
	exec(t, `INSERT INTO stunting.observasi (id, id_anak, ref_encounter, system, kode, display, kategori, tanggal)
		VALUES ('uji-opt-obs-gantung', $1, 'enc-uji-opt-belum', 'http://snomed.info/sct', '1156892006', 'Nutrition assessment', 'survey', '2026-03-05T02:00:00Z')`, a)
	exec(t, `INSERT INTO stunting.observasi (id, id_anak, id_induk, ref_encounter, system, kode, display, nilai_angka, tanggal) VALUES
		('uji-opt-obs-gantung-b', $1, 'uji-opt-obs-gantung', 'enc-uji-opt-belum', 'http://snomed.info/sct', '710999009', 'Monitoring food intake', 0, '2026-03-05T02:00:00Z'),
		('uji-opt-obs-gantung-a', $1, 'uji-opt-obs-gantung', 'enc-uji-opt-belum', 'http://snomed.info/sct', '709261005', 'Assessment of breastfeeding', 1, '2026-03-05T02:00:00Z')`, a)

	// diagnosa menggantung, di-insert tidak urut tanggal_catat
	exec(t, `INSERT INTO stunting.diagnosa (id, id_anak, ref_encounter, system, kode, display, tanggal_catat) VALUES
		('uji-opt-dg-1', $1, 'enc-uji-opt-belum', 'http://hl7.org/fhir/sid/icd-10', 'E44.0', 'Moderate protein-energy malnutrition', '2026-03-03'),
		('uji-opt-dg-2', $1, 'enc-uji-opt-belum', 'http://hl7.org/fhir/sid/icd-10', 'E45', 'Nutritional stunting', '2026-03-01'),
		('uji-opt-dg-3', $1, 'enc-uji-opt-belum', 'http://hl7.org/fhir/sid/icd-10', 'D50.9', 'Iron deficiency anaemia', '2026-03-02')`, a)

	// layanan dan rujukan menggantung tetap tampil di daftar datar
	exec(t, `INSERT INTO stunting.layanan (id, id_anak, ref_encounter, jenis, system, kode, display, status, tanggal)
		VALUES ('uji-opt-ly-gantung', $1, 'enc-uji-opt-belum', 'nutrition_order', 'http://snomed.info/sct', '435801000124108', 'High protein diet', 'active', '2026-03-05T03:00:00Z')`, a)
	exec(t, `INSERT INTO stunting.rujukan (id, id_anak, ref_encounter, jenis, ref_faskes_asal, ref_faskes_tujuan, status, prioritas, tanggal)
		VALUES ('uji-opt-rj-gantung', $1, 'enc-uji-opt-belum', 'rujukan', '100011', '200022', 'active', 'routine', '2026-03-05T04:00:00Z')`, a)

	// kunjungan terhapus beserta turunannya; tidak boleh muncul di bagian mana pun
	exec(t, `INSERT INTO stunting.kunjungan (id, satusehat_id, id_anak, id_faskes, tanggal_pengukuran, berat_badan, "deletedAt")
		VALUES ('uji-opt-kunj-hapus', 'enc-uji-opt-hapus', $1, 'FAS-PKM-KENANGA', '2026-02-10', 7.3, now())`, a)
	exec(t, `INSERT INTO stunting.observasi (id, id_anak, id_kunjungan, system, kode, display, nilai_angka, satuan, tanggal)
		VALUES ('uji-opt-obs-hapus', $1, 'uji-opt-kunj-hapus', 'http://loinc.org', '29463-7', 'Body weight', 7.3, 'kg', '2026-02-10T02:00:00Z')`, a)
	exec(t, `INSERT INTO stunting.diagnosa (id, id_anak, id_kunjungan, system, kode, display, tanggal_catat)
		VALUES ('uji-opt-dg-hapus', $1, 'uji-opt-kunj-hapus', 'http://hl7.org/fhir/sid/icd-10', 'E45', 'Nutritional stunting', '2026-02-10')`, a)
	exec(t, `INSERT INTO stunting.layanan (id, id_anak, id_kunjungan, jenis, system, kode, display, status, tanggal)
		VALUES ('uji-opt-ly-hapus', $1, 'uji-opt-kunj-hapus', 'procedure', 'http://snomed.info/sct', '441041000124100', 'Counseling about nutrition', 'completed', '2026-02-10T03:00:00Z')`, a)
	exec(t, `INSERT INTO stunting.rujukan (id, id_anak, id_kunjungan, jenis, id_faskes_asal, id_faskes_tujuan, ref_faskes_asal, ref_faskes_tujuan, status, prioritas, tanggal)
		VALUES ('uji-opt-rj-hapus', $1, 'uji-opt-kunj-hapus', 'rujukan', 'FAS-PKM-KENANGA', 'FAS-RS-HARAPAN', '100011', '200022', 'active', 'routine', '2026-02-10T04:00:00Z')`, a)

	forgetCache()
}

func fetchSummary(t *testing.T) jawaban {
	t.Helper()
	j := request(t, "GET", "/anak/"+testutil.Scenario1IdAnak+"/summary", testutil.KeyJakantro, nil)
	require.Equal(t, 200, j.Kode, "%s", j.Mentah)
	return j
}

var polaUUID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// kolom audit diisi sistem dan bukan bagian dari bentuk summary
var kolomSistem = map[string]bool{"created_at": true, "updated_at": true, "deleted_at": true, "updated_by": true, "deleted_by": true}

// buang kolom sistem dan ganti uuid acak (dari jalur Kafka) dengan nomor urut kemunculan
func normalize(t *testing.T, mentah []byte) []byte {
	t.Helper()
	var v any
	require.NoError(t, json.Unmarshal(mentah, &v))
	ganti := map[string]string{}
	var jalan func(any) any
	jalan = func(x any) any {
		switch n := x.(type) {
		case map[string]any:
			key := make([]string, 0, len(n))
			for k := range n {
				key = append(key, k)
			}
			sort.Strings(key)
			for _, k := range key {
				if kolomSistem[k] {
					delete(n, k)
					continue
				}
				n[k] = jalan(n[k])
			}
			return n
		case []any:
			for i := range n {
				n[i] = jalan(n[i])
			}
			return n
		case string:
			if polaUUID.MatchString(n) {
				if _, exists := ganti[n]; !exists {
					ganti[n] = fmt.Sprintf("<uuid-%d>", len(ganti)+1)
				}
				return ganti[n]
			}
		}
		return x
	}
	v = jalan(v)

	// urutan diagnosa dan komponen observasi menggantung diuji terpisah
	if m, ok := v.(map[string]any)["menggantung"].(map[string]any); ok {
		sortIds(m["diagnosa"])
		if obs, ok := m["observasi"].([]any); ok {
			for _, o := range obs {
				sortIds(o.(map[string]any)["component"])
			}
		}
	}
	result, err := json.MarshalIndent(v, "", "  ")
	require.NoError(t, err)
	return result
}

// bedaJSON melaporkan path yang berbeda; array berisi objek ber-id dibandingkan per id
func jsonDiff(golden, resp any, route string) []string {
	var result []string
	tambah := func(f string, a ...any) {
		if len(result) < 25 {
			result = append(result, fmt.Sprintf(f, a...))
		}
	}
	var banding func(g, r any, route string)
	banding = func(g, r any, route string) {
		switch gv := g.(type) {
		case map[string]any:
			rv, ok := r.(map[string]any)
			if !ok {
				tambah("%s: golden objek, respons %v", route, r)
				return
			}
			key := map[string]bool{}
			for k := range gv {
				key[k] = true
			}
			for k := range rv {
				key[k] = true
			}
			urut := make([]string, 0, len(key))
			for k := range key {
				urut = append(urut, k)
			}
			sort.Strings(urut)
			for _, k := range urut {
				anak := strings.TrimPrefix(route+"."+k, ".")
				a, adaG := gv[k]
				b, adaR := rv[k]
				switch {
				case !adaR:
					tambah("%s: ada di golden, tidak ada di respons", anak)
				case !adaG:
					tambah("%s: tidak ada di golden, ada di respons (%v)", anak, b)
				default:
					banding(a, b, anak)
				}
			}
		case []any:
			rv, ok := r.([]any)
			if !ok {
				tambah("%s: golden array, respons %v", route, r)
				return
			}
			if ig, ir := indexIds(gv), indexIds(rv); ig != nil && ir != nil {
				var hanyaG, hanyaR []string
				for id := range ig {
					if _, exists := ir[id]; !exists {
						hanyaG = append(hanyaG, id)
					}
				}
				for id := range ir {
					if _, exists := ig[id]; !exists {
						hanyaR = append(hanyaR, id)
					}
				}
				sort.Strings(hanyaG)
				sort.Strings(hanyaR)
				if len(hanyaG) > 0 {
					tambah("%s: id hanya di golden: %v", route, hanyaG)
				}
				if len(hanyaR) > 0 {
					tambah("%s: id hanya di respons: %v", route, hanyaR)
				}
				for id, a := range ig {
					if b, exists := ir[id]; exists {
						banding(a, b, fmt.Sprintf("%s[id=%s]", route, id))
					}
				}
				return
			}
			if len(gv) != len(rv) {
				tambah("%s: golden %d elemen, respons %d", route, len(gv), len(rv))
			}
			for i := 0; i < len(gv) && i < len(rv); i++ {
				banding(gv[i], rv[i], fmt.Sprintf("%s[%d]", route, i))
			}
		default:
			if !reflect.DeepEqual(g, r) {
				tambah("%s: golden %v, respons %v", route, g, r)
			}
		}
	}
	banding(golden, resp, route)
	sort.Strings(result)
	return result
}

func indexIds(xs []any) map[string]any {
	if len(xs) == 0 {
		return nil
	}
	m := map[string]any{}
	for _, x := range xs {
		o, ok := x.(map[string]any)
		if !ok {
			return nil
		}
		id, ok := o["id"].(string)
		if !ok {
			return nil
		}
		m[id] = o
	}
	return m
}

// cariNilai mengembalikan path JSON yang nilainya sama dengan cari
func findValue(x any, find, route string) []string {
	var result []string
	switch n := x.(type) {
	case map[string]any:
		for k, v := range n {
			result = append(result, findValue(v, find, strings.TrimPrefix(route+"."+k, "."))...)
		}
	case []any:
		for i, v := range n {
			result = append(result, findValue(v, find, fmt.Sprintf("%s[%d]", route, i))...)
		}
	case string:
		if n == find {
			result = append(result, route)
		}
	}
	sort.Strings(result)
	return result
}

func sortIds(x any) {
	if d, ok := x.([]any); ok {
		sort.Slice(d, func(i, j int) bool {
			return fmt.Sprint(d[i].(map[string]any)["id"]) < fmt.Sprint(d[j].(map[string]any)["id"])
		})
	}
}

func TestSummaryBentukResponsTetap(t *testing.T) {
	setupSummary(t)
	content := normalize(t, fetchSummary(t).Mentah)

	if os.Getenv("TULIS_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll(filepath.Dir(berkasGolden), 0o755))
		require.NoError(t, os.WriteFile(berkasGolden, content, 0o644))
		t.Logf("golden ditulis ulang: %s", berkasGolden)
	}
	golden, err := os.ReadFile(berkasGolden)
	require.NoError(t, err, "golden belum ada; jalankan sekali dengan TULIS_GOLDEN=1 pada kode yang benar")
	var g, r any
	require.NoError(t, json.Unmarshal(golden, &g))
	require.NoError(t, json.Unmarshal(content, &r))
	if beda := jsonDiff(g, r, ""); len(beda) > 0 {
		t.Errorf("summary berbeda dari golden (%s):\n  %s", filepath.Base(berkasGolden), strings.Join(beda, "\n  "))
	}
}

func TestSummaryTurunanKunjunganTerhapusTidakMuncul(t *testing.T) {
	setupSummary(t)
	mentah := string(fetchSummary(t).Mentah)
	var pohon any
	require.NoError(t, json.Unmarshal([]byte(mentah), &pohon))

	for _, id := range []string{"uji-opt-kunj-hapus", "uji-opt-obs-hapus", "uji-opt-dg-hapus", "uji-opt-ly-hapus", "uji-opt-rj-hapus"} {
		if di := findValue(pohon, id, ""); len(di) > 0 {
			t.Errorf("%s (milik kunjungan terhapus) muncul di: %s", id, strings.Join(di, ", "))
		}
	}

	var s struct {
		Layanan     []struct{ Id string } `json:"layanan"`
		Rujukan     []struct{ Id string } `json:"rujukan"`
		Menggantung struct {
			Layanan []struct{ Id string } `json:"layanan"`
			Rujukan []struct{ Id string } `json:"rujukan"`
		} `json:"menggantung"`
	}
	require.NoError(t, json.Unmarshal([]byte(mentah), &s))
	ids := func(xs []struct{ Id string }) []string {
		var r []string
		for _, x := range xs {
			r = append(r, x.Id)
		}
		return r
	}
	assert.Contains(t, ids(s.Layanan), "uji-opt-ly-gantung", "layanan menggantung tetap di daftar datar")
	assert.Contains(t, ids(s.Rujukan), "uji-opt-rj-gantung", "rujukan menggantung tetap di daftar datar")
	assert.Contains(t, ids(s.Menggantung.Layanan), "uji-opt-ly-gantung")
	assert.Contains(t, ids(s.Menggantung.Rujukan), "uji-opt-rj-gantung")
}

func TestSummaryDiagnosaMenggantungUrutTanggalCatat(t *testing.T) {
	setupSummary(t)
	var s struct {
		Menggantung struct {
			Diagnosa []struct {
				Id           string `json:"id"`
				TanggalCatat string `json:"tanggal_catat"`
			} `json:"diagnosa"`
		} `json:"menggantung"`
	}
	require.NoError(t, json.Unmarshal(fetchSummary(t).Mentah, &s))

	var urutan []string
	for _, d := range s.Menggantung.Diagnosa {
		urutan = append(urutan, d.Id)
	}
	assert.Equal(t, []string{"uji-opt-dg-2", "uji-opt-dg-3", "uji-opt-dg-1"}, urutan)
}

func TestSummaryKomponenObservasiUrutKode(t *testing.T) {
	setupSummary(t)
	var s struct {
		Menggantung struct {
			Observasi []struct {
				Id        string `json:"id"`
				Component []struct {
					Id string `json:"id"`
				} `json:"component"`
			} `json:"observasi"`
		} `json:"menggantung"`
	}
	require.NoError(t, json.Unmarshal(fetchSummary(t).Mentah, &s))

	for _, o := range s.Menggantung.Observasi {
		if o.Id != "uji-opt-obs-gantung" {
			continue
		}
		require.Len(t, o.Component, 2)
		assert.Equal(t, "uji-opt-obs-gantung-a", o.Component[0].Id, "709261005 sebelum 710999009")
		return
	}
	t.Fatal("observasi menggantung tidak ada di summary")
}

// turunan diambil sekali per tabel: 4 query anak/kunjungan/kesehatan/episode + 4 turunan
func TestSummaryJumlahQueryTetap(t *testing.T) {
	setupSummary(t)

	const pattern = `^\s*SELECT .* FROM "stunting"\.`
	expectQueryMax(t, counter.Measure(func() { fetchSummary(t) }), pattern, 8, "query summary")
	expectQuery(t, counter.Measure(func() { fetchSummary(t) }), pattern, 0, "summary kedua harus dari cache")
}

var daftarKolom = regexp.MustCompile(`(?is)^\s*SELECT\s+.*?\s+FROM\s`)

// kolom SELECT diringkas supaya tabel dan WHERE-nya terlihat
func stringsJoin(q testutil.Statement) string {
	var s string
	for i, x := range q {
		x = daftarKolom.ReplaceAllString(x, "SELECT … FROM ")
		if r := []rune(x); len(r) > 200 {
			x = string(r[:200]) + "…"
		}
		s += fmt.Sprintf("%2d. %s\n", i+1, x)
	}
	return s
}
