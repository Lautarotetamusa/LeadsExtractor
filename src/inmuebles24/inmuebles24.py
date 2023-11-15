from time import gmtime, strftime
from datetime import datetime
from dotenv import load_dotenv
import requests
import os
import json

from src.message import generate_mensage
from src.logger import Logger
from src.sheets import Sheet
from src.make_requests import Request

#from .params import cookies, headers
PARAMS_FILE = os.path.dirname(os.path.realpath(__file__)) + "/params.json"
with open(PARAMS_FILE, "r") as f:
    HEADERS = json.load(f)

from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

load_dotenv()

USERNAME = os.getenv("INMUEBLES24_USERNAME")
PASSWORD = os.getenv("INMUEBLES24_PASSWORD")
DATE_FORMAT = "%d/%m/%Y"
SITE_URL = "https://www.inmuebles24.com/"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
PARAMS = {
	#"resolve_captcha": "true",
	"apikey": os.getenv("ZENROWS_APIKEY"),
	"url": "",
    "js_render": "true",
    "antibot": "true",
    "premium_proxy": "true",
    "proxy_country": "mx",
	"session_id": 10,
	"custom_headers": "true",
	"original_status": "true"
}

logger = Logger("inmuebles24.com")

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

	logger.debug(f"POST {login_url}")
	PARAMS["url"] = login_url
	data = request.make(ZENROWS_API_URL, 'POST', params=PARAMS, data=data).json()

	request.headers = {
		"sessionId": data["contenido"]["sessionID"],
		"x-panel-portal": "24MX"
	}

	with open(PARAMS_FILE, "w") as f:
		json.dump(request.headers, f, indent=4)
	logger.success("Sesion iniciada con exito")

request = Request(None, HEADERS, logger, login)

# Get the information about the searchers of the lead
def get_busqueda_info(lead_id):
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
	except requests.exceptions.JSONDecodeError as e:
		logger.error(f"El lead {lead_id} no tiene informacion de busqueda")
		return None

def extract_busqueda_info(data: object) -> object:
	if data == None:
		return {
			"zonas": "",
			"tipo": "",
			"presupuesto": "",
			"cantidad_anuncios": "",
			"contactos": "",
			"inicio_busqueda": ""
		}

	features = data["property_features"]
	lead_info = data["lead_info"]

	busqueda = {
		"zonas": ','.join([i["name"] for i in data["searched_locations"]["streets"]]) if "streets" in data["searched_locations"] else "",
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
def extract_lead_info(data: object) -> object:
	lead_id = data["id"]
	contact_id = data["contact_publisher_user_id"]
	raw_busqueda = get_busqueda_info(contact_id)
	busqueda = extract_busqueda_info(raw_busqueda)
	posting = data.get("posting", {})

	lead_info = {
		"id": lead_id,
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

def send_message(driver, contact_id, msg):
	url = f"https://www.inmuebles24.com/panel/interesados/{contact_id}"
	
	logger.debug(f"Enviando mensaje al lead {contact_id}")
	try:
		driver.get(url)

		WebDriverWait(driver, 10).until(EC.visibility_of_element_located((By.XPATH, "/html/body/div[2]/div[2]/div[2]/div[4]/div[3]/textarea")))

		input = driver.find_element(By.XPATH, "/html/body/div[2]/div[2]/div[2]/div[4]/div[3]/textarea")
		input.send_keys(msg)

		button = driver.find_element(By.XPATH, "/html/body/div[2]/div[2]/div[2]/div[4]/div[3]/button")
		button.click()
		logger.success("Mensaje enviado correctamente")
	except Exception as e:
		logger.error("Ocurrio un error enviando el mensaje")
		logger.error(str(e))

def change_status(contact_id, status="READ"):
	status_url = f"{SITE_URL}leads-api/leads/status/{status}?=&contact_publisher_user_id={contact_id}"

	#res = post_req(status_url, None, logger)
	PARAMS["url"] = status_url
	res = request.make(ZENROWS_API_URL, 'POST', params=PARAMS)
	#res = requests.post(status_url, headers=headers, cookies=cookies)

	if res != None and res.status_code >= 200 and res.status_code < 300:
		logger.success(f"Se marco a lead {contact_id} como {status}")
	else:
		if res != None:
			logger.error(res.content)
			logger.error(res.status_code)
		logger.error(f"Error marcando al lead {contact_id} como {status}")

"""
def send_message(lead_id, msg):
	logger.debug(f"Enviando mensaje a lead {lead_id}")
	msg_url = f"{SITE_URL}leads-api/leads/{lead_id}/messages"

	data = {
		"is_comment": False,
		"message": msg,
		"message_attachments": []
	}
	res = post_req(msg_url, data, logger)
	#res = requests.post(msg_url, headers=headers, cookies=cookies, json=data)

	if res != None and res.status_code >= 200 and res.status_code < 300:
		logger.success(f"Mensaje enviado correctamente a lead {lead_id}")
	else:
		if res != None:
			logger.error(res.content)
			logger.error(res.status_code)
		logger.error(f"Error enviando mensaje al lead {lead_id}")
"""

#Esta sera la primera corrida
def get_all_leads():
	session = "session"
	options = Options()
	options.add_argument(f"--user-data-dir={session}") #Session
	#options.add_argument(f"--headless") #Session
	options.add_argument("--no-sandbox") # Necesario para correrlo como root dentro del container

	driver = webdriver.Chrome(options=options)

	sheet = Sheet(logger)
	headers = sheet.get("A2:Z2")[0]

	status = "nondiscarded" #Filtramos solamente los leads nuevos
	first = True
	offset = 760
	total = 0
	limit = 20
	total_finded = 0

	while first or offset < total:
		leads = []
		leads_url = f"{SITE_URL}leads-api/publisher/leads?offset={offset}&limit={limit}&spam=false&status={status}&sort=last_activity"

		logger.debug(f"GET {leads_url}")
		PARAMS["url"] = leads_url
		data = request.make(ZENROWS_API_URL, 'GET', params=PARAMS).json()

		if first:
			total = data["paging"]["total"]
			logger.debug(f"Total: {total}")
			first = False

		for raw_lead in data["result"]:
			lead = extract_lead_info(raw_lead)
			
			logger.debug(lead)

			#Si nunca le enviamos un mensaje 
			if "last_message" not in raw_lead:
				msg = generate_mensage(lead)
				send_message(driver, raw_lead["contact_publisher_user_id"], msg)
				lead["message"] = msg.replace('\n', ' ')
			else:
				logger.debug(f"Ya le hemos enviado un mensaje al lead {lead['nombre']}, lo salteamos")
				lead["message"] = ""

			row_lead = sheet.map_lead(lead, headers)
			sheet.write([row_lead])
			leads.append(row_lead)

			total_finded += 1
		
		offset += len(leads)
		logger.debug(f"Mensajes enviados: {total_finded}")
		logger.debug(f"len: {len(leads)}")

	logger.success(f"Se encontraron {total_finded} nuevos Leads")
	driver.quit()

def main():
	#send_message("184375921", "Hola muy buenos días")
	#change_status("184375921")
	#driver.quit()
	#return 
	sheet = Sheet(logger)
	headers = sheet.get("A2:Z2")[0]
	logger.debug(f"Extrayendo leads")

	status = "nondiscarded" #Filtramos solamente los leads nuevos
	first = True
	offset = 20
	total = 0
	limit = 20
	finish = False
	total_finded = 0

	while first or (not finish and offset < total):
		leads = []
		leads_url = f"{SITE_URL}leads-api/publisher/leads?offset={offset}&limit={limit}&spam=false&status={status}&sort=unread"

		PARAMS["url"] = leads_url
		data = request.make(ZENROWS_API_URL, 'GET', params=PARAMS)

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
			lead["message"] = msg.replace('\n', ' ')

			row_lead = sheet.map_lead(lead, headers)
			sheet.write([row_lead])
			leads.append(row_lead)
			total_finded += 1
		
		offset += len(leads)
		#sheet.write(leads)

		logger.debug(f"len: {len(leads)}")

	logger.success(f"Se encontraron {total_finded} nuevos Leads")

if __name__ == "__main__":
	main()
	#login(user, password)