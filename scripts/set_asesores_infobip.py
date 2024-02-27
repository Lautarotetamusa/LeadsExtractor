from src.sheets import Logger 
from src.infobip import get_all_person, update_person

import requests

from dotenv import load_dotenv
import os

load_dotenv()
API_KEY = os.getenv("INFOBIP_APIKEY")
API_URL = os.getenv("INFOBIP_APIURL")
if API_KEY == None or API_URL == None:
    print("ERROR: INFOBIP_APIKEY or INFOBIP_APIURL env variables not set")
    exit(1)
HEADERS = {
    'Authorization': API_KEY
}

logger = Logger("Infobip import")

ASESORES = [
    {
        "id": 0,
        "name": "Brenda Diaz",
        "phone": "+523313420733"
    },
    {
        "id": 1,
        "name": "Juan Sanchez",
        "phone": "+523317186543"
    },
    {
        "id": 2,
        "name": "Aldo Salcido",
        "phone": "+523322563353"
    } 
]

if __name__ == "__main__":
    asesor = 0
    logger.debug("Buscando personas")

    page = 1

    #Filtro con url encoder, con este filtro buscamos solamente las personas sin asesor asignado
    #filter = {"customAttributes": {"asesor_name":  None}}
    filter = "%7B%22customAttributes%22%3A%20%7B%22asesor_name%22%3A%20%20null%7D%7D"

    while True:
        logger.debug(f"GET {API_URL}?page={page}")
        res = requests.get(f"{API_URL}?filter={filter}&page={page}", headers=HEADERS)
        if not res.ok:
            logger.error("Error en la request: " + str(res.status_code))
            logger.error(res.json())
        else: 
            data = res.json()
            if len(data.get('persons', [])) == 0:
                logger.debug("Se encontro un pagina vacia, terminando")
                break
           
            for person in data.get("persons", []):
                id = person.get("id", None)
                if id == None:
                    logger.error("La persona no tiene id")
                    continue

                person["customAttributes"]["asesor_name"]  = ASESORES[asesor]["name"]
                person["customAttributes"]["asesor_phone"] = ASESORES[asesor]["phone"] 

                update_person(logger, id, person)

                asesor += 1
                asesor %= len(ASESORES)
        page += 1
