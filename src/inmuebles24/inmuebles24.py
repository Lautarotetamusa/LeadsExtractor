import __init__

import json
from extractor import get_req, post_req
import requests
from logger import Logger
from sheets import Sheet
from time import gmtime, strftime

logger = Logger("inmuebles24.com")

site_url = "https://www.inmuebles24.com/"

def login(user, password):
	login_url = "https://www.inmuebles24.com/login_login.ajax"

	data = {
		"email": user,
		"password": password,
		"recordarme": "true",
		"homeSeeker": "true",
		"urlActual": "https://www.inmuebles24.com"
	}

	res = post_req(login_url, data)
	print(res.text)
	print(res.headers)
	return res


# This functions get all the unreaded leads
def get_leads():
	logger.debug("Comenzando a extraer...")
	offset = 0
	leads_url = f"{site_url}leads-api/publisher/leads?offset={offset}&limit=20&spam=false&status=nondiscarded&sort=last_activity"

	res = get_req(leads_url)
	if res == None:
		logger.error("No se encontraron leads")
		return []

	raw_leads = res.json()["result"]
	logger.success(f"Se encontraron {len(raw_leads)} nuevos leads")
	return raw_leads

# Get the information about the searchers of the lead
def get_busqueda_info(lead_id):
	logger.debug("Extrayendo la informacion de busqueda del lead: "+lead_id)
	busqueda_url = f"{site_url}leads-api/publisher/contact/{lead_id}/user-profile"
	res = get_req(busqueda_url)
	if res == None:
		logger.error("No se pudo obtener la informacion de busqueda para el lead: "+lead_id)
		return None
	try:
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
		"fuente": "Inmuebles24",
		"fecha": strftime('%d/%m/%Y', gmtime()),
		"nombre": data["lead_user"]["name"],
		"link": f"{site_url}panel/interesados/{lead_id}",
		"telefono": data["phone"][0],
		#"telefono_2": data["phone_list"][1],
		"email": data["lead_user"]["email"],
		"propiedad": {
			"titulo": data["posting"]["title"],
			"link": "",
			"precio": data["posting"]["price"]["amount"],
			"ubicacion": data["posting"]["address"],
			"tipo": data["posting"]["real_estate_type"]["name"],
		},
		"busquedas": busqueda
	}

	return lead_info

def main():
	sheet = Sheet(logger)
	headers = sheet.get("A2:Z2")[0]

	raw_leads = get_leads()
	leads = []
	for raw_lead in raw_leads:
		lead = extract_lead_info(raw_lead)
		leads.append(lead)
		print(lead)
		json.dumps(lead, indent=4)

		row_lead = sheet.map_lead(lead, headers)
		sheet.write([row_lead])

	with open("leads.json", "w") as f:
		json.dump(leads, f, indent=4)

if __name__ == "__main__":
	main()
	#login(user, password)