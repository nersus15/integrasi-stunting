package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/helper/validasi"
	"github.com/nersus15/integrasi/mod-stunting/service"
	"github.com/webcore-go/webcore/infra/logger"
)

func (h *HttpHandler) CreateKesehatan(c *fiber.Ctx) error {
	res, err := h.service.CreateKesehatan(c.Body())
	if err != nil {
		return kirimError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}
func (h *HttpHandler) UpdateKesehatanById(c *fiber.Ctx) error {
	id := c.Params("id", "")
	if id == "" {
		return kirimError(c, exceptions.ParameterRequired.Messagef("Id harus dikirimkan sebagai parameter"))
	}

	payload := new(types.KesehatanPayload)

	if err := c.BodyParser(payload); err != nil {
		return kirimError(c, exceptions.BodyRusak.New(err))
	}
	validasi.AbaikanFieldSistem(payload)

	if err := validasi.ValidateKesehatan(*payload); err != nil {
		return kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	res, err := h.service.UpdateKesehatanById(payload, id, utils.StrPtr(c.Locals("orgid").(*string)))

	if err != nil {
		return kirimError(c, exceptions.Classify(err))
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *HttpHandler) FindKesehatanByIdAnak(c *fiber.Ctx) error {
	id := c.Params("id", "")

	if id == "" {
		return kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter id anak harus dikirim")))
	}

	res, err := h.service.KesehatanByAnak(id)

	if errors.Is(err, exceptions.ErrTidakDitemukan) {
		return kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err))
	}
	if err != nil {
		logger.Error("FindKesehatanByIdAnak: " + err.Error())
		return kirimError(c, err)
	}

	if err := h.service.AksesAnak(orgidPemanggil(c), service.AksiBaca, id); err != nil {
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *HttpHandler) FindKesehatanById(c *fiber.Ctx) error {
	id := c.Params("id", "")

	if id == "" {
		return kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter id kesehatan harus dikirim")))
	}

	res, err := h.service.FindKesehatanById(id)

	if errors.Is(err, exceptions.ErrTidakDitemukan) {
		return kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Kesehatan tidak ditemukan", err))
	}
	if err != nil {
		return kirimError(c, err)
	}

	if err := h.service.AksesAnak(orgidPemanggil(c), service.AksiBaca, res.IdAnak); err != nil {
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *HttpHandler) DeleteKesehatanByID(c *fiber.Ctx) error {
	id := c.Params("id", "")
	orgid := c.Locals("orgid").(*string)

	if !utils.IsStrFilled(id) {
		return kirimError(c, exceptions.ParameterRequired.WithMessage("Parameter Id kesehatan required", nil))
	}

	if err := h.service.DeleteKesehatanById(id, *orgid); err != nil {
		return kirimError(c, exceptions.Classify(err))
	}
	res := map[string]string{
		"message": fmt.Sprintf("Berhasil menghapus data kesehatan dengan id #%s", id),
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
