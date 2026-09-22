package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	"github.com/webcore-go/webcore/port"
	"github.com/webcore-go/webcore/port/auth"
)

type HttpHandler struct {
	service    *service.StuntingService
	stream     *service.StreamService
	config     *config.ModuleConfig
	background *backgroundworker.BackgroundWorker
	memory     port.ICacheMemory
}

func NewHttpHandler(wctx *core.AppContext, cfg *config.ModuleConfig, service *service.StuntingService, stream *service.StreamService, backgroundWorker *backgroundworker.BackgroundWorker, memory port.ICacheMemory) *HttpHandler {
	return &HttpHandler{
		service:    service,
		stream:     stream,
		config:     cfg,
		background: backgroundWorker,
		memory:     memory,
	}
}

// faskes diambil dari API key, id_faskes di body diabaikan
func (h *HttpHandler) SimpanPemeriksaanFaskes(c *fiber.Ctx) error {
	orgid, err := h.GetOrgid(c)
	if err != nil {
		return h.kirimError(c, err)
	}
	if orgid == nil || *orgid == "jakantro" {
		return h.kirimError(c, exceptions.Forbidden.WithMessage(
			"endpoint ini hanya untuk faskes, bukan jakantro", nil))
	}

	p := new(types.PemeriksaanFaskes)
	if err := c.BodyParser(p); err != nil {
		return h.kirimError(c, exceptions.BodyRusak.New(err))
	}

	if err := utils.ValidatePemeriksaanFaskes(p); err != nil {
		return h.kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	res, err := h.stream.SimpanPemeriksaanFaskes(p, *orgid)
	if err != nil {
		logger.Error("SimpanPemeriksaanFaskes: " + err.Error())
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(res)
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

func (h *HttpHandler) UpdateOrantua(c *fiber.Ctx) error {
	var orangtua types.OrangtuaPayload

	if err := c.BodyParser(&orangtua); err != nil {
		return h.kirimError(c, exceptions.BodyRusak.New(err))
	}

	if err := utils.ValidateOrangtua(orangtua); err != nil {
		return h.kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	orgid, err := h.GetOrgid(c)
	if err != nil {
		return h.kirimError(c, err)
	}

	err = h.service.VerifyAccess(&orangtua.IdPosyandu, orgid)

	if err != nil {
		logger.ErrorJson("VerifyAccess", err)
		return h.kirimError(c, err)
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

	orgid, err := h.GetOrgid(c)
	if err != nil {
		return h.kirimError(c, err)
	}

	err = h.service.VerifyAccess(orangtua.IdPosyandu, orgid)

	if err != nil {
		logger.ErrorJson("VerifyAccess", err)
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
func (h *HttpHandler) UpdateAnak(c *fiber.Ctx) error {
	var anak types.AnakPayload

	if err := c.BodyParser(&anak); err != nil {
		return h.kirimError(c, exceptions.BodyRusak.New(err))
	}

	if err := utils.ValidateAnak(anak); err != nil {
		return h.kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	orgid, err := h.GetOrgid(c)
	if err != nil {
		return h.kirimError(c, err)
	}

	err = h.service.VerifyAccessByIdOrangtua(anak.IDOrangtua, orgid)

	if err != nil {
		logger.ErrorJson("VerifyAccess", err)
		return h.kirimError(c, err)
	}

	res, err := h.service.UpdateAnakById(anak.ToEntity(), orgid)

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

	if err != nil {
		if errors.Is(err, utils.ErrTidakDitemukan) {
			return h.kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err))
		}
		return h.kirimError(c, err)
	}

	orgid, err := h.GetOrgid(c)
	if err != nil {
		return h.kirimError(c, err)
	}

	err = h.service.VerifyAccessByIdOrangtua(anak.IdOrangtua, orgid)

	if err != nil {
		logger.ErrorJson("VerifyAccess", err)
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

func (h *HttpHandler) CreateKesehatan(c *fiber.Ctx) error {
	res, err := h.service.CreateKesehatan(c.Body())
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
	id := c.Params("id", "")

	if id == "" {
		return h.kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter id kunjungan harus dikirim")))
	}

	res, err := h.service.FindKunjunganById(id)

	if errors.Is(err, utils.ErrTidakDitemukan) {
		return h.kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Kunjungan tidak ditemukan", err))
	}
	if err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *HttpHandler) FindKesehatanByIdAnak(c *fiber.Ctx) error {
	id := c.Params("id", "")

	if id == "" {
		return h.kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter id anak harus dikirim")))
	}

	res, err := h.service.KesehatanByAnak(id)

	if errors.Is(err, utils.ErrTidakDitemukan) {
		return h.kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err))
	}
	if err != nil {
		logger.Error("FindKesehatanByIdAnak: " + err.Error())
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *HttpHandler) FindKesehatanById(c *fiber.Ctx) error {
	id := c.Params("id", "")

	if id == "" {
		return h.kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter id kesehatan harus dikirim")))
	}

	res, err := h.service.FindKesehatanById(id)

	if errors.Is(err, utils.ErrTidakDitemukan) {
		return h.kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Kesehatan tidak ditemukan", err))
	}
	if err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

// SummaryAnak mengembalikan satu anak beserta seluruh riwayat kunjungan dan
// kesehatannya, supaya pemanggil tidak perlu dua request.
func (h *HttpHandler) SummaryAnak(c *fiber.Ctx) error {
	id := c.Params("id", "")

	if id == "" {
		return h.kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter id anak harus dikirim")))
	}

	res, err := h.service.SummaryAnak(id)

	if errors.Is(err, utils.ErrTidakDitemukan) {
		return h.kirimError(c, exceptions.TidakDitemukan.WithMessage("Data Anak tidak ditemukan", err))
	}
	if err != nil {
		logger.Error("SummaryAnak: " + err.Error())
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *HttpHandler) GetOrgid(c *fiber.Ctx) (*string, error) {
	// Cek permission
	var orgid string
	memorgid := new("")
	apikey := auth.GetAPIKey(c)
	memkey := "USERORGID::" + apikey

	if ok := h.memory.Get(memkey, memorgid); !ok {
		groups := auth.GetUserGroups(c)
		if groups == nil {
			return nil, exceptions.Forbidden.WithMessage("Gagal verifikasi hak akses user: user group tidak ditemukan", nil)
		}
		ada := false

		for _, g := range groups {
			if g == "jakantro" {
				ada = true
				orgid = g
				break
			} else if strings.Contains(g, "orgid:") {
				ada = true
				orgid = strings.Replace(g, "orgid:", "", 1)
				break
			}
		}

		if !ada {
			return nil, exceptions.Forbidden.WithMessage("Gagal verifikasi hak akses user", nil)
		}

		err := h.memory.Set(memkey, orgid, 24*time.Minute)

		if err != nil {
			logger.Error("CacheOrgid", err)
		}
	} else {
		logger.Info("Orgid dari cache: " + *memorgid)
		orgid = *memorgid
	}

	if !utils.IsStrFilled(orgid) {
		return nil, exceptions.Forbidden.WithMessage("Gagal verifikasi hak akses user: user group tidak ditemukan", nil)
	}

	return &orgid, nil
}

// DaftarKafkaGagal menampilkan pesan kafka yang belum berhasil diproses.
func (h *HttpHandler) DaftarKafkaGagal(c *fiber.Ctx) error {
	patientId := c.Query("patient_id", "")

	res, err := h.stream.DaftarKafkaTransactionGagal(&patientId)

	if err != nil {
		logger.Error("DaftarKafkaGagal: " + err.Error())
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": res, "jumlah": len(res)})
}

func (h *HttpHandler) RetryKafkaGagal(c *fiber.Ctx) error {
	key := c.Params("id", "")

	if key == "" {
		return h.kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter transaction id harus dikirim")))
	}

	if err := h.stream.RetryKafkaTransaction(key); err != nil {
		logger.Error("RetryKafkaGagal " + key + ": " + err.Error())
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"transaction_id": key, "status": "beres"})
}

// RetrySemuaKafkaGagal memproses ulang seluruh antrean, berurutan.
func (h *HttpHandler) RetrySemuaKafkaGagal(c *fiber.Ctx) error {
	berhasil, gagal, err := h.stream.RetryAllKafkaTransactions()

	if err != nil {
		logger.Error("RetrySemuaKafkaGagal: " + err.Error())
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"berhasil": berhasil, "gagal": gagal})
}
