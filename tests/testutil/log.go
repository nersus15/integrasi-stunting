package testutil

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// LogToTerminal true bila UJI_LOG=stdout: log aplikasi tetap tampil seperti biasa.
var LogToTerminal = envLoaded && os.Getenv("UJI_LOG") == "stdout"

// Log aplikasi (slog webcore, goose, dsb.) dialihkan ke tests/.log/<kelompok>-<paket>.log
// supaya keluaran test hanya berisi hasil test.
func init() {
	if LogToTerminal {
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	name := filepath.Base(filepath.Dir(wd)) + "-" + filepath.Base(wd)
	dir := filepath.Join(RepoRoot(), "tests", ".log")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, strings.TrimPrefix(name, "-")+".log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return
	}
	log.SetOutput(f)
}
