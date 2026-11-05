package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
)

func ms(d time.Duration) string { return fmt.Sprintf("%d", d.Milliseconds()) }

func tulisLaporan(l *laporan) string {
	var w strings.Builder
	f := func(format string, a ...any) { fmt.Fprintf(&w, format, a...) }

	f("# Hasil benchmark %s\n\n", l.Mulai.Format("2006-01-02 15:04"))
	f("## Spesifikasi\n\n```\n")
	key := make([]string, 0, len(l.Spek))
	for k := range l.Spek {
		key = append(key, k)
	}
	sort.Strings(key)
	for _, k := range key {
		f("%s=%s\n", k, l.Spek[k])
	}
	f("```\n\n")

	if len(l.HTTP) > 0 {
		f("## HTTP — titik jenuh: **%.0f req/dtk**\n\n", l.JenuhHTTP)
		tabelHTTP(&w, l.HTTP)
		tabelSumber(&w, l.HTTP)
		for _, h := range ujung(l.HTTP) {
			tabelOp(&w, h)
		}
	}
	if len(l.Kafka) > 0 {
		f("## Kafka — titik jenuh: **%.1f pesan/dtk** (%.2f perjalanan anak/dtk)\n\n", l.JenuhKafka, l.JenuhKafka/6)
		f("| target msg/dtk | diproduksi | diproses | diproses/dtk | lag awal→akhir | gagal | status |\n|---|---|---|---|---|---|---|\n")
		for _, h := range l.Kafka {
			k := h.Kafka
			f("| %.1f | %d | %d | %.1f | %d→%d | %d | %s |\n", h.Laju, k.Diproduksi, k.Diproses, k.DiprosesPerDtk, k.LagAwal, k.LagAkhir, k.PesanGagal, ok(h.Jenuh))
		}
		f("\n")
		if k := l.KapasitasKafka; k != nil {
			f("Kapasitas terukur (antrean %d pesan diproses sekaligus): **%.1f pesan/dtk = %.2f perjalanan anak/dtk**, habis dalam %s, gagal %d.\n\n",
				k.Pesan, k.PesanPerDetik, k.AnakPerDetik, k.Durasi.Round(time.Second), k.Gagal)
		}
		tabelSumber(&w, l.Kafka)
	}
	if len(l.Campuran) > 0 {
		h := l.Campuran[0]
		f("## Campuran — HTTP %.0f req/dtk + Kafka %.1f pesan/dtk selama %s\n\n", h.Laju, h.Kafka.LajuTarget, h.Durasi)
		tabelHTTP(&w, l.Campuran)
		k := h.Kafka
		f("Kafka: diproduksi %d, diproses %d (%.1f/dtk), lag %d→%d, gagal %d\n\n", k.Diproduksi, k.Diproses, k.DiprosesPerDtk, k.LagAwal, k.LagAkhir, k.PesanGagal)
		tabelSumber(&w, l.Campuran)
		tabelOp(&w, h)
	}
	return w.String()
}

func tabelHTTP(w *strings.Builder, hs []hasilTahap) {
	fmt.Fprintf(w, "| target req/dtk | tercapai | p50 ms | p95 ms | p99 ms | maks ms | gagal | status |\n|---|---|---|---|---|---|---|---|\n")
	for _, h := range hs {
		fmt.Fprintf(w, "| %.0f | %.0f | %s | %s | %s | %s | %d/%d | %s |\n", h.Laju, h.Tercapai,
			ms(h.Total.P50), ms(h.Total.P95), ms(h.Total.P99), ms(h.Total.Maks), h.Total.Gagal, h.Total.Jumlah, ok(h.Jenuh))
	}
	fmt.Fprintf(w, "\n")
}

// PSI = persen waktu ada task yang tertahan menunggu sumber daya itu
func tabelSumber(w *strings.Builder, hs []hasilTahap) {
	fmt.Fprintf(w, "Sumber daya box per tahap:\n\n")
	fmt.Fprintf(w, "| target | CPU core | CPU throttled %% | PSI cpu %% | PSI io %% (full) | PSI mem %% | IOPS r/w | MB/dtk r/w | memori MB | cache hit %% | blok dibaca/dtk | koneksi aktif (menunggu: lock/io/lwlock) |\n")
	fmt.Fprintf(w, "|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, h := range hs {
		s := h.Sumber
		fmt.Fprintf(w, "| %.1f | %.2f | %.0f | %.0f | %.0f (%.0f) | %.0f | %.0f / %.0f | %.1f / %.1f | %.0f | %.2f | %.0f | %.1f (%.1f/%.1f/%.1f) |\n",
			h.Laju, s.CPUCore, s.CPUThrottledPersen, s.PSICPU, s.PSIIO, s.PSIIOFull, s.PSIMem, s.ReadIOPS, s.WriteIOPS,
			s.ReadMBs, s.WriteMBs, s.MemoriMB, s.CacheHitPersen, s.BlksReadPerDetik, s.KoneksiAktif, s.TungguLock, s.TungguIO, s.TungguLWLock)
	}
	fmt.Fprintf(w, "\n")
}

// rincian per operasi untuk tahap aman terakhir dan tahap jenuh pertama
func ujung(hs []hasilTahap) []hasilTahap {
	var r []hasilTahap
	for i, h := range hs {
		if h.Jenuh != "" {
			if i > 0 {
				r = append(r, hs[i-1])
			}
			return append(r, h)
		}
	}
	if len(hs) > 0 {
		r = append(r, hs[len(hs)-1])
	}
	return r
}

func tabelOp(w *strings.Builder, h hasilTahap) {
	fmt.Fprintf(w, "Per operasi pada %.0f req/dtk (%s):\n\n", h.Laju, ok(h.Jenuh))
	fmt.Fprintf(w, "| operasi | jumlah | p50 ms | p95 ms | p99 ms | maks ms | kode |\n|---|---|---|---|---|---|---|\n")
	name := make([]string, 0, len(h.PerOp))
	for n := range h.PerOp {
		name = append(name, n)
	}
	sort.Strings(name)
	for _, n := range name {
		r := h.PerOp[n]
		var code []string
		ks := make([]int, 0, len(r.Kode))
		for k := range r.Kode {
			ks = append(ks, k)
		}
		slices.Sort(ks)
		for _, k := range ks {
			label := fmt.Sprint(k)
			switch k {
			case 0:
				label = "timeout/putus"
			case -1:
				label = "antre-klien"
			}
			code = append(code, fmt.Sprintf("%s×%d", label, r.Kode[k]))
		}
		fmt.Fprintf(w, "| %s | %d | %s | %s | %s | %s | %s |\n", n, r.Jumlah, ms(r.P50), ms(r.P95), ms(r.P99), ms(r.Maks), strings.Join(code, " "))
	}
	if len(h.Contoh) > 0 {
		fmt.Fprintf(w, "\nContoh respons gagal:\n\n")
		for _, n := range name {
			if c, exists := h.Contoh[n]; exists {
				fmt.Fprintf(w, "- `%s`: %s\n", n, c)
			}
		}
	}
	fmt.Fprintf(w, "\n")
}
