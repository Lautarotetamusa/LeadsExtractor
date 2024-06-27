# Volcar todo infobip en DB através de la api

import sys
import json
import urllib.parse
sys.path.append('.')

from src.logger import Logger
import src.infobip as infobip
import src.api as api

if __name__ == "__main__":
    logger = Logger("Infobip")

    json_filter = {}
    filter = urllib.parse.quote(json.dumps(json_filter))

    count = 0
    for person in infobip.get_persons(logger, filter):
        count += 1;
        lead = infobip.infobip2lead(person)
        is_new, lead = api.new_communication(logger, lead)

    print(count)
