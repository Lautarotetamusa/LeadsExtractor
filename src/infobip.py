from src.sheets import Logger 
import requests
import phonenumbers

from dotenv import load_dotenv
import os

load_dotenv()
API_KEY = os.getenv("INFOBIP_APIKEY")
API_URL = os.getenv("INFOBIP_APIURL")
if API_KEY == None or API_URL == None:
    print("ERROR: INFOBIP_APIKEY or INFOBIP_APIURL env variables not set")
    exit(1)

def parse_number(logger: Logger, phone: str) -> None | str:
    logger.debug("Parseando numero")
    try:
        number = phonenumbers.parse(phone, "MX")
        parsed_number = phonenumbers.format_number(number, phonenumbers.PhoneNumberFormat.E164)
        logger.success("Numero obtenido: " + parsed_number)
        return parsed_number
    except phonenumbers.NumberParseException:
        logger.error("Error parseando el numero: " + phone)
        return None
        
def create_person(logger: Logger, lead: dict):
    logger.debug("Cargando lead a infobip")

    payload = {
        "firstName": lead['nombre'],
        "lastName": "",
        "customAttributes": {
            "prop_link": lead['propiedad']['link'],
            "prop_precio": str(lead['propiedad']['precio']),
            "prop_ubicacion": lead['propiedad']['ubicacion'],
            "prop_titulo": lead['propiedad']['titulo'],
            "contacted": False,
            "fuente": lead['fuente']
        },
        "contactInformation": {},
        "tags": "Seguimientos"
    }
    if lead['telefono'] != '':
        lead['telefono'] = parse_number(logger, lead['telefono'])
        payload["contactInformation"]['phone'] = [{
            "number": lead['telefono']
        }]
    if lead['email'] != '':
        payload["contactInformation"]['email'] = [{
            "address": lead['email']
        }]
    headers = {
        'Authorization': API_KEY
    }

    logger.debug(payload)
    res = requests.post(API_URL, json=payload, headers=headers)
    if not res.ok:
        logger.error("Error en la request: " + str(res.status_code))
        logger.error(res.json())
        return

    logger.success("Lead cargada correctamente")
