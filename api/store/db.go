package store

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
)

func ConnectDB(ctx context.Context) *sqlx.DB{
    host := os.Getenv("HOST")
    dbUser := os.Getenv("DB_USER")
    dbPort := os.Getenv("DB_PORT")
    dbPass := os.Getenv("DB_PASS")
    dbName := os.Getenv("DB_NAME")

    connectionStr := fmt.Sprintf("%s:%s@(%s:%s)/%s", dbUser, dbPass, host, dbPort, dbName)
    fmt.Printf("db connection: %s\n", connectionStr)

    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    db, err := sqlx.ConnectContext(ctx, "mysql", connectionStr)
    if err != nil {
        log.Fatal("Imposible conectar con la base de datos", err)
    }

    if err := db.PingContext(ctx); err != nil {
        log.Fatal("error pinging database: %w", err)
    }

    return db
}
