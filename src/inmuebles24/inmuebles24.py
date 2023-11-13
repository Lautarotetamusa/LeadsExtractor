from time import gmtime, strftime
from datetime import datetime
import requests
import json

from src.message import generate_mensage
from .extractor import get_req, post_req
from src.logger import Logger
from src.sheets import Sheet

from .params import cookies, headers

DATE_FORMAT = "%d/%m/%Y"
SITE_URL = "https://www.inmuebles24.com/"

logger = Logger("inmuebles24.com")

def login(user, password):
	login_url = f"{SITE_URL}/login_login.ajax"

	data = {
		"email": user,
		"password": password,
		"recordarme": "true",
		"homeSeeker": "true",
		"urlActual": SITE_URL
	}

	res = post_req(login_url, data, logger)
	print(res.text)
	print(res.headers)
	return res

# Get the information about the searchers of the lead
def get_busqueda_info(lead_id):
	logger.debug("Extrayendo la informacion de busqueda del lead: "+lead_id)
	busqueda_url = f"{SITE_URL}leads-api/publisher/contact/{lead_id}/user-profile"
	res = get_req(busqueda_url, logger)
	if res == None:
		logger.error("No se pudo obtener la informacion de busqueda para el lead: "+lead_id)
		return None
	try:
		logger.success("Informacion de busqueda extraida con exito")
		return res.json()
	except requests.exceptions.JSONDecodeError as e:
		logger.error(f"El lead {lead_id} no tiene informacion de busqueda")
		return None

def extract_busqueda_info(data: object) -> object:
	if data == None:
		return None

	features = data["property_features"]

	busqueda = {
		"zonas": ','.join([i["name"] for i in data["searched_locations"]["streets"]]) if "streets" in data["searched_locations"] else "",
		"tipo": data["lead_info"]["search_type"]["type"],
		"presupuesto": str(data["lead_info"]["price"]["min"]) + ", " + str(data["lead_info"]["price"]["max"]),
		"cantidad_anuncios": data["lead_info"]["views"],
		"contactos": data["lead_info"]["contacts"],
		"inicio_busqueda": data["lead_info"]["started_search_days"] 
	}

	range_props = {
		"total_area": "total_area_xm2",
		"covered_area": "covered_area_xm2",
		"banios": "baths",
		"recamaras": "bedrooms",
	}
	for prop in range_props:
		value = features[range_props[prop]]
		busqueda[prop] = str(value["min"]) + ", " + str(value["max"])

	return busqueda

# Takes the JSON object getted from the API and extract the usable information.
def extract_lead_info(data: object) -> object:
	lead_id = data["id"]
	contact_id = data["contact_publisher_user_id"]
	raw_busqueda = get_busqueda_info(contact_id)
	busqueda = extract_busqueda_info(raw_busqueda)
	
	lead_info = {
		"id": lead_id,
		"fuente": "Inmuebles24",
		"fecha_lead": datetime.strptime(data["last_lead_date"], "%Y-%m-%dT%H:%M:%S.%f+00:00").strftime(DATE_FORMAT),
		"fecha": strftime(DATE_FORMAT, gmtime()),
		"nombre": data["lead_user"]["name"],
		"link": f"{SITE_URL}panel/interesados/{contact_id}",
		"telefono": data["phone"],
		#"telefono_2": data["phone_list"][1],
		"email": data["lead_user"]["email"],
		"propiedad": {
			"titulo": data["posting"]["title"],
			"link": "",
			"precio": data["posting"]["price"]["amount"],
			"ubicacion": data["posting"]["address"],
			"tipo": data["posting"]["real_estate_type"]["name"],
			"municipio": data["posting"]["location"]["parent"]["name"] #Ciudad
		},
		"busquedas": busqueda
	}

	return lead_info

def change_status(contact_id, status="READ"):
	status_url = f"{SITE_URL}leads-api/leads/status/READ?=&contact_publisher_user_id={contact_id}"

	#res = post_req(status_url, None, logger)
	res = requests.post(status_url, headers=headers, cookies=cookies)
	print(res.text)

	if res != None:
		logger.success(f"Se marco a lead {contact_id} como {status}")
	else:
		logger.error(f"Error marcando al lead {contact_id} como {status}")

def send_message(lead_id, msg):
	logger.debug(f"Enviando mensaje a lead {lead_id}")
	msg_url = f"{SITE_URL}leads-api/leads/{lead_id}/messages"

	data = {
		"is_comment": False,
		"message": msg,
		"message_attachments": []
	}
	#res = post_req(msg_url, data, logger)
	res = requests.post(msg_url, headers=headers, cookies=cookies, json=data)
	if res != None:
		logger.success(f"Mensaje enviado correctamente a lead {lead_id}")
	else:
		logger.error(f"Error enviando mensaje al lead {lead_id}")

def main():
	#send_message("452528425", "Hola muy buenos días")
	#change_status("184214713")
	#return 
	sheet = Sheet(logger)
	headers = sheet.get("A2:Z2")[0]
	logger.debug(f"Extrayendo leads")

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
		res = get_req(leads_url, logger)
		if res == None: return None
		data = res.json()

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
			msg = generate_mensage(lead)
			send_message(lead["id"], msg)
			change_status(raw_lead["contact_publisher_user_id"])
			lead["message"] = msg

			row_lead = sheet.map_lead(lead, headers)
			leads.append(row_lead)
			total_finded += 1
		
		offset += len(leads)
		sheet.write(leads)

		logger.debug(f"len: {len(leads)}")

	logger.success(f"Se encontraron {total_finded} nuevos Leads")

if __name__ == "__main__":
	main()
	#login(user, password)