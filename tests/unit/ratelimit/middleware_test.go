package ratelimit_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webcore-go/webcore/infra/middleware"
	"github.com/webcore-go/webcore/port"
	"github.com/webcore-go/webcore/port/auth"
)

type storeRusak struct{}

func (storeRusak) RateLimit(context.Context, string, int64, time.Duration) (port.RateLimitResult, error) {
	return port.RateLimitResult{}, errors.New("redis mati")
}

func newApp(store port.IRateLimitStore, failOpen bool) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		if id := c.Get("X-Uji-User"); id != "" {
			auth.PublishIdentity(c, "apikey", "kunci-"+id, &auth.UserAuthInfoRBAC{UserId: id})
		}
		return c.Next()
	})
	app.Use(middleware.NewRateLimit(store, middleware.RateLimitOptions{Limit: 2, Window: time.Minute, FailOpen: failOpen}))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })
	return app
}

func send(t *testing.T, app *fiber.App, user string) (int, string) {
	t.Helper()
	req := httptest.NewRequest("GET", "/", nil)
	if user != "" {
		req.Header.Set("X-Uji-User", user)
	}
	res, err := app.Test(req)
	require.NoError(t, err)
	return res.StatusCode, res.Header.Get("X-RateLimit-Remaining")
}

func TestMiddlewarePerIdentitas(t *testing.T) {
	app := newApp(middleware.NewMemoryRateLimitStore(), true)

	code, remaining := send(t, app, "jakantro")
	assert.Equal(t, 200, code)
	assert.Equal(t, "1", remaining)
	code, remaining = send(t, app, "jakantro")
	assert.Equal(t, 200, code)
	assert.Equal(t, "0", remaining)
	code, _ = send(t, app, "jakantro")
	assert.Equal(t, 429, code)

	code, _ = send(t, app, "faskes")
	assert.Equal(t, 200, code, "identitas lain tidak terpengaruh")
}

func TestMiddlewareTanpaIdentitasTidakDibatasi(t *testing.T) {
	app := newApp(middleware.NewMemoryRateLimitStore(), true)
	for i := 0; i < 5; i++ {
		code, remaining := send(t, app, "")
		assert.Equal(t, 200, code)
		assert.Empty(t, remaining)
	}
}

func TestMiddlewareStoreGagal(t *testing.T) {
	code, _ := send(t, newApp(storeRusak{}, true), "jakantro")
	assert.Equal(t, 200, code, "fail open meloloskan request")

	code, _ = send(t, newApp(storeRusak{}, false), "jakantro")
	assert.Equal(t, 503, code, "fail closed menolak")
}

func TestKeyTidakMenyimpanIdentitasMentah(t *testing.T) {
	k := middleware.RateLimitKey("4a088f54847956e3f828a2938deee1c1")
	assert.NotContains(t, k, "4a088f54847956e3f828a2938deee1c1")
	assert.True(t, strings.HasPrefix(k, "ratelimit:{") && strings.HasSuffix(k, "}"))
	assert.Equal(t, k, middleware.RateLimitKey("4a088f54847956e3f828a2938deee1c1"))
	assert.NotEqual(t, k, middleware.RateLimitKey("lain"))
}
