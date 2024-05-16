import requests
import os
import urllib.parse

from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By

from src.logger import Logger
from src.lead import Lead

API_KEY = os.getenv("JOTFORM_API_KEY")
FORM_ID = os.getenv("JOTFORM_FORM_ID")
assert API_KEY != "" and API_KEY != None, "JOTFORM_API_KEY is not in enviroment"
assert FORM_ID != "" and FORM_ID != None, "JOTFORM_FORM_ID is not in enviroment"

API_URL = "https://api.jotform.com"
PDF_URL = "https://www.jotform.com/pdf-submission/{submissionID}"

def generate_url(logger: Logger, lead: Lead) -> str | None:
    if lead.asesor['email'] == None or lead.asesor['email'] == "":
        logger.error(f"El asesor {lead.asesor['name'] }no tiene email asignado")
        return None

    #Valores por defecto
    if lead.busquedas['covered_area'] == "" or lead.busquedas['covered_area'] == None:
        lead.busquedas['covered_area'] = '350, 350'

    if lead.busquedas['banios'] == "" or lead.busquedas['banios'] == None:
        lead.busquedas['banios'] = '5, 5'

    if lead.busquedas['recamaras'] == "" or lead.busquedas['recamaras'] == None:
        lead.busquedas['recamaras'] = '4, 4'

    props = {
        "covered_area": 0,
        "banios": 0 ,
        "recamaras": 0,
    }
    #La data viene como banios: 5, 6. osea entre 5 y 6, calculamos un promedio entre esos dos valores y usamos eso
    for key in props:
        try:
            [min, max] = lead.busquedas[key].split(', ')
            props[key] = int((int(min) + int(max)) / 2)
        except Exception as e:
            logger.error(f"No se pudo obtener el {key}")
            logger.debug(f"busquedas['{key}']: " + str(lead.busquedas[key]))
            logger.error(str(e))
            return None

    url = f"https://www.jotform.com/{FORM_ID}"
    params = {
        "nombreCliente": lead.nombre,
        "emailCliente": lead.email if lead.email != "" and lead.email != None else "cotizaciones@rebora.com.mx",
        "escribaUna": lead.asesor['name'],
        "email123": lead.asesor['email'],
        "numeroDe[full]": lead.asesor['phone'],
        "pagoInicial": "2,500,000",
        "pagoInicial118": "25%",
        "aCuantos": "16+meses",
        "escribaUna9": "Premium",
        "cuantosCuartos": props['recamaras'],
        "cuantosBanos56": props['banios'],
        "cuantosBanos": "0", #Tamaño del sotano
        "tamanoDe": int(props['covered_area'] * 0.45),   #Tamaño planta baja
        "tamanoDe79": int(props['covered_area'] * 0.45), #Tamaño planta alta
        "tamanoDel": int(props['covered_area'] * 0.10),  #Tamaño roof garden
        "tamanoDe84": "0",  #Tamaño rampa estacionamiento
        "tamanoDe85": "30", #Tamaño jardin exterior
        "tamanoDe86": "0",  #Tamaño de alberca
        "tamanoDe87": "50", #Tamaño muro perimetral
    }

    url = urllib.parse.urljoin(url, '?' + urllib.parse.urlencode(params))
    return url

#Devuelve un link al PDF creado
def new_submission(logger: Logger, lead: Lead) -> None | str:
    url = generate_url(logger, lead)
    print(url)
    if url == None:
        return None

    res = requests.get(url)
    if not res.ok:
        logger.error('Error en la solicitud:' + res.text)
        return None

    options = Options()
    options.add_argument(f"--headless") #Session
    options.add_argument("--no-sandbox") # Necesario para correrlo como root dentro del container

    driver = webdriver.Chrome(options=options)

    logger.debug("Generando pdf")
    try:
        driver.get(url)
        button = driver.find_element(By.ID, "form-pagebreak-next_8")
        button.click()
        button = driver.find_element(By.ID, "form-pagebreak-next_82")
        button.click()
        button = driver.find_element(By.ID, "input_2")
        button.click()
    except Exception as e:
        logger.error("driver error: " + str(e))
        return None
    finally:
        driver.close()

    submission = get_last_submission(logger)
    if submission == None:
        return None
    
    if submission.get('id', None) == None:
        logger.error("No se pudo obtener el id de la submission")
        logger.debug("submission:" + str(submission))
        return None

    return PDF_URL.format(submissionID=submission['id'])

def get_last_submission(logger: Logger, form_id: str = FORM_ID):
    url = urllib.parse.urljoin(API_URL, f"form/{form_id}/submissions")
    payload = {
        'apiKey': API_KEY
    }

    res = requests.get(url, params=payload)

    if not res.ok:
        logger.error('Error en la solicitud:' + res.text)
        return None

    logger.success('Solicitud exitosa')
    data = res.json()
    return data.get('content', [None])[0]

def get_questions_form(logger: Logger, form_id: str = FORM_ID):
    logger.fuente += " > Jotform API"
    url = urllib.parse.urljoin(API_URL, f"form/{form_id}/questions")
    payload = {
        'apiKey': API_KEY
    }

    res = requests.get(url, params=payload)

    if res.ok:
        logger.success('Solicitud exitosa')
        data = res.json()
        print(data)
        return data
    else:
        logger.error('Error en la solicitud:' + res.text)
        return None
