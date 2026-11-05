// Package auth_test menjaga satu sifat: kunci dan identitas yang dipakai satu
// request tidak boleh berasal dari request lain.
//
//	go test -race ./functional/auth/
//
// Latar belakangnya: validator dan StoreWrapper dibuat sekali saat service
// menyala lalu dipakai bersama seluruh request. Dulu keduanya menyimpan kunci
// dan user sebagai field, sehingga request yang datang bersamaan bisa saling
// menimpa — request dengan kunci sampah bisa lolos memakai identitas request
// lain yang sah. Test ini gagal bila pola itu kembali.
package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/adapter/auth/apikey"
	"github.com/webcore-go/webcore/adapter/auth/authn"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/port/auth"
)

// storePalsu menggantikan database. Ia mencatat kunci yang benar-benar sampai
// ke lapisan store pada Locals request bersangkutan, supaya handler di
// belakang bisa membandingkannya dengan kunci yang dikirim.
type storePalsu struct{}

func (s *storePalsu) GetUserLoginInfo(ctx *fiber.Ctx, username, password string) (auth.IUserAuthInfo, error) {
	return nil, fmt.Errorf("tidak dipakai")
}

func (s *storePalsu) GetUserAuthInfo(ctx *fiber.Ctx, validator auth.IAuthValidator, userKey string) (auth.IUserAuthInfo, error) {
	ctx.Locals("kunci_sampai_store", userKey)
	// UserId diisi kunci itu sendiri supaya ApiKeyValidator.VerifyUser lolos.
	return &auth.UserAuthInfoRBAC{UserId: userKey, Roles: []string{"admin"}}, nil
}

// nil resource berarti "tanpa aturan izin" -- Authorization.Check mengizinkan.
func (s *storePalsu) GetResourceInfo(method, path string) (auth.IResourceInfo, error) {
	return nil, nil
}

// rakitHandler menyusun rantai auth sungguhan: validator -> Authenticator ->
// Authorization, persis seperti yang dipasang webcore saat service menyala.
func buildHandler() fiber.Handler {
	cfg := config.AuthConfig{Type: "apikey", Control: "RBAC", APIKeyHeader: "X-API-Key"}

	a := &authn.AuthN{}
	a.SetValidator(apikey.NewApiKeyValidator(cfg))

	pembungkus := auth.NewStoreWrapper(&storePalsu{})
	a.Authenticator = auth.NewAuthenticator(cfg, a.Validator, pembungkus, nil)

	otorisasi, err := auth.NewAuthorization(pembungkus)
	if err != nil {
		panic(err)
	}
	a.Authorizer = otorisasi

	return a.GetAuthenticatonHandler()
}

func TestIdentitasTidakTertukarAntarRequest(t *testing.T) {
	app := fiber.New()
	app.Use(buildHandler())

	var tertukar, dilayani int64
	app.Get("/", func(c *fiber.Ctx) error {
		sent := c.Get("X-API-Key")

		sampai, ok := c.Locals("kunci_sampai_store").(string)
		if !ok {
			t.Error("store tidak pernah dipanggil")
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		atomic.AddInt64(&dilayani, 1)
		if sampai != sent {
			atomic.AddInt64(&tertukar, 1)
		}
		return c.SendString(sampai)
	})

	const n = 2000
	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start // semua berangkat bersamaan, memperbesar peluang tumpang tindih

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-API-Key", fmt.Sprintf("kunci-%06d", i))

			res, err := app.Test(req, 10000)
			if err != nil {
				t.Errorf("request gagal: %v", err)
				return
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Errorf("http %d, harap 200", res.StatusCode)
			}
		}(i)
	}

	close(start)
	wg.Wait()

	if dilayani != n {
		t.Errorf("hanya %d dari %d request terlayani", dilayani, n)
	}
	if tertukar != 0 {
		t.Errorf("%d dari %d request memakai kunci milik request lain", tertukar, dilayani)
	}
}
