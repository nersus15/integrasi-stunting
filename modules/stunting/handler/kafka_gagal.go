package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/webcore-go/webcore/infra/logger"
)

func (h *HttpHandler) DaftarKafkaGagal(c *fiber.Ctx) error {
	patientId := c.Query("patient_id", "")

	res, err := h.stream.DaftarKafkaTransactionGagal(&patientId)

	if err != nil {
		logger.Error("DaftarKafkaGagal: " + err.Error())
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": res, "jumlah": len(res)})
}

func (h *HttpHandler) RetryKafkaGagal(c *fiber.Ctx) error {
	key := c.Params("id", "")

	if key == "" {
		return kirimError(c, exceptions.ParameterRequired.New(fmt.Errorf("Parameter transaction id harus dikirim")))
	}

	if err := h.stream.RetryKafkaTransaction(key); err != nil {
		logger.Error("RetryKafkaGagal " + key + ": " + err.Error())
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"transaction_id": key, "status": "beres"})
}

// diproses berurutan
func (h *HttpHandler) RetrySemuaKafkaGagal(c *fiber.Ctx) error {
	berhasil, gagal, err := h.stream.RetryAllKafkaTransactions()

	if err != nil {
		logger.Error("RetrySemuaKafkaGagal: " + err.Error())
		return kirimError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"berhasil": berhasil, "gagal": gagal})
}
