package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/nersus15/integrasi/mod-stunting/config"
	exceptions "github.com/nersus15/integrasi/mod-stunting/helper/Exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/nersus15/integrasi/mod-stunting/service"
	backgroundworker "github.com/nersus15/lib-background-worker"
	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
	service2 "github.com/semanggilab/lib-go-fhir/service"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

type KafkaHandler struct {
	context       *core.AppContext
	service       *service.StuntingService
	streamService *service.StreamService
	fhirService   *service2.FhirTransactionService
	config        *config.ModuleConfig
	background    *backgroundworker.BackgroundWorker
	memory        *port.ICacheMemory
	// kafka         *port.IKafka
}

func NewKafkaHandler(ctx *core.AppContext, service *service.StuntingService, stream *service.StreamService, config *config.ModuleConfig, background *backgroundworker.BackgroundWorker, memory *port.ICacheMemory) *KafkaHandler {
	return &KafkaHandler{
		context:    ctx,
		service:    service,
		config:     config,
		background: background,
		memory:     memory,
		// kafka:         kafka,
		streamService: stream,
	}
}

func (k *KafkaHandler) Run(ctx context.Context, data []byte) {
	key, err := k.proses(data)
	if err == nil {
		return
	}

	logger.Error(fmt.Sprintf("Gagal proses data transaksi FHIR %s: ", key) + err.Error())
	k.backupMessage(key, data, err)
}

func (k *KafkaHandler) proses(data []byte) (key string, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic saat memproses transaksi FHIR",
				"err", r, "stack", string(debug.Stack()))
			err = exceptions.Internal.Messagef("panic saat memproses: %v", r)
		}
	}()

	transaction := types2.PackMediator{}

	if e := json.Unmarshal(data, &transaction); e != nil {
		return sidikPesan(data), exceptions.BodyRusak.WithMessage("Gagal Unmarshall Data. Json Invalid", e)
	}

	key = utils.StrPtr(transaction.TransactionID)
	if key == "" {
		key = sidikPesan(data)
	}

	// susun ulang bundle dengan request payload dan response
	bundle, _, e := k.streamService.SusunUlangBundle(transaction)

	if e != nil {
		return key, exceptions.Internal.WithMessage("Gagal menerjemahkan data transaksi FHIR: "+e.Error(), e)
	}

	if bundle == nil {
		return key, exceptions.Internal.Messagef("Bundle transaksi FHIR kosong")
	}

	// proses selanjutnya adalah untuk ekstraksi data
	return key, k.streamService.ProsesTransaksiFHIR(bundle, transaction.Patient)
}

func sidikPesan(data []byte) string {
	jumlah := sha256.Sum256(data)
	return "sha256-" + hex.EncodeToString(jumlah[:])[:28]
}

func (k *KafkaHandler) backupMessage(key string, data []byte, err error) {
	if e := k.streamService.SaveKafkaTransaction(key, data, err); e != nil {
		logger.Error("Gagal menyimpan pesan kafka yang gagal diproses "+key, "error", e)
	}
}
