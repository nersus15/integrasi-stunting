// Generator beban untuk tests/benchmark. Dijalankan lewat ../jalankan.sh.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/uptrace/bun/driver/pgdriver"
)

func env(n, bawaan string) string {
	if v := os.Getenv(n); v != "" {
		return v
	}
	return bawaan
}
func envInt(n string, b int) int {
	v, err := strconv.Atoi(env(n, ""))
	if err != nil {
		return b
	}
	return v
}
func envF(n string, b float64) float64 {
	v, err := strconv.ParseFloat(env(n, ""), 64)
	if err != nil {
		return b
	}
	return v
}
func envDur(n string, b time.Duration) time.Duration {
	v, err := time.ParseDuration(env(n, ""))
	if err != nil {
		return b
	}
	return v
}

type laporan struct {
	Mulai                 time.Time
	Spek                  map[string]string
	HTTP                  []hasilTahap `json:",omitempty"`
	Kafka                 []hasilTahap `json:",omitempty"`
	Campuran              []hasilTahap `json:",omitempty"`
	JenuhHTTP, JenuhKafka float64
	KapasitasKafka        *kapasitasKafka `json:",omitempty"`
}

type bench struct {
	ctx      context.Context
	k        *klienHTTP
	p        *populasi
	ops      []operasi
	db       *sql.DB
	cg       *cgroup
	kaf      *kafkaBench
	pj       *perjalanan
	lap      *laporan
	hasil    string
	inflight int
	p95      time.Duration
	errMaks  float64
}

func main() {
	mode := flag.String("mode", "http", "http|kafka|campuran|semua|profil")
	dirHasil := flag.String("hasil", ".", "direktori hasil")
	boxID := flag.String("box", "", "id container box (untuk cgroup)")
	base := flag.String("base", "http://127.0.0.1:22025/api", "base URL API")
	dsn := flag.String("pg", "", "DSN postgres box")
	siapTopic := flag.Bool("siapkan-topic", false, "buat ulang topic lalu keluar")
	remove := flag.Bool("hapus-topic", false, "hapus topic lalu keluar")
	flag.Parse()

	broker, topic := env("KAFKA_BROKERS", "localhost:1092"), env("KAFKA_TOPIC", "bench-stunting-fhir")
	if *siapTopic {
		if err := siapkanTopic(broker, topic, envInt("KAFKA_PARTISI", 6)); err != nil {
			log.Fatal(err)
		}
		return
	}
	if *remove {
		if err := hapusTopic(broker, topic); err != nil {
			log.Fatal(err)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	b := &bench{
		ctx: ctx, hasil: *dirHasil,
		inflight: envInt("MAKS_INFLIGHT", 1000),
		p95:      time.Duration(envInt("BATAS_P95_MS", 1000)) * time.Millisecond,
		errMaks:  envF("BATAS_ERROR_PERSEN", 1),
		p: &populasi{
			anak: envInt("SEED_ANAK", 300000), pkm: envInt("SEED_PUSKESMAS", 40), rs: envInt("SEED_RS", 10),
			persenRujukan: envInt("SEED_PERSEN_RUJUKAN", 10),
			kunciJakantro: env("KUNCI_JAKANTRO", ""), kunciPrefix: env("KUNCI_FASKES_PREFIX", ""),
		},
		lap: &laporan{Mulai: time.Now(), Spek: map[string]string{}},
	}
	// id entitas baru tiap run tidak bentrok dengan run sebelumnya, walau DB/cache server tidak segar
	b.p.urut.Store(time.Now().Unix() % 1e8 * 1000)
	for _, kv := range os.Environ() {
		k, v, _ := strings.Cut(kv, "=")
		for _, pre := range []string{"SERVER_", "DISK_", "PG_", "DB_", "APP_LOG", "RATE_", "SEED_", "HTTP_", "KAFKA_", "DURASI", "MAKS_", "BATAS_", "CAMPURAN_", "BOBOT_"} {
			if strings.HasPrefix(k, pre) && !strings.Contains(k, "KUNCI") {
				b.lap.Spek[k] = v
			}
		}
	}
	b.k = klienBaru(*base, envDur("HTTP_TIMEOUT", 30*time.Second), b.inflight)
	b.ops = daftarOperasi(func(n string) int { return envInt(n, 0) })
	if *dsn != "" {
		b.db = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(*dsn)))
		b.db.SetMaxOpenConns(2)
		defer b.db.Close()
	}
	if *boxID != "" {
		cg, err := cgroupBox(*boxID, env("DISK_DEVICE", "/dev/nvme0n1"))
		if err != nil {
			log.Printf("cgroup box tidak terbaca, metrik sumber daya kosong: %v", err)
		}
		b.cg = cg
	}

	var err error
	switch *mode {
	case "http":
		_, err = b.modeHTTP()
	case "kafka":
		_, err = b.modeKafka()
	case "profil":
		err = b.modeProfil()
	case "campuran":
		err = b.modeCampuran(envF("CAMPURAN_LAJU_HTTP", 0), envF("CAMPURAN_LAJU_KAFKA", 0))
	case "semua":
		var h, k float64
		if h, err = b.modeHTTP(); err == nil {
			if k, err = b.modeKafka(); err == nil {
				persen := envF("CAMPURAN_PERSEN", 70) / 100
				err = b.modeCampuran(h*persen, k*persen)
			}
		}
	default:
		err = fmt.Errorf("mode tidak dikenal: %s", *mode)
	}
	b.tulis()
	if err != nil {
		log.Fatal(err)
	}
}

// mengembalikan laju tertinggi yang masih memenuhi batas
func (b *bench) modeHTTP() (float64, error) {
	if err := cekOperasi(b.k, b.p, b.ops); err != nil {
		return 0, err
	}
	pil := pemilihBaru(b.ops)
	durasi := envDur("DURASI_TAHAP", time.Minute)
	aman := 0.0
	for laju := envF("HTTP_LAJU_AWAL", 10); laju <= envF("HTTP_LAJU_MAKS", 2000) && b.ctx.Err() == nil; laju *= envF("HTTP_FAKTOR", 1.5) {
		c := catatanBaru()
		m := pantau(b.cg, b.db)
		tercapai := bebanHTTP(b.ctx, b.k, b.p, pil, laju, durasi, b.inflight, c)
		h := hasilTahap{Laju: laju, Tercapai: tercapai, Durasi: durasi, Total: c.total(), PerOp: c.perOp(), Contoh: c.contoh, Sumber: m.berhenti()}
		h.Jenuh = b.jenuhHTTP(h)
		b.lap.HTTP = append(b.lap.HTTP, h)
		cetak("http", h)
		b.tulis()
		if h.Jenuh != "" {
			break
		}
		aman = laju
		time.Sleep(5 * time.Second)
	}
	b.lap.JenuhHTTP = aman
	return aman, nil
}

func (b *bench) jenuhHTTP(h hasilTahap) string {
	var reason []string
	if h.Total.Jumlah > 0 {
		if e := 100 * float64(h.Total.Gagal) / float64(h.Total.Jumlah); e > b.errMaks {
			reason = append(reason, fmt.Sprintf("error %.1f%%", e))
		}
	}
	if h.Total.P95 > b.p95 {
		reason = append(reason, fmt.Sprintf("p95 %s", h.Total.P95.Round(time.Millisecond)))
	}
	if h.Tercapai < 0.9*h.Laju {
		reason = append(reason, fmt.Sprintf("hanya %.0f/%.0f req/dtk selesai", h.Tercapai, h.Laju))
	}
	return strings.Join(reason, ", ")
}

func (b *bench) siapkanKafka() error {
	if b.kaf != nil {
		return nil
	}
	pj, err := muatPerjalanan(filepath.Join(env("ROOT", ".."), ""))
	if err != nil {
		return err
	}
	b.pj = pj
	b.kaf, err = kafkaBaru(env("KAFKA_BROKERS", "localhost:1092"), env("KAFKA_TOPIC", "bench-stunting-fhir"), env("KAFKA_GROUP", "bench-stunting"), envInt("KAFKA_PARTISI", 6))
	if err != nil {
		return err
	}
	// satu perjalanan harus tuntas tanpa gagal sebelum beban
	failed := hitungGagal(b.db, b.kaf.group)
	ihs, message := b.pj.untuk(b.p.baru(), b.p.orgPkm(1), b.p.orgRs(1))
	if err := b.kaf.kirim(ihs, message); err != nil {
		return err
	}
	b.kaf.prod.Flush(10000)
	if _, remaining := b.kaf.tungguKosong(90 * time.Second); remaining > 0 {
		return fmt.Errorf("consumer app tidak memproses pesan pemanasan (lag %d)", remaining)
	}
	if g := hitungGagal(b.db, b.kaf.group) - failed; g > 0 {
		return fmt.Errorf("%d pesan pemanasan gagal diproses, cek kafka.transactions", g)
	}
	return nil
}

func (b *bench) modeKafka() (float64, error) {
	if err := b.siapkanKafka(); err != nil {
		return 0, err
	}
	durasi := envDur("DURASI_TAHAP", time.Minute)
	aman := 0.0
	for laju := envF("KAFKA_LAJU_AWAL", 2); laju <= envF("KAFKA_LAJU_MAKS", 500) && b.ctx.Err() == nil; laju *= envF("KAFKA_FAKTOR", 1.5) {
		m := pantau(b.cg, b.db)
		hk, err := bebanKafka(b.ctx, b.kaf, b.pj, b.p, laju, durasi, b.db)
		if err != nil {
			return aman, err
		}
		h := hasilTahap{Laju: laju, Tercapai: hk.DiprosesPerDtk, Durasi: durasi, Kafka: hk, Sumber: m.berhenti()}
		h.Jenuh = jenuhKafka(hk)
		old, remaining := b.kaf.tungguKosong(5 * time.Minute)
		if remaining > 0 && h.Jenuh == "" {
			h.Jenuh = fmt.Sprintf("antrean %d belum habis setelah %s", remaining, old.Round(time.Second))
		}
		b.lap.Kafka = append(b.lap.Kafka, h)
		cetak("kafka", h)
		b.tulis()
		if h.Jenuh != "" {
			break
		}
		aman = laju
	}
	b.lap.JenuhKafka = aman
	if b.ctx.Err() == nil {
		// antrean minimal ~60 detik supaya interval auto-commit 5 detik tidak mendominasi
		n := envInt("KAFKA_BURST_ANAK", 200)
		if len(b.lap.Kafka) > 0 {
			n = max(n, int(b.lap.Kafka[len(b.lap.Kafka)-1].Tercapai*60)/len(b.pj.pesan))
		}
		log.Printf("[kafka] kapasitas: %d perjalanan (%d pesan) sekaligus", n, n*len(b.pj.pesan))
		kap, err := ukurKapasitasKafka(b.kaf, b.pj, b.p, n, b.db)
		if err != nil {
			return aman, err
		}
		b.lap.KapasitasKafka = kap
		log.Printf("[kafka] kapasitas %.1f pesan/dtk (%.2f anak/dtk), habis dalam %s, gagal %d", kap.PesanPerDetik, kap.AnakPerDetik, kap.Durasi.Round(time.Second), kap.Gagal)
		b.tulis()
	}
	return aman, nil
}

func jenuhKafka(h *hasilKafka) string {
	var reason []string
	if h.PesanGagal > 0 {
		reason = append(reason, fmt.Sprintf("%d pesan gagal", h.PesanGagal))
	}
	if naik := h.LagAkhir - h.LagAwal; naik > max(12, h.Diproduksi/20) {
		reason = append(reason, fmt.Sprintf("lag naik %d", naik))
	}
	return strings.Join(reason, ", ")
}

func (b *bench) modeCampuran(lajuHTTP, lajuKafka float64) error {
	if lajuHTTP <= 0 || lajuKafka <= 0 {
		return fmt.Errorf("campuran butuh laju HTTP dan Kafka > 0 (CAMPURAN_LAJU_HTTP, CAMPURAN_LAJU_KAFKA, atau mode semua)")
	}
	if err := b.siapkanKafka(); err != nil {
		return err
	}
	durasi := envDur("DURASI_CAMPURAN", 5*time.Minute)
	c := catatanBaru()
	m := pantau(b.cg, b.db)
	var wg sync.WaitGroup
	var hk *hasilKafka
	var errK error
	wg.Add(1)
	go func() {
		defer wg.Done()
		hk, errK = bebanKafka(b.ctx, b.kaf, b.pj, b.p, lajuKafka, durasi, b.db)
	}()
	tercapai := bebanHTTP(b.ctx, b.k, b.p, pemilihBaru(b.ops), lajuHTTP, durasi, b.inflight, c)
	wg.Wait()
	if errK != nil {
		return errK
	}
	h := hasilTahap{Laju: lajuHTTP, Tercapai: tercapai, Durasi: durasi, Total: c.total(), PerOp: c.perOp(), Contoh: c.contoh, Kafka: hk, Sumber: m.berhenti()}
	h.Jenuh = strings.Trim(b.jenuhHTTP(h)+", "+jenuhKafka(hk), ", ")
	b.lap.Campuran = append(b.lap.Campuran, h)
	cetak("campuran", h)
	return nil
}

func cetak(mode string, h hasilTahap) {
	s := h.Sumber
	if h.Kafka != nil && mode == "kafka" {
		log.Printf("[%s] target %.1f msg/dtk | diproses %.1f msg/dtk | lag %d→%d | gagal %d | CPU %.2f core | IO psi %.0f%% | %s",
			mode, h.Laju, h.Tercapai, h.Kafka.LagAwal, h.Kafka.LagAkhir, h.Kafka.PesanGagal, s.CPUCore, s.PSIIO, ok(h.Jenuh))
		return
	}
	log.Printf("[%s] target %.0f req/dtk | tercapai %.0f | p50 %s p95 %s p99 %s | gagal %d/%d | CPU %.2f core thr %.0f%% | psi cpu %.0f%% io %.0f%% | IOPS r%.0f w%.0f | %s",
		mode, h.Laju, h.Tercapai, h.Total.P50.Round(time.Millisecond), h.Total.P95.Round(time.Millisecond), h.Total.P99.Round(time.Millisecond),
		h.Total.Gagal, h.Total.Jumlah, s.CPUCore, s.CPUThrottledPersen, s.PSICPU, s.PSIIO, s.ReadIOPS, s.WriteIOPS, ok(h.Jenuh))
}

func ok(jenuh string) string {
	if jenuh == "" {
		return "OK"
	}
	return "JENUH: " + jenuh
}

func (b *bench) tulis() {
	if b.hasil == "" {
		return
	}
	j, _ := json.MarshalIndent(b.lap, "", "  ")
	_ = os.WriteFile(filepath.Join(b.hasil, "hasil.json"), j, 0o644)
	_ = os.WriteFile(filepath.Join(b.hasil, "laporan.md"), []byte(tulisLaporan(b.lap)), 0o644)
}
