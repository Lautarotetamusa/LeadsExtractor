// Reasignar desde la db, hacia pipedrive

package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"
	"leadsextractor/pipedrive"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func main(){
    w := os.Stderr

    logger := slog.New(
        tint.NewHandler(w, &tint.Options{
            Level:      slog.LevelDebug,
            TimeFormat: time.DateTime,
        }),
    )

	err := godotenv.Load("../.env")
	if err != nil {
		panic("Error loading .env file")
	}

    pipedriveConfig := pipedrive.Config{
        ClientId: os.Getenv("PIPEDRIVE_CLIENT_ID"),
        ClientSecret: os.Getenv("PIPEDRIVE_CLIENT_SECRET"),
        ApiToken: os.Getenv("PIPEDRIVE_API_TOKEN"),
        RedirectURI: os.Getenv("PIPEDRIVE_REDIRECT_URI"),
    }
    pipe := pipedrive.NewPipedrive(pipedriveConfig, logger)

    users, err := pipe.GetUsers()
    if err != nil{
        logger.Error(err.Error())
        os.Exit(1)
    }

    for _, user := range users{
        fmt.Printf("%#v", user)

        persons, err := pipe.GetUserPersons(&user)
        if err != nil {
            logger.Error(err.Error())
            os.Exit(1)
        }
        fmt.Println(len(persons))
    }
}
