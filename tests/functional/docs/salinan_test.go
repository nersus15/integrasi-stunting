// Package docs_test menjaga agar dokumentasi yang disajikan service tetap
// sama dengan sumbernya.
//
// errors.md ada dua salinan: yang di root repo (dibaca manusia, ditaut dari
// kode) dan yang di modules/stunting/docs/ (ditanam ke biner lewat go:embed,
// karena go:embed tidak bisa menjangkau berkas di luar direktori modul).
//
// Dua salinan berarti bisa melenceng. Test ini yang menahannya.
package docs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func read(t *testing.T, route string) []byte {
	t.Helper()
	content, err := os.ReadFile(route)
	if err != nil {
		t.Fatalf("gagal membaca %s: %v", route, err)
	}
	return content
}

func TestSalinanErrorsMdTidakBasi(t *testing.T) {
	akar := filepath.Join("..", "..", "..")

	source := filepath.Join(akar, "errors.md")
	salinan := filepath.Join(akar, "modules", "stunting", "docs", "errors.md")

	a := read(t, source)
	b := read(t, salinan)

	if string(a) == string(b) {
		return
	}

	t.Errorf("errors.md di root dan di modules/stunting/docs/ sudah berbeda "+
		"(%d vs %d byte). Yang disajikan service adalah salinan di docs/, "+
		"jadi perubahan pada sumber harus disalin ulang:\n\n"+
		"\tcp errors.md modules/stunting/docs/errors.md",
		len(a), len(b))
}

// CSP webcore (default-src 'self') menolak <style> inline, sehingga gaya harus
// tetap berupa berkas terpisah yang ditaut. Kalau suatu saat CSS dikembalikan
// ke dalam kerangka, halaman akan tampil polos tanpa error apa pun -- gagal
// diam-diam. Test ini yang menahannya.
func TestGayaTidakDitanamInline(t *testing.T) {
	akar := filepath.Join("..", "..", "..")
	kerangka := string(read(t, filepath.Join(akar, "modules", "stunting", "docs", "kerangka.html")))

	if strings.Contains(kerangka, "<style") {
		t.Error("kerangka.html memuat <style> inline; CSP webcore (default-src 'self') " +
			"akan menolaknya dan halaman tampil tanpa gaya. Pakai <link> ke berkas terpisah.")
	}
	if !strings.Contains(kerangka, `rel="stylesheet"`) {
		t.Error("kerangka.html tidak menaut stylesheet sama sekali")
	}
}

// Dokumen yang disebut handler harus benar-benar ada. Kalau salah satunya
// dihapus atau diganti nama, endpoint /stunting/docs akan menjawab 500 dan
// baru ketahuan saat dipanggil.
func TestDokumenYangDidaftarkanAda(t *testing.T) {
	akar := filepath.Join("..", "..", "..")

	for _, name := range []string{"index.md", "errors.md", "kerangka.html", "gaya.css"} {
		route := filepath.Join(akar, "modules", "stunting", "docs", name)
		if _, err := os.Stat(route); err != nil {
			t.Errorf("dokumen %s tidak ada: %v", name, err)
		}
	}
}
