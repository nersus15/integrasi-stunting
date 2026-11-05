package docs_endpoint_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var docsURL string

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		os.Exit(0)
	}
	app, err := testutil.StartApp(2094)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi: %v\n", err)
		os.Exit(1)
	}
	docsURL = strings.TrimSuffix(app.Health, "/health") + "/docs"
	code := m.Run()
	app.Stop()
	os.Exit(code)
}

func fetch(t *testing.T, query string) (int, string, string) {
	t.Helper()
	res, err := http.Get(docsURL + query)
	require.NoError(t, err)
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res.StatusCode, res.Header.Get("Content-Type"), string(content)
}

func TestSemuaHalamanDisajikanDanStabil(t *testing.T) {
	for _, name := range []string{"index", "jakantro", "faskes", "baca", "contoh", "errors"} {
		code, tipe, pertama := fetch(t, "?doc="+name)
		require.Equal(t, 200, code, name)
		assert.Contains(t, tipe, "text/html", name)
		assert.Contains(t, pertama, `aria-current="page"`, name)
		assert.NotContains(t, pertama, "__ISI__", "%s: placeholder kerangka terganti", name)

		_, _, kedua := fetch(t, "?doc="+name)
		assert.Equal(t, pertama, kedua, "%s: isi dari cache sama dengan render pertama", name)

		code, tipe, mentah := fetch(t, "?doc="+name+"&raw=1")
		require.Equal(t, 200, code)
		assert.Contains(t, tipe, "text/markdown")
		source, err := os.ReadFile(filepath.Join(testutil.RepoRoot(), "modules", "stunting", "docs", name+".md"))
		require.NoError(t, err)
		assert.Equal(t, string(source), mentah, "%s: markdown mentah sama dengan berkasnya", name)
	}

	code, _, index := fetch(t, "")
	assert.Equal(t, 200, code)
	_, _, indexEksplisit := fetch(t, "?doc=index")
	assert.Equal(t, indexEksplisit, index, "tanpa parameter menyajikan index")
}

func TestAsetDanDokumenTakDikenal(t *testing.T) {
	code, tipe, _ := fetch(t, "?doc=gaya")
	assert.Equal(t, 200, code)
	assert.Contains(t, tipe, "text/css")

	code, tipe, _ = fetch(t, "?doc=cari")
	assert.Equal(t, 200, code)
	assert.Contains(t, tipe, "javascript")

	for _, name := range []string{"laporan", "laporan-gaya", "kerangka", "../../config"} {
		code, _, content := fetch(t, "?doc="+name)
		assert.Equal(t, 404, code, name)
		assert.Contains(t, content, `"errorCode":2001`, name)
	}

	_, _, index := fetch(t, "?doc=index")
	assert.NotContains(t, index, "?doc=laporan", "nav tidak lagi menaut laporan")
}
