#Cambiar el tipo de lead de todas las personas de Infobip
#Actualizado el dia - 15 de Marzo de 2024

import sys
sys.path.append('.')

from src.logger import Logger
import src.infobip as infobip

def USAGE():
    print("USAGE python scripts/update_persons_type.py <TYPE> (customer | lead)")

types = {
    "customer": "CUSTOMER",
    "lead": "LEAD"
}

if __name__ == "__main__":
    logger = Logger("Update persons")

    if len(sys.argv) < 2:
        USAGE()
        exit(1)
    
    person_type = str(sys.argv[1])

    if person_type not in types:
        USAGE()
        exit(1)

    persons = infobip.get_all_person(logger)

    for person in persons:
        infobip.update_person(logger, person['id'], {
            **person,
            "type": person_type
        }) 
