#Asignar fecha lead a los leads que no tienen
#Actualizado el día - 22 marzo 2024

from datetime import datetime
import os
import sys
import json
import urllib.parse

sys.path.append('.')

from src.logger import Logger
from src.logger import Logger
import src.infobip as infobip

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT != None, "DATE_FORMAT is not seted"

if __name__ == "__main__":
    logger = Logger("Update persons")

    #Buscamos las personas que no tengan numero o nombre del asesor asignado
    json_filter = {"customAttributes": {"fecha_lead": None}}
    filtro = urllib.parse.quote(json.dumps(json_filter))

    for page in infobip.get_persons(logger, filtro):
        for person in page:
            fecha_lead = datetime.strptime(person.get('createdAt', ''),"%Y-%m-%dT%H:%M:%S").strftime(DATE_FORMAT)
            infobip.update_person(logger, person['id'], {
                **person,
                "customAttributes": {
                    **person['customAttributes'],
                    "fecha_lead": fecha_lead 
                }
            }) 
