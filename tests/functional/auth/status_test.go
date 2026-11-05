package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/adapter/auth/apikey"
	"github.com/webcore-go/webcore/adapter/auth/authn"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port/auth"
)

// Jalur otorisasi memanggil logger.DebugJson, dan logger webcore panic bila
// belum disiapkan (logDefault() mengembalikan nil, methodnya tidak menjaga
// receiver nil). Service sungguhan menyiapkannya di core.App.Start; test
// harus melakukan hal yang sama.
func init() {
	logger.PrepareLogger(context.Background(), "error")
}

// Tiga kegagalan yang berbeda harus dijawab berbeda:
//
//	kredensial tidak ada / tidak dikenali  -> 401 UNAUTHORIZED  (perbaiki kunci)
//	dikenali tapi hak kurang               -> 403 FORBIDDEN     (jangan ulangi)
//	jalur otorisasi rusak                  -> 500 INTERNAL      (masalah kami)
//
// Sebelumnya ketiganya dijawab 401, sehingga klien yang menerima "User access
// denied" akan menyimpulkan kuncinya salah lalu mencoba ulang selamanya.

// storeAtur adalah store palsu yang bisa diatur untuk tiap skenario.
type storeAtur struct {
	peranUser []string
	resource  auth.IResourceInfo
	errUser   error
	errRes    error
}

func (s *storeAtur) GetUserLoginInfo(ctx *fiber.Ctx, u, p string) (auth.IUserAuthInfo, error) {
	return nil, errors.New("tidak dipakai")
}

func (s *storeAtur) GetUserAuthInfo(ctx *fiber.Ctx, v auth.IAuthValidator, userKey string) (auth.IUserAuthInfo, error) {
	if s.errUser != nil {
		return nil, s.errUser
	}
	return &auth.UserAuthInfoRBAC{UserId: userKey, Roles: s.peranUser}, nil
}

func (s *storeAtur) GetResourceInfo(method, path string) (auth.IResourceInfo, error) {
	return s.resource, s.errRes
}

// handlerAuth merakit rantai autentikasi sungguhan di atas store yang diberikan.
func handlerAuth(store auth.IStore) fiber.Handler {
	cfg := config.AuthConfig{Type: "apikey", Control: "RBAC", APIKeyHeader: "X-API-Key"}

	a := &authn.AuthN{}
	a.SetValidator(apikey.NewApiKeyValidator(cfg))
	pembungkus := auth.NewStoreWrapper(store)
	a.Authenticator = auth.NewAuthenticator(cfg, a.Validator, pembungkus, nil)
	a.Authorizer, _ = auth.NewAuthorization(pembungkus)

	return a.GetAuthenticatonHandler()
}

func buildWith(store auth.IStore) *fiber.App {
	app := fiber.New()
	app.Use(handlerAuth(store))
	app.Post("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	return app
}

type reqBody struct {
	HttpCode  int    `json:"httpCode"`
	ErrorCode int    `json:"errorCode"`
	ErrorName string `json:"errorName"`
	Message   string `json:"message"`
}

func call(t *testing.T, app *fiber.App, key string) (int, reqBody) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	if key != "" {
		req.Header.Set("X-API-Key", key)
	}

	res, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("request gagal: %v", err)
	}
	defer res.Body.Close()

	content, _ := io.ReadAll(res.Body)
	var b reqBody
	_ = json.Unmarshal(content, &b)
	return res.StatusCode, b
}

func TestStatusPenolakanDibedakan(t *testing.T) {
	butuhWrite := &auth.ResourceInfoRBAC{
		Action: "tulis", Path: "/", Method: "POST",
		PermittedRoles: []string{"write"},
	}

	cases := []struct {
		nama      string
		store     auth.IStore
		kunci     string
		http      int
		errorCode int
		errorName string
	}{
		{
			nama:      "tanpa kredensial",
			store:     &storeAtur{peranUser: []string{"write"}},
			kunci:     "",
			http:      http.StatusUnauthorized,
			errorCode: 4001, errorName: "UNAUTHORIZED",
		},
		{
			nama:      "kredensial tidak dikenali",
			store:     &storeAtur{errUser: errors.New("user tidak ada")},
			kunci:     "kunci-asing",
			http:      http.StatusUnauthorized,
			errorCode: 4001, errorName: "UNAUTHORIZED",
		},
		{
			nama:      "dikenali, hak cukup",
			store:     &storeAtur{peranUser: []string{"write"}, resource: butuhWrite},
			kunci:     "kunci-sah",
			http:      http.StatusOK,
			errorCode: 0, errorName: "",
		},
		{
			nama:      "dikenali, hak kurang",
			store:     &storeAtur{peranUser: []string{"read"}, resource: butuhWrite},
			kunci:     "kunci-sah",
			http:      http.StatusForbidden,
			errorCode: 4002, errorName: "FORBIDDEN",
		},
		{
			nama:      "jalur otorisasi rusak",
			store:     &storeAtur{peranUser: []string{"write"}, errRes: errors.New("database mati")},
			kunci:     "kunci-sah",
			http:      http.StatusInternalServerError,
			errorCode: 5001, errorName: "INTERNAL_ERROR",
		},
	}

	for _, k := range cases {
		t.Run(k.nama, func(t *testing.T) {
			code, b := call(t, buildWith(k.store), k.kunci)

			if code != k.http {
				t.Errorf("http %d, harap %d (message: %q)", code, k.http, b.Message)
			}
			if k.errorCode != 0 && b.ErrorCode != k.errorCode {
				t.Errorf("errorCode %d, harap %d", b.ErrorCode, k.errorCode)
			}
			if k.errorName != "" && b.ErrorName != k.errorName {
				t.Errorf("errorName %q, harap %q", b.ErrorName, k.errorName)
			}
		})
	}
}

// Sebab kegagalan di sisi layanan tidak boleh bocor ke klien.
func TestKegagalanServerTidakBocor(t *testing.T) {
	store := &storeAtur{peranUser: []string{"write"}, errRes: errors.New("password database salah")}

	_, b := call(t, buildWith(store), "kunci-sah")

	if b.Message == "" {
		t.Fatal("message kosong")
	}
	if b.Message != "Terjadi kesalahan" {
		t.Errorf("message %q membocorkan sebab aslinya", b.Message)
	}
}
