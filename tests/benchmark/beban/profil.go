package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type statement struct {
	Query string
	Calls float64
	Ms    float64
}

type profilOp struct {
	Nama       string
	Statement  float64
	Transaksi  float64
	Gagal      int
	Rincian    []statement
	LatensiP50 time.Duration
}

// tiap operasi dijalankan berurutan dan dihitung statement-nya lewat pg_stat_statements
func (b *bench) modeProfil() error {
	if b.db == nil {
		return fmt.Errorf("mode profil butuh -pg")
	}
	if _, err := b.db.Exec(`CREATE EXTENSION IF NOT EXISTS pg_stat_statements`); err != nil {
		return fmt.Errorf("pg_stat_statements belum dimuat (PG_PRELOAD): %w", err)
	}
	n := envInt("PROFIL_ULANG", 20)
	all := daftarOperasi(func(string) int { return 1 })
	var result []profilOp
	for _, op := range all {
		for range 3 {
			b.k.kirim(op.buat(b.p))
		}
		time.Sleep(time.Second)
		if _, err := b.db.Exec(`SELECT pg_stat_statements_reset()`); err != nil {
			return err
		}
		po := profilOp{Nama: op.nama}
		var lat []time.Duration
		for range n {
			start := time.Now()
			code, content := b.k.kirim(op.buat(b.p))
			lat = append(lat, time.Since(start))
			if code < 200 || code >= 300 {
				po.Gagal++
				b.lap.Spek["contoh_gagal_"+op.nama] = potong(content, 200)
			}
		}
		// beri waktu job latar (hapus cache dsb.) selesai
		time.Sleep(2 * time.Second)
		rows, err := b.db.Query(`SELECT calls, total_exec_time, regexp_replace(query, '\s+', ' ', 'g')
			FROM pg_stat_statements s JOIN pg_database d ON d.oid = s.dbid
			WHERE d.datname = 'ildki' AND query NOT ILIKE '%pg_stat_statements%'
			ORDER BY calls DESC, total_exec_time DESC`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var s statement
			if err := rows.Scan(&s.Calls, &s.Ms, &s.Query); err != nil {
				rows.Close()
				return err
			}
			s.Calls /= float64(n)
			s.Ms /= float64(n)
			q := strings.ToUpper(strings.TrimSpace(s.Query))
			if strings.HasPrefix(q, "BEGIN") {
				po.Transaksi += s.Calls
			}
			po.Statement += s.Calls
			po.Rincian = append(po.Rincian, s)
		}
		rows.Close()
		po.LatensiP50 = ringkasan(lat, map[int]int{200: len(lat)}).P50
		result = append(result, po)
		fmt.Printf("%-22s %5.1f statement/request, %4.1f transaksi, p50 %s, gagal %d\n", po.Nama, po.Statement, po.Transaksi, po.LatensiP50.Round(time.Millisecond), po.Gagal)
	}
	return tulisProfil(b.hasil, result)
}

func tulisProfil(dir string, hs []profilOp) error {
	var w strings.Builder
	fmt.Fprintf(&w, "# Profil statement per operasi\n\n| operasi | statement/request | transaksi | p50 ms | gagal |\n|---|---|---|---|---|\n")
	for _, h := range hs {
		fmt.Fprintf(&w, "| %s | %.1f | %.1f | %d | %d |\n", h.Nama, h.Statement, h.Transaksi, h.LatensiP50.Milliseconds(), h.Gagal)
	}
	for _, h := range hs {
		fmt.Fprintf(&w, "\n## %s\n\n| per request | ms | statement |\n|---|---|---|\n", h.Nama)
		for _, s := range h.Rincian {
			fmt.Fprintf(&w, "| %.2f | %.3f | `%s` |\n", s.Calls, s.Ms, potong(s.Query, 220))
		}
	}
	return os.WriteFile(filepath.Join(dir, "profil.md"), []byte(w.String()), 0o644)
}
