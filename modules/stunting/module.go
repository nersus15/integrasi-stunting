package stunting

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/handler"
	"github.com/nersus15/integrasi/mod-stunting/repository"
	"github.com/nersus15/integrasi/mod-stunting/service"
	backgroundworker "github.com/nersus15/lib-background-worker"
	cron "github.com/nersus15/lib-go-cron"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/helper"
	appConfig "github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

const (
	ModuleName    = "stunting"
	ModuleVersion = "1.0.0"
)

// Module implements the module.Module interface
type Module struct {
	config     *config.ModuleConfig
	service    *service.StuntingService
	repository *repository.StuntingRepository
	handler    *handler.HttpHandler
	routes     []*core.ModuleRoute
	memory     port.ICacheMemory
	cron       *cron.CronLibrary
	background *backgroundworker.BackgroundWorker
}

// NewModule creates a new Module instance
func NewModule() *Module {
	return &Module{}
}

// Name returns the unique name of the module
func (m *Module) Name() string {
	return ModuleName
}

// Version returns the version of the module
func (m *Module) Version() string {
	return ModuleVersion
}

// Dependencies returns the dependencies of the module to other modules
func (m *Module) Dependencies() []string {
	return []string{}
}

// Init initializes the module with the given app and dependencies
func (m *Module) Init(ctx *core.AppContext) error {
	// Load configuration into ModuleConfig (bind to key)
	m.config = &config.ModuleConfig{}
	if err := appConfig.LoadDefaultConfigModule("fhir", m.config); err != nil {
		return err
	}

	// Register services and repositories
	libMem, ok := core.Instance().Context.GetDefaultSingletonInstance("cache:memory")
	if !ok {
		return fmt.Errorf("Gagal memuat instance Memory")
	}

	m.memory = libMem.(port.ICacheMemory)
	lib, ok := core.Instance().Context.GetDefaultSingletonInstance("database")

	if !ok {
		return fmt.Errorf("Gagal memuat instance database")
	}

	db := lib.(port.IDatabase)

	m.repository = repository.NewStuntingRepository(ctx, m.config, db, m.memory)

	// Jalankan migrasi otomatis hanya jika database driver menggunakan sqlite
	// if ctx.Config.Database.Driver == "sqlite" {
	migrationDir := fmt.Sprintf("migration/stunting/%s", ctx.Config.Database.Driver)
	err := m.repository.StartMigration(db.GetConnection(), ctx.Config.Database.Driver, "stunting", "up", migrationDir, nil)
	if err != nil {
		return fmt.Errorf("Gagal menjalankan migrasi otomatis: %s", err)
	}
	// }

	workerlib, err := ctx.StartSingletonInstance("backgroundworker")

	if err != nil {
		logger.Error("Gagal load library background worker: " + helper.ToLogJSON(err))
	}

	m.background = workerlib.(*backgroundworker.BackgroundWorker)

	logger.Info("Background Stats: " + helper.ToLogJSON(m.background.Stats()))

	m.service = service.NewStuntingService(ctx, m.config, m.repository, m.background)
	m.handler = handler.NewHandler(ctx, m.config, m.service, m.background)

	// Register routes
	m.registerModuleRoute(ctx.Root)
	m.registerRootRoute(ctx.Web)

	if m.config.Cron.Enabled {
		// Inisiai CronJob
		cronLoader, err := ctx.GetDefaultLibraryLoader("cron")

		if err != nil {
			return fmt.Errorf("Gagal memuat instance cronLoader")
		}
		cronLib, err := ctx.LoadSingletonInstance(cronLoader)

		if err != nil {
			return fmt.Errorf("Gagal Load CronLib")
		}
		m.cron = cronLib.(*cron.CronLibrary)

		if err := m.cron.Connect(); err != nil {
			logger.ErrorJson("Gagal Menjalankan Cron", err)
		}

	}

	// These can be accessed through the central registry
	logger.Info("Module Stream initialized successfully")

	return nil
}

func (m *Module) Destroy() error {
	return nil
}

func (m *Module) Config() appConfig.ConfigObject {
	return m.config
}

func (m *Module) Routes() []*core.ModuleRoute {
	return m.routes
}

// Services returns the services provided by this module
func (m *Module) Services() map[string]any {
	// Return services that can be used by other modules
	return map[string]any{
		"stunting": m.service,
	}
}

// Repositories returns the repositories provided by this module
func (m *Module) Repositories() map[string]any {
	// Return repositories that can be used by other modules
	return map[string]any{
		"stunting": m.repository,
	}
}

// registerRoutes registers the module's routes
func (m *Module) registerModuleRoute(root fiber.Router) {
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/orangtua/:id?",
		Handler: m.handler.FindOrangTua,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/orangtua/:id/anak",
		Handler: m.handler.ListAnak,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/anak/:id?",
		Handler: m.handler.FindAnak,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/anak/:id/kunjungan",
		Handler: m.handler.FindKunjunganByIdAnak,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/anak/:id/kesehatan",
		Handler: m.handler.FindKesehatanByIdAnak,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/anak/:id/summary",
		Handler: m.handler.SummaryAnak,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/anak/kunjungan/:id",
		Handler: m.handler.FindKunjunganById,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/anak/kesehatan/:id",
		Handler: m.handler.FindKesehatanById,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/orangtua",
		Handler: m.handler.CreateOrangTua,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/anak",
		Handler: m.handler.CreateAnak,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/kunjungan",
		Handler: m.handler.CreateKunjungan,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/kesehatan",
		Handler: m.handler.CreateKesehatan,
		Root:    root,
	})
}

func (m *Module) registerRootRoute(web *fiber.App) {
	moduleRoot := web.Group("/" + m.Name())
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/health",
		Handler: m.Health,
		Root:    moduleRoot,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/info",
		Handler: m.Info,
		Root:    moduleRoot,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/docs",
		Handler: m.Docs,
		Root:    moduleRoot,
	})
}

// ModuleHealth returns the health status of the module
func (m *Module) Health(c *fiber.Ctx) error {
	health := map[string]any{
		"status":    "healthy",
		"module":    ModuleName,
		"version":   ModuleVersion,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	return c.JSON(health)
}

func (m *Module) Info(c *fiber.Ctx) error {
	endpoints := []string{}
	for _, endpoint := range m.routes {
		endpoint := endpoint.Method + " " + endpoint.Path
		endpoints = append(endpoints, endpoint)
	}

	path := "/" + ModuleName

	info := map[string]any{
		"name":        ModuleName,
		"version":     ModuleVersion,
		"description": "Integrasi data Stunting",
		"path":        path,
		"endpoints":   endpoints,
		"config":      m.config,
	}
	return c.JSON(info)
}
