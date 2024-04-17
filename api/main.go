package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"leadsextractor/infobip"
	"leadsextractor/pipedrive"
	"leadsextractor/pkg"
	"leadsextractor/store"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
    rand.Seed(time.Now().UnixNano())

    err := godotenv.Load("../.env")
    if err != nil {
        log.Printf("Error loading .env file")
    }
    db := store.ConnectDB()

    apiPort := os.Getenv("API_PORT")
    host := fmt.Sprintf("%s:%s", "localhost", apiPort)

    rr := pkg.NewRoundRobin(db)
    infobipApi := infobip.NewInfobipApi()
    
    pipedriveApi := pipedrive.NewPipedrive(
        os.Getenv("PIPEDRIVE_CLIENT_ID"),
        os.Getenv("PIPEDRIVE_CLIENT_SECRET"),
        os.Getenv("PIPEDRIVE_API_TOKEN"),
        os.Getenv("PIPEDRIVE_REDIRECT_URI"),
    )

    server := pkg.NewServer(host, db, rr, infobipApi, pipedriveApi)
    server.Run()
}
