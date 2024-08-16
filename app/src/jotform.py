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

headers = {
      'Origin': 'https://www.jotform.com',
      'Referer': 'https://www.jotform.com/tables/242244461116044',
      'Cookie': '_gj=f61b1aa461bf9b4097b50d92eeaedf345f17fd4a; hblid=idwBHlnuFAtmmILH0v5z60SC12AOKoCK; olfsk=olfsk9475756384112493; FORM_last_folder=allForms; g_state={"i_p":1724073586229,"i_l":3}; SHEET_last_folder=sheets; guest=guest_e5687c44184f0a37; JOTFORM_SESSION=5b169e0f-247f-b44d-3962-00657606; userReferer=https%3A%2F%2Fwww.jotform.com%2F; JF_SESSION_USERNAME=Diego_torres; last_edited_v4_form=242244461116044; DOCUMENT_last_folder=documents; limitAlignment=left_alt; wcsid=tYMhRbpYz3vdg7VA0v5z60SCKKA1B2C1; _oklv=1723802295242%2CtYMhRbpYz3vdg7VA0v5z60SCKKA1B2C1; _okdetect=%7B%22token%22%3A%2217238016935230%22%2C%22proto%22%3A%22about%3A%22%2C%22host%22%3A%22%22%7D; navLang=en-US; _okbk=cd5%3Davailable%2Ccd4%3Dtrue%2Cvi5%3D0%2Cvi4%3D1723801693894%2Cvi3%3Dactive%2Cvi2%3Dfalse%2Cvi1%3Dfalse%2Ccd8%3Dchat%2Ccd6%3D0%2Ccd3%3Dfalse%2Ccd2%3D0%2Ccd1%3D0%2C; _ok=4728-686-10-5570'
}

def get_img_data(img_url: str) -> bytes | None:
    res = requests.get(img_url)
    if not res.ok:
        return None

    return res.content

# TODO: Esto solo sirve para este formulario
# TODO: Podría procesar multiples imagenes en una peticion
def upload_image(form_id: str, submission_id: str, qid: str, img_data: bytes, img_name):
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
        return None

    data = res.json()
    return data

def submit_cotizacion_form(logger: Logger, form_id: str, data, asesor) -> dict | None:
    url = f"https://api.jotform.com/form/{form_id}/submissions?apiKey={API_KEY}"
    data = {
            "10": data["title"],
            "11": data["price"],
            "12": data["type"],
            "13": asesor["name"],
            "14": asesor["phone"],
            "15": asesor["email"], 
            "17": data["building_size"],
            "18": data["size"],
            "19": data["antiguedad"],
            "20": data["url"],
            "21": data["currency"],
            "22": data["location"]["zone"]
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
