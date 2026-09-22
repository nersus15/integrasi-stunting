package entity

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type FailedTransactions struct {
	bun.BaseModel `bun:"table:kafka.transactions,alias:tx"`

	Id            string           `bun:"id,pk,type:varchar(36)" json:"id"`
	TransactionId string           `bun:"transaction_id,type:varchar(36)" json:"transaction_id"`
	GroupId       string           `bun:"group_id,type:varchar(92)" json:"group_id"`
	PatientId     *string          `bun:"patient_id,type:varchar(36),nullzero" json:"patient_id"`
	Message       string           `bun:"message,type:text,notnull" json:"message"`
	Error         *json.RawMessage `bun:"error,type:jsonb,default:null" json:"error"`
	Attempt       int16            `bun:"attempt,type:smallint,notnull,default:1" json:"attempt"`
	ResolvedAt    *time.Time       `bun:"resolved_at,nullzero" json:"resolved_at"`
	CreatedAt     time.Time        `bun:"createdAt,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     *time.Time       `bun:"updatedAt,nullzero" json:"updated_at"`
}

func (t FailedTransactions) TableName() string {
	return "kafka.transactions"
}
