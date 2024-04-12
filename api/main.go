package main

import (
	"fmt"
	"log"
	"os"

	"leadsextractor/infobip"
	"leadsextractor/pkg"
	"leadsextractor/store"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
    err := godotenv.Load("../.env")
    if err != nil {
        log.Printf("Error loading .env file")
    }
    db := store.ConnectDB()

    apiPort := os.Getenv("API_PORT")
    host := fmt.Sprintf("%s:%s", "localhost", apiPort)

    rr := pkg.NewRoundRobin(db)
    infobipApi := infobip.NewInfobipApi()
    server := pkg.NewServer(host, db, rr, infobipApi)
    server.Run()
}
