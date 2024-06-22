// Con este script volcamos la forma en la que están asignados los leads en nuestra base de datos
package main

import (
	"fmt"
	"leadsextractor/models"
	"leadsextractor/pipedrive"
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
	db := store.ConnectDB()
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

    // phone persona -> id persona
    findPersons := make(map[string]uint32)
    count := 0

    for _, asesor := range asesores {
        user, err := pipe.GetUserByEmail(asesor.Email)
        if err != nil {
            logger.Error(err.Error())
            continue
        }

        persons, err := pipe.GetUserPersons(user)
        if err != nil {
            logger.Error(err.Error())
            continue
        }

        fmt.Printf("%s id=%d have %d persons\n", user.Name, user.Id, len(persons))

        for _, person := range persons {
            if person.Phone == nil || len(person.Phone) <= 0 || person.Phone[0].Value == "" {
                logger.Warn("person sin telefono", "person", person)
                continue
            }
            personPhone := person.Phone[0].Value

            if dupPersonId, exists := findPersons[personPhone]; exists {
                logger.Info("se encontro una persona duplicada", "phone", personPhone)

                if _, err := pipe.MergePersons(dupPersonId, person.Id); err != nil {
                    logger.Error(err.Error())
                }
                count += 1
                continue
            }
            findPersons[personPhone] = person.Id 
        }
    }
    logger.Info("se encontraron personas duplicadas", "count", count)
}
