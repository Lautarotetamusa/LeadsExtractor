package main

import (
	"fmt"
	"leadsextractor/models"
	"leadsextractor/store"
	"leadsextractor/pkg/pipedrive"
	"log"
	"log/slog"
	"os"
	"slices"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func oo(){
    if len(os.Args) <= 1 {
        log.Fatal("falta el argumento 'phone'")
    }
    phone := os.Args[1]

    err := reasignAsesor(phone)
    if err != nil {
        log.Fatal(err)
    }
}

func reasignAsesor(phone string) error {
    w := os.Stderr

    logger := slog.New(
        tint.NewHandler(w, &tint.Options{
            Level:      slog.LevelDebug,
            TimeFormat: time.DateTime,
        }),
    )

	err := godotenv.Load("../.env")
	if err != nil {
		return fmt.Errorf("Error loading .env file")
	}
	db := store.ConnectDB()

    pipedriveConfig := pipedrive.Config{
        ClientId: os.Getenv("PIPEDRIVE_CLIENT_ID"),
        ClientSecret: os.Getenv("PIPEDRIVE_CLIENT_SECRET"),
        ApiToken: os.Getenv("PIPEDRIVE_API_TOKEN"),
        RedirectURI: os.Getenv("PIPEDRIVE_REDIRECT_URI"),
    }
	pipe := pipedrive.NewPipedrive(pipedriveConfig, logger)

	wpp := whatsapp.NewWhatsapp(
        os.Getenv("WHATSAPP_ACCESS_TOKEN"),
        os.Getenv("WHATSAPP_NUMBER_ID"),
        logger,
	)

    s := store.NewStore(db, logger) 
    asesorReasignado, err := s.GetOneAsesor(phone)
    if err != nil {
        return fmt.Errorf("no se encontro el asesor con telefono %s", phone)
    }
    asesorReasignado.Active = false
    err = s.UpdateAsesor(asesorReasignado, asesorReasignado.Phone)
    if err != nil {
        return fmt.Errorf("no fue posible actualizar al asesor")
    }

    var asesores []models.Asesor
    err = s.GetAllActiveAsesores(&asesores)
    if err != nil {
        return fmt.Errorf("no fue posible obtener la lista de asesores")
    }
    if slices.Contains(asesores, *asesorReasignado) {
        return fmt.Errorf("el asesor no se desactivo correctamente")
    }

    rr := store.NewRoundRobin(asesores)

    leads, err := s.GetLeadsFromAsesor(phone)
    if err != nil {
        return fmt.Errorf("no fue posible obtener los leads del asesor")
    }
    logger.Info(fmt.Sprintf("Se reasignarán un total de %d leads", len(*leads)))
        
    var pipedriveAsesores map[string]*pipedrive.User = make(map[string]*pipedrive.User)
    pipedriveAsesor, err := pipe.GetUserByEmail(asesorReasignado.Email)
    if err != nil {
        return err
    }
    for _, asesor := range asesores{
        newAsesor, err := pipe.GetUserByEmail(asesor.Email)
        if err != nil {
            return err
        }
        pipedriveAsesores[asesor.Phone] = newAsesor
    }

    for _, lead := range *leads{
        asesor := rr.Next()

        logger.Debug("Reasignando lead en pipedrive")
        person, err := pipe.GetPersonByNumber(lead.Phone)
        if err != nil {
            logger.Error(fmt.Sprintf("%v", err))
            continue
        }
        fmt.Println(person.Id, pipedriveAsesor.Id)

        deal, err := pipe.SearchPersonDeal(person.Id, pipedriveAsesor.Id)
        if err != nil {
            logger.Error(fmt.Sprintf("%v", err))
            return err
        }
        pipe.ReasignDeal(deal.Id, pipedriveAsesores[asesor.Phone].Id)

        msg := fmt.Sprintf(`
        Tienes un nuevo lead reasignado del asesor %s
        Nombre: %s
        Telefono: %s
        `, asesorReasignado.Name, lead.Name, lead.Phone)

        logger.Debug("Enviando mensaje al asesor")
        wpp.SendMessage(asesor.Phone, msg)

        logger.Debug("Actualizando en base de datos")
        s.UpdateAsesor(lead.Phone, &asesor)
    }

    return nil
}
