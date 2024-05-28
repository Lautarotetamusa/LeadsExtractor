import sys
from dotenv import load_dotenv
sys.path.append('.')
load_dotenv()

import json
import src.jotform as jotform
from src.logger import Logger
from src.whatsapp import Whatsapp
from src.lead import Lead
import src.api as api

logger = Logger("Test")
lead = Lead()

def test_portal():
    lead.set_args({
        "fuente": "inmuebles24",
        "fecha_lead": "2024-04-07",
        "id": "461161340",
        "fecha": "2024-04-08",
        "nombre": "Lautaro Teta Musa",
        "link": "https://www.inmuebles24.com/panel/interesados/198059132",
        "telefono": "+5498888888888",
        "email": "cornejoy369@gmail.com",
        "message": "",
        "estado": "",
        "cotizacion": "https://www.jotform.com/pdf-submission/5910676377028343433",
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
    lead.set_busquedas({
        "zonas": "cielo country club",
        "tipo": "Casa",
        "presupuesto": "13750000, 16750000",
        "cantidad_anuncios": 30,
        "contactos": 20,
        "inicio_busqueda": 85,
        "total_area": "371, 571",
        "covered_area": "350, 457",
        "banios": "4, 5",
        "recamaras": "0, 5"
    })

    is_new, res = api.new_communication(logger, lead)
    assert res != None
    assert res.asesor != ""

def test_portal_sin_busquedas():
    lead.set_args({
        "fuente": "inmuebles24",
        "fecha_lead": "2024-04-07",
        "id": "461161340",
        "fecha": "2024-04-08",
        "nombre": "Lautaro Teta Musa",
        "link": "https://www.inmuebles24.com/panel/interesados/198059132",
        "telefono": "+5498888888888",
        "email": "cornejoy369@gmail.com",
        "message": "",
        "estado": "",
        "cotizacion": "https://www.jotform.com/pdf-submission/5910676377028343433",
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
    lead.set_busquedas({
        "zonas": "",
        "tipo": "",
        "presupuesto": "",
        "cantidad_anuncios": 30,
        "contactos": 20,
        "inicio_busqueda": 85,
        "total_area": "",
        "covered_area": "",
        "banios": "",
        "recamaras": ""
    })

    is_new, res = api.new_communication(logger, lead)
    assert res != None
    assert res.asesor != ""

