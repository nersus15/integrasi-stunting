package satusehat

import (
	"github.com/webcore-go/webcore/app/core"
	appConfig "github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/port"

	"github.com/nersus15/integrasi/mod-satusehat/config"
	"github.com/nersus15/integrasi/mod-satusehat/handler"
	"github.com/nersus15/integrasi/mod-satusehat/repository"
	"github.com/nersus15/integrasi/mod-satusehat/service"
)

const (
	ModuleName    = "satusehat"
	ModuleVersion = "1.0.0"
)

type Module struct {
	config     *config.ModuleConfig
	repository *repository.Repository
	service    *service.Service
	handler    *handler.Handler
	routes     []*core.ModuleRoute
}

func NewModule() *Module { return &Module{} }

func (m *Module) Name() string                   { return ModuleName }
func (m *Module) Version() string                { return ModuleVersion }
func (m *Module) Dependencies() []string         { return []string{} }
func (m *Module) Config() appConfig.ConfigObject { return m.config }
func (m *Module) Routes() []*core.ModuleRoute    { return m.routes }
func (m *Module) Destroy() error                 { return nil }

func (m *Module) Services() map[string]any {
	return map[string]any{"service": m.service}
}

func (m *Module) Repositories() map[string]any {
	return map[string]any{"repository": m.repository}
}

func (m *Module) Init(ctx *core.AppContext) error {
	// 1. Muat konfigurasi modul (namespace module.satusehat di config.yaml).
	m.config = &config.ModuleConfig{}
	if err := appConfig.LoadDefaultConfigModule(m.Name(), m.config); err != nil {
		return err
	}

	// 2. Ambil shared library — database harus sudah didaftarkan di
	//    deps/libraries.go (mis. "database:postgres").
	if lib, ok := ctx.GetDefaultSingletonInstance("database"); ok {
		db := lib.(port.IDatabase)
		m.repository = repository.NewRepository(ctx, m.config, db)
		m.service = service.NewService(ctx, m.repository)
		m.handler = handler.NewHandler(ctx, m.service)
	}

	// 3. Registrasi route di bawah ctx.Root (sudah ber-auth & ber-prefix).
	m.registerRoutes(ctx)

	return nil
}

func (m *Module) registerRoutes(ctx *core.AppContext) {
	root := ctx.Root.Group("/" + m.Name())

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method: "GET", Path: "/", Handler: m.handler.ListSatusehats, Root: root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method: "POST", Path: "/", Handler: m.handler.CreateSatusehat, Root: root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method: "GET", Path: "/:id", Handler: m.handler.GetSatusehat, Root: root,
	})
}
