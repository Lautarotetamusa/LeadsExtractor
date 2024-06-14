package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"leadsextractor/flow"
	"leadsextractor/infobip"
	"leadsextractor/models"
	"leadsextractor/whatsapp"
	"log"
	"net/http"
	"os"
	"reflect"
	"text/template"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type GetFlowPayload struct{
    Uuid uuid.NullUUID `json:"uuid"`
}

type FlowHandler struct {
    manager *flow.FlowManager
}

func (h *FlowHandler) GetFlows(w http.ResponseWriter, r *http.Request) error {
    dataResponse(w, h.manager.Flows)
    return nil
}

func (h *FlowHandler) NewFlow(w http.ResponseWriter, r *http.Request) error {
    var f flow.Flow
    if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
        return err
    }

    if err := h.manager.AddFlow(&f); err != nil {
        return err
    }

    dataResponse(w, f)
    return nil
}

func (h *FlowHandler) DeleteFlow(w http.ResponseWriter, r *http.Request) error {
    uuid, err := h.getUUIDFromParam(r)
    if err != nil {
        return err
    }
    if err := h.manager.DeleteFlow(*uuid); err != nil {
        return err
    }

    dataResponse(w, h.manager.Flows)
    return nil
}

func (h *FlowHandler) UpdateFlow(w http.ResponseWriter, r *http.Request) error {
    var f flow.Flow
    if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
        return err
    }

    uuid, err := h.getUUIDFromParam(r)
    if err != nil {
        return err
    }

    if err := h.manager.UpdateFlow(&f, *uuid); err != nil {
        return err
    }

    dataResponse(w, h.manager.Flows[*uuid])
    return nil
}

func (h *FlowHandler) SetFlowAsMain(w http.ResponseWriter, r *http.Request) error {
    uuid, err := h.getUUIDFromBody(r)
    if err != nil {
        return err
    }
    h.manager.SetMain(*uuid)

    dataResponse(w, h.manager.Flows[*uuid])
    return nil
}

func (s *Server) NewBroadcast(w http.ResponseWriter, r *http.Request) error {
    uuid, err := s.flowHandler.getUUIDFromBody(r)
    if err != nil {
        return err
    }
    
    query := "CALL getCommunications('2001-01-01', true)"
	communications := []models.Communication{}
	if err := s.Store.Db.Select(&communications, query); err != nil {
        log.Fatal(err)
	}

    s.flowHandler.manager.Broadcast(communications, *uuid)

	w.Header().Set("Content-Type", "application/json")
	res := struct {
		Success bool    `json:"success"`
		Count   int     `json:"count"`
	}{true, len(communications)}

	json.NewEncoder(w).Encode(res)
	return nil
}

func (h *FlowHandler) getUUIDFromBody(r *http.Request) (*uuid.UUID, error) {
    var body GetFlowPayload
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

    if _, exists := h.manager.Flows[body.Uuid.UUID]; !exists {
        return nil, fmt.Errorf("no existe ningun flow con id %s", body.Uuid.UUID)
    }

    return &body.Uuid.UUID, nil
}

func (h *FlowHandler) getUUIDFromParam(r *http.Request) (*uuid.UUID, error) {
    uuidStr, exists := mux.Vars(r)["uuid"]
    if !exists {
        return nil, fmt.Errorf("se necesita pasar un uuid")
    }
    uuid, err := uuid.Parse(uuidStr)
    if err != nil {
        return nil, err
    }
    if _, exists := h.manager.Flows[uuid]; !exists {
        return nil, fmt.Errorf("no existe ningun flow con id %s", uuid)
    }

    return &uuid, nil
}

func mustReadFile (filepath string) string {
    bytes, err := os.ReadFile(filepath)
    if err != nil {
        panic(fmt.Sprintf("error abriendo el archivo %s", filepath))
    }
    return string(bytes)
}

func formatMsg(tmpl string, c *models.Communication) string{
    t := template.Must(template.New("txt").Parse(tmpl))
    buf := &bytes.Buffer{}
    if err := t.Execute(buf, c); err != nil {
        return ""
    }
    return buf.String()
}

func (s *Server) setupActions(){
    cotizacion1 := mustReadFile("../messages/plantilla_cotizacion_1.txt")
    cotizacion2 := mustReadFile("../messages/plantilla_cotizacion_2.txt")

    flow.DefineAction("wpp.message",
        func(c *models.Communication, params interface{}) error {
            param, ok := params.(*flow.SendWppTextParam)
            if !ok {
                return fmt.Errorf("invalid parameters for wpp.message")
            }

            msg := formatMsg(param.Text, c)
            s.whatsapp.SendMessage(c.Telefono, msg)
            return nil
        },
        reflect.TypeOf(flow.SendWppTextParam{}),
    )

    flow.DefineAction("wpp.template", 
        func(c *models.Communication, params interface{}) error {
            param, ok := params.(*whatsapp.TemplatePayload)
            if !ok {
                return fmt.Errorf("invalid parameters for wpp.message")
            }

            whatsapp.ParseTemplatePayload(param)
            s.whatsapp.SendTemplate(c.Telefono, *param)
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

            s.whatsapp.Send(whatsapp.NewDocumentPayload(
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
            s.whatsapp.SendMsgAsesor(c.Asesor.Phone, c, c.IsNew)
            return nil
        },
        nil,
    )

    flow.DefineAction("wpp.send_image", 
		func(c *models.Communication, params interface{}) error {
            s.whatsapp.SendImage(c.Telefono, os.Getenv("WHATSAPP_IMAGE_ID"))
            return nil
        },
        nil,
    )

    flow.DefineAction("wpp.send_video", 
		func(c *models.Communication, params interface{}) error {
            s.whatsapp.SendVideo(c.Telefono, os.Getenv("WHATSAPP_VIDEO_ID"))
            return nil
        },
        nil,
    )

    flow.DefineAction("wpp.send_response", 
            func(c *models.Communication, params interface{}) error {
            s.whatsapp.SendResponse(c.Telefono, &c.Asesor)
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
            s.whatsapp.Send(whatsapp.NewTemplatePayload(c.Telefono, "broadcast_1", components))
            return nil
        },
        nil,
    )
    

    flow.DefineAction("infobip.save", 
        func(c *models.Communication, params interface{}) error {
            infobipLead := infobip.Communication2Infobip(c)
            s.infobipApi.SaveLead(infobipLead)
            return nil
        },
        nil,
    )

    flow.DefineAction("pipedrive.save", 
        func(c *models.Communication, params interface{}) error {
            s.pipedrive.SaveCommunication(c)
            return nil
        },
        nil,
    )

    s.flowHandler.manager.MustLoad()
}
