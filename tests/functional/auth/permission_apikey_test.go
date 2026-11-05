package auth_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/infra/middleware"
	"github.com/webcore-go/webcore/port/auth"
)

// storePolicy mengembalikan pemakai ABAC dengan kebijakan yang bisa diatur.
type storePolicy struct{ policies []auth.PolicyABAC }

func (s *storePolicy) GetUserLoginInfo(ctx *fiber.Ctx, u, p string) (auth.IUserAuthInfo, error) {
	return nil, nil
}

func (s *storePolicy) GetUserAuthInfo(ctx *fiber.Ctx, v auth.IAuthValidator, userKey string) (auth.IUserAuthInfo, error) {
	return &auth.UserAuthInfoABAC{UserId: userKey, Policies: s.policies}, nil
}

func (s *storePolicy) GetResourceInfo(method, path string) (auth.IResourceInfo, error) {
	return nil, nil
}

func appWithPermission(store auth.IStore, aksi string) *fiber.App {
	app := fiber.New()
	app.Use(handlerAuth(store))
	app.Post("/", middleware.PermissionRequired(aksi), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

func TestPermissionRequiredMenilaiKebijakan(t *testing.T) {
	izinkan := func(aksi string, syarat ...auth.ConditionABAC) auth.PolicyABAC {
		return auth.PolicyABAC{Effect: "Allow", Action: aksi, Condition: syarat}
	}

	cases := []struct {
		nama     string
		policies []auth.PolicyABAC
		aksi     string
		http     int
	}{
		{"aksi cocok", []auth.PolicyABAC{izinkan("tulis")}, "tulis", http.StatusOK},
		{"wildcard", []auth.PolicyABAC{izinkan("*")}, "tulis", http.StatusOK},
		{"aksi lain", []auth.PolicyABAC{izinkan("baca")}, "tulis", http.StatusForbidden},
		{"tanpa kebijakan", nil, "tulis", http.StatusForbidden},
		{
			nama:     "Deny tidak memberi lolos",
			policies: []auth.PolicyABAC{{Effect: "Deny", Action: "tulis"}},
			aksi:     "tulis", http: http.StatusForbidden,
		},
		{
			// Syaratnya butuh atribut resource yang belum ada di titik ini.
			// Bila ragu, jawabannya tidak.
			nama:     "kebijakan bersyarat tidak dinilai di sini",
			policies: []auth.PolicyABAC{izinkan("tulis", auth.ConditionABAC{Attribute: "dept", Operator: "==", Value: "A"})},
			aksi:     "tulis", http: http.StatusForbidden,
		},
	}

	for _, k := range cases {
		t.Run(k.nama, func(t *testing.T) {
			code, b := call(t, appWithPermission(&storePolicy{policies: k.policies}, k.aksi), "kunci-sah")
			if code != k.http {
				t.Errorf("http %d, harap %d (message: %q)", code, k.http, b.Message)
			}
			if k.http == http.StatusForbidden && b.ErrorCode != 4002 {
				t.Errorf("errorCode %d, harap 4002", b.ErrorCode)
			}
		})
	}
}

// Pemakai RBAC tidak punya daftar kebijakan, jadi PermissionRequired menolaknya
// -- pasangan cermin dari RoleRequired yang menolak pemakai ABAC.
func TestPermissionRequiredMenolakRBAC(t *testing.T) {
	store := &storeAtur{peranUser: []string{"write"}}

	code, _ := call(t, appWithPermission(store, "tulis"), "kunci-sah")
	if code != http.StatusForbidden {
		t.Errorf("http %d, harap 403", code)
	}
}

func TestPermissionRequiredDiLuarJangkauanAutentikasi(t *testing.T) {
	app := fiber.New()
	app.Post("/", middleware.PermissionRequired("tulis"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	code, b := call(t, app, "kunci-sah")
	if code != http.StatusUnauthorized || b.ErrorCode != 4001 {
		t.Errorf("http=%d code=%d, harap 401/4001", code, b.ErrorCode)
	}
}

// GetAPIKey harus mengembalikan kunci milik request ini, dan hanya untuk
// jenis auth apikey.
func TestGetAPIKeyTerisiDanMilikRequestSendiri(t *testing.T) {
	store := &storeAtur{peranUser: []string{"write"}}

	app := fiber.New()
	app.Use(handlerAuth(store))
	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendString(middleware.GetAPIKey(c))
	})

	const key = "kunci-uji-123456"
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-API-Key", key)

	res, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("request gagal: %v", err)
	}
	defer res.Body.Close()

	content, _ := io.ReadAll(res.Body)
	if string(content) != key {
		t.Errorf("GetAPIKey = %q, harap %q", string(content), key)
	}
}
