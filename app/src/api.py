import os
import requests

from src.logger import Logger
from src.lead import Lead

API_PORT = os.getenv("API_PORT")
API_HOST = os.getenv("API_HOST")
url = f"http://{API_HOST}:{API_PORT}/communication"

def new_communication(logger: Logger, lead: Lead) -> tuple[bool, Lead | None]:
    res = requests.post(url, json=lead.__dict__)
    if not res.ok:
        logger.error("Error en la peticion: "+str(res.json()))
        return False, None
    logger.success("Communication cargada correctamente")

    json = res.json()
    lead_data = json["data"]
    is_new = json["is_new"]
    lead.set_args(lead_data)
    return is_new, lead
