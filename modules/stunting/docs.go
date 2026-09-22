package stunting

import (
	"bytes"
	"embed"
	"fmt"
	"html"
	"sort"
	"strings"

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
//go:embed docs
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

// nav memuat laporan juga, meski ia bukan markdown dan tidak lewat daftarDokumen.
var judulLaporan = "Laporan Pengujian"

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
	nama := c.Query("doc", "index")

	// Stylesheet disajikan lewat endpoint yang sama, bukan ditanam sebagai
	// <style> inline: webcore memasang Content-Security-Policy "default-src
	// 'self'" yang menolak style inline, tetapi mengizinkan stylesheet
	// same-origin. Kerangka menautnya dengan href relatif "?doc=gaya".
	if nama == "cari" {
		js, err := berkasDocs.ReadFile("docs/cari.js")
		if err != nil {
			return m.docsGagal(c, "membaca cari.js", err)
		}
		c.Set(fiber.HeaderContentType, "application/javascript; charset=utf-8")
		return c.Send(js)
	}

	if nama == "gaya" {
		gaya, err := berkasDocs.ReadFile("docs/gaya.css")
		if err != nil {
			return m.docsGagal(c, "membaca gaya.css", err)
		}
		c.Set(fiber.HeaderContentType, "text/css; charset=utf-8")
		return c.Send(gaya)
	}

	// Laporan pengujian dibangkitkan test suite, bukan ditulis tangan, jadi
	// tidak ikut daftarDokumen yang di-render dari markdown.
	if nama == "laporan" || nama == "laporan-gaya" {
		return m.docsLaporan(c, nama)
	}

	dok, ok := daftarDokumen[nama]
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

	sumber, err := berkasDocs.ReadFile(dok.berkas)
	if err != nil {
		return m.docsGagal(c, "membaca "+dok.berkas, err)
	}

	if c.Query("raw") != "" {
		c.Set(fiber.HeaderContentType, "text/markdown; charset=utf-8")
		return c.Send(sumber)
	}

	var isi bytes.Buffer
	if err := md.Convert(sumber, &isi); err != nil {
		return m.docsGagal(c, "merender "+dok.berkas, err)
	}

	kerangka, err := berkasDocs.ReadFile("docs/kerangka.html")
	if err != nil {
		return m.docsGagal(c, "membaca kerangka.html", err)
	}

	halaman := strings.NewReplacer(
		"__JUDUL__", html.EscapeString(dok.judul),
		"__KET__", html.EscapeString(dok.ket),
		"__NAV__", navDokumen(nama),
		"__ISI__", sorotKotak(bungkusTabel(isi.String())),
		"__SUMBER__", html.EscapeString(dok.berkas),
		"__VERSI__", html.EscapeString(ModuleVersion),
		"__HALAMAN__", daftarHalamanJSON(),
	).Replace(string(kerangka))

	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return c.SendString(halaman)
}

// docsLaporan menyajikan laporan pengujian yang dibangkitkan test suite.
//
// Laporan aslinya menaruh CSS sebagai <style> inline karena juga diterbitkan
// sebagai artifact, yang memang menuntut berkas mandiri. Tapi Content-Security-
// Policy service ("default-src 'self'") menolak style inline, jadi di sini
// blok itu dipisah dan ditaut sebagai stylesheet same-origin.
func (m *Module) docsLaporan(c *fiber.Ctx, nama string) error {
	mentah, err := berkasDocs.ReadFile("docs/laporan.html")
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"httpCode":  fiber.StatusNotFound,
			"errorCode": 2001,
			"errorName": "NOT_FOUND",
			"message": "laporan pengujian belum dibangkitkan; jalankan " +
				"`go test ./functional/api_stunting/` di direktori tests",
		})
	}

	kepala, gaya, badan := belahGaya(string(mentah))

	if nama == "laporan-gaya" {
		c.Set(fiber.HeaderContentType, "text/css; charset=utf-8")
		return c.SendString(gaya)
	}

	halaman := `<!doctype html>
<html lang="id">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
` + kepala + `<link rel="stylesheet" href="?doc=laporan-gaya">
</head>
<body>
` + badan + `
</body>
</html>`

	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return c.SendString(halaman)
}

// belahGaya memisahkan satu blok <style> dari dokumen. Laporan dibangkitkan
// dari kerangka yang kita kendalikan sendiri dan hanya punya satu blok, jadi
// pemisahan sederhana ini memadai. Kalau blok itu tidak ada, dokumen
// dikembalikan apa adanya.
func belahGaya(s string) (kepala, gaya, badan string) {
	awal := strings.Index(s, "<style>")
	if awal < 0 {
		return "", "", s
	}
	akhir := strings.Index(s, "</style>")
	if akhir < awal {
		return "", "", s
	}
	return s[:awal], s[awal+len("<style>") : akhir], s[akhir+len("</style>"):]
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
	fmt.Fprintf(&b, `<a href="?doc=laporan">%s</a>`, html.EscapeString(judulLaporan))
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
