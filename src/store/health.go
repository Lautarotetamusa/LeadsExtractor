package store

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// DBDependency adapta *sqlx.DB al contrato dependency.Dependency.
type DBDependency struct {
	db *sqlx.DB
}

func NewDBDependency(db *sqlx.DB) *DBDependency {
	return &DBDependency{db: db}
}

func (d *DBDependency) Name() string {
	return "mysql"
}

func (d *DBDependency) HealthCheck(ctx context.Context) error {
	return d.db.PingContext(ctx)
}
