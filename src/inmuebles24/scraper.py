import string
import random
import json
import os

from time import gmtime, strftime

from src.logger import Logger
from src.sheets import Sheet
from src.message import generate_post_message
from src.make_requests import ApiRequest

SITE = "https://www.inmuebles24.com"
VIEW_URL = f"{SITE}/rp-api/leads/view"
LIST_URL = f"{SITE}/rplis-api/postings"
CONTACT_URL = f"{SITE}/rp-api/leads/contact"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
DATE_FORMAT = "%d/%m/%Y"
SENDER = {
    "name":  os.getenv("SENDER_NAME"),
    "phone": os.getenv("SENDER_PHONE"),
    "email": os.getenv("SENDER_EMAIL"),
    "id": "",
    "message": "",
    "page": "Listado",
    "postingId": "",
    "publisherId": ""
}

logger = Logger("scraper inmuebles24.com")
request = ApiRequest(logger, ZENROWS_API_URL, {
    "apikey": os.getenv("ZENROWS_APIKEY"),
    "url": "",
    #"js_render": "true",
    #"antibot": "true",
    "premium_proxy": "true",
    "proxy_country": "mx",
})

def get_publisher(post: dict, msg=""):
    detail_data = {
        "email": SENDER["email"],
        "name":  SENDER["name"],
        "phone": SENDER["phone"],
        "page":"Listado",
        "publisherId": post["publisher"]["id"],
        "postingId": post["id"]
    }

    if msg == "":
        #View the phone but not send message
        url = VIEW_URL
        logger.debug(f"Viendo telefono del publisher {post['publisher']['id']}")
    else:
        logger.debug(f"Enviando mensaje a publisher {post['publisher']['id']}")
        url = CONTACT_URL

    #Send a message or get the phone
    while True:
        if msg != "":
            detail_data["message"] = msg

        data = request.make(url, 'POST', json=detail_data).json()[0]
        result = data.get("resultLeadOutput", {})
        if result.get("code", 0) == 409: #El mensaje esta repetido
            logger.debug("Mensaje repetido, reenviando")
            # Agregamos un string random al final del mensaje
            msg += "\n\n"+"".join(random.choice(string.digits) for _ in range(10))
            continue

        publisher = data.get("publisherOutput", {})

        if "mailerror" in publisher: #The sender mail is wrong and server return a 500 code
            return None

        logger.success("Publisher contactado con exito")
        break

    return publisher

def extract_post_data(data):
    posts = []
    for p in data:
        publisher  = p.get("publisher", {})
        if publisher == {}:
            logger.error("No se encontro informacion del 'publisher'")
        location  = p.get("postingLocation", {}).get("location", {})
        if location == {}:
            logger.error("El post no tiene 'location'")
        price_data = p.get("priceOperationTypes", [{}])[0].get("prices", [{}])[0]
        if price_data == {}:
            logger.error("El post no tiene 'price_data'")

        post = {
            "fuente": "INMUEBLES24",
            "id":           p.get("postingId", ""),
            "extraction_date": strftime(DATE_FORMAT, gmtime()),
            "message_date": strftime(DATE_FORMAT, gmtime()),
            "title":        p.get("title", ""),
            "price":        price_data.get("formattedAmount", ""),
            "currency":     price_data.get("currency"),
            "type":         p.get("realEstateType", {}).get("name", ""),
            "url":          SITE + p.get("url", ""),
            "bedrooms": "",
            "bathrooms": "",
            "building_size": "",
            "location":     {
                "full":       "",
                "zone":       location.get("name", ""),
                "city":       location.get("parent", {}).get("name", ""),
                "province":   location.get("parent", {}).get("parent", {}).get("name", ""),
            },
            "publisher":    {
                "id":           publisher.get("publisherId", ""),
                "name":         publisher.get("name", ""),
                "whatsapp":     p.get("whatsApp", ""),
                "phone": "",
                "cellPhone": ""
            }
        }

        #features
        features_keys = [
            ("size", "CFT100"),
            ("building_size", "CFT101"),
            ("bedrooms", "CFT2"),
            ("bathrooms", "CFT3"),
            ("garage", "CFT7"),
            ("antiguedad", "CFT5")]

        mainFeatures = p["mainFeatures"]
        for feature, key in features_keys:
            if key in mainFeatures:
                post[feature] = mainFeatures[key]["value"]
        #-------
        posts.append(post)
    return posts

#Get the all the postings in one search
#Esta es la funcion main enrealidad, pero recibe los filters y no la url
def get_postings(filters):
    sheet = Sheet(logger, 'scraper_mapping.json')
    sheets_headers = sheet.get("Extracciones!A1:Z1")[0]
    
    last_page = False
    page = 1
    total = 0

    while not last_page:
        logger.debug(f"Page nro {page}")        
        data = request.make(LIST_URL, 'POST', json=filters).json()

        #Scrape the data from the JSON
        posts = extract_post_data(data.get("listPostings", []))
        if len(posts) == 0:
            logger.error("No se encontro ningun post")
            logger.error(data)
            last_page = True

        row_ads = [] #La lista que se guarda en el google sheets
        total += len(posts)
        for post in posts:
            msg = generate_post_message(post)
            post["message"] = msg
            #La funciona get_publisher envia el mensaje si se lo pasamos
            publisher = get_publisher(post, msg)

            if publisher != None:
                post["publisher"]["phone"] = publisher.get("phone", "")
                post["publisher"]["cellPhone"] = publisher.get("cellPhone", "")

            logger.debug(post)
            
            #Guardamos los leads como filas para el sheets
            row_ad = sheet.map_lead(post, sheets_headers)
            row_ads.append(row_ad)

        #Save the lead in the sheet
        sheet.write(row_ads, "Extracciones!A2")
        page += 1
        filters["pagina"] = page
        last_page = data["paging"]["lastPage"]

    logger.success(f"Se encontraron {total} ads para la url")

def get_filters(url):
    from selenium import webdriver
    from selenium.webdriver.chrome.options import Options
    from selenium.webdriver.common.by import By
    from selenium.webdriver.chrome.options import Options
    from selenium.webdriver.support.ui import WebDriverWait
    from selenium.webdriver.support import expected_conditions as EC
    from selenium.webdriver.common.desired_capabilities import DesiredCapabilities

    options = Options()
    #options.add_argument(f"--headless") #Session
    options.add_argument("--no-sandbox") # Necesario para correrlo como root dentro del container
    caps = DesiredCapabilities.CHROME.copy()
    options.set_capability('goog:loggingPrefs', {'performance': 'ALL'})
    driver = webdriver.Chrome(options=options)

    logger.debug("Obteniendo filtros")
    try:
        driver.get(url)
        WebDriverWait(driver, 10).until(EC.visibility_of_element_located((By.XPATH, "//a[@data-qa='PAGING_2']")))
        next_page = driver.find_element("//a[@data-qa='PAGING_2']")
        next_page.click()

        browser_log = driver.get_log('performance') 
        print(browser_log)

        WebDriverWait(driver, 10).until(EC.visibility_of_element_located((By.XPATH, "//a[@data-qa='PAGING_1']")))
        next_page = driver.find_element("//a[@data-qa='PAGING_1']")

        browser_log = driver.get_log('performance') 
        print(browser_log)
    except Exception as e:
        logger.error("Ocurrio un error obteniendo los filtros")
        logger.error(str(e))
    finally:
        driver.quit()

def main(url: str):
    filters = {
        "ambientesmaximo": 0,
        "ambientesminimo": 0,
        "amenidades": "",
        "antiguedad": None,
        "areaComun": "",
        "areaPrivativa": "",
        "auctions": None,
        "banks": "",
        "banos": None,
        "caracteristicasprop": None,
        "city": None,
        "comodidades": "",
        "condominio": "",
        "coordenates": None,
        "direccion": None,
        "disposicion": None,
        "etapaDeDesarrollo": "",
        "excludePostingContacted": "",
        "expensasmaximo": None,
        "expensasminimo": None,
        "garages": None,
        "general": "",
        "grupoTipoDeMultimedia": "",
        "habitacionesmaximo": 0,
        "habitacionesminimo": 0,
        "idInmobiliaria": None,
        "idunidaddemedida": 1,
        "metroscuadradomax": None,
        "metroscuadradomin": None,
        "moneda": 10,
        "multipleRets": "",
        "outside": "",
        "pagina": 1,
        "places": "",
        "polygonApplied": None,
        "preciomax": None,
        "preciomin": "5000000",
        "province": None,
        "publicacion": None,
        "q": None,
        "roomType": "",
        "searchbykeyword": "",
        "services": "",
        "sort": "relevance",
        "subtipoDePropiedad": None,
        "subZone": None,
        "superficieCubierta": 1,
        "tipoAnunciante": "ALL",
        "tipoDeOperacion": "1",
        "tipoDePropiedad": "1,101,12",
        "valueZone": None,
        "zone": "47735"
    }
    get_postings(filters)
    #url = "https://www.inmuebles24.com/departamentos-en-renta-en-ciudad-de-mexico.html"
    #get_filters(url)
