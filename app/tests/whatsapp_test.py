import sys
from dotenv import load_dotenv
sys.path.append('.')
load_dotenv()

from src.whatsapp import Whatsapp
from src.lead import Lead

if __name__ == "__main__":
    lead = Lead()

    lead.set_args({
        "fuente": "inmuebles24",
        "fecha_lead": "2024-04-07",
        "id": "461161340",
        "fecha": "2024-04-08",
        "nombre": "Lautaro",
        "link": "https://www.inmuebles24.com/panel/interesados/198059132",
        "telefono": "5493415854220",
        "email": "cornejoy369@gmail.com",
        "message": "",
        "estado": "",
    })
    lead.set_asesor({
        "name": "",
        "phone": ""
    })
    lead.set_propiedad({
        "id": "a78c1555-f684-4de7-bbf1-a7288461fe51",
        "titulo": "Casa en venta en El Cielo Country Club Incre\u00edble dise\u00f1o y amplitud",
        "link": "",
        "precio": "16117690",
        "ubicacion": "cielo country club",
        "tipo": "Casa",
        "municipio": "Tlajomulco de Z\u00fa\u00f1iga"
    })

    wpp = Whatsapp()
    #wpp.send_message(lead.telefono, "Mensaje de prueba")
    #wpp.send_response("5493415854220", {"name": "Diego", "phone": "4444"})
    wpp.send_template("5493415854220", "8_sobreprecio", [], "es_ES")
    #wpp.send_msg_asesor("5493415854220", lead, False)

    pdf_url = "https://www.jotform.com/pdf-submission/5903736415724129773"
    plantilla = "🏡😀Te comparto la cotización de la casa que consultaste, si tienes alguna duda, con mucho gusto tu asesor Rebora esta feliz de ayudarte, que tengas un gran día!"
    wpp.send_document(lead.telefono, pdf_url, 
                      filename=f"Cotizacion para {lead.nombre}",
                      caption=plantilla
                      )
