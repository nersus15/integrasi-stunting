package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/integrasi/mod-stunting/config"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/service"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/out"
)

type HttpHandler struct {
	service *service.StuntingService
	stream  *service.StreamService
	config  *config.ModuleConfig
}

func NewHttpHandler(wctx *core.AppContext, cfg *config.ModuleConfig, service *service.StuntingService, stream *service.StreamService) *HttpHandler {
	return &HttpHandler{
		service: service,
		stream:  stream,
		config:  cfg,
	}
}

func kirimError(c *fiber.Ctx, err error) error {
	k := exceptions.Classify(err)

	return c.Status(k.HttpCode).JSON(out.ErrorDetail(
		k.HttpCode,
		k.ErrorCode,
		k.Name,
		k.Message,
		k,
	))
}

func orgidPemanggil(c *fiber.Ctx) string {
	orgid, _ := c.Locals("orgid").(*string)
	return utils.StrPtr(orgid)
}
