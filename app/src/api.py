import os
import requests
from dotenv import load_dotenv

from src.logger import Logger
from src.lead import Lead

load_dotenv()
API_PROTOCOL = os.getenv("API_PROTOCOL")
API_PORT = os.getenv("API_PORT")
API_HOST = os.getenv("API_HOST")
assert API_PORT is not None, "Error: 'API_PORT env variable not set'"
assert API_HOST is not None, "Error: 'API_HOST env variable not set'"
assert API_PROTOCOL is not None, "Error: 'API_PROTOCOL env variable not set'"


def assign_asesor(logger: Logger, lead: Lead) -> tuple[bool, Lead | None]:
    url = f"{API_PROTOCOL}://{API_HOST}:{API_PORT}/assign"
    res = requests.post(url, json=lead.__dict__)
    if not res.ok:
        logger.error("Error en la peticion: "+str(res.json()))
        return False, None
    logger.success("Asesor obtenido correctamente")

    json = res.json()
    lead_data = json["data"]
    is_new = json["is_new"]
    lead.set_args(lead_data)
    return is_new, lead


def new_communication(logger: Logger, lead: Lead) -> tuple[bool, Lead | None]:
    url = f"{API_PROTOCOL}://{API_HOST}:{API_PORT}/communication"
    res = requests.post(url, json=lead.__dict__)
    if not res.ok:
        logger.error("Error en la peticion: "+str(res.text))
        return False, None
    logger.success("Communication cargada correctamente")

    json = res.json()
    lead_data = json["data"]
    is_new = lead_data["is_new"]
    lead.set_args(lead_data)
    return is_new, lead


def get_communications(logger: Logger, date: str, is_new: bool | None=None) -> list[Lead]:
    url = f"{API_PROTOCOL}://{API_HOST}:{API_PORT}/communications?date={date}"

    if is_new is not None:
        url += f"&is_new={'true' if is_new else 'false'}"

    res = requests.get(url)
    if not res.ok:
        logger.error("Error en la peticion: "+str(res.json()))
        return []

    leads = []
    for data in res.json()["data"]:
        lead = Lead()
        lead.set_args(data)
        lead.set_asesor(data["asesor"])
        lead.set_propiedad(data["propiedad"])
        lead.set_busquedas(data["busquedas"])
        leads.append(lead)

    return leads

def update_publication(property_id: int, portal: str, status: str, publication_id: str | None = None):
    url = f"{API_PROTOCOL}://{API_HOST}:{API_PORT}/property/{property_id}/publications/{portal}"
    payload = {
        "status": status,
    }
    if publication_id is not None:
        payload["publication_id"] = publication_id

    res = requests.put(url, json=payload)
    print(res.json())
