package app

import (
	authstoragedb "github.com/nersus15/lib-authstoredb"
	backgroundworker "github.com/nersus15/lib-background-worker"
	cron "github.com/nersus15/lib-go-cron"
	sqlite "github.com/nersus15/lib-sqlchiper"
	memory "github.com/webcore-go/lib-memory"
	postgres "github.com/webcore-go/lib-postgres"
	"github.com/webcore-go/webcore/adapter/auth/apikey"
	"github.com/webcore-go/webcore/app/core"
)

var APP_LIBRARIES = map[string]core.LibraryLoader{
	"cache:memory":          &memory.MemoryLoader{},
	"database:sqlite":       &sqlite.SqliteLoader{},
	"database:postgres":     &postgres.PostgresLoader{},
	"authstorage:db":        &authstoragedb.DBLoader{},
	"authentication:apikey": &apikey.ApiKeyLoader{},
	// Module.Init membutuhkan loader ini ketika module.fhir.cron.enabled aktif
	// (nilai default modul), jadi tanpa entri ini aplikasi gagal start.
	"cron":             &cron.CronLoader{},
	"backgroundworker": &backgroundworker.BackgroundWorkerLoader{},
}
