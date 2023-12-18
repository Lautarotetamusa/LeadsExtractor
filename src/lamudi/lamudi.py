from time import gmtime, strftime
from datetime import datetime
from dotenv import load_dotenv
#import requests
import json
import uuid
import os

from src.message import generate_mensage
from src.logger import Logger
from src.sheets import Sheet
from src.make_requests import Request

load_dotenv()
logger = Logger("lamudi.com")

API_URL = "https://api.proppit.com"
DATE_FORMAT = "%d/%m/%Y"
USERNAME=os.getenv('LAMUDI_USERNAME')
PASSWORD=os.getenv('LAMUDI_PASSWORD')
PARAMS_FILE = os.path.dirname(os.path.realpath(__file__)) + "/params.json"

if (not os.path.exists(PARAMS_FILE)):
    logger.error("El archivo params.json no existe")
    with open(PARAMS_FILE, "a") as f:
        json.dump({
            "authToken": ""
        }, f, indent=4)

with open(PARAMS_FILE, "r") as f:
    COOKIES = json.load(f)

def login():
    logger.debug("Iniciando sesion")
    login_url = f"{API_URL}/login"

    data = {
        "email": USERNAME,
        "password": PASSWORD
    }
    res = request.make(login_url, 'POST', json=data)
    
    request.cookies = {
        "authToken": res.cookies["authToken"]
    }
    with open(PARAMS_FILE, "w") as f:
        json.dump(request.cookies, f, indent=4)
    logger.success("Sesion iniciada con exito")

request = Request(COOKIES, None, logger, login)

def send_email_message(lead_id, msg):
    logger.debug(f"Enviando mensaje por email a lead {lead_id}")
    msg_url = f"{API_URL}/leads/{lead_id}/emails"

    data = {
        "leadId": lead_id,
        "message": msg
    }
    request.make(msg_url, 'POST', json=data)
    logger.success(f"Mensaje enviado correctamente a lead {lead_id}")

def send_message(lead_id, msg):
    logger.debug(f"Enviando mensaje a lead {lead_id}")
    #Tenemos que pasarle un id del mensaje, lo generamos nosotros con uuid()
    msg_id = str(uuid.uuid4())
    msg_url = f"{API_URL}/leads/{lead_id}/notes/{msg_id}"

    data = {
        "id": msg_id,
        "message": msg
    }
    request.make(msg_url, 'POST', json=data)
    logger.success(f"Mensaje enviado correctamente a lead {lead_id}")

def make_contacted(lead_id):
    logger.debug(f"Marcando como contactacto a lead {lead_id}")
    read_url = f"{API_URL}/leads/{lead_id}/status"

    data = {
        "status": "contacted"
    }
    request.make(read_url, 'PUT', json=data)
    logger.success(f"Se contacto correctamente a lead {lead_id}")

def get_lead_property(lead_id):
    property_url = f"{API_URL}/leads/{lead_id}/properties"
    props = request.make(property_url).json().get("data")

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
        "fecha_lead": datetime.strptime(lead["lastActivity"], "%Y-%m-%dT%H:%M:%SZ").strftime(DATE_FORMAT),
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

def get_leads(max=-1, status="new lead"):
    page = 1
    limit = 25
    url = f"{API_URL}/leads?_limit={limit}&_order=desc&_page={page}&_sort=lastActivity&status={status}"
    logger.debug(f"Extrayendo leads")

    data = request.make(url, 'GET').json()["data"]

    total = data["totalFilteredRows"]
    logger.debug(f"total: {total}")

    leads = []
    while (len(leads) < total and (len(leads) < max or max == -1)):
        logger.debug(f"len: {len(leads)}")
        leads += data["rows"]
        page += 1

        url = f"{API_URL}/leads?_limit={limit}&_order=desc&_page={page}&_sort=lastActivity&status={status}"
        logger.debug(url)
        data = request.make(url, 'GET').json()["data"]

    logger.success(f"Se encontraron {len(leads)} nuevos Leads")
    return leads

# Esta es la funcion que se ejecuta todo el tiempo
def main():
    leads = get_leads()
    sheet = Sheet(logger, "mapping.json")
    headers = sheet.get("A2:Z2")[0]

    with open('messages/gmail.html', 'r') as f:
        gmail_spin = f.read()
    with open('messages/gmail_subject.html', 'r') as f:
        gmail_subject = f.read()

    # Adjuntar archivo PDF
    with open('messages/attachment.pdf', 'rb') as pdf_file:
        attachment = MIMEApplication(pdf_file.read(), _subtype="pdf")
        attachment.add_header('Content-Disposition', 'attachment', 
              filename='Bienvenido a Rebora! Seguridad, Confort y Placer - Casas de gran diseño y alta calidad.pdf'
        )

    leads_info = []
    for lead_res in leads:
        lead = get_lead_info(lead_res)
        if lead['email'] != '':
            if lead["propiedad"]["ubicacion"] == "":
                lead["propiedad"]["ubicacion"] = "que consultaste"
            else:
                lead["propiedad"]["ubicacion"] = "ubicada en " + lead["propiedad"]["ubicacion"]

        gmail_msg = generate_mensage(lead, gmail_spin)
        subject = generate_mensage(lead, gmail_subject)
        gmail.send_message(gmail_msg, subject, lead["email"], attachment)

        msg = generate_mensage(lead)
        send_message(lead["id"], msg)
        lead["message"] = msg.replace('\n', '')
        make_contacted(lead["id"])

        leads_info.append(lead)

        #Save the lead in the sheet
        row_lead = sheet.map_lead(lead, headers)
        sheet.write([row_lead])

#Esta funcion ejecuta la primera corrida
def first_run():
    leads = get_leads(status="")
    sheet = Sheet(logger, "mapping.json")
    headers = sheet.get("A2:Z2")[0]

    for lead_res in leads:
        lead = get_lead_info(lead_res)
        msg = generate_mensage(lead)
        lead["message"] = msg.replace('\n', '')

        #Save the lead in the sheet
        row_lead = sheet.map_lead(lead, headers)
        sheet.write([row_lead])

if __name__ == "__main__":
    from sheets import Gmail
    gmail = Gmail({'email': 'Prueba'}, logger)
    gmail.send_message('mensaje de prueba', 'lautarotetamusa@gmail.com', 'Subject de prueba')
