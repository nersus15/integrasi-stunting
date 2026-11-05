package logtransaksi_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/nersus15/integrasi/tests/testutil"
	redislib "github.com/webcore-go/lib-redis"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/infra/logger"
)

// penampung catatan tertunda, dipakai bersama seperti di cluster
var rdb *redis.Client

// tanpa aplikasi webcore, logger global harus disiapkan sendiri
func TestMain(m *testing.M) {
	logger.PrepareLogger(context.Background(), "error")
	if err := os.Chdir(testutil.RepoRoot()); err != nil {
		panic(err)
	}
	cfg := config.Config{}
	if err := config.LoadDefaultConfig(&cfg); err != nil {
		panic(err)
	}
	rdb = redislib.NewRedis(cfg.Redis).Client
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		fmt.Fprintln(os.Stderr, "redis tidak hidup, test dilewati:", err)
		os.Exit(0)
	}
	code := m.Run()
	rdb.Close()
	os.Exit(code)
}
