package stunting

import (
	"bytes"
	"embed"
	"fmt"
	"html"
	"sort"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

// berkasDocs memuat dokumentasi ke dalam biner, sehingga yang disajikan selalu
// sama dengan versi yang sedang berjalan -- tidak bergantung pada berkas di
// disk yang bisa tertinggal saat deploy.
//
//go:embed docs/*.md docs/kerangka.html docs/gaya.css docs/cari.js
var berkasDocs embed.FS

type dokumen struct {
	berkas string
	judul  string
	ket    string
}

// Hanya nama yang terdaftar di sini yang dilayani. Tanpa daftar ini, parameter
// "doc" bisa dipakai menjelajah berkas lain di dalam embed.FS.
var daftarDokumen = map[string]dokumen{
	"index": {
		berkas: "docs/index.md",
		judul:  "Dokumentasi API",
		ket:    "Autentikasi, daftar endpoint, dan bentuk payload yang diterima layanan.",
	},
	"jakantro": {
		berkas: "docs/jakantro.md",
		judul:  "Jalur Jakantro",
		ket:    "Endpoint kirim untuk posyandu: kunjungan, kesehatan, orangtua, dan anak.",
	},
	"faskes": {
		berkas: "docs/faskes.md",
		judul:  "Jalur Faskes",
		ket:    "Endpoint kirim untuk puskesmas dan RS, alternatif dari jalur SatuSehat.",
	},
	"baca": {
		berkas: "docs/baca.md",
		judul:  "Endpoint GET",
		ket:    "Seluruh endpoint baca, termasuk riwayat lengkap anak lintas posyandu, puskesmas, dan RS.",
	},
	"contoh": {
		berkas: "docs/contoh.md",
		judul:  "Contoh Payload",
		ket:    "Payload lengkap siap salin untuk setiap endpoint POST.",
	},
	"errors": {
		berkas: "docs/errors.md",
		judul:  "Katalog Error",
		ket:    "Setiap kode error beserta artinya dan tindakan yang perlu diambil pemanggil.",
	},
}

// md merender markdown ke HTML. GFM dinyalakan karena dokumentasi banyak
// memakai tabel; auto heading id supaya tiap bagian bisa ditaut langsung.
// Raw HTML sengaja TIDAK diizinkan -- dokumen ini tidak membutuhkannya, dan
// mematikannya menutup satu jalan masuk seandainya isinya kelak datang dari
// sumber yang kurang tepercaya.
var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

// Docs menyajikan dokumentasi sebagai halaman HTML bergaya.
//
//	GET /stunting/docs              -> index, dirender
//	GET /stunting/docs?doc=errors   -> katalog error, dirender
//	GET /stunting/docs?raw=1        -> markdown mentah, untuk curl dan skrip
func (m *Module) Docs(c *fiber.Ctx) error {
	docsSekali.Do(siapkanDocs)
	if docsGalat != nil {
		return m.docsGagal(c, "menyiapkan dokumentasi", docsGalat)
	}

	nama := c.Query("doc", "index")

	if a, ok := docsAset[nama]; ok {
		c.Set(fiber.HeaderContentType, a.tipe)
		return c.Send(a.isi)
	}

	hal, ok := docsHalaman[nama]
	if !ok {
		tersedia := make([]string, 0, len(daftarDokumen))
		for k := range daftarDokumen {
			tersedia = append(tersedia, k)
		}
		sort.Strings(tersedia)

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"httpCode":  fiber.StatusNotFound,
			"errorCode": 2001,
			"errorName": "NOT_FOUND",
			"message": fmt.Sprintf("dokumen %q tidak ada, yang tersedia: %s",
				nama, strings.Join(tersedia, ", ")),
		})
	}

	if c.Query("raw") != "" {
		c.Set(fiber.HeaderContentType, "text/markdown; charset=utf-8")
		return c.Send(hal.mentah)
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return c.Send(hal.html)
}

type halamanDocs struct {
	html   []byte
	mentah []byte
}

type asetDocs struct {
	isi  []byte
	tipe string
}

// isi docs ter-embed dan tidak berubah selama proses hidup, cukup dirender sekali
var (
	docsSekali  sync.Once
	docsHalaman map[string]halamanDocs
	docsAset    map[string]asetDocs
	docsGalat   error
)

func siapkanDocs() {
	aset := map[string]asetDocs{}
	for nama, a := range map[string]struct{ berkas, tipe string }{
		"cari": {"docs/cari.js", "application/javascript; charset=utf-8"},
		"gaya": {"docs/gaya.css", "text/css; charset=utf-8"},
	} {
		isi, err := berkasDocs.ReadFile(a.berkas)
		if err != nil {
			docsGalat = fmt.Errorf("membaca %s: %w", a.berkas, err)
			return
		}
		aset[nama] = asetDocs{isi: isi, tipe: a.tipe}
	}

	kerangka, err := berkasDocs.ReadFile("docs/kerangka.html")
	if err != nil {
		docsGalat = fmt.Errorf("membaca kerangka.html: %w", err)
		return
	}

	halaman := make(map[string]halamanDocs, len(daftarDokumen))
	for nama, dok := range daftarDokumen {
		sumber, err := berkasDocs.ReadFile(dok.berkas)
		if err != nil {
			docsGalat = fmt.Errorf("membaca %s: %w", dok.berkas, err)
			return
		}

		var isi bytes.Buffer
		if err := md.Convert(sumber, &isi); err != nil {
			docsGalat = fmt.Errorf("merender %s: %w", dok.berkas, err)
			return
		}

		jadi := strings.NewReplacer(
			"__JUDUL__", html.EscapeString(dok.judul),
			"__KET__", html.EscapeString(dok.ket),
			"__NAV__", navDokumen(nama),
			"__ISI__", sorotKotak(bungkusTabel(isi.String())),
			"__SUMBER__", html.EscapeString(dok.berkas),
			"__VERSI__", html.EscapeString(ModuleVersion),
			"__HALAMAN__", daftarHalamanJSON(),
		).Replace(string(kerangka))

		halaman[nama] = halamanDocs{html: []byte(jadi), mentah: sumber}
	}

	docsAset, docsHalaman = aset, halaman
}

func (m *Module) docsGagal(c *fiber.Ctx, apa string, err error) error {
	logger.Error("Docs: gagal " + apa + ": " + err.Error())
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"httpCode":  fiber.StatusInternalServerError,
		"errorCode": 5001,
		"errorName": "INTERNAL_ERROR",
		"message":   "Terjadi kesalahan",
	})
}

// urutanDokumen menentukan urutan tampil di nav. Sengaja ditulis manual, bukan
// diurut abjad: index adalah pintu masuk dan harus memimpin.
var urutanDokumen = []string{"index", "jakantro", "faskes", "baca", "contoh", "errors"}

// daftarHalamanJSON dipakai skrip pencarian untuk tahu halaman apa saja yang
// boleh diambil versi mentahnya.
func daftarHalamanJSON() string {
	var b strings.Builder
	b.WriteByte('[')
	for i, n := range urutanDokumen {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"nama":%q,"judul":%q}`, n, daftarDokumen[n].judul)
	}
	b.WriteByte(']')
	return b.String()
}

func navDokumen(aktif string) string {
	var b strings.Builder
	for _, n := range urutanDokumen {
		aria := ""
		if n == aktif {
			aria = ` aria-current="page"`
		}
		fmt.Fprintf(&b, `<a href="?doc=%s"%s>%s</a>`,
			html.EscapeString(n), aria, html.EscapeString(daftarDokumen[n].judul))
	}
	return b.String()
}

// sorotKotak memberi warna pada blockquote yang diawali penanda, supaya hal
// penting tidak tenggelam. Blockquote lain dibiarkan apa adanya.
func sorotKotak(s string) string {
	for penanda, kelas := range map[string]string{
		"<strong>Penting.</strong>":   "penting",
		"<strong>Awas.</strong>":      "awas",
		"<strong>Perhatian.</strong>": "awas",
		"<strong>Catatan.</strong>":   "catatan",
	} {
		s = strings.ReplaceAll(s,
			"<blockquote>\n<p>"+penanda,
			`<blockquote class="`+kelas+`">`+"\n<p>"+penanda)
	}
	return s
}

// bungkusTabel membungkus setiap tabel dengan pembungkus bergulir, supaya tabel
// lebar tidak membuat seluruh halaman bergeser mendatar. goldmark tidak punya
// opsi untuk ini, dan markdown-nya milik kita sendiri, jadi penggantian teks
// biasa sudah memadai.
func bungkusTabel(s string) string {
	s = strings.ReplaceAll(s, "<table>", `<div class="tabel"><table>`)
	return strings.ReplaceAll(s, "</table>", "</table></div>")
}
