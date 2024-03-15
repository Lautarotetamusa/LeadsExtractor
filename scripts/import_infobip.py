#Este script sirve para volcar toda la informacion del sheets en el CDP de Infobip
#Actualizado el día - 15 marzo 2024

import sys
sys.path.append('.')

from src.sheets import Sheet
from src.logger import Logger
import src.infobip as infobip
from src.numbers import parse_number

THREADS = 16

if __name__ == "__main__":
    logger = Logger("SHEETS into INFOBIP")
    sheet = Sheet(logger, "mapping.json")

    headers = sheet.get("A2:Z2")[0]

    init = 3000
    stop = init + THREADS
    stop = 3500

    rows = sheet.get(f"A{init}:Z{stop}")

    news = 0
    duplicateds = 0
    for row in rows:
        lead = sheet.get_lead(row, headers)
        telefono = parse_number(logger, lead.telefono)
        if not telefono:
            telefono = parse_number(logger, lead.telefono, "MX")
        if not telefono:
            logger.debug("No se pudo parsear el numero de telefono")
            continue
        lead.telefono = telefono

        load = infobip.create_person(logger, lead)
        if not load: #La persona no se cargó
            duplicateds += 1
            logger.debug("Se encontro un Lead duplicado")
        else:
            news += 1

        logger.debug(f"news: {news}. duplicateds: {duplicateds}")
