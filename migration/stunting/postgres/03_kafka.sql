-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS kafka;

CREATE TABLE IF NOT EXISTS kafka.transactions (
    id varchar(36) NOT NULL,
    transaction_id varchar(36) NOT NULL,
    group_id varchar(92) NOT NULL,
    patient_id varchar(36) DEFAULT NULL,
    message text NOT NULL,
    error jsonb DEFAULT NULL,
    attempt smallint NOT NULL DEFAULT 1,
    resolved_at timestamp(6) with time zone DEFAULT NULL,
    "createdAt" timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updatedAt" timestamp(6) with time zone,

    CONSTRAINT pk_kafka_transactions PRIMARY KEY (id),
    CONSTRAINT uq_kafka_transactions_key UNIQUE (transaction_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_kafka_transactions_belum_selesai
    ON kafka.transactions (group_id, "createdAt") WHERE resolved_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_kafka_transactions_patient ON kafka.transactions (patient_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS kafka.transactions;

-- +goose StatementEnd
