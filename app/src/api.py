from datetime import date
import os
import requests

from src.logger import Logger
from src.lead import Lead

API_PORT = os.getenv("API_PORT")
API_HOST = os.getenv("API_HOST")

def assign_asesor(logger: Logger, lead: Lead) -> tuple[bool, Lead | None]:
    url = f"http://{API_HOST}:{API_PORT}/assign"
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

def new_communication(logger: Logger, lead: Lead) -> tuple[bool, Lead | None]:
    url = f"https://{API_HOST}:{API_PORT}/communication"
    res = requests.post(url, json=lead.__dict__)
    if not res.ok:
        logger.error("Error en la peticion: "+str(res.text))
        return False, None
    logger.success("Communication cargada correctamente")

    json = res.json()
    lead_data = json["data"]
    is_new = json["is_new"]
    lead.set_args(lead_data)
    return is_new, lead

def get_communications(logger: Logger, date: str, is_new: bool | None=None) -> list[Lead]:
    url = f"https://{API_HOST}:{API_PORT}/communications?date={date}"

    if is_new != None:
        url += f"&is_new={'true' if is_new else 'false'}"
    print(url)

    res = requests.get(url)
    if not res.ok:
        logger.error("Error en la peticion: "+str(res.json()))
        return []

    print(res)
    leads = []
    for data in res.json()["data"]:
        lead = Lead()
        lead.set_args(data) 
        lead.set_asesor(data["asesor"])
        lead.set_propiedad(data["propiedad"])
        lead.set_busquedas(data["busquedas"])
        leads.append(lead)

    return leads
