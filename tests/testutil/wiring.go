package testutil

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/handler"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/repository"
	"github.com/nersus15/integrasi/mod-stunting/service"
	backgroundworker "github.com/nersus15/lib-background-worker"
	"github.com/webcore-go/webcore/app/core"
	appConfig "github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/infra/logger"
)

const APIPrefix = "/api"

func Ptr[T any](v T) *T { return &v }

// siapkanLogger memastikan logger global webcore sudah diinisialisasi.
// Di produksi ini dilakukan core.NewApp; tanpa ini logger.Error() panic.
var loggerOnce sync.Once

func setupLogger() {
	loggerOnce.Do(func() {
		logger.PrepareLogger(context.Background(), "error")
	})
}

func NewAppContext() *core.AppContext {
	return NewAppContextWith(context.Background())
}

func NewAppContextWith(ctx context.Context) *core.AppContext {
	return &core.AppContext{Context: ctx, Config: NewAppConfig()}
}

func NewAppConfig() *appConfig.Config {
	cfg := &appConfig.Config{}
	cfg.App.Name = "integrasi-stunting-test"
	cfg.App.Environment = "development"
	cfg.App.Features.Recovery = true
	cfg.Server.PathPrefix = APIPrefix
	cfg.Server.Host = "127.0.0.1"
	cfg.Auth.Type = "none"
	cfg.Database.Driver = "postgres"
	return cfg
}

func NewModuleConfig() *config.ModuleConfig {
	cfg := &config.ModuleConfig{}
	cfg.Cron.Enabled = false
	return cfg
}

// NewBackgroundWorker membuat worker sungguhan dengan pool kecil supaya
// DrainBackground bisa menunggunya sampai bersih.
func NewBackgroundWorker(t testing.TB) *backgroundworker.BackgroundWorker {
	t.Helper()

	w := backgroundworker.New(backgroundworker.Config{
		Workers:           1,
		PriorityWorkers:   1,
		QueueSize:         32,
		PriorityQueueSize: 32,
		ShutdownTimeout:   10 * time.Second,
	})
	w.Start()
	t.Cleanup(func() { _ = w.Stop() })

	return w
}

type Stack struct {
	Context    *core.AppContext
	Config     *config.ModuleConfig
	DB         *FakeDatabase
	Memory     *FakeMemory
	Repository *repository.StuntingRepository
	Service    *service.StuntingService
	Stream     *service.StreamService
	Handler    *handler.HttpHandler
	Background *backgroundworker.BackgroundWorker
}

// NewStack merangkai repository -> service -> handler meniru urutan Module.Init.
func NewStack(t testing.TB, db *FakeDatabase) *Stack {
	t.Helper()
	return NewStackWithContext(t, NewAppContext(), db)
}

func NewStackWithContext(t testing.TB, ctx *core.AppContext, db *FakeDatabase) *Stack {
	t.Helper()
	setupLogger()

	cfg := NewModuleConfig()
	worker := NewBackgroundWorker(t)

	mem := NewFakeMemory()
	repo := repository.NewStuntingRepository(ctx, cfg, db, mem)
	svc := service.NewStuntingService(ctx, cfg, repo, worker)
	stream := service.NewStreamService(ctx, svc, cfg, repo, worker)
	hdl := handler.NewHttpHandler(ctx, cfg, svc, stream)

	return &Stack{
		Context:    ctx,
		Config:     cfg,
		DB:         db,
		Memory:     mem,
		Repository: repo,
		Service:    svc,
		Stream:     stream,
		Handler:    hdl,
		Background: worker,
	}
}

// DrainBackground menunggu seluruh job background selesai supaya assertion
// terhadap efek sampingnya tidak balapan dengan worker.
func (s *Stack) DrainBackground(t testing.TB) {
	t.Helper()

	if err := s.Background.Stop(); err != nil {
		t.Fatalf("gagal menunggu background worker: %v", err)
	}
}

// CatchPanic menjalankan fn dan mengembalikan nilai panic-nya, atau nil.
func CatchPanic(fn func()) (value any) {
	defer func() { value = recover() }()
	fn()
	return nil
}

// CreateKunjungan memanggil service dan menggagalkan test bila panic, bukan
// mematikan seluruh paket. Tanpa ini satu panic menyembunyikan hasil test lain.
func (s *Stack) CreateKunjungan(t testing.TB, body []byte) (*types.KunjunganAnak, error) {
	res, err, _, _ := s.CreateKunjunganFull(t, body)
	return res, err
}

// CreateKunjunganFull juga mengembalikan status dan pesan yang dipilih service,
// karena keduanya yang dipakai handler untuk menyusun response.
func (s *Stack) CreateKunjunganFull(t testing.TB, body []byte) (*types.KunjunganAnak, error, int, string) {
	t.Helper()

	var res *types.KunjunganAnak
	var err error

	if value := CatchPanic(func() { res, err = s.Service.CreateKunjungan(body) }); value != nil {
		t.Fatalf("service panic, seharusnya mengembalikan error: %v", value)
	}

	code, message := http.StatusCreated, ""
	if err != nil {
		k := exceptions.Classify(err)
		code, message = k.HttpCode, k.Message
	}

	return res, err, code, message
}

// JSONBody merapikan map jadi body request.
func JSONBody(t testing.TB, v any) []byte {
	t.Helper()

	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("gagal marshal body: %v", err)
	}
	return raw
}

var counter struct {
	sync.Mutex
	n int
}

// UniqueId menghasilkan id 36 karakter yang unik, meniru panjang kolom
// character(36) pada skema.
func UniqueId(prefix string) string {
	counter.Lock()
	counter.n++
	n := counter.n
	counter.Unlock()

	id := prefix + "-"
	for i := len(id); i < 35; i++ {
		id += "0"
	}
	return id + string(rune('a'+n%26))
}

func SetupLogger() { setupLogger() }
