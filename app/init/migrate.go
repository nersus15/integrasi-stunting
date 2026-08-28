package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	memory "github.com/webcore-go/lib-memory"
	postgres "github.com/webcore-go/lib-postgres"
	sql "github.com/webcore-go/lib-sql"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/out"
	"github.com/webcore-go/webcore/infra/config"
)

var APP_LIBRARIES = map[string]core.LibraryLoader{
	"cache:memory":      &memory.MemoryLoader{},
	"database:postgres": &postgres.PostgresLoader{},
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: migrate [stunting] [command]")
		os.Exit(1)
	}

	service := os.Args[1]
	validServices := map[string]bool{
		"stunting": true,
	}

	if !validServices[service] {
		log.Fatal("Usage: migrate [stunting]")
		os.Exit(1)
	}

	var flags = flag.NewFlagSet("migrate", flag.ExitOnError)
	dir := flags.String("dir", "migration/stunting/postgres", "direktori file migrasi")

	command := os.Args[2]
	flags.Parse(os.Args[3:])
	args := flags.Args()

	ctx := context.Background()

	// Load configuration
	cfg := config.Config{}
	if err := config.LoadDefaultConfig(&cfg); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	out.SetEnvironment(cfg.App.Environment)

	// Reset Config
	cfg.Redis.Host = ""
	cfg.Kafka.Brokers = []string{}

	// Initialize application
	application := core.NewApp(ctx, &cfg, APP_LIBRARIES, []core.Module{})

	// Start the application
	if err := application.Context.Start(); err != nil {
		log.Fatalf("Failed to start migration %s: %v", service, err)
		os.Exit(1)
	}

	lib, ok := core.Instance().Context.GetDefaultSingletonInstance("database")
	if !ok {
		log.Fatal("Gagal memuat instance database")
		os.Exit(1)
	}

	db := lib.(*sql.SQLDatabase)
	db.StartMigration(ctx, service, command, *dir, args)

	fmt.Printf("Migration %s %s run successfully\n", service, command)
}
