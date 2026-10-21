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

func (h *HttpHandler) CreateKunjungan(c *fiber.Ctx) error {
	res, err := h.service.CreateKunjungan(c.Body())
	if err != nil {
		return kirimError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}

// khusus jakantro: id kunjungan = id pengukuran internal jakantro
func (h *HttpHandler) UpdateKunjunganById(c *fiber.Ctx) error {

	id := c.Params("id", "")
	if id == "" {
		return kirimError(c, exceptions.ParameterRequired.Messagef("Id harus dikirimkan sebagai parameter"))
	}

	payload := new(types.KunjunganPayload)

	if err := c.BodyParser(payload); err != nil {
		return kirimError(c, exceptions.BodyRusak.New(err))
	}
	validasi.AbaikanFieldSistem(payload)

	if err := validasi.ValidateKunjungan(*payload); err != nil {
		return kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	res, err := h.service.UpdateKunjunganById(payload, id, utils.StrPtr(c.Locals("orgid").(*string)))

	if err != nil {
		return kirimError(c, exceptions.Classify(err))
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
func (h *HttpHandler) FindKunjunganByIdAnak(c *fiber.Ctx) error {
	id := c.Params("id", "")

	if id == "" {
		return kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter id anak harus dikirim")))
	}

	res, err := h.service.KunjunganByAnak(id)

	if errors.Is(err, exceptions.ErrTidakDitemukan) {
		return kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err))
	}
	if err != nil {
		logger.Error("FindKunjunganByIdAnak: " + err.Error())
		return kirimError(c, err)
	}

	if err := h.service.AksesAnak(orgidPemanggil(c), service.AksiBaca, id); err != nil {
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)

}

func (h *HttpHandler) FindKunjunganById(c *fiber.Ctx) error {
	id := c.Params("id", "")

	if id == "" {
		return kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter id kunjungan harus dikirim")))
	}

	res, err := h.service.FindKunjunganById(id)

	if errors.Is(err, exceptions.ErrTidakDitemukan) {
		return kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Kunjungan tidak ditemukan", err))
	}
	if err != nil {
		return kirimError(c, err)
	}

	if err := h.service.AksesAnak(orgidPemanggil(c), service.AksiBaca, res.IdAnak); err != nil {
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *HttpHandler) DeleteKunjunganByID(c *fiber.Ctx) error {
	id := c.Params("id", "")
	orgid := c.Locals("orgid").(*string)

	if !utils.IsStrFilled(id) {
		return kirimError(c, exceptions.ParameterRequired.WithMessage("Parameter Id kunjungan required", nil))
	}

	if err := h.service.DeleteKunjunganById(id, *orgid); err != nil {
		return kirimError(c, exceptions.Classify(err))
	}
	res := map[string]string{
		"message": fmt.Sprintf("Berhasil menghapus data kunjungan dengan id #%s", id),
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
