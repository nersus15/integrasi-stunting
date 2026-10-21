package handler

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/helper/validasi"
	"github.com/webcore-go/webcore/infra/logger"
)

// faskes diambil dari API key, id_faskes di body diabaikan
func (h *HttpHandler) SimpanPemeriksaanFaskes(c *fiber.Ctx) error {
	orgid := c.Locals("orgid").(*string)

	p := new(types.PemeriksaanFaskes)
	if err := c.BodyParser(p); err != nil {
		return kirimError(c, exceptions.BodyRusak.New(err))
	}
	validasi.AbaikanFieldSistem(p)

	if err := validasi.ValidatePemeriksaanFaskes(p); err != nil {
		return kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}
	if err := validasi.ValidateZscoreSatuSumber(p); err != nil {
		return kirimError(c, exceptions.BentukPayload.WithMessage(err.Error(), err))
	}

	res, err := h.stream.SimpanPemeriksaanFaskes(p, utils.StrPtr(orgid))
	if err != nil {
		logger.Error("SimpanPemeriksaanFaskes: " + err.Error())
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *HttpHandler) UpdateResourceBySatusehatId(c *fiber.Ctx) error {
	resource := c.Params("resourceName", "")

	orgid := c.Locals("orgid").(*string)

	if resource == "" {
		return kirimError(c, exceptions.TidakDitemukan.Messagef("Endpoint tidak ditemukan"))
	}

	resource = strings.ToLower(resource)
	var res any
	var err error

	switch resource {
	case "anak":
		res, err = h.UpdateAnakFaskes(c, orgid)
	case "orangtua":
		res, err = h.UpdateOrangtuaFaskes(c, orgid)
	case "kunjungan":
		res, err = perbaruiResource(c, resource, utils.StrPtr(orgid), validasi.ValidateKunjunganFaskes, (*types.KunjunganPayload).ToEntity, h.service.UpdateKunjunganBySatusehatId)
	case "observasi":
		res, err = perbaruiResource(c, resource, utils.StrPtr(orgid), validasi.ValidateObservasiFaskes, (*types.Observasi).ToEntity, h.service.UpdateObservasiBySatusehatId)
	case "diagnosa":
		res, err = perbaruiResource(c, resource, utils.StrPtr(orgid), validasi.ValidateDiagnosaFaskes, (*types.Diagnosa).ToEntity, h.service.UpdateDiagnosaBySatusehatId)
	case "layanan":
		res, err = perbaruiResource(c, resource, utils.StrPtr(orgid), validasi.ValidateLayananFaskes, (*types.Layanan).ToEntity, h.service.UpdateLayananBySatusehatId)
	case "rujukan":
		res, err = perbaruiResource(c, resource, utils.StrPtr(orgid), validasi.ValidateRujukanFaskes, (*types.Rujukan).ToEntity, h.service.UpdateRujukanBySatusehatId)
	case "episode":
		res, err = perbaruiResource(c, resource, utils.StrPtr(orgid), validasi.ValidateEpisodeFaskes, (*types.Episode).ToEntity, h.service.UpdateEpisodeBySatusehatId)
	default:
		return kirimError(c, exceptions.TidakDitemukan.Messagef("Resource '%s' tidak ditemukan", resource))
	}

	if err == nil {
		return c.Status(fiber.StatusOK).JSON(res)
	} else {
		return kirimError(c, err)
	}
}

func (h *HttpHandler) UpdateAnakFaskes(c *fiber.Ctx, orgid *string) (any, error) {
	payload := new(types.AnakPayload)
	if err := c.BodyParser(payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}

	validasi.AbaikanFieldSistem(payload)

	anak, err := h.service.TemukanAnak(utils.PtrStr(payload.Id, true), payload.SatusehatId)
	if err != nil {
		return nil, err
	}
	payload.Id = anak.ID

	if err := validasi.ValidateAnakUpdate(*payload); err != nil {
		return nil, exceptions.Validasi.WithMessage(err.Error(), err)
	}

	res, err := h.service.UpdateAnakFaskes(utils.StrPtr(orgid), payload.ToEntity())
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (h *HttpHandler) UpdateOrangtuaFaskes(c *fiber.Ctx, orgid *string) (any, error) {
	payload := new(types.OrangtuaPayload)
	if err := c.BodyParser(payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}

	validasi.AbaikanFieldSistem(payload)

	orangtua, err := h.service.TemukanOrangtua(utils.PtrStr(payload.Id, true), payload.SatusehatId)
	if err != nil {
		return nil, err
	}
	payload.Id = orangtua.ID
	if err := validasi.ValidateOrangtuaUpdate(*payload); err != nil {
		return nil, exceptions.Validasi.WithMessage(err.Error(), err)
	}

	res, err := h.service.UpdateOrangtuaFaskes(utils.StrPtr(orgid), payload.ToEntity())
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (h *HttpHandler) CreateOrangTuaFaskes(c *fiber.Ctx) error {
	var orangtua types.OrangtuaPayload

	if err := c.BodyParser(&orangtua); err != nil {
		return kirimError(c, exceptions.BodyRusak.New(err))
	}

	validasi.AbaikanFieldSistem(&orangtua)
	orgid := c.Locals("orgid").(*string)

	if err := validasi.ValidateOrangtuaFaskes(orangtua); err != nil {
		return kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	res, err := h.service.CreateOrangTua(orangtua.ToEntity(), utils.StrPtr(orgid))
	if err != nil {
		return kirimError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}
func (h *HttpHandler) CreateAnakFaskes(c *fiber.Ctx) error {

	var anak types.AnakPayload
	if err := c.BodyParser(&anak); err != nil {
		return kirimError(c, exceptions.BodyRusak.New(err))
	}
	orgid := c.Locals("orgid").(*string)
	validasi.AbaikanFieldSistem(&anak)

	if err := validasi.ValidateAnakFaskes(anak); err != nil {
		return kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	res, err := h.service.CreateAnak(anak.ToEntity(), utils.StrPtr(orgid))
	if err != nil {
		return kirimError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}

// PUT /faskes/:resource: parse, buang field sistem, validasi, simpan
func perbaruiResource[P any, E any, R any](c *fiber.Ctx, jenis, orgid string, periksa func(P) error, keEntity func(*P) E, simpan func(string, E) (R, error)) (any, error) {
	payload := new(P)
	if err := c.BodyParser(payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}
	validasi.AbaikanFieldSistem(payload)
	if err := periksa(*payload); err != nil {
		return nil, exceptions.Validasi.WithMessage(err.Error(), err)
	}
	res, err := simpan(orgid, keEntity(payload))
	if err != nil {
		logger.Error("http:perbarui " + jenis + ": " + err.Error())
		return nil, err
	}
	return res, nil
}
