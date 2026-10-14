package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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

// Region alternatif jalur ekstract payload satusehat untuk faskes
// faskes diambil dari API key, id_faskes di body diabaikan
func (h *HttpHandler) SimpanPemeriksaanFaskes(c *fiber.Ctx) error {
	orgid := c.Locals("orgid").(*string)

	p := new(types.PemeriksaanFaskes)
	if err := c.BodyParser(p); err != nil {
		return h.kirimError(c, exceptions.BodyRusak.New(err))
	}
	utils.AbaikanFieldSistem(p)

	if err := utils.ValidatePemeriksaanFaskes(p); err != nil {
		return h.kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}
	if err := utils.ValidateZscoreSatuSumber(p); err != nil {
		return h.kirimError(c, exceptions.BentukPayload.WithMessage(err.Error(), err))
	}

	res, err := h.stream.SimpanPemeriksaanFaskes(p, utils.StrPtr(orgid))
	if err != nil {
		logger.Error("SimpanPemeriksaanFaskes: " + err.Error())
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *HttpHandler) UpdateResourceBySatusehatId(c *fiber.Ctx) error {
	resource := c.Params("resourceName", "")

	orgid := c.Locals("orgid").(*string)

	if resource == "" {
		return h.kirimError(c, exceptions.TidakDitemukan.Messagef("Endpoint tidak ditemukan"))
	}

	resource = strings.ToLower(resource)
	var res any
	var err error

	switch resource {
	case "kunjungan":
		res, err = h.updateKunjunganBySatusehatId(c, orgid)
	case "observasi":
		res, err = h.updateObservasiBySatusehatId(c, orgid)
	case "diagnosa":
		res, err = h.updateDiagnosaBySatusehatId(c, orgid)
	case "layanan":
		res, err = h.updateLayananBySatusehatId(c, orgid)
	case "rujukan":
		res, err = h.updateRujukanBySatusehatId(c, orgid)
	case "episode":
		res, err = h.updateEpisodeBySatusehatId(c, orgid)
	default:
		return h.kirimError(c, exceptions.TidakDitemukan.Messagef("Resource '%s' tidak ditemukan", resource))
	}

	if err == nil {
		return c.Status(fiber.StatusOK).JSON(res)
	} else {
		return h.kirimError(c, err)
	}
}

func (h *HttpHandler) updateKunjunganBySatusehatId(c *fiber.Ctx, orgid *string) (any, error) {
	payload := new(types.KunjunganPayload)
	if err := c.BodyParser(payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}
	utils.AbaikanFieldSistem(payload)
	if err := utils.ValidateKunjunganFaskes(*payload); err != nil {
		return nil, exceptions.Validasi.WithMessage(err.Error(), err)
	}

	res, err := h.service.UpdateKunjunganBySatusehatId(utils.StrPtr(orgid), payload.ToEntity())
	if err != nil {
		logger.Error("http:updateKunjunganBySatusehatId: " + err.Error())
		return nil, err
	}

	return res, nil
}

func (h *HttpHandler) updateObservasiBySatusehatId(c *fiber.Ctx, orgid *string) (any, error) {
	payload := new(types.Observasi)
	if err := c.BodyParser(payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}
	utils.AbaikanFieldSistem(payload)

	if err := utils.ValidateObservasiFaskes(*payload); err != nil {
		return nil, exceptions.Validasi.WithMessage(err.Error(), err)
	}

	res, err := h.service.UpdateObservasiBySatusehatId(utils.StrPtr(orgid), payload.ToEntity())
	if err != nil {
		logger.Error("http:updateObservasiBySatusehatId: " + err.Error())
		return nil, err
	}

	return res, nil
}

func (h *HttpHandler) updateDiagnosaBySatusehatId(c *fiber.Ctx, orgid *string) (any, error) {
	payload := new(types.Diagnosa)
	if err := c.BodyParser(payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}
	utils.AbaikanFieldSistem(payload)

	if err := utils.ValidateDiagnosaFaskes(*payload); err != nil {
		return nil, exceptions.Validasi.WithMessage(err.Error(), err)
	}

	res, err := h.service.UpdateDiagnosaBySatusehatId(utils.StrPtr(orgid), payload.ToEntity())
	if err != nil {
		logger.Error("http:updateDiagnosaBySatusehatId: " + err.Error())
		return nil, err
	}

	return res, nil
}

func (h *HttpHandler) updateLayananBySatusehatId(c *fiber.Ctx, orgid *string) (any, error) {

	payload := new(types.Layanan)
	if err := c.BodyParser(payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}
	utils.AbaikanFieldSistem(payload)

	if err := utils.ValidateLayananFaskes(*payload); err != nil {
		return nil, exceptions.Validasi.WithMessage(err.Error(), err)
	}

	res, err := h.service.UpdateLayananBySatusehatId(utils.StrPtr(orgid), payload.ToEntity())
	if err != nil {
		logger.Error("http:pdateLayananBySatusehatId: " + err.Error())
		return nil, err
	}

	return res, nil
}

func (h *HttpHandler) updateRujukanBySatusehatId(c *fiber.Ctx, orgid *string) (any, error) {
	payload := new(types.Rujukan)

	if err := c.BodyParser(payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}
	utils.AbaikanFieldSistem(payload)

	if err := utils.ValidateRujukanFaskes(*payload); err != nil {
		return nil, exceptions.Validasi.WithMessage(err.Error(), err)
	}

	res, err := h.service.UpdateRujukanBySatusehatId(utils.StrPtr(orgid), payload.ToEntity())
	if err != nil {
		logger.Error("http:updateRujukanBySatusehatId: " + err.Error())
		return nil, err
	}

	return res, nil
}
func (h *HttpHandler) updateEpisodeBySatusehatId(c *fiber.Ctx, orgid *string) (any, error) {
	payload := new(types.Episode)

	if err := c.BodyParser(payload); err != nil {
		return nil, exceptions.BodyRusak.New(err)
	}
	utils.AbaikanFieldSistem(payload)

	if err := utils.ValidateEpisodeFaskes(*payload); err != nil {
		return nil, exceptions.Validasi.WithMessage(err.Error(), err)
	}

	res, err := h.service.UpdateEpisodeBySatusehatId(utils.StrPtr(orgid), payload.ToEntity())
	if err != nil {
		logger.Error("http:updateEpisodeBySatusehatId: " + err.Error())
		return nil, err
	}

	return res, nil
}

// End Region

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
	utils.AbaikanFieldSistem(&orangtua)
	orangtua.SatusehatId = nil

	if err := utils.ValidateOrangtua(orangtua); err != nil {
		return h.kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	res, err := h.service.CreateOrangTua(orangtua.ToEntity())
	if err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}

func (h *HttpHandler) CreateOrangTuaFaskes(c *fiber.Ctx) error {
	var orangtua types.OrangtuaPayload

	if err := c.BodyParser(&orangtua); err != nil {
		return h.kirimError(c, exceptions.BodyRusak.New(err))
	}
	utils.AbaikanFieldSistem(&orangtua)
	orangtua.SatusehatId = nil

	if err := utils.ValidateOrangtuaFaskes(orangtua); err != nil {
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
	utils.AbaikanFieldSistem(&orangtua)

	if err := utils.ValidateOrangtuaUpdate(orangtua); err != nil {
		return h.kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	orgid := c.Locals("orgid").(*string)

	if err := h.service.VerifyAksesOrangtua(orgid, orangtua.Id); err != nil {
		return h.kirimError(c, err)
	}

	res, err := h.service.UpdateOrangTuaById(orangtua.ToEntity(), orgid)
	if err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
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

	orgid := c.Locals("orgid").(*string)

	if err := h.service.VerifyAksesOrangtua(orgid, orangtua.Id); err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(orangtua)
}

func (h *HttpHandler) CreateAnak(c *fiber.Ctx) error {

	var anak types.AnakPayload
	if err := c.BodyParser(&anak); err != nil {
		return h.kirimError(c, exceptions.BodyRusak.New(err))
	}
	utils.AbaikanFieldSistem(&anak)
	anak.SatusehatId = nil

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
	utils.AbaikanFieldSistem(&anak)

	if err := utils.ValidateAnakUpdate(anak); err != nil {
		return h.kirimError(c, exceptions.Validasi.WithMessage(err.Error(), err))
	}

	orgid := c.Locals("orgid").(*string)

	if err := h.service.VerifyAksesAnak(orgid, anak.Id); err != nil {
		return h.kirimError(c, err)
	}

	res, err := h.service.UpdateAnakById(anak.ToEntity(), orgid)

	if err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(http.StatusOK).JSON(res)
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

	orgid := c.Locals("orgid").(*string)

	if err := h.service.VerifyAksesAnak(orgid, anak.Id); err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(anak)

}

func (h *HttpHandler) FindAnakByNIK(c *fiber.Ctx) error {
	return nil
}

func (h *HttpHandler) CreateKunjungan(c *fiber.Ctx) error {
	res, err := h.service.CreateKunjungan(c.Body())
	if err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}

// UpdateKunjunganById ini adalah handler yang dikhususkan untuk jakantro, karena id kunjungan dari jakantro = id pengukurna internal jakantro
func (h *HttpHandler) UpdateKunjunganById(c *fiber.Ctx) error {

	id := c.Params("id", "")
	if id == "" {
		return h.kirimError(c, exceptions.ParameterRequired.Messagef("Id harus dikirimkan sebagai parameter"))
	}

	payload := new(types.KunjunganPayload)

	if err := c.BodyParser(payload); err != nil {
		return h.kirimError(c, exceptions.BodyRusak.New(err))
	}
	utils.AbaikanFieldSistem(payload)

	if err := utils.ValidateKunjungan(*payload); err != nil {
		return h.kirimError(c, exceptions.Validasi.New(err))
	}

	res, err := h.service.UpdateKunjunganById(payload, id)

	if err != nil {
		return h.kirimError(c, exceptions.Classify(err))
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
func (h *HttpHandler) CreateKesehatan(c *fiber.Ctx) error {
	res, err := h.service.CreateKesehatan(c.Body())
	if err != nil {
		return h.kirimError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(res)
}
func (h *HttpHandler) UpdateKesehatanById(c *fiber.Ctx) error {
	id := c.Params("id", "")
	if id == "" {
		return h.kirimError(c, exceptions.ParameterRequired.Messagef("Id harus dikirimkan sebagai parameter"))
	}

	payload := new(types.KesehatanPayload)

	if err := c.BodyParser(payload); err != nil {
		return h.kirimError(c, exceptions.BodyRusak.New(err))
	}
	utils.AbaikanFieldSistem(payload)

	if err := utils.ValidateKesehatan(*payload); err != nil {
		return h.kirimError(c, exceptions.Validasi.New(err))
	}

	res, err := h.service.UpdateKesehatanById(payload, id)

	if err != nil {
		return h.kirimError(c, exceptions.Classify(err))
	}

	return c.Status(fiber.StatusOK).JSON(res)
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

// Region Delete
func (h *HttpHandler) DeleteAnakByID(c *fiber.Ctx) error {
	id := c.Params("id", "")
	uid := auth.GetUserID(c)

	if !utils.IsStrFilled(id) {
		return h.kirimError(c, exceptions.ParameterRequired.WithMessage("Parameter ID anak required", nil))
	}

	if err := h.service.DeleteAnakByID(id, uid); err != nil {
		return h.kirimError(c, exceptions.Classify(err))
	}
	res := map[string]string{
		"message": fmt.Sprintf("Berhasil menghapus anak dengan id #%s", id),
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *HttpHandler) DeleteOrangtuaByID(c *fiber.Ctx) error {
	return nil
}

func (h *HttpHandler) DeleteKunjunganByID(c *fiber.Ctx) error {
	return nil
}

func (h *HttpHandler) DeleteKesehatanBYID(c *fiber.Ctx) error {
	return nil
}

// End Region

// Region Delete Untuk Faskes (Alternatif jalur satusehat, kemungkinan untuk delete tidak bisa melalui jalur kafka stream)

// End Region

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
