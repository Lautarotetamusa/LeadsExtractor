import requests
from dotenv import load_dotenv
import os

from src.lead import Lead
from src.sheets import Logger 

load_dotenv()
API_KEY = os.getenv("INFOBIP_APIKEY")
API_URL = os.getenv("INFOBIP_APIURL")
if API_KEY == None or API_URL == None:
    print("ERROR: INFOBIP_APIKEY or INFOBIP_APIURL env variables not set")
    exit(1)
HEADERS = {
    'Authorization': API_KEY
}

def get_all_person(logger: Logger) -> list[dict]:
    logger.debug("Buscando personas")

    page = 1
    persons = []

    while True:
        logger.debug(f"GET {API_URL}?page={page}")
        res = requests.get(API_URL, params={"page": page}, headers=HEADERS)
        if not res.ok:
            logger.error("Error en la request: " + str(res.status_code))
            logger.error(res.json())
        else: 
            data = res.json()
            if len(data.get('persons', [])) == 0:
                logger.debug("Se encontro un pagina vacia, terminando")
                break

            persons.append(data.get('persons', []))
        page += 1

    return persons

#@filter -> url encoder filter
def search_person(logger: Logger, phone: str) -> None | dict:
    res = requests.get(f"{API_URL}?type=PHONE&identifier={phone}", headers=HEADERS)
    try:
        if not res.ok:
            logger.error("Error en la request: " + str(res.status_code))
            logger.error(res.json())
            return None
    except Exception as e:
        logger.error("Error peticion infobip "+str(e))
        return None

    persons = res.json().get('persons', [])
    
    if len(persons) == 0:
        return None

    return persons[0]

def update_person(logger: Logger, id: int, payload: dict):
    logger.debug(f"Actualizando persona {id}")

    res = requests.put(API_URL, params={"id": id}, json=payload, headers=HEADERS)
    if not res.ok:
        logger.error("Error en la request: " + str(res.status_code))
        logger.error(res.json())
        return

    logger.success(f"Persona {id} actualizada correctamente")

#Si valid_number es True, no se parseara el numero para no generar problemas
def create_person(logger: Logger, lead: Lead):
    logger.debug("Cargando lead a infobip")

    payload = {
        "firstName": lead.nombre,
        "lastName": "",
        "customAttributes": {
            "prop_link": lead.propiedad['link'],
            "prop_precio": str(lead.propiedad['precio']),
            "prop_ubicacion": lead.propiedad['ubicacion'],
            "prop_titulo": lead.propiedad['titulo'],
            "contacted": False,
            "fuente": lead.fuente,
            "asesor_name": lead.asesor['name'],
            "asesor_phone": lead.asesor['phone']
        },
        "contactInformation": {},
        "tags": "Seguimientos"
    }
    if lead.telefono != '':
        payload["contactInformation"]['phone']= [{
            "number": lead.telefono
        }]
    if lead.email != '':
        payload["contactInformation"]['email'] = [{
            "address": lead.email
        }]

    try:
        res = requests.post(API_URL, json=payload, headers=HEADERS)
        if not res.ok:
            logger.error("Error en la request: " + str(res.status_code))
            logger.error(res.json())
            logger.error(payload)
            return
    except Exception as e:
        logger.error("Error cargando lead"+str(e))
        return

    logger.success("Lead cargada correctamente")
