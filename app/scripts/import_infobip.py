#Este script sirve para volcar toda la informacion del sheets en el CDP de Infobip
#Actualizado el día - 22 marzo 2024

from datetime import datetime
import sys
import os
sys.path.append('.')

from src.sheets import Sheet
from src.logger import Logger
import src.infobip as infobip
from src.asesor import assign_asesor
from src.numbers import parse_number

THREADS = 1000
DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT != None, "DATE_FORMAT is not seted"

if __name__ == "__main__":
    logger = Logger("SHEETS into INFOBIP")
    sheet = Sheet(logger, "mapping.json")

    headers = sheet.get("A2:Z2")[0]
    asesores = {
        "Brenda Diaz":	"523313420733",
        "Aldo Salcido":	"523322563353",
        "Maggie Escobedo":	"523314299454",
        "Diego Rubio":	"523317186543",
        "Onder Sotomayor":	"523318940377",
        "Juan Sanchez": "523317186543",
        "Juan Sánchez": "523317186543"
    }

    init = 3

    while init < 2823: 
        stop = init + THREADS
    
        print(init, stop)
        rows = sheet.get(f"2da-corrida!A{init}:Z{stop}")
        init += THREADS
        print(len(rows))

        news = 0
        duplicateds = 0

        leads = []

        for row in rows:
            lead = sheet.get_lead(row, headers)

            telefono = parse_number(logger, lead.telefono)
            if not telefono:
                telefono = parse_number(logger, lead.telefono, "MX")
            if not telefono:
                logger.debug("No se pudo parsear el numero de telefono")
                continue

            lead.telefono = telefono
            try:
                lead.fecha_lead = datetime.strptime(lead.fecha_lead, "%d/%m/%Y").strftime(DATE_FORMAT)
            except Exception as e:
                logger.debug("Fecha en el correcto formato")
                pass
    
            if lead.asesor['phone'] == '':
                if lead.asesor['name'] != '':
                    if lead.asesor['name'] in asesores:
                        lead.asesor['phone'] = asesores[lead.asesor['name']]
                    else:
                        is_new, lead = assign_asesor(lead)
                else:
                    print("sin asesor")
                    is_new, lead = assign_asesor(lead)

            lead.validate()
            leads.append(lead)

        results = infobip.update_persons(logger, leads)

        logger.debug(results)
        if type(results) is bool:
            logger.debug("Ocurrio un error inesperado, vamos a crear las personas")
            exit(1)
            #TODO: Crear todas las personas
        new_phones = []
        for result in results:
            if 'errors' in result: #Buscamos los leads que no se pudieron cargar
                phone = result.get('query', {}).get('identifier')
                new_phones.append(phone)
                logger.debug(f"El lead con telefono no existe: {phone}, cargando")
                #infobip.create_persons(logger, leads)
            else:
                logger.debug("Todas las personas se actualizaron correctamente")
            #for phone in new_phones:
            #    for lead in leads:
            #        if lead.telefono.replace('+', '') == phone:
            #            infobip.create_person(logger, lead)
            #            break
