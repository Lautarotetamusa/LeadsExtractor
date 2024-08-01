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

    // Lista de users (Asesores) para los cuales no tenenmos que reasignar
    blackList := map[string]any{
        "brenda.diaz@rebora.com.mx": true,
        "hernan.guerrero@rebora.com.mx": true,
    }

    // Vamos a reasignar todas las personas a este user (Hernan)
    newOwnerId := uint32(21828798)

    for _, user := range users{
        if _, ok := blackList[user.Email]; ok {
            continue
        }
        logger.Debug("Buscando deals de usuario", "userId", user.Id, "name", user.Name)

        deals, err := pipe.GetUserDeals(&user)
        if err != nil {
            logger.Error(err.Error())
            os.Exit(1)
        }
        logger.Debug("Deals encontrados", "cantidad", len(deals))

        for _, deal := range deals {
            if _, err := pipe.ReasignDeal(deal.Id, newOwnerId); err != nil {
                logger.Error(fmt.Sprintf("error actualizando deal %d, \n err=%s", deal.Id, err.Error()))
            }else{
                logger.Info(fmt.Sprintf("deal %d actualizado correctamente", deal.Id))
            }
        }
    }
}
