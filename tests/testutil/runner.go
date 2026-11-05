package testutil

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	modconfig "github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/handler"
	"github.com/nersus15/integrasi/mod-stunting/repository"
	"github.com/nersus15/integrasi/mod-stunting/service"
	appstunting "github.com/nersus15/integrasi/stunting"
	backgroundworker "github.com/nersus15/lib-background-worker"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/port"
)

type App struct {
	app     *core.App
	BaseURL string
	Health  string
	selesai chan error
}

func StartApp(port int) (*App, error) {
	return StartAppWith(port, nil)
}

func StartAppWith(port int, modify func(*config.Config)) (*App, error) {
	if core.Instance() != nil {
		return nil, fmt.Errorf("aplikasi hanya boleh dijalankan sekali per paket test")
	}
	if err := os.Chdir(RepoRoot()); err != nil {
		return nil, err
	}
	db := OpenDB()
	err := EnsureAPIKeys(db)
	db.Close()
	if err != nil {
		return nil, fmt.Errorf("gagal menyiapkan key test: %w", err)
	}

	cfg := config.Config{}
	if err := config.LoadDefaultConfig(&cfg); err != nil {
		return nil, err
	}
	cfg.Server.Port = port
	cfg.Kafka.Enabled = false
	cfg.App.Features.Logging = LogToTerminal
	if modify != nil {
		modify(&cfg)
	}

	app := core.NewApp(context.Background(), &cfg, appstunting.APP_LIBRARIES, appstunting.APP_PACKAGES)

	a := &App{
		app:     app,
		BaseURL: fmt.Sprintf("http://localhost:%d/api", port),
		Health:  fmt.Sprintf("http://localhost:%d/stunting/health", port),
		selesai: make(chan error, 1),
	}
	go func() { a.selesai <- app.Start() }()

	client := &http.Client{Timeout: time.Second}
	limit := time.Now().Add(30 * time.Second)
	for time.Now().Before(limit) {
		select {
		case err := <-a.selesai:
			return nil, fmt.Errorf("aplikasi berhenti saat start: %v", err)
		default:
		}
		if res, err := client.Get(a.Health); err == nil {
			res.Body.Close()
			if res.StatusCode == http.StatusOK {
				return a, nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil, fmt.Errorf("aplikasi tidak siap dalam 30 detik")
}

func (a *App) Singleton(name string) (port.Library, bool) {
	return a.app.Context.GetSingletonInstance(name)
}

func (a *App) Stop() {
	if a != nil && a.app != nil {
		_ = a.app.Stop()
	}
}

type Wiring struct {
	Config     *modconfig.ModuleConfig
	Repository *repository.StuntingRepository
	Service    *service.StuntingService
	Stream     *service.StreamService
	Kafka      *handler.KafkaHandler
	Memory     port.ICacheMemory
}

func (a *App) Wiring(groupID string) (*Wiring, error) {
	ctx := a.app.Context

	cfg := &modconfig.ModuleConfig{}
	if err := config.LoadDefaultConfigModule("fhir", cfg); err != nil {
		return nil, err
	}
	if groupID != "" {
		cfg.Kafka.GroupID = groupID
	}

	libDB, ok := ctx.GetDefaultSingletonInstance("database")
	if !ok {
		return nil, fmt.Errorf("database tidak termuat")
	}
	libMem, ok := ctx.GetDefaultSingletonInstance("cache:redis")
	if !ok {
		return nil, fmt.Errorf("cache tidak termuat")
	}
	var bg *backgroundworker.BackgroundWorker
	if lib, ok := ctx.GetSingletonInstance("backgroundworker"); ok {
		bg, _ = lib.(*backgroundworker.BackgroundWorker)
	}

	var mem port.ICacheMemory = &CountingMemory{ICacheMemory: libMem.(port.ICacheMemory)}
	repo := repository.NewStuntingRepository(ctx, cfg, libDB.(port.IDatabase), mem)
	svc := service.NewStuntingService(ctx, cfg, repo, bg)
	stream := service.NewStreamService(ctx, svc, cfg, repo, bg)

	return &Wiring{
		Config:     cfg,
		Repository: repo,
		Service:    svc,
		Stream:     stream,
		Kafka:      handler.NewKafkaHandler(ctx, svc, stream, cfg, bg, &mem),
		Memory:     mem,
	}, nil
}
