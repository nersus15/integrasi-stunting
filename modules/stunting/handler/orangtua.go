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
)

func (h *HttpHandler) CreateOrangTua(c *fiber.Ctx) error {

	var orangtua types.OrangtuaPayload
	if err := c.BodyParser(&orangtua); err != nil {
		return kirimError(c, exceptions.BodyRusak.New(err))
	}
	validasi.AbaikanFieldSistem(&orangtua)
	orangtua.SatusehatId = nil

	if err := validasi.ValidateOrangtua(orangtua); err != nil {
		return kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	res, err := h.service.CreateOrangTua(orangtua.ToEntity())
	if err != nil {
		return kirimError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}

func (h *HttpHandler) UpdateOrangtua(c *fiber.Ctx) error {
	var orangtua types.OrangtuaPayload
	if err := c.BodyParser(&orangtua); err != nil {
		return kirimError(c, exceptions.BodyRusak.New(err))
	}
	validasi.AbaikanFieldSistem(&orangtua)

	if err := validasi.ValidateOrangtuaUpdate(orangtua); err != nil {
		return kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	orgid := c.Locals("orgid").(*string)

	if _, err := h.service.VerifyAksesOrangtua(orgid, nil, utils.PtrStr(orangtua.Id, true), orangtua.SatusehatId); err != nil {
		return kirimError(c, err)
	}

	res, err := h.service.UpdateOrangTuaById(orangtua.ToEntity(), orgid)
	if err != nil {
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *HttpHandler) FindOrangTua(c *fiber.Ctx) error {
	id := c.Params("id", "")
	nik := c.Query("nik", "")
	nokk := c.Query("nokk", "")
	nama_ayah := c.Query("nama_ayah", "")
	nama_ibu := c.Query("nama_ibu", "")

	if err := validasi.ValidateCariOrangtua(id, nik, nokk, nama_ayah, nama_ibu); err != nil {
		return kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	orangtua, err := h.service.FindOrangTua(&id, &nik, &nokk, &nama_ayah, &nama_ibu)
	if errors.Is(err, exceptions.ErrTidakDitemukan) {
		return kirimError(c, exceptions.TidakDitemukan.WithMessage("Data orang tua tidak ditemukan", err))
	}
	if err != nil {
		return kirimError(c, err)
	}

	orgid := c.Locals("orgid").(*string)

	if _, err := h.service.VerifyAksesOrangtua(orgid, orangtua, utils.PtrStr(orangtua.Id, true), orangtua.IdSatusehat); err != nil {
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(orangtua)
}

func (h *HttpHandler) DeleteOrangtuaByID(c *fiber.Ctx) error {
	id := c.Params("id", "")
	orgid := c.Locals("orgid").(*string)

	if !utils.IsStrFilled(id) {
		return kirimError(c, exceptions.ParameterRequired.Messagef("Parameeter Id orang tua harus dikirim"))
	}

	if err := h.service.DeleteOrangtuaByID(id, *orgid); err != nil {
		return kirimError(c, exceptions.Classify(err))
	}
	res := map[string]string{
		"message": fmt.Sprintf("Berhasil menghapus orang tua dengan id #%s", id),
	}
	return c.Status(fiber.StatusOK).JSON(res)
}
