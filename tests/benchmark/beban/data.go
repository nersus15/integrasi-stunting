package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"
)

// Pola id harus sama dengan box/seed.sql.

type populasi struct {
	anak, pkm, rs, persenRujukan int
	kunciJakantro, kunciPrefix   string
	urut                         atomic.Int64
}

func (p *populasi) acak() int { return rand.IntN(p.anak) + 1 }

func idOrangtua(i int) string  { return fmt.Sprintf("00000000-0000-4000-a000-%012d", i) }
func idAnak(i int) string      { return fmt.Sprintf("00000000-0000-4000-b000-%012d", i) }
func nikAnak(i int) string     { return fmt.Sprintf("33%014d", i) }
func nikOrangtua(i int) string { return fmt.Sprintf("31%014d", i) }
func ihsAnak(i int) string     { return fmt.Sprintf("P%011d", i) }
func encSeed(i int) string     { return fmt.Sprintf("E%011d", i) }

func (p *populasi) orgPkm(i int) string { return fmt.Sprintf("9%05d", i%p.pkm+1) }
func (p *populasi) orgRs(i int) string  { return fmt.Sprintf("8%05d", i%p.rs+1) }
func (p *populasi) adaRujukan(i int) bool {
	return i%100 < p.persenRujukan
}

// pemegang akses faskes: RS tujuan rujukan, selain itu puskesmas posyandunya
func (p *populasi) orgPemegang(i int) string {
	if p.adaRujukan(i) {
		return p.orgRs(i)
	}
	return p.orgPkm(i)
}
func (p *populasi) posyandu(i int) string {
	return fmt.Sprintf("BPOS-%d-%d", i%p.pkm+1, i%5+1)
}
func (p *populasi) kunciFaskes(org string) string { return p.kunciPrefix + org }

// nomor unik untuk entitas baru selama satu run; volume selalu segar per run
func (p *populasi) baru() int64 { return p.urut.Add(1) }

func uuidBaru(jenis string, n int64) string {
	return fmt.Sprintf("10000000-%s-4000-8000-%012d", jenis, n)
}

func tanggalLahir(i int) string {
	return fmt.Sprintf("%d-%02d-%02d", 2022+i%3, i%12+1, i%28+1)
}

func tanggalAcak() string {
	return fmt.Sprintf("2026-%02d-%02d", rand.IntN(9)+1, rand.IntN(28)+1)
}

func kunjunganIsi(id, idAnak string) map[string]any {
	k := map[string]any{
		"tanggal_pengukuran": tanggalAcak(), "cara_ukur": "telentang",
		"berat_badan": 7 + rand.Float64()*4, "tinggi_badan": 68 + rand.Float64()*12,
		"lingkar_lengan": 12.5, "lingkar_kepala": 44.1, "lingkar_dada": 46.0,
		"asi_bulan_0": 1, "asi_bulan_1": 1, "asi_bulan_2": 1, "asi_bulan_3": 1,
		"asi_bulan_4": 1, "asi_bulan_5": 0, "asi_bulan_6": 0,
		"vit_biru": 1, "vit_merah": 0, "pitting_edema": 0, "kelas_ibu_balita": 1,
		"status_bbu": "Berat Badan Kurang", "zscore_bbu": -2.4,
		"status_tbu": "Pendek", "zscore_tbu": -2.6,
		"status_bbtb": "Gizi Baik", "zscore_bbtb": -0.9,
	}
	if id != "" {
		k["id"] = id
	}
	if idAnak != "" {
		k["id_anak"] = idAnak
	}
	return k
}

func (p *populasi) registrasi() map[string]any {
	n := p.baru()
	i := p.acak()
	ortu, anak := uuidBaru("0001", n), uuidBaru("0002", n)
	return map[string]any{
		"orangtua": map[string]any{
			"id": ortu, "id_posyandu": p.posyandu(i), "no_kk": fmt.Sprintf("35%014d", n),
			"nik": fmt.Sprintf("34%014d", n), "nama_ayah": "Ayah Baru", "nama_ibu": "Ibu Baru",
			"telepon": "081200000000", "rt": "001", "rw": "002", "alamat": "Jl. Baru", "kia": 1,
		},
		"anak": map[string]any{
			"id": anak, "id_orangtua": ortu, "nik": fmt.Sprintf("36%014d", n), "nama": "Anak Baru",
			"tanggal_lahir": tanggalLahir(int(n)), "jenis_kelamin": "L", "anak_ke": 1, "imd": 1,
			"bb_lahir": 3.1, "tb_lahir": 49.0, "lk_lahir": 34.0,
			"source_data": "20f11c5c-5a29-11f0-a136-5749ce66d036",
		},
		"kunjungan": kunjunganIsi(uuidBaru("0003", n), anak),
	}
}

func (p *populasi) kunjunganSaja(i int) map[string]any {
	return map[string]any{"kunjungan": kunjunganIsi(uuidBaru("0004", p.baru()), idAnak(i))}
}

func (p *populasi) kesehatan(i int) map[string]any {
	return map[string]any{
		"id": uuidBaru("0005", p.baru()), "id_anak": idAnak(i), "tanggal_pemantauan": tanggalAcak(),
		"tbc_batuk": 0, "tbc_demam": 0, "tbc_bb": 0, "tbc_kontak": 0,
		"layanan_asi_eks": 1, "layanan_imunisasi": 1, "layanan_vit_a": 1, "penyuluhan_edukasi": 1,
	}
}

func putAnak(i int) map[string]any {
	return map[string]any{
		"id": idAnak(i), "nik": nikAnak(i), "nama": fmt.Sprintf("Anak %d Ubah", i),
		"tanggal_lahir": "2023-01-01", "jenis_kelamin": map[bool]string{true: "L", false: "P"}[i%2 == 0],
	}
}

func putOrangtua(i int) map[string]any {
	return map[string]any{
		"id": idOrangtua(i), "nik": nikOrangtua(i), "nama_ibu": fmt.Sprintf("Ibu %d", i),
		"telepon": fmt.Sprintf("0813%08d", rand.IntN(1e8)),
	}
}

func (p *populasi) pemeriksaan(i int) map[string]any {
	n := p.baru()
	enc := fmt.Sprintf("enc-b-%d", n)
	tgl := tanggalAcak()
	k := kunjunganIsi("", "")
	k["id_satusehat"] = enc
	k["tanggal_pengukuran"] = tgl
	obs := func(suf, code, display string, value float64, satuan string) map[string]any {
		return map[string]any{
			"id_satusehat": fmt.Sprintf("obs-b-%d-%s", n, suf), "system": "http://loinc.org",
			"kode": code, "display": display, "kategori": "vital-signs",
			"nilai_angka": value, "satuan": satuan, "tanggal": tgl + "T02:10:00Z",
		}
	}
	body := map[string]any{
		"anak": map[string]any{
			"id_satusehat": ihsAnak(i), "nik": nikAnak(i), "nama": fmt.Sprintf("Anak %d", i),
			"tanggal_lahir": "2023-01-01", "jenis_kelamin": map[bool]string{true: "L", false: "P"}[i%2 == 0], "anak_ke": 1,
		},
		"kunjungan": k,
		"observasi": []any{
			obs("bb", "29463-7", "Body weight", k["berat_badan"].(float64), "kg"),
			obs("tb", "8306-3", "Body height (lying)", k["tinggi_badan"].(float64), "cm"),
			obs("lk", "9843-4", "Head circumference", 44.1, "cm"),
		},
		"diagnosa": []any{map[string]any{
			"id_satusehat": fmt.Sprintf("cond-b-%d", n), "jenis": "diagnosis",
			"system": "http://hl7.org/fhir/sid/icd-10", "kode": "E45", "display": "Nutritional stunting",
			"kategori": "encounter-diagnosis", "clinical_status": "active", "verification_status": "confirmed",
			"tanggal_catat": tgl,
		}},
		"layanan": []any{map[string]any{
			"id_satusehat": fmt.Sprintf("proc-b-%d", n), "jenis": "procedure", "system": "http://snomed.info/sct",
			"kode": "441041000124100", "display": "Counseling about nutrition", "status": "completed",
			"tanggal": tgl + "T03:00:00Z",
		}},
	}
	if rand.IntN(10) == 0 {
		body["rujukan"] = []any{map[string]any{
			"id_satusehat": fmt.Sprintf("sr-b-%d", n), "jenis": "rujukan",
			// tujuan = pengirim supaya pemegang akses tidak berubah selama beban
			"ref_faskes_asal": p.orgPemegang(i), "ref_faskes_tujuan": p.orgPemegang(i),
			"system": "http://snomed.info/sct", "kode": "737481003", "display": "Inpatient care management",
			"status": "active", "prioritas": "routine", "tanggal": tgl + "T04:00:00Z",
		}}
	}
	return body
}

func putKunjunganFaskes(i int) map[string]any {
	k := kunjunganIsi("", "")
	delete(k, "status_tbu")
	delete(k, "zscore_tbu")
	k["id_satusehat"] = encSeed(i)
	return k
}

// --- Kafka: perjalanan satu anak dari simulasi/skenario1 dengan id unik ---

type perjalanan struct {
	pesan  []string
	tokens []string
}

var polaToken = regexp.MustCompile(`\b(?:cond|enc|obs|sr|md|proc|no)-[a-z]+-[a-z0-9-]*[0-9]\b|\bsim1-[a-z0-9-]+\b|\bep-stunting-sim1\b`)

func muatPerjalanan(root string) (*perjalanan, error) {
	file := []string{
		"02-kafka-patient.json", "03-kafka-relatedperson.json", "04-kafka-puskesmas.json",
		"05-kafka-rs.json", "06-kafka-rujuk-balik.json", "07-kafka-puskesmas-lanjutan.json",
	}
	pj := &perjalanan{}
	unik := map[string]bool{}
	for _, b := range file {
		content, err := os.ReadFile(filepath.Join(root, "simulasi", "skenario1", b))
		if err != nil {
			return nil, err
		}
		var check map[string]any
		if err := json.Unmarshal(content, &check); err != nil {
			return nil, fmt.Errorf("%s: %w", b, err)
		}
		pj.pesan = append(pj.pesan, string(content))
		for _, t := range polaToken.FindAllString(string(content), -1) {
			unik[t] = true
		}
	}
	for t := range unik {
		pj.tokens = append(pj.tokens, t)
	}
	sort.Slice(pj.tokens, func(a, b int) bool { return len(pj.tokens[a]) > len(pj.tokens[b]) })
	return pj, nil
}

// kunci partisi = IHS anak supaya urutan pesan satu anak terjaga
func (pj *perjalanan) untuk(n int64, pkm, rs string) (string, []string) {
	ihs := fmt.Sprintf("K%011d", n)
	ganti := []string{
		"P99001100022", ihs,
		"R99001100033", fmt.Sprintf("Q%011d", n),
		"3175042003240007", fmt.Sprintf("37%014d", n),
		"3175040202900008", fmt.Sprintf("38%014d", n),
		"Organization/100011", "Organization/" + pkm,
		"Organization/200022", "Organization/" + rs,
	}
	for _, t := range pj.tokens {
		ganti = append(ganti, t, fmt.Sprintf("%s-k%d", t, n))
	}
	r := strings.NewReplacer(ganti...)
	result := make([]string, len(pj.pesan))
	for i, m := range pj.pesan {
		result[i] = r.Replace(m)
	}
	return ihs, result
}
