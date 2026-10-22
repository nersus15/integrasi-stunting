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

	res, err := h.service.CreateOrangTua(orangtua.ToEntity(), utils.StrPtr(c.Locals("orgid").(*string)))
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

	// by nik terbuka untuk semua, jalan faskes mendapatkan id dan id_satusehat
	if nik == "" {
		if err := h.service.AksesOrangtua(orgidPemanggil(c), service.AksiBaca, orangtua.Id); err != nil {
			return kirimError(c, err)
		}
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
