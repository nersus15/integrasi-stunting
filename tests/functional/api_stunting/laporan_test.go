package api_stunting_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

//go:embed kerangka.html
var kerangka string

var kelompok = map[byte][2]string{
	'1': {"perbaiki payload", "k1"},
	'2': {"tidak ditemukan", "k2"},
	'3': {"bentrok data", "k3"},
	'5': {"sisi layanan", "k5"},
}

func httpClass(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "ok"
	case code == 409:
		return "c9"
	case code >= 400 && code < 500:
		return "c4"
	}
	return "c5"
}

func block(v any) string {
	if v == nil {
		return `<pre class="blok">(tidak ada body)</pre>`
	}
	var teks string
	if s, ok := v.(string); ok {
		teks = s
	} else {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			teks = fmt.Sprint(v)
		} else {
			teks = string(b)
		}
	}
	if teks == "" {
		teks = "(kosong)"
	}
	return `<pre class="blok">` + html.EscapeString(teks) + `</pre>`
}

func table(judul string, kepala []string, row [][]string, catatan string) string {
	var b strings.Builder
	b.WriteString(`<div class="tabel-bungkus"><h3>` + html.EscapeString(judul) + `</h3>`)
	if catatan != "" {
		b.WriteString(`<p class="catatan">` + html.EscapeString(catatan) + `</p>`)
	}
	b.WriteString(`<div class="gulir"><table><thead><tr>`)
	for _, k := range kepala {
		b.WriteString(`<th>` + html.EscapeString(k) + `</th>`)
	}
	b.WriteString(`</tr></thead><tbody>`)
	for _, r := range row {
		b.WriteString(`<tr>`)
		for _, sel := range r {
			kelas := ""
			switch sel {
			case "<NULL>", "NULL":
				kelas = ` class="null"`
			case "KOSONG":
				kelas = ` class="warn"`
			}
			b.WriteString(`<td` + kelas + `>` + html.EscapeString(sel) + `</td>`)
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table></div></div>`)
	return b.String()
}

// snapshotDatabase mengambil keadaan akhir tabel untuk ditampilkan di laporan.
func snapshotDatabase(t *testing.T) string {
	t.Helper()
	db := openDB(t)

	ortu := fetchTable(t, db,
		`SELECT id, nik, no_kk FROM stunting.orangtua ORDER BY nik`, 3)

	anak := fetchTable(t, db, `
		SELECT id, id_orangtua, nama, anak_ke::text, COALESCE(nik,'<NULL>'),
		       CASE WHEN nik IS NULL THEN 'NULL' WHEN nik='' THEN 'KOSONG' ELSE 'terisi' END
		FROM stunting.anak ORDER BY anak_ke, id`, 6)

	kunj := fetchTable(t, db, `
		SELECT k.id, k.id_anak, a.nama, k.tanggal_pengukuran::text
		FROM stunting.kunjungan k JOIN stunting.anak a ON a.id = k.id_anak
		ORDER BY k.tanggal_pengukuran`, 4)

	nul, kosong := 0, 0
	for _, r := range anak {
		switch r[5] {
		case "NULL":
			nul++
		case "KOSONG":
			kosong++
		}
	}

	return table("Orangtua", []string{"id", "nik", "no_kk"}, ortu, "") +
		table("Anak", []string{"id", "id_orangtua", "nama", "anak_ke", "nik", "bentuk nik"}, anak,
			fmt.Sprintf("Kolom nik boleh NULL: %d baris tersimpan NULL, %d sebagai string kosong.", nul, kosong)) +
		table("Kunjungan", []string{"id", "id_anak", "menempel ke anak", "tanggal"}, kunj, "")
}

func writeReport(t *testing.T) {
	t.Helper()
	if len(rekam) == 0 {
		return
	}

	httpHitung := map[int]int{}
	codeHitung := map[int]int{}
	cocok := 0
	var urutanBagian []string
	terlihat := map[string]bool{}

	for _, r := range rekam {
		httpHitung[r.Http]++
		if ec := number(r.Response["errorCode"]); ec != 0 {
			codeHitung[ec]++
		}
		if r.Salah == "" {
			cocok++
		}
		if !terlihat[r.Bagian] {
			terlihat[r.Bagian] = true
			urutanBagian = append(urutanBagian, r.Bagian)
		}
	}

	var content strings.Builder
	for _, bagian := range urutanBagian {
		sid := "bag-" + strings.TrimSpace(strings.SplitN(bagian, ".", 2)[0])
		count := 0
		for _, r := range rekam {
			if r.Bagian == bagian {
				count++
			}
		}
		fmt.Fprintf(&content,
			`<section class="bagian" id="%s"><header class="bagian-kepala">`+
				`<h2>%s</h2><span class="hitung">%d skenario</span></header>`,
			sid, html.EscapeString(bagian), count)

		for _, r := range rekam {
			if r.Bagian != bagian {
				continue
			}
			ec := number(r.Response["errorCode"])
			en, _ := r.Response["errorName"].(string)
			message, _ := r.Response["message"].(string)

			var stack []any
			resp := map[string]any{}
			for k, v := range r.Response {
				if k == "stack" {
					if s, ok := v.([]any); ok {
						stack = s
					}
					continue
				}
				resp[k] = v
			}

			kid, judul, _ := strings.Cut(r.Label, " ")
			badge := ""
			if ec != 0 {
				badge = fmt.Sprintf(`<span class="ecode %s">%d</span>`,
					kelompok[fmt.Sprint(ec)[0]][1], ec)
			}
			isiPesan := `<em>sukses</em>`
			if message != "" {
				isiPesan = html.EscapeString(message)
			}
			verdikt, kelasVerdikt := "sesuai", "setuju"
			if r.Salah != "" {
				verdikt, kelasVerdikt = html.EscapeString(truncate(r.Salah, 36)), "beda"
			}
			stackHTML := ""
			if len(stack) > 0 {
				stackHTML = fmt.Sprintf(
					`<details class="stack"><summary>stack trace &mdash; %d baris</summary>%s</details>`,
					len(stack), block(stack))
			}
			var respTampil any = resp
			if len(resp) == 0 {
				respTampil = r.Mentah
			}

			fmt.Fprintf(&content,
				`<details class="skenario" data-http="%d" data-code="%s"><summary>`+
					`<span class="tag">%s</span>`+
					`<span class="judul">%s<span class="rute">%s %s</span></span>`+
					`<span class="pesan">%s</span>%s`+
					`<span class="kode %s">%d</span>`+
					`<span class="verdikt %s">%s</span></summary>`+
					`<div class="isi">`+
					`<div class="kolom"><h3>Request &mdash; %s %s</h3><p class="kredensial">Kredensial: %s</p>%s</div>`+
					`<div class="kolom"><h3>Response &mdash; %d%s</h3>%s%s</div>`+
					`</div></details>`,
				r.Http, codeOrEmpty(ec),
				html.EscapeString(kid), html.EscapeString(judul),
				html.EscapeString(r.Metode), html.EscapeString(r.Path),
				isiPesan, badge,
				httpClass(r.Http), r.Http,
				kelasVerdikt, verdikt,
				html.EscapeString(r.Metode), html.EscapeString(r.Path),
				html.EscapeString(r.Kunci), block(r.Payload),
				r.Http, namaTambahan(en), block(respTampil), stackHTML)
		}
		content.WriteString(`</section>`)
	}

	var kartu strings.Builder
	fmt.Fprintf(&kartu,
		`<div class="kartu utama"><span class="angka">%d/%d</span><span class="ket">sesuai harapan</span></div>`,
		cocok, len(rekam))
	for _, k := range sortedKeys(httpHitung) {
		fmt.Fprintf(&kartu,
			`<div class="kartu"><span class="angka %s">%d</span><span class="ket">%d skenario</span></div>`,
			httpClass(k), k, httpHitung[k])
	}

	var nav, fhttp, fcode strings.Builder
	for _, b := range urutanBagian {
		fmt.Fprintf(&nav, `<a href="#bag-%s">%s</a>`,
			strings.TrimSpace(strings.SplitN(b, ".", 2)[0]), html.EscapeString(b))
	}
	nav.WriteString(`<a href="#basis-data">Data tersimpan</a>`)
	for _, k := range sortedKeys(httpHitung) {
		fmt.Fprintf(&fhttp, `<button data-t="http" data-v="%d">%d</button>`, k, k)
	}
	for _, k := range sortedKeys(codeHitung) {
		fmt.Fprintf(&fcode, `<button data-t="code" data-v="%d">%d</button>`, k, k)
	}

	var legenda strings.Builder
	for _, k := range []byte{'1', '2', '3', '5'} {
		fmt.Fprintf(&legenda, `<li><span class="ecode %s">%cxxx</span> %s</li>`,
			kelompok[k][1], k, html.EscapeString(kelompok[k][0]))
	}

	keluaran := kerangka
	for find, ganti := range map[string]string{
		"__TOTAL__":   fmt.Sprint(len(rekam)),
		"__KARTU__":   kartu.String(),
		"__LEGENDA__": legenda.String(),
		"__NAV__":     nav.String(),
		"__FHTTP__":   fhttp.String(),
		"__FCODE__":   fcode.String(),
		"__BARIS__":   content.String(),
		"__DB__":      snapshotDatabase(t),
	} {
		keluaran = strings.Replace(keluaran, find, ganti, 1)
	}

	if err := writeFile(berkasHTML, keluaran); err != nil {
		t.Fatalf("gagal menulis laporan: %v", err)
	}
	abs, _ := filepath.Abs(berkasHTML)
	t.Logf("laporan HTML: %s (%d KB)", abs, len(keluaran)/1024)
}

func codeOrEmpty(ec int) string {
	if ec == 0 {
		return ""
	}
	return fmt.Sprint(ec)
}

func namaTambahan(en string) string {
	if en == "" {
		return ""
	}
	return " &middot; " + html.EscapeString(en)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func sortedKeys(m map[int]int) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}
