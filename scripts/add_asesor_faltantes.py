#Este script asigna asesores a todos los leads de Infobip que todavia no lo tengan
#Actualizado el día - 15 marzo 2024

import sys
import json
import urllib.parse

sys.path.append('.')

from src.whatsapp import Whatsapp
from src.logger import Logger
from src.logger import Logger
from src.sheets import Sheet
import src.infobip as infobip

if __name__ == "__main__":
    logger = Logger("Update persons")
    sheet = Sheet(logger, "mapping.json")
    wpp = Whatsapp()

    #Buscamos las personas que no tengan numero o nombre del asesor asignado
    json_filter = {"#or": [
        {"customAttributes": {"asesor_phone": None}}, 
        {"customAttributes": {"asesor_name": None}}
    ]}
    filtro = urllib.parse.quote(json.dumps(json_filter))
    
    rows = sheet.get('Asesores!A2:C25')
    activos = [row for row in rows if row[2] == "Activo"]
    logger.debug(f"Activos: {activos}")

    asesor_i = 0
    for page in infobip.get_persons(logger, filtro):
        for person in page:
            asesor = {
                "name":  activos[asesor_i][0],
                "phone": activos[asesor_i][1]
            }
            infobip.update_person(logger, person['id'], {
                **person,
                "customAttributes": {
                    "asesor_name":  asesor['name'],
                    "asesor_phone": asesor['phone']
                }
            }) 
            lead = infobip.infobip2lead(person)
            wpp.send_msg_asesor(asesor['phone'], lead)
            asesor_i += 1
            asesor_i %= len(activos)
