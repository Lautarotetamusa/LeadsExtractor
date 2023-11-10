import __init__

import requests
import json
import os
from dotenv import load_dotenv
from time import gmtime, strftime
from datetime import datetime

from logger import Logger
from sheets import Sheet

#dotenv_path = Path('../../.env')
#print(dotenv_path)
load_dotenv()
logger = Logger("casasyterrenos.com")

DATE_FORMAT = "%d/%m/%Y"
API_URL = "https://cytpanel.casasyterrenos.com/api/v1"
CLIENT_ID = "4je1v2kfou9e9plpv6vf0vmnll"
PARAMS_FILE = os.path.dirname(os.path.realpath(__file__)) + "/params.json"
USERNAME=os.getenv('CASASYTERRENOS_USERNAME')
PASSWORD=os.getenv('CASASYTERRENOS_PASSWORD')
logger.debug(f"params file: {PARAMS_FILE}")

if (not os.path.exists(PARAMS_FILE)):
    logger.error("El archivo params.json no existe")
    with open(PARAMS_FILE, "a") as f:
        json.dump({
            "Authorization": ""
        }, f, indent=4)
    
with open(PARAMS_FILE, "r") as f:
    HEADERS = json.load(f)

# Takes the JSON object getted from the API and extract the usable information.
def extract_lead_info(data: object) -> object:
    lead_info = {
		"fuente": "Casas y terrenos",
        "fecha_lead": datetime.strptime(data["created"], '%d-%m-%Y %H:%M:%S').strftime(DATE_FORMAT),
        "id": data["id"],
		"fecha": strftime(DATE_FORMAT, gmtime()),
		"nombre": data["name"],
		"link": "",
		"telefono": data["phone"],
		#"telefono_2": data["phone_list"][1],
		"email": data["email"],
		"propiedad": {
            "id": data["property_id"],
			"titulo": data["property_title"],
			"link": f"https://www.casasyterrenos.com/propiedad/{data['property_id']}",
			"precio": "",
			"ubicacion": "",
			"tipo": data["property_type"],
		},
		"busquedas": {
            "zonas": "",
            "tipo": "",
            "presupuesto": "",
            "cantidad_anuncios": "",
            "contactos": "",
            "inicio_busqueda": "",
            "total_area": "",
            "covered_area": "",
            "banios": "",
            "recamaras": "",
        }
	}

    return lead_info

# Obtener el access token necesario para las requests posteriores
# Si la sesion no esta iniciada, la iniciamos.
# Si la sesión ya está iniciada, generamos el access tokens
def get_access_token(username: str, password: str, session="session"):
    global HEADERS

    from selenium import webdriver
    from selenium.webdriver.chrome.options import Options
    from selenium.webdriver.common.by import By
    from selenium.webdriver.chrome.options import Options
    from selenium.webdriver.support.ui import WebDriverWait
    from selenium.webdriver.support import expected_conditions as EC

    login_url = "https://panel-pro.casasyterrenos.com/login"
    panel_url = "https://panel-pro.casasyterrenos.com/contacts"
    options = Options()
    options.add_argument(f"--user-data-dir={session}") #Session

    driver = webdriver.Chrome(options=options)
    access_token = None

    logger.debug("Generando un nuevo token de acceso")
    try:
        driver.get(panel_url)

        if (driver.current_url == login_url):
            logger.debug("La sesion no esta iniciada, iniciando..")
            driver.get(login_url)

            #Esperar que cargue la pagina
            WebDriverWait(driver, 10).until(EC.visibility_of_element_located((By.XPATH, "//input[@name='username']")))

            username_input = driver.find_element(By.XPATH, "//input[@name='username']")
            password_input = driver.find_element(By.XPATH, "//input[@name='password']")
            send_btn = driver.find_element(By.XPATH, "//button[@data-splitbee-event='log-in']")

            username_input.send_keys(username)
            password_input.send_keys(password)
            send_btn.click()

            WebDriverWait(driver, 10).until(EC.visibility_of_element_located((By.XPATH, "/html/body/div/div/div/div[1]/div/nav/a")))
            logger.success("Sesion iniciada con exito")

        logger.debug("Obteniendo access token")
        access_token = driver.execute_script(f"return window.localStorage.getItem('CognitoIdentityServiceProvider.{CLIENT_ID}.{username}.idToken');")
        logger.success("Access token obtenido con exito")
        logger.success(access_token)
    except Exception as e:
        logger.error("Ocurrio un error generando el access token")
        logger.error(str(e))
    finally:
        driver.quit()

    HEADERS = {
        "Authorization": f"Bearer {access_token}"
    }
    with open(PARAMS_FILE, "w") as f:
        json.dump(HEADERS, f, indent=4)

def get_data(url) -> object:
    res = requests.get(url, headers=HEADERS)
    if (res.status_code != 200):
        if (res.status_code == 401):
            logger.error("El token de acceso expiro")
            get_access_token(USERNAME, PASSWORD)
            return None
        else: 
            logger.error(res.status_code)
            logger.error(res.text)
            return None
    return res.json()

def get_leads(page=1):
    logger.debug(f"Extrayendo leads")
    status = "1" #Filtramos solamente los leads nuevos
    leads = []
    first = True
    url = f"{API_URL}/list_contact/?page={page}&status={status}"

    while first == True or url != None:
        logger.debug(url)

        data = get_data(url)
        if data == None: return None

        url = data["next"]

        leads += data["results"]
        logger.debug(f"len: {len(leads)}")
        first = False

    logger.success(f"Se encontraron {len(leads)} nuevos Leads")
    return leads

def make_contacted(lead_id):
    logger.debug(f"Marcando como contactacto a lead {lead_id}")
    url = f"{API_URL}/contact/{lead_id}"

    data = {
        "id": lead_id,
        "status": 2, # 2 -> Contactado por correo
        "status_description": "Correo"
    }

    res = requests.patch(url, headers=HEADERS, data=data)
    if (res.status_code != 200):
        logger.error(res.status_code)
        logger.error(res.text)
    else:
        logger.success(f"Se contacto correctamente a lead {lead_id}")

def main():
    sheet = Sheet(logger)
    headers = sheet.get("A2:Z2")[0]
    logger.debug(f"Extrayendo leads")

    status = "1" #Filtramos solamente los leads nuevos
    first = True
    page = 1
    url = f"{API_URL}/list_contact/?page={page}&status={status}"

    while first == True or url != None:
        leads = []
        logger.debug(url)

        data = get_data(url)
        if data == None: return None

        url = data["next"]
        total = data["count"]
        logger.debug(f"total: {total}")

        for raw_lead in data["results"]:
            lead = extract_lead_info(raw_lead)
            logger.debug(lead)

            lead["message"] = ""
            make_contacted(lead["id"])

            row_lead = sheet.map_lead(lead, headers)
            leads.append(row_lead)

        sheet.write(leads)

        logger.debug(f"len: {len(leads)}")
        first = False

    logger.success(f"Se encontraron {len(leads)} nuevos Leads")

if __name__ == "__main__":
    main()