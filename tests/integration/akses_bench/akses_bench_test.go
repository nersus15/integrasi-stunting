package akses_bench_test

import (
	"testing"

	"github.com/nersus15/integrasi/mod-stunting/service"
)

const (
	orgPkm = "uji-bench-pkm"
	orgRs  = "uji-bench-rs"
	ortu   = "uji-bench-ortu"
	anak1  = "uji-bench-anak-1" // belum dirujuk: dipegang puskesmas posyandu
	anak2  = "uji-bench-anak-2" // dirujuk ke RS
)

// cache data transaksi yang dibaca cek akses, untuk semua versi; master faskes/posyandu tetap hangat
func forgetTransaksi() {
	keys := []string{"OrangtuaByID::" + ortu, "LIST:ANAK:ORTU:" + ortu, "LIST:IDANAK:ORTU:" + ortu}
	for _, a := range []string{anak1, anak2} {
		keys = append(keys, "ANAK:ID:"+a, "RUJUKAN:TERAKHIR:"+a, "PEMBUAT::anak::"+a)
		for _, o := range []string{orgPkm, orgRs} {
			keys = append(keys, "AKSESKLINIS:"+o+":"+a+":", "AKSESKLINIS:"+o+"::"+ortu)
		}
	}
	for _, o := range []string{orgPkm, orgRs} {
		keys = append(keys, "AKSESKLINIS:"+o+"::"+ortu)
	}
	for _, k := range keys {
		_ = wiring.Memory.Delete(k)
	}
}

type kasus struct {
	nama string
	cek  func() error
}

var semuaKasus = []kasus{
	{"anak/puskesmas-posyandu", func() error { return wiring.Service.AksesAnak(orgPkm, service.AksiBaca, anak1) }},
	{"anak/rs-tujuan-rujukan", func() error { return wiring.Service.AksesAnak(orgRs, service.AksiBaca, anak2) }},
	{"orangtua/puskesmas-posyandu", func() error { return wiring.Service.AksesOrangtua(orgPkm, service.AksiBaca, ortu) }},
	{"orangtua/rs-tujuan-rujukan", func() error { return wiring.Service.AksesOrangtua(orgRs, service.AksiBaca, ortu) }},
}

// jumlah statement DB dan Get cache per cek, dihitung dari 20 cek di luar waktu ukur;
// dilaporkan setelah b.Loop karena Loop mereset metrik tambahan
func hitungRoundTrip(b *testing.B, k kasus, dingin bool) (query, get float64) {
	b.Helper()
	const n = 20
	gets0, _ := memory.CacheGets()
	var statement int
	for range n {
		if dingin {
			forgetTransaksi()
		}
		statement += len(counter.Measure(func() {
			if err := k.cek(); err != nil {
				b.Fatalf("%s: %v", k.nama, err)
			}
		}))
	}
	gets1, _ := memory.CacheGets()
	return float64(statement) / n, float64(gets1-gets0) / n
}

func BenchmarkAkses(b *testing.B) {
	for _, k := range semuaKasus {
		b.Run(k.nama+"/hangat", func(b *testing.B) {
			forgetTransaksi()
			if err := k.cek(); err != nil {
				b.Fatalf("%s: %v", k.nama, err)
			}
			query, get := hitungRoundTrip(b, k, false)
			for b.Loop() {
				if err := k.cek(); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(query, "query/op")
			b.ReportMetric(get, "cache-get/op")
		})
		b.Run(k.nama+"/dingin", func(b *testing.B) {
			query, get := hitungRoundTrip(b, k, true)
			for b.Loop() {
				b.StopTimer()
				forgetTransaksi()
				b.StartTimer()
				if err := k.cek(); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(query, "query/op")
			b.ReportMetric(get, "cache-get/op")
		})
	}
}
