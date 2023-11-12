from time import gmtime, strftime
from dotenv import load_dotenv
from datetime import datetime
import requests
import json
import os
from enum import IntEnum

from src.logger import Logger
from src.sheets import Sheet

#mis propiedades:
#https://propiedades.com/api/v3/property/MyProperties

#En este archivo tenemos todas las propieades previamente extraidas
with open("src/propiedades_com/properties_obj.json") as f:
    PROPS = json.load(f)

load_dotenv()

logger = Logger("propiedades.com")

DATE_FORMAT = "%d/%m/%Y"
API_URL = "https://ggcmh0sw5f.execute-api.us-east-2.amazonaws.com"
PARAMS_FILE = os.path.dirname(os.path.realpath(__file__)) + "/params.json"
with open(PARAMS_FILE, "r") as f:
    HEADERS = json.load(f)

#Lista de estados posibles de un lead
#Los tomamos de la pagina
class Status(IntEnum):
    CONTACTADO = 2,
    CALIFICADO = 3,
    EN_PROCESO = 4,
    CONVERTIDO = 5,
    CERRADA = 6

if (not os.path.exists(PARAMS_FILE)):
    logger.error("El archivo params.json no existe")
    with open(PARAMS_FILE, "a") as f:
        json.dump({
            "Authorization": "",
            "x-api-key": ""
        }, f, indent=4)

def get_data(url) -> object:
    res = requests.get(url, headers=HEADERS)
    if (res.status_code != 200):
        if (res.status_code == 401):
            logger.error("El token de acceso expiro")
            #get_access_token(USERNAME, PASSWORD)
            return None
        else: 
            logger.error(res.status_code)
            logger.error(res.text)
            return None
    return res.json()

def get_lead_property(property_id: str):
    if property_id not in PROPS:
        logger.error(f"No se encontro la propiedad con id {property_id}")
        return {
            "id": "",
            "titulo": "",
            "link": "",
            "precio": "",
            "ubicacion": "",
            "tipo": "",
            "municipio": ""
        }
    data = PROPS[property_id]

    return {
        "id": data["id"],
        "titulo": data["short_address"],
        "link": data["url"],
        "precio": data["price"],
        "ubicacion": data["address"],
        "tipo": data["type_children_string"],
        "municipio": data["municipality"]
    }

# Takes the JSON object getted from the API and extract the usable information.
def extract_lead_info(data: object) -> object:
    prop = get_lead_property(str(data["property_id"]))

    prop["address"] = data["address"]
    if prop["titulo"] == "": prop["titulo"] = prop["address"]

    lead_info = {
		"fuente": "Propiedades.com",
        "fecha_lead": datetime.strptime(data["updated_at"], '%Y-%m-%d').strftime(DATE_FORMAT),
        "id": data["id"],
		"fecha": strftime(DATE_FORMAT, gmtime()),
		"nombre": data["name"],
		"link": "",
		"telefono": data["phone"],
		#"telefono_2": data["phone_list"][1],
		"email": data["email"],
		"propiedad": prop,
		"busquedas": {
            "zonas": "",
            "tipo": "",
            "presupuesto": "",
            "cantidad_anuncios": "",
            "contactos": "",
            "inicio_busqueda": "",
            "total_area": "",
            "covered_area": "",
            "banios": "",
            "recamaras": "",
        }
	}

    return lead_info

def change_status(lead_id, status:Status):
    logger.debug(f"Marcando como {status.name} al lead {lead_id}")
    
    url = f"{API_URL}/prod/leads/status"

    req = {
        "lead_id": lead_id,
        "status": status,
        "country": "MX"
    }

    res = requests.put(url, headers=HEADERS, json=req)
    if (res.status_code != 200):
        logger.error(res.status_code)
        logger.error(res.text)
    else:
        logger.success(f"Se marco a lead {lead_id} como {status.name}")

def main():
    sheet = Sheet(logger)
    headers = sheet.get("A2:Z2")[0]
    logger.debug(f"Extrayendo leads")

    first = True
    page = 1
    url = f"{API_URL}/prod/get/leads?page={page}&country=MX"

    while first == True or page != None:
        leads = []
        logger.debug(url)

        data = get_data(url)["leads"]
        if data == None: return None

        page = data["page"]["next_page"]
        url = f"{API_URL}/prod/get/leads?page={page}&country=MX"
        total = data["page"]["items"]
        logger.debug(f"total: {total}")

        for raw_lead in data["properties"]:
            if raw_lead["status"] == Status.CONTACTADO:
                logger.debug("lead ya contactado, ignoramos")
                continue

            lead = extract_lead_info(raw_lead)
            logger.debug(lead)

            lead["message"] = ""
            change_status(lead["id"], Status.CONTACTADO)

            row_lead = sheet.map_lead(lead, headers)
            leads.append(row_lead)

        sheet.write(leads)

        logger.debug(f"Leads leidos: {len(leads)}")
        first = False

    logger.success(f"Se encontraron {len(leads)} nuevos Leads")

def change_all_status(status: Status):
    first = True
    page = 1
    url = f"{API_URL}/prod/get/leads?page={page}&country=MX"

    while first == True or url != None:
        leads = []
        logger.debug(url)

        data = get_data(url)["leads"]
        if data == None: return None

        page = data["page"]["next_page"]
        url = f"{API_URL}/prod/get/leads?page={page}&country=MX"
        total = data["page"]["items"]
        logger.debug(f"total: {total}")

        for raw_lead in data["properties"]:
            if raw_lead["status"] == status:
                logger.debug(f"lead ya  se encuentra en estado {status.name} contactado, ignoramos")
                continue

            change_status(raw_lead["id"], status)

        logger.debug(f"Leads leidos: {len(leads)}")
        first = False

    logger.success(f"Se encontraron {len(leads)} nuevos Leads")

if __name__ == "__main__":
    main()
    #change_all_status(Status.EN_PROCESO)