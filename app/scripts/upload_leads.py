# Cargar leads desde un csv a la tabla
# USAGE 
# python -m venv .venv
# python scripts/upload_leads.py <file_path>

import os
import sys
sys.path.append('.')

import csv

from src.lead import Lead
from src.logger import Logger
from src.api import new_communication

if __name__ == "__main__":
    logger = Logger("CSV to leads")
    if (len(sys.argv) < 2):
        logger.error("Numero incorrecto de argumentos")
        exit(1)

    file_path = sys.argv[1]
    [file_name, format] = os.path.basename(file_path).split('.')
    with open(file_path, 'r', encoding='latin1') as file:
        reader = csv.DictReader(file)
        for row in reader:
            lead = Lead()
            lead.set_args({
                "nombre": row["Nombre"],
                "telefono": row["Telefono"],
                "email": row["Email"],
                "fuente": "ivr",
                "utm": {
                    "utm_campaign": file_name 
                }
            })
            lead.print()
            new_communication(logger, lead)
