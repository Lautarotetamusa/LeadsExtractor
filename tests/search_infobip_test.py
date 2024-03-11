import sys
import json
import urllib.parse
sys.path.append('..')

from src.lead import Lead
from src.logger import Logger
import src.infobip as infobip

logger = Logger("infobip search test")

#Si la persona no existe en infobip hacemos round robin para otorgarle un asesor
#Si ya existe devolvemos el asesor que ya tiene
def assign_asesor(lead: Lead):
    json_filter = {"#contains": {"contactInformation": {"phone": {"number": lead.telefono}}, "email": lead.email}}
    filtro = urllib.parse.quote(json.dumps(json_filter))
    person = infobip.search_person(logger, filtro)
    is_new = person == None
    return is_new, person

if __name__ == "__main__":
    lead = Lead()
    lead.telefono = "+523335708376"
    lead.email = "lenalizan@aol.com"

    is_new, person = assign_asesor(lead)
    print(json.dumps(person, indent=4))
    print(is_new)
