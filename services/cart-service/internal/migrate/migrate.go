package migrate

import (
	"GoCart/migrations"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func Up(pgURL string) {
	src, err := iofs.New(migrations.Files, ".")
	if err != nil {
		log.Fatalf("migrate: iofs new: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, pgURL)
	if err != nil {
		log.Fatalf("migrate: new with source: %v", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrate: up: %v", err)
	}

	log.Println("migrate: up OK (or no change)")
}
