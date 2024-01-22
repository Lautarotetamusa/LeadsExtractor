from time import gmtime, strftime
from dotenv import load_dotenv
from datetime import datetime
import json
import os
from enum import IntEnum

from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

from src.message import generate_mensage
import src.infobip as infobip
from src.logger import Logger
from src.sheets import Gmail, Sheet
from src.make_requests import Request
from email.mime.application import MIMEApplication
#mis propiedades:
#https://propiedades.com/api/v3/property/MyProperties

#En este archivo tenemos todas las propieades previamente extraidas
with open("src/propiedades_com/properties_obj.json") as f:
    PROPS = json.load(f)

load_dotenv()

logger = Logger("propiedades.com")
gmail = Gmail({
    "email": os.getenv("EMAIL_CONTACT"),
}, logger)
SUBJECT = os.getenv("SUBJECT") or "subject"

USERNAME=os.getenv('PROPIEDADES_USERNAME')
PASSWORD=os.getenv('PROPIEDADES_PASSWORD')

DATE_FORMAT = "%d/%m/%Y"
API_URL = "https://ggcmh0sw5f.execute-api.us-east-2.amazonaws.com"
PARAMS_FILE = os.path.dirname(os.path.realpath(__file__)) + "/params.json"
with open(PARAMS_FILE, "r") as f:
    HEADERS = json.load(f)

def login(session="propiedades_com/session"):
    login_url = "https://propiedades.com/login"
    options = Options()
    options.add_argument(f"--user-data-dir={session}") #Session
    #options.add_argument(f"--headless") #Session
    options.add_argument("--no-sandbox") # Necesario para correrlo como root dentro del container

    driver = webdriver.Chrome(options=options)
    access_token = None

    logger.debug("Iniciando sesion")
    try:
        driver.get(login_url)

        #Esperar que cargue la pagina
        WebDriverWait(driver, 20).until(EC.visibility_of_element_located((By.XPATH, "//input[@data-gtm='text_field_email']")))

        username_input = driver.find_element(By.XPATH, "//input[@data-gtm='text_field_email']")
        username_input.send_keys(USERNAME)
        driver.find_element(By.XPATH, "//button[@data-id='login_correo']").click()

        password_input = driver.find_element(By.XPATH, "//input[@data-gtm='text_field_password']")
        password_input.send_keys(PASSWORD)
        driver.find_element(By.XPATH, "//button[@data-id='login_password']").click()

        WebDriverWait(driver, 10).until(EC.visibility_of_element_located((By.XPATH, "//img[@data-testid='user-avatar']")))
        logger.success("Sesion iniciada con exito")

        logger.debug("Obteniendo access token")
        #access_token = driver.execute_script(f"return window.localStorage.getItem('CognitoIdentityServiceProvider.{CLIENT_ID}.{USERNAME}.idToken');")
        cookies = driver.get_cookies()
        print(cookies)
        logger.success("Access token obtenido con exito")
        logger.success(access_token)
    except Exception as e:
        logger.error("Ocurrio un error generando el access token")
        logger.error(str(e))
    finally:
        driver.quit()

request = Request(None, HEADERS, logger, login)

#Lista de estados posibles de un lead
#Los tomamos de la pagina
class Status(IntEnum):
    CONTACTADO = 2,
    CALIFICADO = 3,
    EN_PROCESO = 4,
    CONVERTIDO = 5,
    CERRADA = 6

if (not os.path.exists(PARAMS_FILE)):
    logger.error("El archivo params.json no existe")
    with open(PARAMS_FILE, "a") as f:
        json.dump({
            "Authorization": "",
            "x-api-key": ""
        }, f, indent=4)

def get_lead_property(property_id: str):
    if property_id not in PROPS:
        logger.error(f"No se encontro la propiedad con id {property_id}")
        return {
            "id": "",
            "titulo": "",
            "link": "",
            "precio": "",
            "ubicacion": "",
            "tipo": "",
            "municipio": ""
        }
    data = PROPS[property_id]

    return {
        "id": data["id"],
        "titulo": data["short_address"],
        "link": data["url"],
        "precio": data["price"],
        "ubicacion": data["address"],
        "tipo": data["type_children_string"],
        "municipio": data["municipality"]
    }

# Takes the JSON object getted from the API and extract the usable information.
def extract_lead_info(data: dict) -> dict:
    prop = get_lead_property(str(data["property_id"]))

    prop["address"] = data["address"]
    if prop["titulo"] == "": prop["titulo"] = prop["address"]

    lead_info = {
            "fuente": "Propiedades.com",
            "fecha_lead": datetime.strptime(data["updated_at"], '%Y-%m-%d').strftime(DATE_FORMAT),
            "id": data["id"],
            "fecha": strftime(DATE_FORMAT, gmtime()),
            "nombre": data["name"],
            "link": "",
            "telefono": data["phone"],
            #"telefono_2": data["phone_list"][1],
            "email": data["email"],
            "propiedad": prop,
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

def change_status(lead_id, status:Status):
    logger.debug(f"Marcando como {status.name} al lead {lead_id}")
    
    url = f"{API_URL}/prod/leads/status"

    req = {
        "lead_id": lead_id,
        "status": status,
        "country": "MX"
    }

    res = request.make(url, "PUT", json=req)
    #res = requests.put(url, headers=HEADERS, json=req)
    if (res.status_code != 200):
        logger.error(res.status_code)
        logger.error(res.text)
    else:
        logger.success(f"Se marco a lead {lead_id} como {status.name}")

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

    first = True
    page = 1
    end = False
    url = f"{API_URL}/prod/get/leads?page={page}&country=MX"

    while (not end) and (first == True or page != None):
        leads = []
        logger.debug(url)

        data = request.make(url).json()["leads"]
        #data = get_data(url)["leads"]
        #if data == None: return None
        
        page = data["page"]["next_page"]
        url = f"{API_URL}/prod/get/leads?page={page}&country=MX"
        total = data["page"]["items"]
        logger.debug(f"total: {total}")

        for raw_lead in data["properties"]:
            if raw_lead["status"] == Status.CONTACTADO:
                logger.debug("Se encontro un lead ya conctactado, paramos")
                end = True #Cuando encontramos un lead conctado paramos
                break

            lead = extract_lead_info(raw_lead)
            logger.debug(lead)

            msg = generate_mensage(lead)
            lead["message"] = msg.replace('\n', '')
            
            if lead['email'] != '':
                if lead["propiedad"]["ubicacion"] == "":
                    lead["propiedad"]["ubicacion"] = "que consultaste"
                else:
                    lead["propiedad"]["ubicacion"] = "ubicada en " + lead["propiedad"]["ubicacion"]

                gmail_msg = generate_mensage(lead, gmail_spin)
                subject = generate_mensage(lead, gmail_subject)
                gmail.send_message(gmail_msg, subject, lead["email"], attachment)
                infobip.create_person(logger, lead)

            change_status(lead["id"], Status.CONTACTADO)

            row_lead = sheet.map_lead(lead, headers)
            leads.append(row_lead)

        sheet.write(leads)

        logger.debug(f"Leads leidos: {len(leads)}")
        first = False

    logger.success(f"Se encontraron {len(leads)} nuevos Leads")

if __name__ == "__main__":
    main()
    #change_all_status(Status.EN_PROCESO)
