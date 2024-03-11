import sys
sys.path.append('..')

from src.logger import Logger
import src.infobip as infobip

if __name__ == "__main__":
    logger = Logger("Update persons")

    persons = infobip.get_all_person(logger)

    for person in persons:
        infobip.update_person(logger, person['id'], {
            **person,
            "type": "LEAD"
        }) 
