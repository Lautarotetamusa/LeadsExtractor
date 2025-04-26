// Con este script volcamos la forma en la que están asignados los leads en nuestra base de datos
// Reasignar desde pipedrive hacia la BD
package main

import (
	"context"
	"fmt"
	"leadsextractor/pkg/numbers"
	"leadsextractor/pkg/pipedrive"
	"leadsextractor/store"
	"log/slog"
	"os"
	"strings"
	"time"

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
	db := store.ConnectDB(context.Background())

    asesorStore := store.NewAsesorDBStore(db)
    leadStore := store.NewLeadStore(db)

    pipedriveConfig := pipedrive.Config{
        ClientId: os.Getenv("PIPEDRIVE_CLIENT_ID"),
        ClientSecret: os.Getenv("PIPEDRIVE_CLIENT_SECRET"),
        ApiToken: os.Getenv("PIPEDRIVE_API_TOKEN"),
        RedirectURI: os.Getenv("PIPEDRIVE_REDIRECT_URI"),
    }
	pipe := pipedrive.NewPipedrive(pipedriveConfig, logger)

    asesores, err := asesorStore.GetAllActive()
    if err != nil {
        panic("no fue posible obtener la lista de asesores")
    }

    for _, asesor := range asesores {
        user, err := pipe.GetUserByEmail(asesor.Email)
        if err != nil {
            logger.Error(err.Error())
            continue
        }

        deals, err := pipe.GetUserDeals(user)
        if err != nil {
            logger.Error(err.Error())
        }

        fmt.Printf("%s id=%d have %d deals\n", user.Name, user.Id, len(deals))

        for _, deal := range deals {
            person := deal.Person
            if person.Phone == nil || len(person.Phone) <= 0 {
                logger.Warn("person sin telefono", "person", person)
                continue
            }
            rawPhone := person.Phone[0].Value
            phone, err := numbers.NewPhoneNumber(strings.ReplaceAll(rawPhone, " ", ""))

            if err != nil {
                logger.Warn("invalid phone", "phone", person.Phone[0].Value)
                continue
            }

            dbLead, err := leadStore.GetOne(*phone)
            if err != nil {
                logger.Warn("no se encontro el lead", "err", err.Error())
                continue
            }
           
            logger.Info("Reasignando", "lead", dbLead.Name, "old asesor", dbLead.Asesor.Name, "new asesor", asesor.Name)

            // asesor, err := asesorStore.GetFromEmail(person.Owner.Email) 
            // if err != nil {
            //     logger.Warn("no se encontro el asesor", "err", err.Error())
            //     continue
            // }

            if err := leadStore.UpdateAsesor(dbLead.Phone, asesor); err != nil {
                logger.Error(err.Error())
            }
        }
    }
}
