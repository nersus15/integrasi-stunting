package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/app/out"

	"github.com/nersus15/integrasi/mod-satusehat/entity"
	"github.com/nersus15/integrasi/mod-satusehat/service"
)

type Handler struct {
	service service.SatusehatService
}

func NewHandler(ctx *core.AppContext, svc service.SatusehatService) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) ListSatusehats(c *fiber.Ctx) error {
	items, err := h.service.GetSatusehats(c.Context())
	if err != nil {
		return err // ditangani ErrorHandler global
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)
	paged, pagination := helper.Paginate(items, page, pageSize)
	return c.JSON(helper.NewPaginatedResponse(paged, pagination))
}

func (h *Handler) GetSatusehat(c *fiber.Ctx) error {
	item, err := h.service.GetSatusehat(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).
			JSON(out.Error(fiber.StatusNotFound, 1, "NOT_FOUND", err.Error()))
	}
	return c.JSON(out.SuccessData(item))
}

func (h *Handler) CreateSatusehat(c *fiber.Ctx) error {
	var input entity.SatusehatInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(out.Error(fiber.StatusBadRequest, 10, "BAD_REQUEST", err.Error()))
	}

	item, err := h.service.CreateSatusehat(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(out.Error(fiber.StatusBadRequest, 11, "VALIDATION", err.Error()))
	}
	return c.JSON(out.SuccessData(item))
}
