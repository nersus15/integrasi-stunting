package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/port/auth"
)

// Memastikan identitas request benar-benar sampai ke handler modul, dan
// identitas itu milik request yang bersangkutan.
//
// Dulu ini tidak ada test-nya, dan akibatnya tidak ada yang sadar bahwa
// ApiKeyValidator sama sekali tidak pernah mengisi konteks -- handler modul
// selalu menerima kosong untuk auth.type: apikey. Sekarang penerbitan
// identitas dilakukan satu kali oleh handler autentikasi lewat
// auth.PublishIdentity, sehingga berlaku seragam untuk semua jenis auth.
func TestIdentitasSampaiKeHandler(t *testing.T) {
	app := fiber.New()
	app.Use(buildHandler())

	var (
		dilayani    int64
		jenisSalah  int64
		userKosong  int64
		idSalah     int64
		peranKosong int64
	)

	app.Get("/", func(c *fiber.Ctx) error {
		sent := c.Get("X-API-Key")
		atomic.AddInt64(&dilayani, 1)

		if auth.GetAuthType(c) != "apikey" {
			atomic.AddInt64(&jenisSalah, 1)
		}
		if auth.GetUser(c) == nil {
			atomic.AddInt64(&userKosong, 1)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		// storePalsu mengisi UserId dengan kunci itu sendiri, jadi id yang
		// terbaca harus sama dengan kunci yang dikirim request ini.
		if auth.GetUserID(c) != sent {
			atomic.AddInt64(&idSalah, 1)
		}
		peran := auth.GetUserRoles(c)
		if len(peran) != 1 || peran[0] != "admin" {
			atomic.AddInt64(&peranKosong, 1)
		}
		return c.SendString(auth.GetUserID(c))
	})

	const n = 1000
	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start

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
		t.Errorf("hanya %d dari %d request sampai ke handler", dilayani, n)
	}
	if jenisSalah != 0 {
		t.Errorf("%d request melihat auth_type bukan \"apikey\"", jenisSalah)
	}
	if userKosong != 0 {
		t.Errorf("%d request tidak menerima identitas sama sekali", userKosong)
	}
	if idSalah != 0 {
		t.Errorf("%d request menerima id milik request lain", idSalah)
	}
	if peranKosong != 0 {
		t.Errorf("%d request tidak menerima daftar peran", peranKosong)
	}
}

// Tanpa kredensial, handler tidak boleh terpanggil sama sekali dan konteks
// harus tetap kosong.
func TestTanpaKredensialTidakAdaIdentitas(t *testing.T) {
	app := fiber.New()
	app.Use(buildHandler())

	var terpanggil int64
	app.Get("/", func(c *fiber.Ctx) error {
		atomic.AddInt64(&terpanggil, 1)
		return c.SendStatus(fiber.StatusOK)
	})

	res, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil), 10000)
	if err != nil {
		t.Fatalf("request gagal: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("http %d, harap 401", res.StatusCode)
	}
	if terpanggil != 0 {
		t.Errorf("handler terpanggil %d kali padahal tanpa kredensial", terpanggil)
	}
}
