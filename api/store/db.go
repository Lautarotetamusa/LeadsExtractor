package store

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
)

func ConnectDB() *sqlx.DB{
    host := os.Getenv("HOST")
    dbUser := os.Getenv("DB_USER")
    dbPort := os.Getenv("DB_PORT")
    dbPass := os.Getenv("DB_PASS")
    dbName := os.Getenv("DB_NAME")

    connectionStr := fmt.Sprintf("%s:%s@(%s:%s)/%s", dbUser, dbPass, host, dbPort, dbName)
    fmt.Printf("str: %s\n", connectionStr)

    db, err := sqlx.Connect("mysql", connectionStr)
    if err != nil {
        log.Fatal("Imposible conectar con la base de datos", err)
    }
    return db
}
