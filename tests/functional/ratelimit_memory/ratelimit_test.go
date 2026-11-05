package ratelimit_memory_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nersus15/integrasi/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	memlib "github.com/webcore-go/lib-memory"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/infra/middleware"
)

var (
	app *testutil.App
	db  *sql.DB
	key = "uji-rlm-" + strings.ReplaceAll(uuid.NewString(), "-", "")
)

func TestMain(m *testing.M) {
	if !testutil.DatabaseAvailable() {
		os.Exit(0)
	}
	db = testutil.OpenDB()
	_, err := db.Exec(`WITH u AS (
		INSERT INTO access.auth_users (username, password_hash, groups, roles)
		VALUES ($1, 'x', '{jakantro}', '{read,write}') RETURNING id)
		INSERT INTO access.auth_apikey (api_key, user_id) SELECT $1, id FROM u`, key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal membuat api key uji: %v\n", err)
		os.Exit(1)
	}

	app, err = testutil.StartAppWith(2092, func(c *config.Config) {
		c.Memory.Enabled = true
		c.App.RateLimit.Enabled = true
		c.App.RateLimit.Max = 3
		c.App.RateLimit.Window = time.Minute
		c.App.RateLimit.Backend = "memory"
	})
	code := 1
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal menjalankan aplikasi: %v\n", err)
	} else {
		code = m.Run()
		app.Stop()
	}
	_, _ = db.Exec(`DELETE FROM access.auth_users WHERE username = $1`, key)
	db.Close()
	os.Exit(code)
}

func request(t *testing.T, key string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest("GET", app.BaseURL+"/kafka/gagal", nil)
	req.Header.Set("X-API-Key", key)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	res.Body.Close()
	return res
}

func TestBackendMemoryMemakaiLibMemory(t *testing.T) {
	for _, remaining := range []string{"2", "1", "0"} {
		res := request(t, key)
		require.Equal(t, 200, res.StatusCode)
		assert.Equal(t, remaining, res.Header.Get("X-RateLimit-Remaining"))
	}
	assert.Equal(t, 429, request(t, key).StatusCode)

	lib, ok := app.Singleton("cache:memory")
	require.True(t, ok, "library cache:memory dimuat")
	cache := lib.(*memlib.MemoryCache)

	sekarang := time.Now().UnixMilli()
	initial := sekarang - sekarang%time.Minute.Milliseconds()
	e, ok := cache.Cache.GetEntryQuietly(fmt.Sprintf("%s:%d", middleware.RateLimitKey(key), initial))
	require.True(t, ok, "counter tersimpan di lib-memory, bukan di penyimpanan bawaan webcore")
	assert.Equal(t, "3", e.Value)
}

func TestKunciTidakSahTidakDihitung(t *testing.T) {
	for i := 0; i < 10; i++ {
		res := request(t, "acak-"+uuid.NewString())
		assert.Equal(t, 401, res.StatusCode)
		assert.Empty(t, res.Header.Get("X-RateLimit-Limit"))
	}
}
