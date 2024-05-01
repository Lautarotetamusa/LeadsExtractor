import sys
from dotenv import load_dotenv
sys.path.append('.')
load_dotenv()

import json
import src.jotform as jotform
from src.logger import Logger
from src.whatsapp import Whatsapp
from src.lead import Lead

logger = Logger("Test")
lead = Lead()
lead.set_args({
    "fuente": "inmuebles24",
    "fecha_lead": "2024-04-07",
    "id": "461161340",
    "fecha": "2024-04-08",
    "nombre": "Yolanda",
    "link": "https://www.inmuebles24.com/panel/interesados/198059132",
    "telefono": "5493415854220",
    "email": "cornejoy369@gmail.com",
    "message": "",
    "estado": "",
})
lead.set_asesor({
    "name": "Asesor de prueba",
    "phone": "+54 9 999999999"
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
lead.set_busquedas({
    "zonas": "cielo country club",
    "tipo": "Casa",
    "presupuesto": "13750000, 16750000",
    "cantidad_anuncios": 30,
    "contactos": 20,
    "inicio_busqueda": 85,
    "total_area": "371, 571",
    "covered_area": "",
    "banios": "4, 5",
    "recamaras": "0, 5"
})

def test_new_submission():
    pdf_url = jotform.new_submission(logger, lead)
    print(pdf_url)
    assert type(pdf_url) == str

def save_questions():
    questions = jotform.get_questions_form(logger)
    with open("src/jotform.json", "w") as f:
        json.dump(questions, f, indent=4)

if __name__ == "__main__":
    logger = Logger("Test")
    pdf_url = jotform.new_submission(logger, lead)
    print(pdf_url)
    wpp = Whatsapp(logger)
    if pdf_url != None:
        wpp.send_document(lead.telefono, pdf_url, 
                          filename=f"Cotizacion para {lead.nombre}",
                          caption="hola"
                          )
