import string
import random
import json
import os
import re

from src.logger import Logger
from src.message import generate_post_message
from src.make_requests import ApiRequest

site = "https://www.inmuebles24.com"
VIEW_URL = "https://www.inmuebles24.com/rp-api/leads/view"
LIST_URL = "https://www.inmuebles24.com/rplis-api/postings"
CONTACT_URL = "https://www.inmuebles24.com/rp-api/leads/contact"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
SENDER = {
	"email": "ventas.rebora@gmail.com",
	"id": "",
	"message": "",
	"name": "Brenda Diaz",
	"page": "Listado",
	"phone": "3313420733",
	"postingId": "",
	"publisherId": ""
}

logger = Logger("inmuebles24.com")
request = ApiRequest(logger, ZENROWS_API_URL, {
	"apikey": os.getenv("ZENROWS_APIKEY"),
	"url": "",
    "js_render": "true",
    #"antibot": "true",
    #"premium_proxy": "true",
    #"proxy_country": "mx",
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
        logger.debug("Viewing phone")
        publisher = request.make(VIEW_URL, 'POST', json=detail_data).json()[0].get("publisherOutput", {})
        print("publisher: ", publisher)

        if "mailerror" in publisher or publisher == {}:
            return None
        return publisher

    #Send a message and get the phone
    while True:
        detail_data["message"] = generate_post_message(post)

        logger.debug("Contancting publisher")
        publisher = request.make(CONTACT_URL, 'POST', json=detail_data).json().get("publisherOutput", {})
        logger.debug("Message: "+detail_data["message"])

        # SI el publisher es {} es porque ya se envio un mensaje igual a este publisher
        if "mailerror" in publisher or publisher == {}: #The sender mail is wrong and server return a 500 code
            return None

        # Agregamos un string random al final del mensaje
        logger.debug("Mensaje repetido, reenviando")
        msg += "\n\n"+"".join(random.choice(string.digits) for i in range(10))
        return publisher


#Get the all the postings in one search
def get_postings(filters, msg=""):
    last_page = False
    page = 1

    posts = []
    while not last_page:
        logger.debug(f"Page nro {page}")        
        data = request.make(LIST_URL, 'POST', json=filters).json()

        #Scrape the data from the JSON
        #print(data)
        posts = data.get("listPostings", [])
        if len(posts) == 0:
            logger.error("No se encontro ningun post")
            last_page = True
            return posts

        for p in posts:
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
                "id":           p.get("postingId", ""),
                "title":        p.get("title", ""),
                "price":        price_data.get("formattedAmount", ""),
                "currency":     price_data.get("currency"),
                "type":         p.get("realEstateType", {}).get("name", ""),
                "url":          site + p.get("url", ""),
                "location":     {
                    "zona":         location.get("name", ""),
                    "ciudad":       location.get("parent", {}).get("name", ""),
                    "provinicia":   location.get("parent", {}).get("parent", {}).get("name", ""),
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
                ("terreno", "CFT100"),
                ("construido", "CFT101"),
                ("recamaras", "CFT2"),
                ("banios", "CFT3"),
                ("garage", "CFT7"),
                ("antiguedad", "CFT5")]

            mainFeatures = p["mainFeatures"]
            for feature, key in features_keys:
                if key in mainFeatures:
                    post[feature] = mainFeatures[key]["value"]
            #-------

            #Get the publisher data
            #Si no ponemos nada como mensaje solamente se ve el telefono
            #sender = senders[sender_i%len(senders)]
            #sender_i += 1
            publisher = get_publisher(post, msg)

            if publisher != None:
                post["publisher"]["phone"] = publisher["phone"]
                post["publisher"]["cellPhone"] = publisher["cellPhone"]

            dumped = json.dumps(post, indent=4)
            logger.debug(dumped)
            posts.append(post)

        page += 1
        filters["pagina"] = page
        last_page = data["paging"]["lastPage"]

    return posts

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

def main():
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
        "moneda": None,
        "multipleRets": "",
        "outside": "",
        "places": "",
        "polygonApplied": None,
        "preciomax": None,
        "preciomin": None,
        "province": "69",
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
        "tipoDeOperacion": "2",
        "tipoDePropiedad": "2",
        "valueZone": None,
        "zone": None
    }
    #get_postings(filters)
    url = "https://www.inmuebles24.com/departamentos-en-renta-en-ciudad-de-mexico.html"
    get_filters(url)

if __name__ == '__main__':
    main()
