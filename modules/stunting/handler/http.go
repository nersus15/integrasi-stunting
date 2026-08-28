package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/integrasi/mod-stunting/config"
	exceptions "github.com/nersus15/integrasi/mod-stunting/helper/Exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/service"
	backgroundworker "github.com/nersus15/lib-background-worker"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/out"
	"github.com/webcore-go/webcore/infra/logger"
)

type HttpHandler struct {
	service    *service.StuntingService
	config     *config.ModuleConfig
	background *backgroundworker.BackgroundWorker
}

func NewHandler(wctx *core.AppContext, cfg *config.ModuleConfig, service *service.StuntingService, backgroundWorker *backgroundworker.BackgroundWorker) *HttpHandler {
	return &HttpHandler{
		service:    service,
		config:     cfg,
		background: backgroundWorker,
	}
}

func (h *HttpHandler) kirimError(c *fiber.Ctx, err error) error {
	k := exceptions.Classify(err)

	return c.Status(k.HttpCode).JSON(out.ErrorDetail(
		k.HttpCode,
		k.ErrorCode,
		k.Name,
		k.Message,
		k,
	))
}

func (h *HttpHandler) CreateOrangTua(c *fiber.Ctx) error {
	var orangtua types.OrangtuaPayload

	if err := c.BodyParser(&orangtua); err != nil {
		return h.kirimError(c, exceptions.BodyRusak.New(err))
	}

	if err := utils.ValidateOrangtua(orangtua); err != nil {
		return h.kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	res, err := h.service.CreateOrangTua(orangtua.ToEntity())
	if err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}

func (h *HttpHandler) FindOrangTua(c *fiber.Ctx) error {
	id := c.Params("id", "")
	nik := c.Query("nik", "")
	nokk := c.Query("nokk", "")
	nama_ayah := c.Query("nama_ayah", "")
	nama_ibu := c.Query("nama_ibu", "")

	if err := utils.ValidateCariOrangtua(id, nik, nokk, nama_ayah, nama_ibu); err != nil {
		return h.kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	orangtua, err := h.service.FindOrangTua(&id, &nik, &nokk, &nama_ayah, &nama_ibu)
	if errors.Is(err, utils.ErrTidakDitemukan) {
		return h.kirimError(c, exceptions.TidakDitemukan.WithMessage("Data orang tua tidak ditemukan", err))
	}
	if err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(orangtua)
}

func (h *HttpHandler) CreateAnak(c *fiber.Ctx) error {
	var anak types.AnakPayload

	if err := c.BodyParser(&anak); err != nil {
		return h.kirimError(c, exceptions.BodyRusak.New(err))
	}

	if err := utils.ValidateAnak(anak); err != nil {
		return h.kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	res, err := h.service.CreateAnak(anak.ToEntity())
	if err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}

func (h *HttpHandler) ListAnak(c *fiber.Ctx) error {
	idorangtua := c.Params("id", "")

	if idorangtua == "" {
		return h.kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter id orangtua harus dikirim")))
	}

	res, err := h.service.ListAnak(&idorangtua)

	if errors.Is(err, utils.ErrTidakDitemukan) {
		return h.kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Orang Tua tidak ditemukan", err))
	}
	if err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

// Saat ini hanya by id
func (h *HttpHandler) FindAnak(c *fiber.Ctx) error {
	id := c.Params("id", "")

	if err := utils.ValidateCariAnak(id); err != nil {
		return h.kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	anak, err := h.service.FindAnak(&id, nil, nil, nil)

	if errors.Is(err, utils.ErrTidakDitemukan) {
		return h.kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err))
	}
	if err != nil {
		return h.kirimError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(anak)

}
func (h *HttpHandler) CreateKunjungan(c *fiber.Ctx) error {
	res, err := h.service.CreateKunjungan(c.Body())
	if err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}

func (h *HttpHandler) FindKunjunganByIdAnak(c *fiber.Ctx) error {
	id := c.Params("id", "")

	if id == "" {
		return h.kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter id anak harus dikirim")))
	}

	res, err := h.service.KunjunganByAnak(id)

	if errors.Is(err, utils.ErrTidakDitemukan) {
		return h.kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err))
	}
	if err != nil {
		logger.Error("FindKunjunganByIdAnak: " + err.Error())
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)

}

func (h *HttpHandler) FindKunjunganById(c *fiber.Ctx) error {
	return nil
}
