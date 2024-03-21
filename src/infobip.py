from datetime import datetime
import json
from typing import Iterator
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

#@filter -> url encoder filter
def get_persons(logger: Logger, filter: str | None) -> Iterator[list[dict]]:
    logger.debug("Buscando personas")

    page = 1
    while True:
        url = f"{API_URL}?filter={filter}&page={page}" if filter != None else f"{API_URL}?page={page}" 
        logger.debug("GET "+url)
        res = requests.get(url, headers=HEADERS)
        if not res.ok:
            logger.error("Error en la request: " + str(res.status_code))
            logger.error(res.json())
        else: 
            data = res.json()
            if len(data.get('persons', [])) == 0:
                logger.debug("Se encontro un pagina vacia, terminando")
                break
            yield data.get('persons', [])
        page += 1

def search_person(logger: Logger, phone: str) -> None | Lead:
    res = requests.get(f"{API_URL}?type=PHONE&identifier={phone}", headers=HEADERS)
    try:
        if not res.ok:
            logger.error("Error en la request: " + str(res.status_code))
            logger.error(res.json())
            return None
    except Exception as e:
        logger.error("Error peticion infobip "+str(e))
        return None

    person = res.json()
    
    if person.get("errorCode", None) != None:
        return None
    
    return infobip2lead(person)

def infobip2lead(person: dict) -> Lead:
    lead = Lead()
    attrs = person.get('customAttributes', {})
    attrs = person.get('customAttributes', {})
    lead.set_asesor({
        'name': attrs.get('asesor_name', None),
        'phone': attrs.get('asesor_phone', None)
    })
    lead.set_args({
        'nombre': person.get('firstName', '') + ' ' + person.get('lastName', ''),
        'telefono': person.get('contactInformation', {}).get('phone', [{}])[0].get('number', None),
        'fecha_lead': attrs.get('fecha_lead', None),
        'estado': 'Duplicado'
    })
    return lead

def update_person(logger: Logger, id: int, payload: dict):
    logger.debug(f"Actualizando persona {id}")

    try:
        res = requests.put(API_URL, params={"id": id}, json=payload, headers=HEADERS)
        if not res.ok:
            logger.error("Error en la request: " + str(res.status_code))
            logger.error(res.json())
            return False
    except Exception as e:
        logger.error("Error cargando lead"+str(e))
        return False

    logger.success(f"Persona {id} actualizada correctamente")
    return True

def add_fecha_lead(logger: Logger, leads: list[Lead]) -> bool | list[dict]:
    persons = []
    phones = []
    for lead in leads:
        #Si intentamos actualizar una persona dos veces en el mismo payload dice que no existe el numero aunque si exista
        #De este modo eliminamos duplicados y solucionamos el problema
        if lead.telefono in phones: #ERROR muy raro de infobip
            continue

        person = {
            "query": {
                "identifier": lead.telefono[1:], #Sacamos el +
                "type": "PHONE"
            },
            "update": {
                "customAttributes":{
                    "fecha_lead": lead.fecha_lead
                }
            }
        }
        persons.append(person)
        phones.append(lead.telefono)
    payload = {'people': persons}
    print(payload)

    try:
        res = requests.patch(API_URL+'/batch', json=payload, headers=HEADERS)
        if not res.ok:
            logger.error("Error en la request: " + str(res.status_code))
            logger.error(res.json())
            return False
    except Exception as e:
        logger.error("Error cargando lead"+str(e))
        return False

    #Cuando sale todo bien devuelve null
    if res.json() == None:
        return []
    
    return res.json()['results']

#Si valid_number es True, no se parseara el numero para no generar problemas
def create_person(logger: Logger, lead: Lead):
    logger.debug("Cargando lead a infobip")

    payload = {
        "firstName": lead.nombre,
        "lastName": "",
        "type": "LEAD", #Para que los leads entren en el flow
        "customAttributes": {
            "prop_link": lead.propiedad['link'],
            "prop_precio": str(lead.propiedad['precio']),
            "prop_ubicacion": lead.propiedad['ubicacion'],
            "prop_titulo": lead.propiedad['titulo'],
            "contacted": False,
            "fuente": lead.fuente,
            "asesor_name": lead.asesor['name'],
            "asesor_phone": lead.asesor['phone'],
            "fecha_lead": lead.fecha_lead
        },
        "contactInformation": {},
        "tags": "Seguimientos"
    }
    if lead.telefono != '' and lead.telefono != None:
        payload["contactInformation"]['phone']= [{
            "number": lead.telefono
        }]
    if lead.email != '' and lead.email != None:
        payload["contactInformation"]['email'] = [{
            "address": lead.email
        }]
    print("PAYLOAD:")
    print(json.dumps(payload, indent=4))

    try:
        res = requests.post(API_URL, json=payload, headers=HEADERS)
        if not res.ok:
            logger.error("Error en la request: " + str(res.status_code))
            logger.error(res.json())
            logger.error(payload)
            return False
    except Exception as e:
        logger.error("Error cargando lead"+str(e))
        return False

    logger.success("Lead cargada correctamente")
    return True
