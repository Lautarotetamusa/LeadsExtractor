from time import gmtime, strftime
import urllib.parse
import json

import src.infobip as infobip
from src.logger import Logger
from src.sheets import Sheet

logger = Logger("Asesor")
sheet = Sheet(logger, "mapping.json")

ASESOR_i = 0
def next_asesor():
    global ASESOR_i
    rows = sheet.get('Asesores!A2:C25')
    activos = [row for row in rows if row[2] == "Activo"]
    print("Activos: ", activos)
    ASESOR_i += 1
    ASESOR_i %= len(activos)
    
    asesor = {
        "name":  activos[ASESOR_i][0],
        "phone": activos[ASESOR_i][1]
    }
    logger.debug("Asesor asignado: "+str(asesor['name']))
    return asesor

#Si la persona no existe en infobip hacemos round robin para otorgarle un asesor
#Si ya existe devolvemos el asesor que ya tiene
def assign_asesor(phone, fuente) -> dict:
    json_filter = {"#contains": {"contactInformation": {"phone": {"number": phone}}}}
    filtro = urllib.parse.quote(json.dumps(json_filter))
    person = infobip.search_person(logger, filtro)
    
    if person == None:
        logger.debug(f"Un nuevo lead se comunico via: {fuente}")
        return next_asesor()
    else:
        logger.debug(f"Un lead existente se volvio a comunicar via: {fuente}")
        asesor = {
            'name': person.get('customAttributes', {}).get('asesor_name', None),
            'phone': person.get('customAttributes', {}).get('asesor_phone', None)
        }
        return asesor
