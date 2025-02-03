package store

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	db *sqlx.DB
    logger *slog.Logger
}

func NewStore(db *sqlx.DB, logger *slog.Logger) *Store {
	return &Store{
		db: db,
        logger: logger.With("module", "store"),
	}
}
