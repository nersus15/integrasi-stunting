package testutil

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// dibaca saat paket dimuat, sebelum variabel paket test lain diisi; env dari shell tetap menang
var envLoaded = loadEnvFiles()

func loadEnvFiles() bool {
	for _, name := range []string{".test.env"} {
		readEnvFile(filepath.Join(RepoRoot(), "tests", name))
	}
	return true
}

func readEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		row := strings.TrimSpace(sc.Text())
		if row == "" || strings.HasPrefix(row, "#") {
			continue
		}
		k, v, ok := strings.Cut(strings.TrimPrefix(row, "export "), "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		if _, exists := os.LookupEnv(k); exists {
			continue
		}
		os.Setenv(k, strings.Trim(strings.TrimSpace(v), `"'`))
	}
}

func Env(name, bawaan string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return bawaan
}

// PG_DSN bila diisi, selain itu disusun dari PG_*
func DSN() string {
	if v := os.Getenv("PG_DSN"); v != "" {
		return v
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		Env("PG_USER", "postgres"), Env("PG_PASSWORD", "postgres"), Env("PG_HOST", "localhost"),
		Env("PG_PORT", "6434"), Env("PG_NAME", "ildki"), Env("PG_SSLMODE", "disable"))
}

// service yang sudah berjalan, hanya bila API_BASE diisi; health bawaannya diturunkan dari API_BASE
func ExternalService() (base, health string, exists bool) {
	base = os.Getenv("API_BASE")
	if base == "" {
		return "", "", false
	}
	return base, Env("API_HEALTH", strings.TrimSuffix(base, "/api")+"/stunting/health"), true
}
