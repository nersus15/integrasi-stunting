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

func (h *HttpHandler) CreateAnak(c *fiber.Ctx) error {

	var anak types.AnakPayload
	if err := c.BodyParser(&anak); err != nil {
		return kirimError(c, exceptions.BodyRusak.New(err))
	}
	validasi.AbaikanFieldSistem(&anak)
	anak.SatusehatId = nil

	if err := validasi.ValidateAnak(anak); err != nil {
		return kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	res, err := h.service.CreateAnak(anak.ToEntity(), utils.StrPtr(c.Locals("orgid").(*string)))
	if err != nil {
		return kirimError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}
func (h *HttpHandler) UpdateAnak(c *fiber.Ctx) error {
	var anak types.AnakPayload

	if err := c.BodyParser(&anak); err != nil {
		return kirimError(c, exceptions.BodyRusak.New(err))
	}
	validasi.AbaikanFieldSistem(&anak)

	if err := validasi.ValidateAnakUpdate(anak); err != nil {
		return kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	orgid := c.Locals("orgid").(*string)
	res, err := h.service.UpdateAnakById(anak.ToEntity(), orgid)

	if err != nil {
		return kirimError(c, err)
	}

	return c.Status(http.StatusOK).JSON(res)
}

func (h *HttpHandler) ListAnak(c *fiber.Ctx) error {
	idorangtua := c.Params("id", "")

	if idorangtua == "" {
		return kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter id orangtua harus dikirim")))
	}

	res, err := h.service.ListAnak(&idorangtua)

	if errors.Is(err, exceptions.ErrTidakDitemukan) {
		return kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Orang Tua tidak ditemukan", err))
	}
	if err != nil {
		return kirimError(c, err)
	}

	if err := h.service.AksesOrangtua(orgidPemanggil(c), service.AksiBaca, idorangtua); err != nil {
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

// Saat ini hanya by id
func (h *HttpHandler) FindAnak(c *fiber.Ctx) error {
	id := c.Params("id", "")

	if err := validasi.ValidateCariAnak(id); err != nil {
		return kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	anak, err := h.service.FindAnak(&id, nil, nil, nil)

	if err != nil {
		if errors.Is(err, exceptions.ErrTidakDitemukan) {
			return kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err))
		}
		return kirimError(c, err)
	}

	if err := h.service.AksesAnak(orgidPemanggil(c), service.AksiBaca, anak.Id); err != nil {
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(anak)

}

func (h *HttpHandler) FindAnakByNIK(c *fiber.Ctx) error {
	return nil
}

// satu anak beserta seluruh riwayat kunjungan dan kesehatannya
func (h *HttpHandler) SummaryAnak(c *fiber.Ctx) error {
	id := c.Params("id", "")

	if id == "" {
		return kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter id anak harus dikirim")))
	}

	res, err := h.service.SummaryAnak(id)

	if errors.Is(err, exceptions.ErrTidakDitemukan) {
		return kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err))
	}
	if err != nil {
		logger.Error("SummaryAnak: " + err.Error())
		return kirimError(c, err)
	}

	if err := h.service.AksesAnak(orgidPemanggil(c), service.AksiBaca, id); err != nil {
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *HttpHandler) DeleteAnakByID(c *fiber.Ctx) error {
	id := c.Params("id", "")
	orgid := c.Locals("orgid").(*string)

	if !utils.IsStrFilled(id) {
		return kirimError(c, exceptions.ParameterRequired.WithMessage("Parameter Id anak required", nil))
	}

	if err := h.service.DeleteAnakByID(id, *orgid); err != nil {
		return kirimError(c, exceptions.Classify(err))
	}
	res := map[string]string{
		"message": fmt.Sprintf("Berhasil menghapus anak dengan id #%s", id),
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
