package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nersus15/integrasi/mod-stunting/entity"
	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/nersus15/integrasi/mod-stunting/helper/utils"
	"github.com/uptrace/bun"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

func (d *StuntingRepository) FindKafkaTransaction(TransactionId string) (*entity.FailedTransactions, error) {
	filter := []port.DbExpression{
		{
			Expr: "transaction_id = ? AND group_id = ? AND resolved_at IS NULL",
			Args: []any{TransactionId, d.Config.Kafka.GroupID},
		},
	}

	tmp := new(entity.FailedTransactions)

	err := d.Connection.FindOne(d.Context.Context, tmp, tmp.TableName(), []string{"*"}, filter, map[string]int{})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, exceptions.ErrTidakDitemukan
	}
	if err != nil {
		logger.Error(fmt.Sprintf("FindKafkaTransaction (%s): ", TransactionId) + err.Error())
		return nil, err
	}

	return tmp, nil
}

func (d *StuntingRepository) FindKafkaTransactions(patientId *string) ([]*entity.FailedTransactions, error) {
	filter := []port.DbExpression{
		{
			Expr: "group_id = ? AND resolved_at IS NULL",
			Args: []any{d.Config.Kafka.GroupID},
		},
	}

	if utils.IsFilled(patientId) {
		filter = append(filter, port.DbExpression{Expr: "patient_id = ?", Args: []any{*patientId}})
	}

	tmp := make([]*entity.FailedTransactions, 0)
	err := d.Connection.Find(d.Context.Context, &tmp, entity.FailedTransactions{}.TableName(),
		[]string{"*"}, filter, map[string]int{"createdAt": 1}, 0, 0)

	if err != nil {
		logger.Error("FindKafkaTransactions: " + err.Error())
		return nil, err
	}

	return tmp, nil
}

func (d *StuntingRepository) SaveKafkaTransaction(e *entity.FailedTransactions) error {
	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		logger.Error("Gagal konversi: bukan merupakan *bun.DB")
		return exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	_, err := bunDB.NewInsert().Model(e).
		On("CONFLICT (transaction_id, group_id) DO UPDATE").
		Set("message = EXCLUDED.message").
		Set("error = EXCLUDED.error").
		Set("patient_id = EXCLUDED.patient_id").
		Set("attempt = tx.attempt + 1").
		Set("resolved_at = NULL").
		Set(`"updatedAt" = ?`, time.Now()).
		Exec(d.Context.Context)

	if err != nil {
		logger.Error(fmt.Sprintf("SaveKafkaTransaction (%s): ", e.TransactionId) + err.Error())
	}

	return err
}

func (d *StuntingRepository) TandaiKafkaTransactionSelesai(transactionId string) error {
	bunDB, ok := d.Connection.GetConnection().(*bun.DB)
	if !ok {
		return exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	_, err := bunDB.NewUpdate().Model((*entity.FailedTransactions)(nil)).
		Set("resolved_at = ?", time.Now()).
		Set("error = NULL").
		Set(`"updatedAt" = ?`, time.Now()).
		Where("transaction_id = ? AND group_id = ?", transactionId, d.Config.Kafka.GroupID).
		Exec(d.Context.Context)

	if err != nil {
		logger.Error(fmt.Sprintf("TandaiKafkaTransactionBeres (%s): ", transactionId) + err.Error())
	}

	return err
}
