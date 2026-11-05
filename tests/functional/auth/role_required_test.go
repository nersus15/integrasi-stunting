package auth_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/infra/middleware"
	"github.com/webcore-go/webcore/port/auth"
)

// RoleRequired dipakai per route, sebagai pelengkap tabel access.auth_resources
// untuk aturan yang tidak bisa dinyatakan tabel.
//
// Versi lama fungsi ini menolak SEMUA orang: cabang apikey tidak pernah kena
// (tidak ada yang mengisi auth_type) dan cabang jwt selalu nil (membaca kunci
// "user_role" sementara yang ditulis "role"). Test ini memastikan versi
// sekarang benar-benar membedakan.

// appDenganRoleRequired merakit rantai lengkap: autentikasi lalu RoleRequired.
func appWithRoleRequired(peranUser []string, wajib ...string) *fiber.App {
	store := &storeAtur{peranUser: peranUser}

	app := fiber.New()
	app.Use(handlerAuth(store))
	app.Post("/", middleware.RoleRequired(wajib...), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

func TestRoleRequiredMembedakanPeran(t *testing.T) {
	cases := []struct {
		nama      string
		peranUser []string
		wajib     []string
		http      int
		errorCode int
	}{
		{"peran cocok", []string{"write"}, []string{"write"}, http.StatusOK, 0},
		{"salah satu cocok", []string{"read", "write"}, []string{"write"}, http.StatusOK, 0},
		{"cocok salah satu dari beberapa", []string{"admin"}, []string{"write", "admin"}, http.StatusOK, 0},
		{"peran tidak cocok", []string{"read"}, []string{"write"}, http.StatusForbidden, 4002},
		{"tanpa peran sama sekali", nil, []string{"write"}, http.StatusForbidden, 4002},
	}

	for _, k := range cases {
		t.Run(k.nama, func(t *testing.T) {
			code, b := call(t, appWithRoleRequired(k.peranUser, k.wajib...), "kunci-sah")

			if code != k.http {
				t.Errorf("http %d, harap %d (message: %q)", code, k.http, b.Message)
			}
			if k.errorCode != 0 {
				if b.ErrorCode != k.errorCode {
					t.Errorf("errorCode %d, harap %d", b.ErrorCode, k.errorCode)
				}
				if b.ErrorName != "FORBIDDEN" {
					t.Errorf("errorName %q, harap FORBIDDEN", b.ErrorName)
				}
			}
		})
	}
}

// Dipasang di luar jangkauan autentikasi, identitas tidak pernah terbit.
// Jawabannya harus 401 -- bukan 403, karena layanan belum tahu siapa pemanggil,
// dan bukan 200, yang berarti middleware-nya bisa dilewati begitu saja.
func TestRoleRequiredDiLuarJangkauanAutentikasi(t *testing.T) {
	app := fiber.New()
	app.Post("/", middleware.RoleRequired("write"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	code, b := call(t, app, "kunci-sah")

	if code != http.StatusUnauthorized {
		t.Errorf("http %d, harap 401", code)
	}
	if b.ErrorCode != 4001 || b.ErrorName != "UNAUTHORIZED" {
		t.Errorf("code=%d name=%q, harap 4001 UNAUTHORIZED", b.ErrorCode, b.ErrorName)
	}
}

// Pemakai ABAC tidak punya daftar peran, jadi RoleRequired menolaknya.
func TestRoleRequiredMenolakABAC(t *testing.T) {
	store := &storeABAC{}

	app := fiber.New()
	app.Use(handlerAuth(store))
	app.Post("/", middleware.RoleRequired("write"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	code, _ := call(t, app, "kunci-sah")
	if code != http.StatusForbidden {
		t.Errorf("http %d, harap 403", code)
	}
}

type storeABAC struct{}

func (s *storeABAC) GetUserLoginInfo(ctx *fiber.Ctx, u, p string) (auth.IUserAuthInfo, error) {
	return nil, nil
}

func (s *storeABAC) GetUserAuthInfo(ctx *fiber.Ctx, v auth.IAuthValidator, userKey string) (auth.IUserAuthInfo, error) {
	return &auth.UserAuthInfoABAC{UserId: userKey}, nil
}

func (s *storeABAC) GetResourceInfo(method, path string) (auth.IResourceInfo, error) {
	return nil, nil
}
