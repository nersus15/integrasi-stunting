package stunting

import (
	"fmt"

	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/handler"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/repository"
	"github.com/nersus15/integrasi/mod-stunting/service"
	backgroundworker "github.com/nersus15/lib-background-worker"
	cron "github.com/nersus15/lib-go-cron"
	kafka "github.com/webcore-go/lib-kafka"
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

type Module struct {
	config     *config.ModuleConfig
	service    *service.StuntingService
	repository *repository.StuntingRepository
	handler    *handler.HttpHandler
	routes     []*core.ModuleRoute
	memory     port.ICacheMemory
	cron       *cron.CronLibrary
	background *backgroundworker.BackgroundWorker
	kafka      *handler.KafkaHandler
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return ModuleName
}

func (m *Module) Version() string {
	return ModuleVersion
}

func (m *Module) Dependencies() []string {
	return []string{}
}

func (m *Module) Init(ctx *core.AppContext) error {
	m.config = &config.ModuleConfig{}
	if err := appConfig.LoadDefaultConfigModule("fhir", m.config); err != nil {
		return err
	}

	libMem, ok := core.Instance().Context.GetDefaultSingletonInstance("cache:redis")
	if !ok {
		return fmt.Errorf("Gagal memuat instance Redis")
	}

	utils.SetDumpAktif(ctx.Config.App.Logging.Level)

	m.memory = libMem.(port.ICacheMemory)
	lib, ok := core.Instance().Context.GetDefaultSingletonInstance("database")

	if !ok {
		return fmt.Errorf("Gagal memuat instance database")
	}

	db := lib.(port.IDatabase)

	m.repository = repository.NewStuntingRepository(ctx, m.config, db, m.memory)

	migrationDir := fmt.Sprintf("migration/stunting/%s", ctx.Config.Database.Driver)
	err := m.repository.StartMigration(db.GetConnection(), ctx.Config.Database.Driver, "stunting", "up", migrationDir, nil)
	if err != nil {
		return fmt.Errorf("Gagal menjalankan migrasi otomatis: %s", err)
	}

	workerlib, err := ctx.StartSingletonInstance("backgroundworker")

	if err != nil {
		logger.Error("Gagal load library background worker: " + helper.ToLogJSON(err))
	}

	m.background = workerlib.(*backgroundworker.BackgroundWorker)

	logger.Info("Background Stats: " + helper.ToLogJSON(m.background.Stats()))

	m.service = service.NewStuntingService(ctx, m.config, m.repository, m.background)

	streamService := service.NewStreamService(ctx, m.service, m.config, m.repository, m.background)
	m.handler = handler.NewHttpHandler(ctx, m.config, m.service, streamService)

	if ctx.Config.Kafka.Enabled {
		m.kafka = handler.NewKafkaHandler(ctx, m.service, streamService, m.config, m.background, &m.memory)
		kafkaLoader, err := core.Instance().Context.GetDefaultLibraryLoader("kafka:consumer")

		if err != nil {
			return err
		}

		libkakfka, err := core.Instance().Context.LoadSingletonInstance(kafkaLoader, ctx.Config.Kafka, m.kafka)
		if err != nil {
			return err
		}

		kafkaConsumer := libkakfka.(*kafka.KafkaConsumer)

		go kafkaConsumer.Run(ctx.Context)
	}

	ctx.Web.Use("/api", handler.Otorisasi(m.memory))

	m.registerModuleRoute(ctx.Root)
	m.registerApiJakantroRoute(ctx.Root)
	m.registerApiFaskes(ctx.Root)
	m.registerRootRoute(ctx.Web)

	if m.config.Cron.Enabled {
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

	logger.Info("Module Integrasi Stunting initialized successfully")

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

func (m *Module) Services() map[string]any {
	return map[string]any{
		"stunting": m.service,
	}
}

func (m *Module) Repositories() map[string]any {
	return map[string]any{
		"stunting": m.repository,
	}
}
