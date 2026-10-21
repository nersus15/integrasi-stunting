-- +goose Up
-- +goose StatementBegin

ALTER TABLE stunting.orangtua ADD COLUMN IF NOT EXISTS created_by varchar(100);
ALTER TABLE stunting.anak ADD COLUMN IF NOT EXISTS created_by varchar(100);
ALTER TABLE stunting.kunjungan ADD COLUMN IF NOT EXISTS created_by varchar(100);
ALTER TABLE stunting.kesehatan
    ADD COLUMN IF NOT EXISTS created_by varchar(100),
    ADD COLUMN IF NOT EXISTS updated_by varchar(100),
    ADD COLUMN IF NOT EXISTS deleted_by varchar(100);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE stunting.kesehatan DROP COLUMN IF EXISTS deleted_by, DROP COLUMN IF EXISTS updated_by, DROP COLUMN IF EXISTS created_by;
ALTER TABLE stunting.kunjungan DROP COLUMN IF EXISTS created_by;
ALTER TABLE stunting.anak DROP COLUMN IF EXISTS created_by;
ALTER TABLE stunting.orangtua DROP COLUMN IF EXISTS created_by;

-- +goose StatementEnd
