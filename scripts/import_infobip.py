
from src.sheets import Sheet, Logger, set_prop
from src.schema import LEAD_SCHEMA
import requests
import phonenumbers

def create_person(lead: dict):
    logger.debug("Cargando lead a infobip")
    api_url = 'https://8gzewr.api.infobip.com/people/2/persons'
    payload = {
        "firstName": lead['nombre'],
        "lastName": "",
        "customAttributes": {
            "prop_link": lead['propiedad']['link'],
            "prop_precio": lead['propiedad']['precio'],
            "prop_ubicacion": lead['propiedad']['ubicacion'],
            "prop_titulo": lead['propiedad']['titulo'],
            "contacted": False,
            "fuente": lead['fuente']
        },
        "contactInformation": {},
        "tags": "Seguimientos"
    }
    if lead['telefono'] != '':
        payload["contactInformation"]['phone'] = [{
            "number": lead['telefono']
        }]
    if lead['email'] != '':
        payload["contactInformation"]['email'] = [{
            "address": lead['email']
        }]
    headers = {
        'Authorization': 'App 97a01ad468d7489a60d851446c30725f-4f33fbc8-67eb-4443-8603-d26ab0c58298'
    }

    logger.debug(payload)
    res = requests.post(api_url, json=payload, headers=headers)
    if not res.ok:
        logger.error("Error en la request: " + str(res.status_code))
        logger.error(res.json())
        return

    logger.success("Lead cargada correctamente")

logger = Logger("Infobip import")
sheets = Sheet(logger, "mapping.json")

headers = sheets.get_dict_headers("A2:Y2")
rows = sheets.get("A3:4065")

mappings = []
for header in headers:
    mappings.append(sheets.mapping[header])

fila = 0
for row in rows:
    fila += 1
    lead = LEAD_SCHEMA.copy()
    for i, col in enumerate(row):
        lead = set_prop(lead, mappings[i], col)
    try:
        number = phonenumbers.parse(lead['telefono'], "MX")
        parsed_number = phonenumbers.format_number(number, phonenumbers.PhoneNumberFormat.E164)
        lead['telefono'] = parsed_number
        logger.debug("Numero obtenido: " + parsed_number)
    except phonenumbers.NumberParseException:
        logger.error("Error parseando el numero: " + lead['telefono'])
    create_person(lead)
    #exit(1)
