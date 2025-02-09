// Con este script volcamos la forma en la que están asignados los leads en nuestra base de datos
// Reasignar desde pipedrive hacia la BD
package main

import (
	"context"
	"fmt"
	"leadsextractor/models"
	"leadsextractor/pkg/pipedrive"
	"leadsextractor/store"
	"log/slog"
	"os"
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
    s := store.NewStore(db, logger) 

    pipedriveConfig := pipedrive.Config{
        ClientId: os.Getenv("PIPEDRIVE_CLIENT_ID"),
        ClientSecret: os.Getenv("PIPEDRIVE_CLIENT_SECRET"),
        ApiToken: os.Getenv("PIPEDRIVE_API_TOKEN"),
        RedirectURI: os.Getenv("PIPEDRIVE_REDIRECT_URI"),
    }
	pipe := pipedrive.NewPipedrive(pipedriveConfig, logger)

    var asesores []models.Asesor
    err = s.GetAllAsesores(&asesores)
    if err != nil {
        panic("no fue posible obtener la lista de asesores")
    }

    for _, asesor := range asesores {
        user, err := pipe.GetUserByEmail(asesor.Email)
        if err != nil {
            logger.Error(err.Error())
            continue
        }

        persons, err := pipe.GetUserPersons(user)
        if err != nil {
            logger.Error(err.Error())
        }

        fmt.Printf("%s id=%d have %d persons\n", user.Name, user.Id, len(persons))

        for _, person := range persons {
            if person.Phone == nil || len(person.Phone) <= 0 {
                logger.Warn("person sin telefono", "person", person)
                continue
            }

            dbLead, err := s.GetOne(person.Phone[0].Value)
            if err != nil {
                //logger.Warn("no se encontro el lead", "err", err.Error())
                continue
            }
           
            if dbLead.Asesor.Email != person.Owner.Email {
                logger.Info("Reasignando", "lead", dbLead.Name, "old asesor", dbLead.Asesor.Name, "new asesor", asesor.Name)

                asesor, err := s.GetAsesorFromEmail(person.Owner.Email) 
                if err != nil {
                    logger.Warn("no se encontro el asesor", "err", err.Error())
                    continue
                }

                if err := s.UpdateAsesor(dbLead.Phone, asesor); err != nil {
                    logger.Error(err.Error())
                }
            }
        }
    }
}
