package service

import (
	"encoding/json"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/types"
	types2 "github.com/semanggilab/lib-go-fhir/helper/types"
	"github.com/webcore-go/webcore/infra/logger"
)

func (s *StreamService) RetryKafkaTransaction(key string) error {
	tr, err := s.repository.FindKafkaTransaction(key)
	if err != nil {
		return err
	}

	err = s.prosesUlang((*types.KafkaTransaction)(nil).FromEntity(tr))
	if err != nil {
		if e := s.SaveKafkaTransaction(key, []byte(tr.Message), err); e != nil {
			logger.Error("RetryKafkaTransaction: gagal memperbarui catatan "+key, "error", e)
		}
		return err
	}

	return s.repository.TandaiKafkaTransactionSelesai(key)
}

func (s *StreamService) RetryAllKafkaTransactions() (berhasil int, gagal int, err error) {
	daftar, err := s.repository.FindKafkaTransactions(nil)
	if err != nil {
		return 0, 0, err
	}

	for _, tr := range daftar {
		if tr == nil {
			continue
		}
		if e := s.RetryKafkaTransaction(tr.TransactionId); e != nil {
			logger.Error("RetryAllKafkaTransactions: "+tr.TransactionId, "error", e)
			gagal++
			continue
		}
		berhasil++
	}

	return berhasil, gagal, nil
}

func (s *StreamService) prosesUlang(tr *types.KafkaTransaction) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic saat memproses ulang transaksi FHIR",
				"err", r, "stack", string(debug.Stack()))
			err = exceptions.Internal.Messagef("panic saat memproses ulang: %v", r)
		}
	}()

	if tr == nil || len(tr.Message) == 0 {
		return exceptions.TidakDitemukan.Messagef("pesan kafka kosong, tidak ada yang bisa diproses ulang")
	}

	transaction := types2.PackMediator{}
	if e := json.Unmarshal([]byte(tr.Message), &transaction); e != nil {
		return exceptions.BodyRusak.WithMessage("Gagal Unmarshall Data. Json Invalid", e)
	}

	bundle, _, e := s.SusunUlangBundle(transaction)
	if e != nil {
		return exceptions.Internal.WithMessage("Gagal menerjemahkan data transaksi FHIR: "+e.Error(), e)
	}
	if bundle == nil {
		return exceptions.Internal.Messagef("Bundle transaksi FHIR kosong")
	}

	return s.ProsesTransaksiFHIR(bundle, transaction.Patient)
}

func (s *StreamService) DaftarKafkaTransactionGagal(patientId *string) ([]*types.KafkaTransaction, error) {
	daftar, err := s.repository.FindKafkaTransactions(patientId)
	if err != nil {
		return nil, err
	}
	res := make([]*types.KafkaTransaction, 0, len(daftar))
	for _, t := range daftar {
		res = append(res, (*types.KafkaTransaction)(nil).FromEntity(t))
	}
	return res, nil
}

func (s *StreamService) SaveKafkaTransaction(key string, message []byte, err error) error {
	pesan := "tanpa keterangan"
	if err != nil {
		pesan = err.Error()
	}

	eBytes, _ := json.Marshal(pesan)
	e := json.RawMessage(eBytes)

	patientId := ekstrakPatientId(message)

	data := types.KafkaTransactionPayload{
		Id:            uuid.NewString(),
		TransactionId: key,
		GroupId:       s.Config.Kafka.GroupID,
		PatientId:     patientId,
		Message:       string(message),
		Error:         &e,
		Attempt:       1,
		Created:       time.Now(),
	}
	return s.repository.SaveKafkaTransaction(data.ToEntity())
}
