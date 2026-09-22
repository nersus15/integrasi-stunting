package types

import (
	"encoding/json"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
)

type KafkaTransaction struct {
	Id            string           `json:"id"`
	TransactionId string           `json:"transaction_id"`
	GroupId       string           `json:"group_id"`
	PatientId     *string          `json:"patient_id"`
	Message       string           `json:"message"`
	Error         *json.RawMessage `json:"error"`
	Attempt       int16            `json:"attempt"`
	Resolved      *string          `json:"resolved_at"`
	Created       string           `json:"created_at"`
}

type KafkaTransactionPayload struct {
	Id            string           `json:"id"`
	TransactionId string           `json:"transaction_id"`
	GroupId       string           `json:"group_id"`
	PatientId     *string          `json:"patient_id"`
	Message       string           `json:"message"`
	Error         *json.RawMessage `json:"error"`
	Attempt       int16            `json:"attempt"`
	Resolved      *time.Time       `json:"resolved_at"`
	Created       time.Time        `json:"created_at"`
}

// waktu disimpan sampai detik: urutan kejadian yang dibutuhkan saat menelusuri
// pesan gagal, bukan tanggalnya saja
const formatWaktu = time.RFC3339

func (tr *KafkaTransaction) FromEntity(data *entity.FailedTransactions) *KafkaTransaction {
	if data == nil {
		return nil
	}

	res := &KafkaTransaction{
		Id:            data.Id,
		TransactionId: data.TransactionId,
		GroupId:       data.GroupId,
		PatientId:     data.PatientId,
		Message:       data.Message,
		Error:         data.Error,
		Attempt:       data.Attempt,
		Created:       data.CreatedAt.Format(formatWaktu),
	}

	if data.ResolvedAt != nil {
		s := data.ResolvedAt.Format(formatWaktu)
		res.Resolved = &s
	}

	return res
}

func (tr *KafkaTransaction) ToPayload() *KafkaTransactionPayload {
	if tr == nil {
		return nil
	}

	created, err := time.Parse(formatWaktu, tr.Created)
	if err != nil {
		created = time.Now()
	}

	res := &KafkaTransactionPayload{
		Id:            tr.Id,
		TransactionId: tr.TransactionId,
		GroupId:       tr.GroupId,
		PatientId:     tr.PatientId,
		Message:       tr.Message,
		Error:         tr.Error,
		Attempt:       tr.Attempt,
		Created:       created,
	}

	if tr.Resolved != nil {
		if t, err := time.Parse(formatWaktu, *tr.Resolved); err == nil {
			res.Resolved = &t
		}
	}

	return res
}

func (tr *KafkaTransactionPayload) ToEntity() *entity.FailedTransactions {
	if tr == nil {
		return nil
	}

	return &entity.FailedTransactions{
		Id:            tr.Id,
		TransactionId: tr.TransactionId,
		GroupId:       tr.GroupId,
		PatientId:     tr.PatientId,
		Message:       tr.Message,
		Error:         tr.Error,
		Attempt:       tr.Attempt,
		ResolvedAt:    tr.Resolved,
		CreatedAt:     tr.Created,
	}
}
