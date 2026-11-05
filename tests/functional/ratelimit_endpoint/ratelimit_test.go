package ratelimit_endpoint_test

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
	"github.com/webcore-go/webcore/infra/config"
)

var (
	app *testutil.App
	db  *sql.DB
	key = "uji-rl-" + strings.ReplaceAll(uuid.NewString(), "-", "")
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

	app, err = testutil.StartAppWith(2093, func(c *config.Config) {
		c.App.RateLimit.Enabled = true
		c.App.RateLimit.Max = 3
		c.App.RateLimit.Window = time.Minute
		c.App.RateLimit.Backend = "redis"
		c.App.RateLimit.FailOpen = true
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

func request(t *testing.T, url, key string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest("GET", url, nil)
	if key != "" {
		req.Header.Set("X-API-Key", key)
	}
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	res.Body.Close()
	return res
}

func TestKuotaPerApiKey(t *testing.T) {
	for _, remaining := range []string{"2", "1", "0"} {
		res := request(t, app.BaseURL+"/kafka/gagal", key)
		require.Equal(t, 200, res.StatusCode)
		assert.Equal(t, "3", res.Header.Get("X-RateLimit-Limit"))
		assert.Equal(t, remaining, res.Header.Get("X-RateLimit-Remaining"))
	}
	assert.Equal(t, 429, request(t, app.BaseURL+"/kafka/gagal", key).StatusCode)
}

func TestKunciTidakSahTidakMembuatBucket(t *testing.T) {
	for i := 0; i < 10; i++ {
		res := request(t, app.BaseURL+"/kafka/gagal", "acak-"+uuid.NewString())
		assert.Equal(t, 401, res.StatusCode, "ditolak autentikasi, bukan dihitung limiter")
		assert.Empty(t, res.Header.Get("X-RateLimit-Limit"))
	}
	assert.Equal(t, 401, request(t, app.BaseURL+"/kafka/gagal", "").StatusCode)
}

func TestEndpointPublikTidakDibatasi(t *testing.T) {
	base := strings.TrimSuffix(app.Health, "/health")
	for i := 0; i < 10; i++ {
		for _, path := range []string{"/health", "/docs"} {
			res := request(t, base+path, "")
			assert.Equal(t, 200, res.StatusCode, path)
			assert.Empty(t, res.Header.Get("X-RateLimit-Limit"), path)
		}
	}
}
