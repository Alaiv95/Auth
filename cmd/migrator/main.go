package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var storageDsn, migrationsPath, migrationsTable string

	flag.StringVar(&storageDsn, "storage-dsn", "", "Dsn to access storage")
	flag.StringVar(&migrationsPath, "migrations-path", "", "Path to store migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "Migrations table name")
	flag.Parse()

	if storageDsn == "" || migrationsPath == "" {
		panic("StoragePath and migrationsPath are required")
	}

	m, err := migrate.New("file://"+migrationsPath, storageDsn)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No change")
			return
		}
		panic(err)
	}
}
