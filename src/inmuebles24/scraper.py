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
CONTACT_URL =  f"{SITE}/rp-api/leads/contact"
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
    cookies = {
		"__cf_bm": "hoDyQ_mtImkdY5HD29Se6J2Aqv_Yz44dNM5FJh0n9SI-1710160201-1.0.1.1-tFEriZezq30JgcOPw135QgEzZV3.OXKocDieHa_VY14FKL1celv9Om5.o81Ae2WyJoWmW8rF3tXUxSP.t2oOKcxibwb7e98sqSGbfJ_Y_1k",
		"_ga": "GA1.1.666477216.1710160203",
		"_ga_8XFRKTEF9J": "GS1.1.1710160203.1.1.1710160222.41.0.0",
		"_gcl_au": "1.1.894966547.1710160202",
		"cf_clearance": "Q7RkPoggohxEnA5JIdYYfyYZyHUQ3uMC_im1hbjItzM-1710160201-1.0.1.1-YmcpkWwDvzKyb3a54VxdeB8VmSjpcb9OQZxLWbrumJ2vQ4NQDwrTeiaTWfXhPlpVTVrPJ9nWKKsy7Si_dSi7Pw",
		"g_state": "{\"i_p\":1710167409819,\"i_l\":1}",
		"JSESSIONID": "FCB57AE475324D4D0F664D11D17B678F",
		"mousestats_si": "0e23e2327c539da7d990",
		"mousestats_vi": "fe5d6cc48141971925a4",
		"sessionId": "6898964a-b6d4-44fc-af0b-07306ebdf91a"
    }
    headers = {
        "User-Agent": "Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0"
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

        res = request.make(url, 'POST', json=detail_data, cookies=cookies, headers=headers)
        if res == None: 
            logger.error("No se pudo enviar mensaje al post")
            return None
        logger.debug(res.json())
        data = res.json()[0]

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
def get_postings(filters: dict, spin_msg: str):
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
            msg = generate_post_message(post, spin_msg)
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

def main(filters: dict, spin_msg):
    print(type(filters))
    get_postings(filters, spin_msg)
    #url = "https://www.inmuebles24.com/departamentos-en-renta-en-ciudad-de-mexico.html"
    #get_filters(url)

    # filters = {
    #     "ambientesmaximo": 0,
    #     "ambientesminimo": 0,
    #     "amenidades": "",
    #     "antiguedad": null,
    #     "areaComun": "",
    #     "areaPrivativa": "",
    #     "auctions": null,
    #     "banks": "",
    #     "banos": null,
    #     "caracteristicasprop": null,
    #     "city": null,
    #     "comodidades": "",
    #     "condominio": "",
    #     "coordenates": null,
    #     "direccion": null,
    #     "disposicion": null,
    #     "etapaDeDesarrollo": "",
    #     "excludePostingContacted": "",
    #     "expensasmaximo": null,
    #     "expensasminimo": null,
    #     "garages": null,
    #     "general": "",
    #     "grupoTipoDeMultimedia": "",
    #     "habitacionesmaximo": 0,
    #     "habitacionesminimo": 0,
    #     "idInmobiliaria": null,
    #     "idunidaddemedida": 1,
    #     "metroscuadradomax": null,
    #     "metroscuadradomin": null,
    #     "moneda": 10,
    #     "multipleRets": "",
    #     "outside": "",
    #     "pagina": 1,
    #     "places": "",
    #     "polygonApplied": null,
    #     "preciomax": null,
    #     "preciomin": "5000000",
    #     "province": null,
    #     "publicacion": null,
    #     "q": null,
    #     "roomType": "",
    #     "searchbykeyword": "",
    #     "services": "",
    #     "sort": "relevance",
    #     "subtipoDePropiedad": null,
    #     "subZone": null,
    #     "superficieCubierta": 1,
    #     "tipoAnunciante": "ALL",
    #     "tipoDeOperacion": "1",
    #     "tipoDePropiedad": "1,101,12",
    #     "valueZone": null,
    #     "zone": "47735"
    # }
