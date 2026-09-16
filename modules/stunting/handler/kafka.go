package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nersus15/integrasi/mod-stunting/config"
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
	transaction := types2.PackMediator{}

	err := json.Unmarshal(data, &transaction)

	if err != nil {
		logger.Error("Gagal Unmarshall Data. Json Invalid")
		return
	}

	// sinkron supaya urutan pesan dalam satu partisi terjaga; recover karena
	// panic di sini tidak lagi ditangkap worker pool
	defer k.background.Recover("ProsesTransaksiFHIR")

	// susun ulang bundle dengan request payload dan response
	bundle, _, err := k.streamService.SusunUlangBundle(transaction)

	if err != nil {
		logger.Error("Gagal menerjemahkan data transaksi FHIR", "error", err)
		return
	}

	if bundle == nil {
		logger.Error("Bundle transaksi FHIR kosong")
		return
	}

	// proses selanjutnya adalah untuk ekstraksi data
	if err := k.streamService.ProsesTransaksiFHIR(bundle, transaction.Patient); err != nil {
		logger.Error(fmt.Sprintf("Gagal proses data transaksi FHIR %s: ", utils.Nilai(transaction.TransactionID)) + err.Error())
	}
}
