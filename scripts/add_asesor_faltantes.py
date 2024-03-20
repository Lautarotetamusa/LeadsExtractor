#Este script asigna asesores a todos los leads de Infobip que todavia no lo tengan
#Actualizado el día - 15 marzo 2024

import sys
import json
import urllib.parse
sys.path.append('.')

from src.logger import Logger
from src.logger import Logger
from src.sheets import Sheet
import src.infobip as infobip

if __name__ == "__main__":
    logger = Logger("Update persons")
    sheet = Sheet(logger, "mapping.json")

    json_filter = {"customAttributes": {"asesor_phone": None}}
    filtro = urllib.parse.quote(json.dumps(json_filter))
    persons = infobip.get_all_with_filter(logger, filtro)
    logger.debug(f"Personas sin asesor: {len(persons)}")
    
    rows = sheet.get('Asesores!A2:C25')
    activos = [row for row in rows if row[2] == "Activo"]
    logger.debug(f"Activos: {activos}")

    asesor_i = 0
    for person in persons:
        infobip.update_person(logger, person['id'], {
            **person,
            "customAttributes": {
                "asesor_name":  activos[asesor_i][0],
                "asesor_phone": activos[asesor_i][1]
            }
        }) 
        asesor_i += 1
        asesor_i %= len(activos)
