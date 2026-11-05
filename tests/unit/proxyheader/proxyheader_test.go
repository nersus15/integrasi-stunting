package proxyheader_test

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webcore-go/webcore/infra/config"
)

func readIP(t *testing.T, proxyHeader string, header map[string]string) string {
	t.Helper()
	cfg := &config.Config{}
	cfg.Server.ProxyHeader = proxyHeader

	app := fiber.New(cfg.GetFiberConfig(nil))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString(c.IP()) })

	req := httptest.NewRequest("GET", "/", nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	res, err := app.Test(req)
	require.NoError(t, err)
	content, _ := io.ReadAll(res.Body)
	return string(content)
}

func TestIPKlienDariHeaderProxy(t *testing.T) {
	alamatKoneksi := readIP(t, "", nil)

	assert.Equal(t, "10.15.71.94", readIP(t, "X-Client-IP", map[string]string{"X-Client-IP": "10.15.71.94"}))
	assert.Equal(t, "2001:db8::1", readIP(t, "X-Client-IP", map[string]string{"X-Client-IP": "2001:db8::1"}))

	assert.Equal(t, alamatKoneksi, readIP(t, "X-Client-IP", nil), "tanpa header, pakai alamat koneksi")
	assert.Equal(t, alamatKoneksi, readIP(t, "X-Client-IP", map[string]string{"X-Client-IP": "bukan-ip"}), "nilai bukan IP diabaikan")
}

func TestTanpaProxyHeaderPerilakuLamaTetap(t *testing.T) {
	alamatKoneksi := readIP(t, "", nil)
	assert.Equal(t, alamatKoneksi, readIP(t, "", map[string]string{"X-Client-IP": "10.15.71.94", "X-Forwarded-For": "1.2.3.4"}),
		"header diabaikan kalau proxy_header kosong")
}
