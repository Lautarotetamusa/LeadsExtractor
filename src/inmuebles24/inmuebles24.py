from time import gmtime, strftime
from datetime import datetime
from dotenv import load_dotenv
import requests
import os
import json

from src.message import generate_mensage
from src.logger import Logger
from src.sheets import Gmail, Sheet
from src.make_requests import Request

from email.mime.application import MIMEApplication
load_dotenv()

#from .params import cookies, headers
PARAMS_FILE = os.path.dirname(os.path.realpath(__file__)) + "/params.json"
with open(PARAMS_FILE, "r") as f:
    HEADERS = json.load(f)

USERNAME = os.getenv("INMUEBLES24_USERNAME")
PASSWORD = os.getenv("INMUEBLES24_PASSWORD")
DATE_FORMAT = "%d/%m/%Y"
SITE_URL = "https://www.inmuebles24.com/"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
PARAMS = {
	#"resolve_captcha": "true",
	"apikey": os.getenv("ZENROWS_APIKEY"),
	"url": "",
    #"js_render": "true",
    #"antibot": "true",
    "premium_proxy": "true",
    "proxy_country": "mx",
	#"session_id": 10,
	"custom_headers": "true",
	"original_status": "true",
	"autoparse": "true"
}

logger = Logger("inmuebles24.com")
gmail = Gmail({
    "email": os.getenv("EMAIL_CONTACT"),
}, logger)
SUBJECT = os.getenv("SUBJECT") or "subject"

def login():
	logger.debug("Iniciando sesion")
	login_url = f"{SITE_URL}login_login.ajax"

	data = {
		"email": USERNAME,
		"password": PASSWORD,
		"recordarme": "true",
		"homeSeeker": "true",
		"urlActual": SITE_URL
	}
	HEADERS["Content-Type"] = "application/x-www-form-urlencoded;charset=UTF-8"

	logger.debug(f"POST {login_url}")
	params = PARAMS.copy()
	params["url"] = login_url
	res = request.make(ZENROWS_API_URL, 'POST', params=params, data=data)
	data = res.json()
	
	print(data)

	request.headers = {
		"sessionId": data["contenido"]["sessionID"],
		"x-panel-portal": "24MX",
		"content-type": "application/json;charset=UTF-8"
	}

	with open(PARAMS_FILE, "w") as f:
		json.dump(request.headers, f, indent=4)
	logger.success("Sesion iniciada con exito")

request = Request(None, HEADERS, logger, login)

# Get the information about the searchers of the lead
def get_busqueda_info(lead_id) -> dict | None:
	logger.debug("Extrayendo la informacion de busqueda del lead: "+lead_id)
	busqueda_url = f"{SITE_URL}leads-api/publisher/contact/{lead_id}/user-profile"

	logger.debug(f"GET {busqueda_url}")
	PARAMS["url"] = busqueda_url
	res = request.make(ZENROWS_API_URL, 'GET', params=PARAMS)

	if res == None:
		logger.error("No se pudo obtener la informacion de busqueda para el lead: "+lead_id)
		return None
	try:
		logger.success("Informacion de busqueda extraida con exito")
		return res.json()
	except requests.exceptions.JSONDecodeError:
		logger.error(f"El lead {lead_id} no tiene informacion de busqueda")
		return None

def extract_busqueda_info(data: dict | None) -> dict:
	if data == None:
		return {
			"zonas": "",
			"tipo": "",
			"presupuesto": "",
			"cantidad_anuncios": "",
			"contactos": "",
			"inicio_busqueda": ""
		}

	features = data.get("property_features", {})
	lead_info = data.get("lead_info", {})

	busqueda = {
		"zonas": ','.join([i["name"] for i in data["searched_locations"]["streets"]]) if "streets" in data.get("searched_locations", {}) else "",
		"tipo": lead_info.get("search_type", {}).get("type", ""),
		"presupuesto": f"{lead_info.get('price', {}).get('min', '')}, {lead_info.get('price', {}).get('max', '')}",
		"cantidad_anuncios": lead_info.get("views", ""),
		"contactos": lead_info.get("contacts", ""),
		"inicio_busqueda": lead_info.get("started_search_days", "")
	}

	range_props = {
		"total_area": "total_area_xm2",
		"covered_area": "covered_area_xm2",
		"banios": "baths",
		"recamaras": "bedrooms",
	}
	for prop in range_props:
		try:
			value = features[range_props[prop]]
			busqueda[prop] = str(value["min"]) + ", " + str(value["max"])
		except KeyError:
			busqueda[prop] = ""

	return busqueda

# Takes the JSON object getted from the API and extract the usable information.
def extract_lead_info(data: dict) -> dict:
	lead_id = data["id"]
	contact_id = data["contact_publisher_user_id"]
	raw_busqueda = get_busqueda_info(contact_id)
	busqueda = extract_busqueda_info(raw_busqueda)
	posting = data.get("posting", {})

	lead_info = {
		"id": lead_id,
		"contact_id": contact_id, 
		"fuente": "Inmuebles24",
		"fecha_lead": datetime.strptime(data["last_lead_date"], "%Y-%m-%dT%H:%M:%S.%f+00:00").strftime(DATE_FORMAT),
		"fecha": strftime(DATE_FORMAT, gmtime()),
		"nombre": data.get("lead_user", {}).get("name"),
		"link": f"{SITE_URL}panel/interesados/{contact_id}",
		"telefono": data.get("phone", ""),
		#"telefono_2": data["phone_list"][1],
		"email": data.get("lead_user", {}).get("email"),
		"propiedad": {
			"titulo": posting.get("title", ""),
			"link": "",
			"precio": posting.get("price", {}).get("amount", ""),
			"ubicacion": posting.get("address", ""),
			"tipo": posting.get("real_estate_type", {}).get("name"),
			"municipio": posting.get("location", {}).get("parent", {}).get("name", "") #Ciudad
		},
		"busquedas": busqueda
	}

	return lead_info

def change_status(contact_id, status="READ"):
	status_url = f"{SITE_URL}leads-api/leads/status/{status}?=&contact_publisher_user_id={contact_id}"

	PARAMS["url"] = status_url
	PARAMS["autoparse"] = False
	res = request.make(ZENROWS_API_URL, 'POST', params=PARAMS)

	if res != None and res.status_code >= 200 and res.status_code < 300:
		logger.success(f"Se marco a lead {contact_id} como {status}")
	else:
		if res != None:
			logger.error(res.content)
			logger.error(res.status_code)
		logger.error(f"Error marcando al lead {contact_id} como {status}")
	PARAMS["autoparse"] = True

def send_message(lead_id, msg):
	logger.debug(f"Enviando mensaje a lead {lead_id}")
	msg_url = f"{SITE_URL}leads-api/leads/{lead_id}/messages"

	data = {
		"is_comment": False,
		"message": msg,
		"message_attachments": []
	}

	PARAMS["url"] = msg_url
	res = request.make(ZENROWS_API_URL, 'POST', params=PARAMS, json=data)

	if res != None and res.status_code >= 200 and res.status_code < 300:
		logger.success(f"Mensaje enviado correctamente a lead {lead_id}")
	else:
		logger.error(f"Error enviando mensaje al lead {lead_id}")

def main():
	sheet = Sheet(logger, "mapping.json")
	headers = sheet.get("A2:Z2")[0]
	logger.debug(f"Extrayendo leads")

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

	status = "nondiscarded" #Filtramos solamente los leads nuevos
	first = True
	offset = 0
	total = 0
	limit = 20
	finish = False
	total_finded = 0

	while first or (not finish and offset < total):
		leads = []
		leads_url = f"{SITE_URL}leads-api/publisher/leads?offset={offset}&limit={limit}&spam=false&status={status}&sort=unread"

		logger.debug(f"GET {leads_url}")
		PARAMS["url"] = leads_url
		data = request.make(ZENROWS_API_URL, 'GET', params=PARAMS).json()

		if first:
			total = data["paging"]["total"]
			logger.debug(f"Total: {total}")
			first = False

		for raw_lead in data["result"]:
			if raw_lead["statuses"][0] == "READ": #Como los leads estan ordenandos, al encontrar uno con estado READ. paramos
				logger.debug("Se encontro un lead con status READ, deteniendo")
				finish = True
				break

			lead = extract_lead_info(raw_lead)
			logger.debug(lead)

			#Si nunca le enviamos un mensaje, la segunda condicion es para validar que no sea un mensaje enviado por el lead a notros
			if "last_message" not in raw_lead or (raw_lead.get("last_message", {}).get("to") == "57036554"):
				msg = generate_mensage(lead)
				#send_message(lead["id"], msg)
				if lead['email'] != '':
					if lead["propiedad"]["ubicacion"] == "":
						lead["propiedad"]["ubicacion"] = "que consultaste"
					else:
						lead["propiedad"]["ubicacion"] = "ubicada en " + lead["propiedad"]["ubicacion"]

					gmail_msg = generate_mensage(lead, gmail_spin)
					subject = generate_mensage(lead, gmail_subject)
					gmail.send_message(gmail_msg, subject, lead["email"], attachment)

				lead["message"] = msg.replace('\n', ' ')
			else:
				logger.debug(f"Ya le hemos enviado un mensaje al lead {lead['nombre']}, lo salteamos")
				lead["message"] = ""
				change_status(lead["contact_id"], "READ")

			row_lead = sheet.map_lead(lead, headers)
			sheet.write([row_lead])
			leads.append(row_lead)
			total_finded += 1
		
		offset += len(leads)
		logger.debug(f"Mensajes enviados: {total_finded}")
		logger.debug(f"len: {len(leads)}")

	logger.success(f"Se encontraron {total_finded} nuevos Leads")

if __name__ == "__main__":
	main()
	#login(user, password)
