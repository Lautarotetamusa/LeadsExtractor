from message import format_msg
from src.whatsapp import Whatsapp
from src.lead import Lead

with open('messages/bienvenida_1.txt') as f:
    bienvenida_1 = f.read()
with open('messages/bienvenida_2.txt') as f:
    bienvenida_2 = f.read()
with open(f"messages/plantilla_cotizacion_1.txt", 'r') as f:
    cotizacion_1 = f.read()
with open(f"messages/plantilla_cotizacion_2.txt", 'r') as f:
    cotizacion_2 = f.read()

def new_lead_action(whatsapp: Whatsapp, lead: Lead) -> str:
    if lead.busquedas['covered_area'] == "" or lead.busquedas['covered_area'] == None:
        cotizacion_msj = cotizacion_2
    else:
        cotizacion_msj = cotizacion_1

    whatsapp.send_image(lead.telefono)
    whatsapp.send_message(lead.telefono, bienvenida_1)
    msg_2 = format_msg(lead, bienvenida_2)
    whatsapp.send_message(lead.telefono, msg_2)
    whatsapp.send_video(lead.telefono)

    if lead.cotizacion != "" and lead.cotizacion != None:
        whatsapp.send_document(lead.telefono, lead.cotizacion, 
            filename=f"Cotizacion para {lead.nombre}",
            caption=cotizacion_msj
        )

    return bienvenida_1 + '\n' + msg_2
