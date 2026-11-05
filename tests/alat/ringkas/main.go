// Meringkas keluaran `go test -json` menjadi daftar lulus/gagal ala jest.
// Test gagal dan dilewati selalu ditampilkan beserta keluaran milik test itu sendiri
// (t.Error, t.Skip, assertion testify, panic), bukan log aplikasi.
//
//	go test -json ./... | go run ./alat/ringkas [-daftar]
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type event struct {
	Action      string
	Package     string
	ImportPath  string
	Test        string
	Elapsed     float64
	Output      string
	FailedBuild string
}

type testCase struct {
	nama   string
	status string
	lama   float64
	keluar []string
}

type pkgResult struct {
	nama     string
	status   string
	lama     float64
	uji      map[string]*testCase
	urutan   []string
	keluar   []string
	dilewati string
	gagalDi  string
}

const (
	maxLines = 60
	maxWidth = 400
)

var (
	useColor = os.Getenv("NO_COLOR") == "" && terminal()
	green    = ansi("32")
	red      = ansi("31")
	yellow   = ansi("33")
	dim      = ansi("2")
	bold     = ansi("1")
	normal   = ansi("0")
	showAll  = flag.Bool("daftar", false, "tampilkan setiap test, bukan hanya yang gagal")
)

func terminal() bool {
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func ansi(k string) string {
	if !useColor {
		return ""
	}
	return "\033[" + k + "m"
}

func shortName(pkg string) string {
	if i := strings.Index(pkg, "/tests/"); i >= 0 {
		return pkg[i+len("/tests/"):]
	}
	return pkg
}

func main() {
	flag.Parse()
	start := time.Now()
	packages := map[string]*pkgResult{}
	var packageOrder []string
	build := map[string][]string{}

	getPackage := func(name string) *pkgResult {
		p, exists := packages[name]
		if !exists {
			p = &pkgResult{nama: name, uji: map[string]*testCase{}}
			packages[name] = p
			packageOrder = append(packageOrder, name)
		}
		return p
	}

	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for sc.Scan() {
		var e event
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			fmt.Println(sc.Text()) // keluaran non-json (mis. galat go vet)
			continue
		}
		if e.Action == "build-output" {
			build[e.ImportPath] = append(build[e.ImportPath], e.Output)
			continue
		}
		if e.Package == "" {
			continue
		}
		p := getPackage(e.Package)
		if e.Test == "" {
			switch e.Action {
			case "output":
				p.keluar = append(p.keluar, e.Output)
				if strings.Contains(e.Output, "dilewati") {
					p.dilewati = strings.TrimSpace(e.Output)
				}
			case "pass", "fail", "skip":
				p.status, p.lama, p.gagalDi = e.Action, e.Elapsed, e.FailedBuild
				printPackage(p, build)
			}
			continue
		}
		u, exists := p.uji[e.Test]
		if !exists {
			u = &testCase{nama: e.Test}
			p.uji[e.Test] = u
			p.urutan = append(p.urutan, e.Test)
		}
		switch e.Action {
		case "output":
			u.keluar = append(u.keluar, e.Output)
		case "pass", "fail", "skip":
			u.status, u.lama = e.Action, e.Elapsed
		}
	}

	printSummary(packages, packageOrder, time.Since(start))
}

func printPackage(p *pkgResult, build map[string][]string) {
	if p.status == "skip" && len(p.uji) == 0 {
		return // paket tanpa file test
	}
	label := map[string]string{
		"pass": green + bold + " LULUS " + normal,
		"fail": red + bold + " GAGAL " + normal,
	}[p.status]
	if p.dilewati != "" && p.status == "pass" && len(p.uji) == 0 {
		label = yellow + bold + " LEWAT " + normal
	}
	fmt.Printf("%s %s %s(%.1fs)%s\n", label, shortName(p.nama), dim, p.lama, normal)

	if p.dilewati != "" && len(p.uji) == 0 {
		fmt.Printf("        %s%s%s\n", yellow, p.dilewati, normal)
	}

	failed := 0
	for _, n := range p.urutan {
		u := p.uji[n]
		if u.status == "fail" {
			failed++
		}
		if !*showAll && u.status != "fail" && u.status != "skip" {
			continue
		}
		indent := strings.Repeat("  ", strings.Count(n, "/"))
		switch u.status {
		case "pass":
			fmt.Printf("    %s%s✓%s %s %s(%.2fs)%s\n", indent, green, normal, n[strings.LastIndex(n, "/")+1:], dim, u.lama, normal)
		case "fail":
			fmt.Printf("    %s%s✗ %s%s %s(%.2fs)%s\n", indent, red, n[strings.LastIndex(n, "/")+1:], normal, dim, u.lama, normal)
		case "skip":
			fmt.Printf("    %s%s○ %s (dilewati)%s\n", indent, yellow, n[strings.LastIndex(n, "/")+1:], normal)
			for _, b := range testOutputLines(u.keluar) {
				fmt.Printf("    %s  %s%s%s\n", indent, dim, b, normal)
			}
		}
	}

	for _, n := range p.urutan {
		u := p.uji[n]
		if u.status != "fail" {
			continue
		}
		row := testOutputLines(u.keluar)
		if len(row) == 0 && hasFailedSubtest(p, n) {
			continue // kegagalannya ditampilkan di sub-test
		}
		fmt.Printf("\n  %s● %s%s\n\n", red+bold, n, normal)
		for _, b := range row {
			fmt.Printf("    %s\n", b)
		}
	}

	if p.status == "fail" && failed == 0 {
		// gagal di luar test: build, TestMain, atau panic sebelum test jalan
		fmt.Printf("\n  %s● paket gagal sebelum test berjalan%s\n\n", red+bold, normal)
		source := build[p.gagalDi]
		if len(source) == 0 {
			source = p.keluar
		}
		for _, b := range tail(source, 25) {
			fmt.Printf("    %s\n", truncate(strings.TrimRight(b, "\n")))
		}
	}
	if failed > 0 || p.status == "fail" {
		fmt.Println()
	}
}

func hasFailedSubtest(p *pkgResult, n string) bool {
	for _, m := range p.urutan {
		if strings.HasPrefix(m, n+"/") && p.uji[m].status == "fail" {
			return true
		}
	}
	return false
}

// keluaran milik test diindentasi oleh paket testing; log aplikasi tidak
func testOutputLines(output []string) []string {
	var result []string
	panik := false
	for _, b := range output {
		b = strings.TrimRight(b, "\n")
		t := strings.TrimSpace(b)
		switch {
		case t == "":
			continue
		case strings.HasPrefix(t, "=== "), strings.HasPrefix(t, "--- PASS"), strings.HasPrefix(t, "--- FAIL"), strings.HasPrefix(t, "--- SKIP"):
			continue
		case strings.HasPrefix(t, "panic:"):
			panik = true
		}
		if panik || strings.HasPrefix(b, "    ") || strings.HasPrefix(b, "\t") {
			result = append(result, truncate(strings.TrimPrefix(b, "    ")))
		}
		if len(result) >= maxLines {
			result = append(result, dim+"… dipotong; jalankan dengan -v untuk keluaran lengkap"+normal)
			break
		}
	}
	return result
}

func tail(xs []string, n int) []string {
	var cleanup []string
	for _, x := range xs {
		if strings.TrimSpace(x) != "" {
			cleanup = append(cleanup, x)
		}
	}
	if len(cleanup) > n {
		return cleanup[len(cleanup)-n:]
	}
	return cleanup
}

func truncate(s string) string {
	if r := []rune(s); len(r) > maxWidth {
		return string(r[:maxWidth]) + dim + "…" + normal
	}
	return s
}

func printSummary(packages map[string]*pkgResult, urut []string, old time.Duration) {
	var pl, pg, pd, tl, tg, td int
	for _, n := range urut {
		p := packages[n]
		switch {
		case p.status == "fail":
			pg++
		case p.dilewati != "" && len(p.uji) == 0:
			pd++
		case p.status == "pass":
			pl++
		}
		for _, u := range p.uji {
			if strings.Contains(u.nama, "/") {
				continue // hitung test utama saja
			}
			switch u.status {
			case "pass":
				tl++
			case "fail":
				tg++
			case "skip":
				td++
			}
		}
	}
	count := func(g, l, d int) string {
		s := fmt.Sprintf("%s%d gagal%s, %s%d lulus%s", red, g, normal, green, l, normal)
		if d > 0 {
			s += fmt.Sprintf(", %s%d dilewati%s", yellow, d, normal)
		}
		return s + fmt.Sprintf(", %d total", g+l+d)
	}
	fmt.Printf("\n%sPaket:%s %s\n", bold, normal, count(pg, pl, pd))
	fmt.Printf("%sTest :%s %s\n", bold, normal, count(tg, tl, td))
	fmt.Printf("%sWaktu:%s %.1fs\n", bold, normal, old.Seconds())
	if pg > 0 {
		fmt.Printf("%sLog aplikasi:%s tests/.log/ (UJI_LOG=stdout untuk menampilkannya di terminal)\n", dim, normal)
	}
}
