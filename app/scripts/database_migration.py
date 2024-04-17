#Este script migra todo el google sheets a la base de datos
#Actualizado el día - 16 abril 2024

import sys
import os
sys.path.append('.')

from src.sheets import Sheet
from src.logger import Logger
import src.api as api

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT != None, "DATE_FORMAT is not seted"

THREADS = 1000

if __name__ == "__main__":
    logger = Logger("SHEETS into DATABASE")
    sheet = Sheet(logger, "mapping.json")

    headers = sheet.get("A2:Z2")[0]

    init = 3
    while init < 2823: 
        stop = init + THREADS
    
        print(init, stop)
        rows = sheet.get(f"BD-General!A{init}:Z{stop}")
        print(len(rows))

        for row in rows:
            lead = sheet.get_lead(row, headers)
            api.new_communication(logger, lead)
