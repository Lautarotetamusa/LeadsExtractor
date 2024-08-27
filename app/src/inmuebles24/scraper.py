if __name__ == "__main__":
    import sys
    from dotenv import load_dotenv
    sys.path.append('.')
    load_dotenv()

from bs4 import BeautifulSoup, Tag

import string
import random
import json
import os
import urllib.parse

from time import gmtime, strftime
from pypdf import PdfMerger

import requests

from src.logger import Logger
from src.sheets import Sheet
from src.message import generate_post_message
from src.make_requests import ApiRequest
import src.jotform as jotform

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
    "antibot": "true",
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

def extract_post_data(p: dict) -> dict:
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
        else:
            post[feature] = ""
    return post

def sanitaze_str(text: str) -> str:
    return text.strip().replace("\n", "").replace("\t", "")

# Extraer los datos de la propiedad através del link a la propiedad directamente
def get_post_data(url: str) -> dict | None:
    request.api_params['js_render'] = 'true'
    # Lo agregamos para poder opbtener la imagen con la ubicacion
    request.api_params["js_instructions"] =  """[{"wait":1000},{"scroll_y":400}]"""
    res = request.make(url, "GET") 
    del request.api_params['js_render']
    del request.api_params['js_instructions']
    if res is None:
        return
    soup = BeautifulSoup(res.text, "html.parser")

    images = []
    images_containers = soup.find("div", id="new-gallery-portal").find_all("img")
    for img in images_containers:
        images.append(img["src"])

    map_url = ""
    map_container = soup.find("img", class_="static-map")
    if type(map_container) is Tag:
        map_url = map_container["src"][2:]
        map_url = "https://" + map_url

    attrs = {
        "antiguedad":   "icon-antiguedad",
        "size":         "icon-stotal",
        "building_size": "icon-scubierta",
    }

    post = {
        "extraction_date": strftime(DATE_FORMAT, gmtime()),
        "message_date": strftime(DATE_FORMAT, gmtime()),
        "title":        sanitaze_str(soup.find("h1", class_="title-property").text),
        "price":        sanitaze_str(soup.find("div", class_="price-value").find("span").find("span").text.strip()),
        "currency":     "",
        "type":         "",
        "url":          url,
        "location":     {
            "full":       "",
            "zone":       sanitaze_str(soup.find("section", id="map-section").find("h4").text),
            "city":       "",
            "province":   "",
        },
        "images_urls": images,
        "map_url": map_url
    }

    for key in attrs:
        container = soup.find("i", class_=attrs[key])
        if type(container) is not Tag:
            continue

        if container.nextSibling is None:
            continue

        post[key] = sanitaze_str(container.nextSibling.text.strip())

    return post

# Get the all the postings in one search
def get_postings(filters: dict, spin_msg: str):
    last_page = False
    page = 1
    total = 0

    while not last_page:
        logger.debug(f"Page nro {page}")        
        data = request.make(LIST_URL, 'POST', json=filters).json()

        #Scrape the data from the JSON
        posts = []
        for post_data in data.get("listPostings", []):
            posts.append(extract_post_data(post_data))
        if len(posts) == 0:
            logger.error("No se encontro ningun post")
            logger.error(data)
            last_page = True

        total += len(posts)
        for post in posts:
            yield post
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

def extract_images(post_data: dict):
    pictures = post_data.get("visiblePictures", {}).get("pictures", [])
    if len(pictures) == 0:
        print("No se encontraron imagenes para la publicacion")

    images_urls = []
    for picture_data in pictures:
        images_urls.append(picture_data.get("url730x532", "")) 

    return images_urls

def extract_ubication_image(post_data: dict) -> str | None:
    location_data = post_data.get("postingLocation", {}).get("postingGeolocation", {})
    print("ld", location_data)
    if location_data is None:
        print("No se encontro un mapa para la propiedad")
        return None

    url_str = location_data.get("urlStaticMap")
    if url_str is None:
        print("No se encontro un mapa para la propiedad")
        return None

    # Las urls comienzan con //. lo sacamos
    url_str = url_str[2:]
    url_str = "https://" + url_str

    url = urllib.parse.urlparse(url_str)
    query = urllib.parse.parse_qs(url.query)

    geo = location_data.get("geolocation")
    if geo == None:
        print("No se encontro informacion de la geolocalizacion para la propiedad")
        return None
    #markers=19.363745800000000,-99.279810700000000
    query["markers"] = [f"{geo.get('latitude')},{geo.get('longitude')}"]
    query.pop("signature") # Generamos una nueva par aque google acepte la peticion

    url = url._replace(query=urllib.parse.urlencode(query, doseq=True))
    return url.geturl()

def main(filters: dict, spin_msg):
    sheet = Sheet(logger, 'scraper_mapping.json')
    sheets_headers = sheet.get("Extracciones!A1:Z1")[0]
    
    row_ads = [] #La lista que se guarda en el google sheets
    for post in get_postings(filters, spin_msg):
        msg = generate_post_message(post, spin_msg)
        post["message"] = msg.replace('\n', '')
        publisher = get_publisher(post, msg)

        if publisher != None:
            post["publisher"]["phone"] = publisher.get("phone", "")
            post["publisher"]["cellPhone"] = publisher.get("cellPhone", "")
        
        row_ad = sheet.map_lead(post, sheets_headers)
        row_ads.append(row_ad)

    sheet.write(row_ads, "Extracciones!A2")

def combine_pdfs(pdfs: list[str], file_name):
    merger = PdfMerger()

    for pdf in pdfs:
        merger.append(pdf)

    merger.write(file_name)
    merger.close()

# Genera cotizaciones en pdf para los postings en la lista
def cotizacion(asesor: dict, posts: list[dict]):
    form_id = "242244461116044"  # TODO: No hardcodear
    pdfs = []
    max_images = 3

    for post in posts:
        map_url = post["map_url"]
        images_urls = post["images_urls"]

        res = jotform.submit_cotizacion_form(logger, form_id, post, asesor)
        if res is None:
            logger.error("No fue posible subir la cotizacion a jotform")
            continue

        submission_id = res["content"]["submissionID"]
        logger.debug("Obteniendo imagen: " + map_url)
        res = requests.get(map_url)
        if res is None or not res.ok:
            logger.error("Imposible obtener la imagen: "+ map_url)
        else:
            # 70 es el qid del campo map img
            err = jotform.upload_image(form_id, submission_id, "70", res.content, "map")
            if err is None:
                logger.success("Imagen ubicacion subida correctamente")
            else: 
                logger.error("No fue posible subir la imagen de la ubicacion" + str(err))

        img_count = 0
        for image_url in images_urls:
            if img_count >= max_images:
                break
            img_count += 1
            logger.debug("Uploading: " + image_url)
            img_data = jotform.get_img_data(image_url)
            if img_data is None:
                logger.error("Imposible de obtener la imagen: "+ image_url)
                continue
            # 41 es el qid del campo de las imagenes
            err = jotform.upload_image(form_id, submission_id, "41", img_data, "property")
            if err is None:
                logger.success("Imagen subida correctamente")
            else: 
                logger.error("No fue posible subir la imagen: " + str(err))

        logger.debug("Generating PDF")
        res = jotform.generate_pdf(form_id, submission_id)
        if res != None and res.get("content", "") != "":
            logger.success("PDF Generado con exito")
            logger.success(res.get("content"))
            res = requests.get(res.get("content"))

            pdf_path = f"./src/inmuebles24/pdfs/${submission_id}.pdf" 

            with open(pdf_path, 'wb') as f:
                f.write(res.content)

            pdfs.append(pdf_path)

        else:
            logger.error("No se pudo generar el PDF")
            logger.error(res)

    combine_pdfs(pdfs, "/app/pdfs/result.pdf")

def posts_from_list(res) -> list[dict]:
    posts_count = 0
    max_posts = 3
    posts = []
    for post in res.get("listPostings", []):
        if posts_count >= max_posts:
            break
        posts_count += 1
        #print(json.dumps(post, indent=4))
        post_data = extract_post_data(post) 

        post_data["images_urls"] = extract_images(post)
        post_data["map_url"] = extract_ubication_image(post)

        if post_data["map_url"] is None:
            logger.error("No fue obtener la imagen del mapa")
            continue

        posts.append(post_data)
    return posts

if __name__ == "__main__":
    res = []
    max_posts = 6

    asesor = {
        "name": "Brenda Irene Diaz Castillo",
        "phone": "3313420733",
        "email": "brenda.diaz@rebora.com.mx"
    }

    #cotizacion(asesor, res)
    url = "https://www.inmuebles24.com/propiedades/clasificado/alclapin-departamento-a-la-renta-en-el-country-club-60428488.html"
    post = get_post_data(url)
    print(json.dumps(post, indent=4))
