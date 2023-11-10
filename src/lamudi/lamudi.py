import __init__

from time import gmtime, strftime
import requests
import json
import uuid

from message import generate_mensage
from logger import Logger
from sheets import Sheet

api_url = "https://api.proppit.com"

auth_token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VyIjp7ImlkIjoiZmFmYjA4NTEtZjlkOC00MGNkLTg0N2QtZDBjNzA2YzE2YjkxIiwicHVibGlzaGVySWQiOiJkODk3Y2RjYS03NjliLTQyNWItYmM5My0yYzNjM2Y4YjU2MjkiLCJhY2NlcHRlZFRlcm1zQW5kQ29uZGl0aW9ucyI6dHJ1ZSwiZW1haWwiOiJyZWJvcmFteEBnbWFpbC5jb20iLCJsYXN0UGFzc3dvcmRNb2RpZmljYXRpb25EYXRlIjoxNjg2NjA5MzQ5fSwiZXhwIjoxNjk5OTIxNDcxfQ.b5TLPFlEzti4XZRDzPYePnHAMO_GXPC1b9rjzO351v0"

cookies = {
    "authToken": auth_token
}

logger = Logger("lamudi.com")

def send_email_message(lead_id, msg):
    logger.debug(f"Enviando mensaje por email a lead {lead_id}")
    msg_url = f"{api_url}/leads/{lead_id}/emails"

    data = {
        "leadId": lead_id,
        "message": msg
    }

    res = requests.post(msg_url, cookies=cookies, json=data)
    if (res.status_code != 201):
        logger.error(res.status_code)
        logger.error(res.text)
        return None

    logger.success(f"Mensaje enviado correctamente a lead {lead_id}")

def send_message(lead_id, msg):
    logger.debug(f"Enviando mensaje a lead {lead_id}")
    #Tenemos que pasarle un id del mensaje, lo generamos nosotros con uuid()
    msg_id = str(uuid.uuid4())
    msg_url = f"{api_url}/leads/{lead_id}/notes/{msg_id}"

    data = {
        "id": msg_id,
        "message": msg
    }

    res = requests.post(msg_url, cookies=cookies, json=data)
    if (res.status_code != 201):
        logger.error(res.status_code)
        logger.error(res.text)
        return None

    logger.success(f"Mensaje enviado correctamente a lead {lead_id}")

def make_contacted(lead_id):
    logger.debug(f"Marcando como contactacto a lead {lead_id}")
    read_url = f"{api_url}/leads/{lead_id}/status"

    data = {
        "status": "contacted"
    }

    res = requests.put(read_url, cookies=cookies, json=data)
    if (res.status_code != 200):
        logger.error(res.status_code)
        logger.error(res.text)
        return None

    logger.success(f"Se contacto correctamente a lead {lead_id}")

def get_lead_property(lead_id):
    property_url = f"{api_url}/leads/{lead_id}/properties"

    res = requests.get(property_url, cookies=cookies)
    if (res.status_code != 200):
        logger.error(res.status_code)
        logger.error(res.text)
        return None

    props = res.json()["data"]
    formatted_props = []
    for p in props:
        formatted_props.append({
            "titulo": p["title"],
            "id": p['id'],
            "link": f"https://www.lamudi.com.mx/detalle/{p['id']}",
            "precio": p["price"]["amount"],
            "ubicacion": p["address"], #Direccion completa
            "tipo": p["propertyType"],
            "municipio": p["geoLevels"][0]["name"] if len(p["geoLevels"]) > 0 else "" #Solamente el municipio, lo usamos para generar el mensaje
        })

    return formatted_props
        
def get_lead_info(lead):
    lead_info = {
        "id": lead["id"],
        "fuente": "Lamudi",
        "fecha": strftime('%d/%m/%Y', gmtime()),
        "nombre": lead["name"],
        "link": f"https://proppit.com/leads/{lead['id']}",
        "telefono": lead['phone'],
        "telefono_2": "",
        "email": lead['email'],
        "propiedad": get_lead_property(lead["id"])[0],
        "busquedas": {
            "zonas": "",
            "tipo": "",
            "total_area": "",
            "covered_area": "",
            "banios": "",
            "recamaras": "",
            "presupuesto": "",
            "cantidad_anuncios": "",
            "contactos": "",
            "inicio_busqueda": "" 
        }
    }
    return lead_info

def get_leads(max=-1):
    status = "new lead" #Filtramos solamente los leads nuevos
    page = 1
    limit = 25
    url = f"{api_url}/leads?_limit={limit}&_order=desc&_page={page}&_sort=lastActivity&status={status}"
    logger.debug(f"Extrayendo leads")
    res = requests.get(url, cookies=cookies)

    if (res.status_code != 200):
        logger.error(res.status_code)
        logger.error(res.text)
        return None
    data = res.json()["data"]

    total = data["totalFilteredRows"]
    print("total: ", total)

    leads = []
    while (len(leads) < total and (len(leads) < max or max == -1)):
        logger.debug(f"len: {len(leads)}")
        leads += data["rows"]
        page += 1

        url = f"{api_url}/leads?_limit={limit}&_order=desc&_page={page}&_sort=lastActivity&status={status}"
        logger.debug(url)
        res = requests.get(url, cookies=cookies)

        if (res.status_code != 200):
            logger.error(res.status_code)
            logger.error(res.text)
            return None
        data = res.json()["data"]

    logger.success(f"Se encontraron {len(leads)} nuevos Leads")
    return leads

def main():
    leads = get_leads()
    sheet = Sheet(logger)
    headers = sheet.get("A2:Z2")[0]

    leads_info = []
    for lead_res in leads:
        lead = get_lead_info(lead_res)

        msg = generate_mensage(lead)
        send_message(lead["id"], msg)
        lead["message"] = msg
        make_contacted(lead["id"])

        leads_info.append(lead)

        #Save the lead in the sheet
        row_lead = sheet.map_lead(lead, headers)
        sheet.write([row_lead])

    with open('./lamudi.json', 'w') as f:
        json.dump(leads_info, f)

if __name__ == "__main__":
    main()
    #send_message("4156ae85-36d0-4981-9e2d-fb1c0c08f750", "Hola muy buenos dias como estas")
    #make_contacted("eab4a707-7585-4725-b447-c7f18f1d23af")

    #mandar email a uno sin email
    #send_email_message("4156ae85-36d0-4981-9e2d-fb1c0c08f750", "Hola muy buenos dias")
    #{"error":{"message":"The parameters passed to the API were invalid. Check your inputs!\n\nto parameter is missing"}}

    #mandar email a uno con email
    #send_email_message("eab4a707-7585-4725-b447-c7f18f1d23af", "Hola muy buenos dias")