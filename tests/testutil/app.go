package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	stunting "github.com/nersus15/integrasi/mod-stunting"
	backgroundworker "github.com/nersus15/lib-background-worker"
	cron "github.com/nersus15/lib-go-cron"
	"github.com/uptrace/bun/driver/pgdriver"
	postgres "github.com/webcore-go/lib-postgres"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/infra/middleware"
	"github.com/webcore-go/webcore/port"
)

// RepoRoot dihitung dari lokasi file ini, bukan dari working directory.
func RepoRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("tidak bisa menentukan lokasi testutil")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

// staticLoader mengembalikan instance library yang sudah disiapkan test.
type staticLoader struct {
	name    string
	library port.Library
}

var _ core.LibraryLoader = (*staticLoader)(nil)

func (l *staticLoader) SetName(name string)                    { l.name = name }
func (l *staticLoader) Name() string                           { return l.name }
func (l *staticLoader) Init(args ...any) (port.Library, error) { return l.library, nil }

// FunctionalApp adalah aplikasi lengkap dengan modul stunting yang asli.
type FunctionalApp struct {
	Web        *fiber.App
	Module     *stunting.Module
	Background *backgroundworker.BackgroundWorker

	// terisi kalau dibuat lewat NewFunctionalAppAuth
	Key      string
	aplikasi *App
}

// NewFunctionalAppAuth menjalankan aplikasi dengan auth API key asli. Request
// lewat Test() memakai Kunci, jadi middleware otorisasi berjalan seperti produksi.
func NewFunctionalAppAuth(port int, key string) (*FunctionalApp, error) {
	a, err := StartApp(port)
	if err != nil {
		return nil, err
	}
	if err := SeedReferenceRows(); err != nil {
		a.Stop()
		return nil, fmt.Errorf("gagal menyiapkan baris acuan: %w", err)
	}

	fa := &FunctionalApp{Web: a.app.Context.Web, Key: key, aplikasi: a}
	if lib, ok := a.Singleton("backgroundworker"); ok {
		fa.Background, _ = lib.(*backgroundworker.BackgroundWorker)
	}
	return fa, nil
}

func (a *FunctionalApp) Test(req *http.Request) (*http.Response, error) {
	if a.Key != "" && req.Header.Get("X-API-Key") == "" {
		req.Header.Set("X-API-Key", a.Key)
	}
	return a.Web.Test(req, -1)
}

// NewFunctionalApp membangun aplikasi lewat jalur bootstrap webcore yang asli:
// library dimuat LibraryManager, lalu Module.Init mendaftarkan route-nya sendiri.
// Yang tidak dijalankan hanya Web.Listen(), karena test memakai Web.Test().
//
// core.NewApp memakai singleton global, jadi hanya boleh dipanggil sekali per
// paket test - lewat TestMain. Pemanggil wajib memanggil Close().
func NewFunctionalApp() (*FunctionalApp, error) {
	if core.Instance() != nil {
		return nil, fmt.Errorf("NewFunctionalApp hanya boleh dipanggil sekali per paket test")
	}

	// Module.Init menyelesaikan path relatif terhadap working directory.
	if err := os.Chdir(RepoRoot()); err != nil {
		return nil, fmt.Errorf("gagal pindah ke root repo: %w", err)
	}

	cfg := NewAppConfig()

	cfg.Database.Driver = Env("PG_DRIVER", "postgres")
	cfg.Database.Host = Env("PG_HOST", "localhost")
	cfg.Database.Port, _ = strconv.Atoi(Env("PG_PORT", "6434"))
	cfg.Database.User = Env("PG_USER", "postgres")
	cfg.Database.Password = Env("PG_PASSWORD", "postgres")
	cfg.Database.Name = Env("PG_NAME", "ildki")
	cfg.Database.SchemaName = Env("PG_SCHEMA", "stunting")
	cfg.Database.SSLMode = Env("PG_SSLMODE", "disable")
	cfg.Memory.Enabled = true
	cfg.Redis.Host = Env("REDIS_HOST", "localhost")

	cache := NewFakeMemory()

	loaders := map[string]core.LibraryLoader{
		"database:postgres": &postgres.PostgresLoader{},
		"cache:memory":      &staticLoader{library: cache},
		"cache:redis":       &staticLoader{library: cache},
		"cron":              &cron.CronLoader{},
		"backgroundworker":  &backgroundworker.BackgroundWorkerLoader{},
	}

	module := stunting.NewModule()
	app := core.NewApp(context.Background(), cfg, loaders, []core.Module{module})
	app.Context.Web = fiber.New(cfg.GetFiberConfig(middleware.ErrorHandler))

	if err := app.Context.Start(); err != nil {
		return nil, fmt.Errorf("gagal memuat dependensi bersama: %w", err)
	}

	middleware.SetupGlobalMiddleware(app.Context.Web, cfg)

	// meniru App.setupAuthMiddleware() untuk auth.type = "none"
	app.Context.AuthHandler = func(c *fiber.Ctx) error { return c.Next() }
	app.Context.Root = app.Context.Web.Group(cfg.Server.PathPrefix, app.Context.AuthHandler)

	if err := app.ModuleManager.InitializeModulesWithDependencies(); err != nil {
		return nil, fmt.Errorf("gagal inisialisasi modul: %w", err)
	}

	if err := SeedReferenceRows(); err != nil {
		return nil, fmt.Errorf("gagal menyiapkan baris acuan: %w", err)
	}

	fa := &FunctionalApp{Web: app.Context.Web, Module: module}
	if lib, ok := app.Context.GetSingletonInstance("backgroundworker"); ok {
		fa.Background, _ = lib.(*backgroundworker.BackgroundWorker)
	}

	return fa, nil
}

func (a *FunctionalApp) Close() {
	if a.aplikasi != nil {
		a.aplikasi.Stop()
		return
	}
	if a.Web != nil {
		_ = a.Web.Shutdown()
	}
}

func SeedReferenceRows() error {
	dsn := DSN()

	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	defer db.Close()

	if err := cleanFixtures(db); err != nil {
		return err
	}

	_, err := db.Exec(`
		INSERT INTO stunting.posyandu (id, id_puskesmas, nama, rt, rw)
		SELECT $1, f.id, 'Posyandu Fixture', '001', '002'
		FROM stunting.faskes f WHERE f.satusehat_id = $2
		ON CONFLICT (id) DO NOTHING`, IdPosyandu, Env("ORGID_FASKES", "100011"))

	return err
}

func CleanFixtures() error {
	dsn := DSN()

	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	defer db.Close()

	return cleanFixtures(db)
}

func cleanFixtures(db *sql.DB) error {
	perintah := []struct {
		sql  string
		args []any
	}{
		{`DELETE FROM stunting.kunjungan WHERE id = $1 OR id_anak = $2`, []any{IdKunjungan, IdAnak}},
		{`DELETE FROM stunting.kesehatan WHERE id_anak = $1`, []any{IdAnak}},
		{`DELETE FROM stunting.anak WHERE id = $1 OR nik = $2`, []any{IdAnak, NikAnak}},
		{`DELETE FROM stunting.orangtua WHERE id = $1 OR nik = $2 OR no_kk = $3`, []any{IdOrangtua, NikOrangtua, NoKK}},
	}

	for _, p := range perintah {
		if _, err := db.Exec(p.sql, p.args...); err != nil {
			return err
		}
	}

	return nil
}

func SeedOrangtua() error {
	dsn := DSN()
	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO stunting.orangtua (id, id_posyandu, no_kk, nik, nama_ayah, nama_ibu, telepon, rt, rw, alamat, kia)
		VALUES ($1, $2, $3, $4, 'Budi', 'Siti', '08123', '003', '002', 'Jl. Mawar', 1)
		ON CONFLICT (id) DO NOTHING`, IdOrangtua, IdPosyandu, NoKK, NikOrangtua)

	return err
}

func SeedOrangtuaAndAnak() error {
	if err := SeedOrangtua(); err != nil {
		return err
	}

	dsn := DSN()
	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO stunting.anak (id, id_orangtua, nama, nik, tanggal_lahir, jenis_kelamin, anak_ke, imd, bb_lahir, tb_lahir, lk_lahir)
		VALUES ($1, $2, 'Ananda', $3, '2024-05-10', 'L', 1, 1, 3.2, 49.5, 33.0)
		ON CONFLICT (id) DO NOTHING`, IdAnak, IdOrangtua, NikAnak)

	return err
}

func OpenDB() *sql.DB {
	dsn := DSN()
	return sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
}

func DatabaseAvailable() bool {
	db := OpenDB()
	defer db.Close()
	ctx, batal := context.WithTimeout(context.Background(), 3*time.Second)
	defer batal()
	return db.PingContext(ctx) == nil
}
