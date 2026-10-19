package repository

import (
	"fmt"

	"github.com/nersus15/integrasi/mod-stunting/helper/exceptions"
	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
	"github.com/webcore-go/webcore/infra/logger"
)

func (d *StuntingRepository) StartMigration(DB interface{}, dialect string, service string, command string, dir string, args []string) error {
	bunDB, ok := DB.(*bun.DB)
	if !ok {
		logger.Error("Gagal konversi: objek yang dikirim bukan merupakan *bun.DB")
		return exceptions.Internal.Messagef("koneksi database tidak dalam bentuk yang diharapkan")
	}

	switch dialect {
	case "sqlite":
		goose.SetDialect("sqlite3")
	case "postgres":
		goose.SetDialect("postgres")
		logger.Info(fmt.Sprintf("Membuat Schema %s Untuk postgres", d.Context.Config.Database.SchemaName))

		_, err := bunDB.ExecContext(d.Context.Context, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", d.Context.Config.Database.SchemaName))
		if err != nil {
			logger.Error(fmt.Sprintf("Gagal membuat schema %s: %v", d.Context.Config.Database.SchemaName, err))
			return err
		}
	}

	if service != "" {
		goose.SetTableName("__migration_" + service + "_logs")
	} else {
		goose.SetTableName("__migration_webcore_logs")
	}

	logger.Info(fmt.Sprintf("Mengeksekusi Goose %s pada folder: %s", command, dir))

	if err := goose.RunContext(d.Context.Context, command, bunDB.DB, dir, args...); err != nil {
		logger.Error(fmt.Sprintf("goose run %s: %v", command, err))
		return err
	}

	return nil
}
