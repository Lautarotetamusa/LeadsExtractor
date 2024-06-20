package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"time"

	"leadsextractor/flow"
	"leadsextractor/infobip"
	"leadsextractor/models"
	"leadsextractor/pipedrive"
	"leadsextractor/pkg"
	"leadsextractor/store"
	"leadsextractor/whatsapp"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func main() {
    ctx := context.Background()

    w := os.Stderr

    logger := slog.New(
        tint.NewHandler(w, &tint.Options{
            Level:      slog.LevelDebug,
            TimeFormat: time.DateTime,
        }),
    )

	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	db := store.ConnectDB(ctx)

	apiPort := os.Getenv("API_PORT")
	host := fmt.Sprintf("%s:%s", "localhost", apiPort)

	infobipApi := infobip.NewInfobipApi(
        os.Getenv("INFOBIP_APIURL"),
        os.Getenv("INFOBIP_APIKEY"),
        "5213328092850",
        logger,
    )

	pipedriveApi := pipedrive.NewPipedrive(
		os.Getenv("PIPEDRIVE_CLIENT_ID"),
		os.Getenv("PIPEDRIVE_CLIENT_SECRET"),
		os.Getenv("PIPEDRIVE_API_TOKEN"),
		os.Getenv("PIPEDRIVE_REDIRECT_URI"),
        logger,
	)

	wpp := whatsapp.NewWhatsapp(
        os.Getenv("WHATSAPP_ACCESS_TOKEN"),
        os.Getenv("WHATSAPP_NUMBER_ID"),
        logger,
	)

    flowManager := flow.NewFlowManager("actions.json", logger)
    defineActions(wpp, pipedriveApi, infobipApi)
    flowManager.MustLoad()

    webhook := whatsapp.NewWebhook(logger)

	router := mux.NewRouter()

    router.Use(loggingMiddleware)

    flowHandler := pkg.NewFlowHandler(flowManager)

	server := pkg.NewServer(host, logger, db, flowHandler)
    server.SetRoutes(router)
    
    go webhook.ConsumeEntries(server.NewCommunication)

	router.HandleFunc("/pipedrive", pkg.HandleErrors(pipedriveApi.HandleOAuth)).Methods("GET")

    router.HandleFunc("/webhooks", pkg.HandleErrors(webhook.ReciveNotificaction)).Methods("POST", "OPTIONS")
    router.HandleFunc("/webhooks", pkg.HandleErrors(webhook.Verify)).Methods("GET", "OPTIONS")

	router.HandleFunc("/actions", pkg.HandleErrors(flowHandler.NewFlow)).Methods("POST", "OPTIONS")
	router.HandleFunc("/actions/{uuid}", pkg.HandleErrors(flowHandler.UpdateFlow)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/actions", pkg.HandleErrors(flowHandler.GetFlows)).Methods("GET", "OPTIONS")
	router.HandleFunc("/actions/{uuid}", pkg.HandleErrors(flowHandler.DeleteFlow)).Methods("DELETE", "OPTIONS")

	server.Run(router)
}

//Definimos las acciones permitidas dentro de un flow
func defineActions(wpp *whatsapp.Whatsapp, pipedriveApi *pipedrive.Pipedrive, infobipApi *infobip.InfobipApi) {
    cotizacion1 := mustReadFile("../messages/plantilla_cotizacion_1.txt")
    cotizacion2 := mustReadFile("../messages/plantilla_cotizacion_2.txt")

    flow.DefineAction("wpp.message",
        func(c *models.Communication, params interface{}) error {
            param, ok := params.(*flow.SendWppTextParam)
            if !ok {
                return fmt.Errorf("invalid parameters for wpp.message")
            }

            msg := pkg.FormatMsg(param.Text, c)
            wpp.SendMessage(c.Telefono, msg)
            return nil
        },
        reflect.TypeOf(flow.SendWppTextParam{}),
    )

    flow.DefineAction("wpp.template", 
        func(c *models.Communication, params interface{}) error {
            payload, ok := params.(*whatsapp.TemplatePayload)
            if !ok {
                return fmt.Errorf("invalid parameters for wpp.message")
            }

            payload.Components[0].ParseParameters(c)
            wpp.SendTemplate(c.Telefono, *payload)
            return nil
        },
        reflect.TypeOf(whatsapp.TemplatePayload{}), 
    )

    flow.DefineAction("wpp.cotizacion", 
        func(c *models.Communication, params interface{}) error {
            if c.Cotizacion == "" {
                return fmt.Errorf("el lead %s no tiene cotizacion", c.Nombre)
            }

            var caption string
            if !c.Busquedas.CoveredArea.Valid {
                caption = cotizacion2
            }else{
                caption = cotizacion1
            }

            wpp.Send(whatsapp.NewDocumentPayload(
                c.Telefono,
                c.Cotizacion,
                caption,
                fmt.Sprintf("Cotizacion para %s", c.Nombre),
            ))
            return nil
        },
        nil,
    )

    flow.DefineAction("wpp.send_message_asesor", 
		func(c *models.Communication, params interface{}) error {
            wpp.SendMsgAsesor(c.Asesor.Phone, c, c.IsNew)
            return nil
        },
        nil,
    )

    flow.DefineAction("wpp.send_image", 
		func(c *models.Communication, params interface{}) error {
            wpp.SendImage(c.Telefono, os.Getenv("WHATSAPP_IMAGE_ID"))
            return nil
        },
        nil,
    )

    flow.DefineAction("wpp.send_video", 
		func(c *models.Communication, params interface{}) error {
            wpp.SendVideo(c.Telefono, os.Getenv("WHATSAPP_VIDEO_ID"))
            return nil
        },
        nil,
    )

    flow.DefineAction("wpp.send_response", 
            func(c *models.Communication, params interface{}) error {
            wpp.SendResponse(c.Telefono, &c.Asesor)
            return nil
        },
        nil,
    )

    flow.DefineAction("wpp.broadcast",         
        func(c *models.Communication, params interface{}) error {
            components := []whatsapp.Components{{
                Type:       "body",
                Parameters: []whatsapp.Parameter{{
                    Type: "text",
                    Text: c.Nombre,
                }},
            }}
            wpp.Send(whatsapp.NewTemplatePayload(c.Telefono, "broadcast_1", components))
            return nil
        },
        nil,
    )

    flow.DefineAction("infobip.save", 
        func(c *models.Communication, params interface{}) error {
            infobipLead := infobip.Communication2Infobip(c)
            infobipApi.SaveLead(infobipLead)
            return nil
        },
        nil,
    )

    flow.DefineAction("pipedrive.save", 
        func(c *models.Communication, params interface{}) error {
            pipedriveApi.SaveCommunication(c)
            return nil
        },
        nil,
    )
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Println(r.Method, r.RequestURI)
        next.ServeHTTP(w, r)
    })
}

func mustReadFile (filepath string) string {
    bytes, err := os.ReadFile(filepath)
    if err != nil {
        panic(fmt.Sprintf("error abriendo el archivo %s", filepath))
    }
    return string(bytes)
}
