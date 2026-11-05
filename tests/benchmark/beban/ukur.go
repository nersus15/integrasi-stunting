package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type catatan struct {
	mu      sync.Mutex
	latensi map[string][]time.Duration
	kode    map[string]map[int]int
	contoh  map[string]string
}

func catatanBaru() *catatan {
	return &catatan{latensi: map[string][]time.Duration{}, kode: map[string]map[int]int{}, contoh: map[string]string{}}
}

// kode 0 = gagal transport/timeout, -1 = tidak terkirim karena inflight penuh
func (c *catatan) tambah(op string, d time.Duration, code int, content string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if code >= 200 && code < 300 {
		c.latensi[op] = append(c.latensi[op], d)
	} else if _, exists := c.contoh[op]; !exists && code > 0 {
		c.contoh[op] = fmt.Sprintf("%d %s", code, potong(content, 300))
	}
	if c.kode[op] == nil {
		c.kode[op] = map[int]int{}
	}
	c.kode[op][code]++
}

type ringkas struct {
	Jumlah, Sukses, Gagal int
	P50, P95, P99, Maks   time.Duration
	Kode                  map[int]int
}

func ringkasan(lat []time.Duration, code map[int]int) ringkas {
	r := ringkas{Kode: code}
	for k, n := range code {
		r.Jumlah += n
		if k >= 200 && k < 300 {
			r.Sukses += n
		} else {
			r.Gagal += n
		}
	}
	if len(lat) == 0 {
		return r
	}
	s := slices.Clone(lat)
	slices.Sort(s)
	q := func(p float64) time.Duration { return s[min(len(s)-1, int(p*float64(len(s))))] }
	r.P50, r.P95, r.P99, r.Maks = q(0.50), q(0.95), q(0.99), s[len(s)-1]
	return r
}

func (c *catatan) total() ringkas {
	c.mu.Lock()
	defer c.mu.Unlock()
	var all []time.Duration
	code := map[int]int{}
	for op, l := range c.latensi {
		all = append(all, l...)
		_ = op
	}
	for _, m := range c.kode {
		for k, n := range m {
			code[k] += n
		}
	}
	return ringkasan(all, code)
}

func (c *catatan) perOp() map[string]ringkas {
	c.mu.Lock()
	defer c.mu.Unlock()
	h := map[string]ringkas{}
	for op, m := range c.kode {
		h[op] = ringkasan(c.latensi[op], m)
	}
	return h
}

func potong(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

// --- cgroup box ---

type cgroup struct{ dir, dev string }

func cgroupBox(id, device string) (*cgroup, error) {
	dir := filepath.Join("/sys/fs/cgroup/system.slice", "docker-"+id+".scope")
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}
	dev := ""
	if fi, err := os.Stat(filepath.Join("/sys/class/block", filepath.Base(device), "dev")); err == nil && !fi.IsDir() {
		b, _ := os.ReadFile(filepath.Join("/sys/class/block", filepath.Base(device), "dev"))
		dev = strings.TrimSpace(string(b))
	}
	return &cgroup{dir: dir, dev: dev}, nil
}

type potret struct {
	waktu                              time.Time
	cpuUsec, throttledUsec             float64
	psiCPU, psiIO, psiIOFull           float64
	psiMem                             float64
	rios, wios, rbytes, wbytes         float64
	memori                             float64
	blksRead, blksHit, xact            float64
	koneksiAktif, koneksiTunggu        float64
	tungguLock, tungguIO, tungguLWLock float64
}

func bacaKV(path string) map[string]float64 {
	m := map[string]float64{}
	f, err := os.Open(path)
	if err != nil {
		return m
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 2 {
			v, _ := strconv.ParseFloat(f[1], 64)
			m[f[0]] = v
		}
	}
	return m
}

// total stall (usec) dari baris "some"/"full" file PSI
func bacaPSI(path, jenis string) float64 {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	for _, row := range strings.Split(string(b), "\n") {
		if !strings.HasPrefix(row, jenis+" ") {
			continue
		}
		for _, f := range strings.Fields(row) {
			if v, ok := strings.CutPrefix(f, "total="); ok {
				n, _ := strconv.ParseFloat(v, 64)
				return n
			}
		}
	}
	return 0
}

func (c *cgroup) potret(p *potret) {
	cpu := bacaKV(filepath.Join(c.dir, "cpu.stat"))
	p.cpuUsec, p.throttledUsec = cpu["usage_usec"], cpu["throttled_usec"]
	p.psiCPU = bacaPSI(filepath.Join(c.dir, "cpu.pressure"), "some")
	p.psiIO = bacaPSI(filepath.Join(c.dir, "io.pressure"), "some")
	p.psiIOFull = bacaPSI(filepath.Join(c.dir, "io.pressure"), "full")
	p.psiMem = bacaPSI(filepath.Join(c.dir, "memory.pressure"), "some")
	if b, err := os.ReadFile(filepath.Join(c.dir, "memory.current")); err == nil {
		p.memori, _ = strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	}
	if b, err := os.ReadFile(filepath.Join(c.dir, "io.stat")); err == nil {
		for _, row := range strings.Split(string(b), "\n") {
			f := strings.Fields(row)
			if len(f) == 0 || f[0] != c.dev {
				continue
			}
			for _, kv := range f[1:] {
				k, v, _ := strings.Cut(kv, "=")
				n, _ := strconv.ParseFloat(v, 64)
				switch k {
				case "rios":
					p.rios = n
				case "wios":
					p.wios = n
				case "rbytes":
					p.rbytes = n
				case "wbytes":
					p.wbytes = n
				}
			}
		}
	}
}

func potretPG(ctx context.Context, db *sql.DB, p *potret) {
	if db == nil {
		return
	}
	_ = db.QueryRowContext(ctx, `SELECT blks_read, blks_hit, xact_commit FROM pg_stat_database WHERE datname = 'ildki'`).
		Scan(&p.blksRead, &p.blksHit, &p.xact)
	_ = db.QueryRowContext(ctx, `SELECT
		count(*) FILTER (WHERE state = 'active'),
		count(*) FILTER (WHERE state = 'active' AND wait_event_type IS NOT NULL),
		count(*) FILTER (WHERE wait_event_type = 'Lock'),
		count(*) FILTER (WHERE wait_event_type = 'IO'),
		count(*) FILTER (WHERE wait_event_type = 'LWLock')
		FROM pg_stat_activity WHERE datname = 'ildki' AND pid <> pg_backend_pid()`).
		Scan(&p.koneksiAktif, &p.koneksiTunggu, &p.tungguLock, &p.tungguIO, &p.tungguLWLock)
}

// rata-rata sampel sesaat (koneksi) dan selisih kumulatif (sisanya) selama satu tahap
type sumberDaya struct {
	CPUCore, CPUThrottledPersen            float64
	PSICPU, PSIIO, PSIIOFull, PSIMem       float64
	ReadIOPS, WriteIOPS, ReadMBs, WriteMBs float64
	MemoriMB                               float64
	CacheHitPersen, BlksReadPerDetik       float64
	KoneksiAktif, KoneksiMenunggu          float64
	TungguLock, TungguIO, TungguLWLock     float64
}

type pemantau struct {
	cg      *cgroup
	db      *sql.DB
	mu      sync.Mutex
	sampel  []potret
	stop    chan struct{}
	selesai chan struct{}
}

func pantau(cg *cgroup, db *sql.DB) *pemantau {
	m := &pemantau{cg: cg, db: db, stop: make(chan struct{}), selesai: make(chan struct{})}
	go func() {
		defer close(m.selesai)
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			m.ambil()
			select {
			case <-m.stop:
				m.ambil()
				return
			case <-t.C:
			}
		}
	}()
	return m
}

func (m *pemantau) ambil() {
	p := potret{waktu: time.Now()}
	if m.cg != nil {
		m.cg.potret(&p)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	potretPG(ctx, m.db, &p)
	cancel()
	m.mu.Lock()
	m.sampel = append(m.sampel, p)
	m.mu.Unlock()
}

func (m *pemantau) berhenti() sumberDaya {
	close(m.stop)
	<-m.selesai
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.sampel
	if len(s) < 2 {
		return sumberDaya{}
	}
	a, z := s[0], s[len(s)-1]
	dt := z.waktu.Sub(a.waktu).Seconds() * 1e6
	r := sumberDaya{
		CPUCore:            (z.cpuUsec - a.cpuUsec) / dt,
		CPUThrottledPersen: 100 * (z.throttledUsec - a.throttledUsec) / dt,
		PSICPU:             100 * (z.psiCPU - a.psiCPU) / dt,
		PSIIO:              100 * (z.psiIO - a.psiIO) / dt,
		PSIIOFull:          100 * (z.psiIOFull - a.psiIOFull) / dt,
		PSIMem:             100 * (z.psiMem - a.psiMem) / dt,
		ReadIOPS:           (z.rios - a.rios) / dt * 1e6,
		WriteIOPS:          (z.wios - a.wios) / dt * 1e6,
		ReadMBs:            (z.rbytes - a.rbytes) / dt,
		WriteMBs:           (z.wbytes - a.wbytes) / dt,
		MemoriMB:           z.memori / 1e6,
		BlksReadPerDetik:   (z.blksRead - a.blksRead) / dt * 1e6,
	}
	if h, rd := z.blksHit-a.blksHit, z.blksRead-a.blksRead; h+rd > 0 {
		r.CacheHitPersen = 100 * h / (h + rd)
	}
	for _, p := range s {
		r.KoneksiAktif += p.koneksiAktif
		r.KoneksiMenunggu += p.koneksiTunggu
		r.TungguLock += p.tungguLock
		r.TungguIO += p.tungguIO
		r.TungguLWLock += p.tungguLWLock
	}
	n := float64(len(s))
	r.KoneksiAktif /= n
	r.KoneksiMenunggu /= n
	r.TungguLock /= n
	r.TungguIO /= n
	r.TungguLWLock /= n
	return r
}
