#Reasignar asesores que ya no estan mas activos
#Actualizado el día - 23 marzo 2024

import sys
import json
import urllib.parse

sys.path.append('.')

from src.logger import Logger
from src.sheets import Sheet
import src.infobip as infobip

if __name__ == "__main__":
    logger = Logger("Update persons")
    sheet = Sheet(logger, "mapping.json")

    rows = sheet.get('Asesores!A2:C25')
    asesores = [row for row in rows if row[2] == "Activo"]
    asesor_i = 0
    logger.debug(f"Activos: {asesores}")

    #Buscamos las personas que no tengan numero o nombre del asesor asignado
    json_filter = {"customAttributes": {"asesor_name": "Juan Sanchez"}}
    filtro = urllib.parse.quote(json.dumps(json_filter))

    for page in infobip.get_persons(logger, filtro):
        for person in page:
            infobip.update_person(logger, person['id'], {
                **person,
                "customAttributes": {
                    **person['customAttributes'],
                    'asesor_name': asesores[asesor_i][0],
                    'asesor_phone': asesores[asesor_i][1],
                }
            }) 

            asesor_i += 1
            asesor_i %= len(asesores)
