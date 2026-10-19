package stunting

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/app/core"
)

func (m *Module) registerModuleRoute(root fiber.Router) {
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/kafka/gagal",
		Handler: m.handler.DaftarKafkaGagal,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/kafka/gagal/retry",
		Handler: m.handler.RetrySemuaKafkaGagal,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/kafka/gagal/:id/retry",
		Handler: m.handler.RetryKafkaGagal,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/orangtua/:id?",
		Handler: m.handler.FindOrangTua,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/orangtua/:id/anak",
		Handler: m.handler.ListAnak,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/anak/:id?",
		Handler: m.handler.FindAnak,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/anak/:id/kunjungan",
		Handler: m.handler.FindKunjunganByIdAnak,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/anak/:id/kesehatan",
		Handler: m.handler.FindKesehatanByIdAnak,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/anak/:id/summary",
		Handler: m.handler.SummaryAnak,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/kunjungan/:id",
		Handler: m.handler.FindKunjunganById,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/kesehatan/:id",
		Handler: m.handler.FindKesehatanById,
		Root:    root,
	})
}

func (m *Module) registerApiJakantroRoute(root fiber.Router) {
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/orangtua",
		Handler: m.handler.CreateOrangTua,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/anak",
		Handler: m.handler.CreateAnak,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/kunjungan",
		Handler: m.handler.CreateKunjungan,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/kesehatan",
		Handler: m.handler.CreateKesehatan,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "PUT",
		Path:    "/orangtua",
		Handler: m.handler.UpdateOrangtua,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "PUT",
		Path:    "/anak",
		Handler: m.handler.UpdateAnak,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "PUT",
		Path:    "/kunjungan/:id",
		Handler: m.handler.UpdateKunjunganById,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "PUT",
		Path:    "/kesehatan/:id",
		Handler: m.handler.UpdateKesehatanById,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "DELETE",
		Path:    "/anak/:id",
		Handler: m.handler.DeleteAnakByID,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "DELETE",
		Path:    "/orangtua/:id",
		Handler: m.handler.DeleteOrangtuaByID,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "DELETE",
		Path:    "/kunjungan/:id",
		Handler: m.handler.DeleteKunjunganByID,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "DELETE",
		Path:    "/kesehatan/:id",
		Handler: m.handler.DeleteKesehatanByID,
		Root:    root,
	})
}

func (m *Module) registerApiFaskes(root fiber.Router) {
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/faskes/anak",
		Handler: m.handler.CreateAnakFaskes,
		Root:    root,
	})
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/faskes/orangtua",
		Handler: m.handler.CreateOrangTuaFaskes,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "POST",
		Path:    "/faskes/pemeriksaan",
		Handler: m.handler.SimpanPemeriksaanFaskes,
		Root:    root,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "PUT",
		Path:    "/faskes/:resourceName",
		Handler: m.handler.UpdateResourceBySatusehatId,
		Root:    root,
	})
}
func (m *Module) registerRootRoute(web *fiber.App) {
	moduleRoot := web.Group("/" + m.Name())
	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/health",
		Handler: m.Health,
		Root:    moduleRoot,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/info",
		Handler: m.Info,
		Root:    moduleRoot,
	})

	m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
		Method:  "GET",
		Path:    "/docs",
		Handler: m.Docs,
		Root:    moduleRoot,
	})
}

func (m *Module) Health(c *fiber.Ctx) error {
	health := map[string]any{
		"status":    "healthy",
		"module":    ModuleName,
		"version":   ModuleVersion,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	return c.JSON(health)
}

func (m *Module) Info(c *fiber.Ctx) error {
	endpoints := []string{}
	for _, endpoint := range m.routes {
		endpoint := endpoint.Method + " " + endpoint.Path
		endpoints = append(endpoints, endpoint)
	}

	path := "/" + ModuleName

	info := map[string]any{
		"name":        ModuleName,
		"version":     ModuleVersion,
		"description": "Integrasi data Stunting",
		"path":        path,
		"endpoints":   endpoints,
		"config":      m.config,
	}
	return c.JSON(info)
}
