package main

import (
	"fmt"
	"leadsextractor/models"
	"leadsextractor/store"
	"leadsextractor/whatsapp"
	"log"
	"log/slog"
	"os"
	"slices"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func main(){
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

    for _, lead := range *leads{
        asesor := rr.Next()
        s.UpdateLeadAsesor(lead.Phone, &asesor)

        msg := fmt.Sprintf(`
        Tienes un nuevo lead reasignado del asesor %s
        Nombre: %s
        Telefono: %s
        `, asesorReasignado.Name, lead.Name, lead.Phone)

        wpp.SendMessage(asesor.Phone, msg)
    }

    return nil
}
