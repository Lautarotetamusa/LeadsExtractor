package store

import (
	"fmt"

	"leadsextractor/db/migrations"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

// RunMigrations aplica las migraciones pendientes embebidas en
// db/migrations. Se corre una vez al arrancar el servicio principal.
func RunMigrations(db *sqlx.DB) error {
	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("error configurando el dialecto de goose: %w", err)
	}

	if err := goose.Up(db.DB, "."); err != nil {
		return fmt.Errorf("error aplicando migraciones: %w", err)
	}

	return nil
}
