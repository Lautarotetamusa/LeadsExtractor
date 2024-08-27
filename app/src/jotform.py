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

# TODO: Renovar token 
cookies = os.getenv("JOTFORM_COOKIE")

headers = {
      'Origin': 'https://www.jotform.com',
      'Referer': 'https://www.jotform.com/tables/242244461116044',
      'Cookie': cookies
}

def get_img_data(img_url: str) -> bytes | None:
    res = requests.get(img_url)
    if not res.ok:
        return None

    return res.content

# TODO: Esto solo sirve para este formulario
# TODO: Podría procesar multiples imagenes en una peticion
def upload_image(form_id: str, submission_id: str, qid: str, img_data: bytes, img_name) -> str | None:
    url = f"https://www.jotform.com/API/sheets/{form_id}/form/{form_id}/submission/{submission_id}/files?from=sheets"
    payload = {
        'submissionID': submission_id, 
        'qid': qid
    }
    files = [
      ('uploadedFile[]',(f"{img_name}.jpg", img_data, 'image/jpeg'))
    ]

    res = requests.post(url, headers=headers, data=payload, files=files)
    if not res.ok:
        return "Request error"

    # Devuelve codigo 200 con responseCode != 200 xD
    data = res.json()
    if data.get("responseCode", 200) != 200:
        return data

# Obtener las lista de preguntas del form:
# https://api.jotform.com/form/242244461116044/questions?apiKey=
def submit_cotizacion_form(logger: Logger, form_id: str, data, asesor) -> dict | None:
    url = f"https://api.jotform.com/form/{form_id}/submissions?apiKey={API_KEY}"
    data = {
            "10": data.get("title", ""),
            "26": data.get("price", ""),
            "12": data.get("type", ""),
            "13": asesor.get("name", ""),
            "61": asesor.get("phone", ""),
            "15": asesor.get("email", ""), 
            "17": data.get("building_size", ""),
            "18": data.get("size", ""),
            "19": data.get("antiguedad", ""),
            "20": data.get("url", ""),
            "21": data.get("currency", ""),
            "22": data.get("location", {}).get("zone", "")
    }

    res = requests.post(url, json=data)

    if res.ok:
        logger.success('Solicitud exitosa')
        data = res.json()
        return data
    else:
        logger.error('Error en la solicitud:' + res.text)
        return None

def generate_url(logger: Logger, lead: Lead) -> tuple[str, str | None]:
    if lead.asesor['email'] is None or lead.asesor['email'] == "":
        logger.error(f"El asesor {lead.asesor['name']} no tiene email asignado")
        return "", f"El asesor {lead.asesor['name']} no tiene email asignado"
    if lead.asesor['phone'] is None or lead.asesor['phone'] == "":
        logger.error(f"El asesor {lead.asesor['name']} no tiene phone asignado")
        return "", f"El asesor {lead.asesor['name']} no tiene phone asignado"
    if lead.asesor['name'] is None or lead.asesor['name'] == "":
        logger.error("El asesor no tiene name")
        return "", "El asesor no tiene name"

    # Valores por defecto
    if lead.busquedas['covered_area'] == "" or lead.busquedas['covered_area'] is None:
        lead.busquedas['covered_area'] = '350, 350'

    if lead.busquedas['banios'] == "" or lead.busquedas['banios'] is None:
        lead.busquedas['banios'] = '5, 5'

    if lead.busquedas['recamaras'] == "" or lead.busquedas['recamaras'] is None:
        lead.busquedas['recamaras'] = '4, 4'

    props = {
        "covered_area": 0,
        "banios": 0,
        "recamaras": 0,
    }
    # La data viene como banios: 5, 6. osea entre 5 y 6, calculamos un promedio entre esos dos valores y usamos eso
    for key in props:
        try:
            [min, max] = lead.busquedas[key].split(', ')
            props[key] = int((int(min) + int(max)) / 2)
        except Exception as e:
            logger.error(f"No se pudo obtener el {key}")
            logger.debug(f"busquedas['{key}']: " + str(lead.busquedas[key]))
            logger.error(str(e))
            return "", f"No se pudo obtener el {key}"

    url = f"https://www.jotform.com/{FORM_ID}"
    params = {
        "nombreCliente": lead.nombre,
        "emailCliente": lead.email if lead.email != "" and lead.email != None else "cotizaciones@rebora.com.mx",
        "escribaUna": lead.asesor['name'],
        "email123": lead.asesor['email'],
        "numeroDe[full]": lead.asesor['phone'][3:], #Sacar la caracteristica de pais
        "pagoInicial": "2,500,000",
        "pagoInicial118": "25%",
        "aCuantos": "16+meses",
        "escribaUna9": "Premium",
        "cuantosCuartos": props['recamaras'],
        "cuantosBanos56": props['banios'],
        "cuantosBanos": "0",  # Tamaño del sotano
        "tamanoDe": int(props['covered_area'] * 0.45),    # Tamaño planta baja
        "tamanoDe79": int(props['covered_area'] * 0.45),  # Tamaño planta alta
        "tamanoDel": int(props['covered_area'] * 0.10),   # Tamaño roof garden
        "tamanoDe84": "0",   # Tamaño rampa estacionamiento
        "tamanoDe85": "30",  # Tamaño jardin exterior
        "tamanoDe86": "0",   # Tamaño de alberca
        "tamanoDe87": "50",  # Tamaño muro perimetral
    }

    url = urllib.parse.urljoin(url, '?' + urllib.parse.urlencode(params))
    return url, None

def generate_pdf(form_id: str, submission_id: str):
    url = "https://www.jotform.com/API/sheets/generatePDF"
    params = {
        "formid": form_id,
        "submissionid":submission_id,
        #"reportid": ,
        #"sheetID": ,
    }
    
    res = requests.get(url, params=params, headers=headers)
    if res.ok:
        data = res.json()
        return data
    else:
        return None

# Devuelve un link al PDF creado
def new_submission(logger: Logger, lead: Lead) -> [str, str | None]:
    url, err = generate_url(logger, lead)
    if err is not None:
        return "", err
    logger.debug("url: "+url)

    res = requests.get(url)
    if not res.ok:
        logger.error('Error en la solicitud:' + res.text)
        return "", "Error en la solicitud" + res.text

    options = Options()
    options.add_argument("--headless")   # Session
    # Necesario para correrlo como root dentro del container
    options.add_argument("--no-sandbox")

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
        return "", "driver error: " + str(e)
    finally:
        driver.close()

    submission = get_last_submission(logger)
    if submission is None:
        return "", "Submission is None"

    if submission.get('id', None) is None:
        logger.error("No se pudo obtener el id de la submission")
        logger.debug("submission:" + str(submission))
        return "", "No se pudo obtener el id de la submission"

    logger.success("Pdf generado con exito")

    return PDF_URL.format(submissionID=submission['id']), None


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
        return data
    else:
        logger.error('Error en la solicitud:' + res.text)
        return None
