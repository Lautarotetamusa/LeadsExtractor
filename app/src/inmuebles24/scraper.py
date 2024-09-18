if __name__ == "__main__":
    import sys
    from dotenv import load_dotenv
    sys.path.append('.')
    load_dotenv()

import datetime
from multiprocessing.pool import ThreadPool
from bs4 import BeautifulSoup, Tag

import string
import random
import json
import os
import re
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
    # "js_render": "true",
    "antibot": "true",
    "premium_proxy": "true",
    "proxy_country": "mx",
})

def get_publisher(post: dict, msg=""):
    detail_data = {
        "email": SENDER["email"],
        "name":  SENDER["name"],
        "phone": SENDER["phone"],
        "page": "Listado",
        "publisherId": post["publisher"]["id"],
        "postingId": post["id"]
    }
    cookies = {
        "sessionId": "6898964a-b6d4-44fc-af0b-07306ebdf91a"
    }
    headers = {
        "User-Agent": "Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0"
    }

    if msg == "":
        # View the phone but not send message
        url = VIEW_URL
        logger.debug(f"Viendo telefono del publisher {post['publisher']['id']}")
    else:
        logger.debug(f"Enviando mensaje a publisher {post['publisher']['id']}")
        url = CONTACT_URL

    # Send a message or get the phone
    while True:
        if msg != "":
            detail_data["message"] = msg

        res = request.make(url, 'POST', json=detail_data, cookies=cookies, headers=headers)
        if res is None:
            logger.error("No se pudo enviar mensaje al post")
            return None
        logger.debug(res.json())
        data = res.json()[0]

        result = data.get("resultLeadOutput", {})
        if result.get("code", 0) == 409:  # El mensaje esta repetido
            logger.debug("Mensaje repetido, reenviando")
            # Agregamos un string random al final del mensaje
            msg += "\n\n"+"".join(random.choice(string.digits) for _ in range(10))
            continue

        publisher = data.get("publisherOutput", {})

        if "mailerror" in publisher:  # The sender mail is wrong and server return a 500 code
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

    # features
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


def safe_find(soup: BeautifulSoup, name: str, **args):
    tag = soup.find(name, **args)
    if tag is None:
        raise KeyError(str(**args))

    return tag

# Extraer los datos de la propiedad através del link a la propiedad directamente
# Este link no anda: https://www.inmuebles24.com/propiedades/desarrollo/ememvein-ri-a-americas-143720777.html
def get_post_data(url: str) -> dict | None:
    request.api_params['js_render'] = 'true'
    # Lo agregamos para poder opbtener la imagen con la ubicacion
    request.api_params["js_instructions"] = """[{"wait":1500},{"scroll_y":400}]"""
    res = request.make(url, "GET")
    if 'js_render' in request.api_params:       # Como es mutlithreading aveces se bugeaba. TODO: fix
        del request.api_params['js_render']
    if 'js_instructions' in request.api_params:  # Como es mutlithreading aveces se bugeaba. TODO: fix
        del request.api_params['js_instructions']
    if res is None:
        return
    soup = BeautifulSoup(res.text, "html.parser")

    images = []
    MAX_IMAGES = 4
    DEFAULT_DIMENSIONS = "360x266"
    gallery = soup.find("div", id="new-gallery-portal")
    dimensions_regex = re.compile(r'\d+x\d+')
    if type(gallery) is Tag:
        for img in gallery.find_all("img"):
            if len(images) >= MAX_IMAGES:
                break
            # https://img10.naventcdn.com/avisos-va/vamx-pt10-ads/ad/1200x1200/ad4827e9-3c44-4558-b5f2-3e5a3de42f1c?isFirstImage=true
            src = dimensions_regex.sub(DEFAULT_DIMENSIONS, img["src"])
            src = src.replace("resize/", "")
            images.append(src)
    else:
        logger.error("cannot find gallery div")
        return

    map_url = ""
    map_container = soup.find("img", class_="static-map")
    if type(map_container) is Tag:
        map_url = map_container["src"][2:]
        map_url = "https://" + map_url

    attrs = {
        "antiguedad":   "icon-antiguedad",
        "size":         "icon-stotal",
        "building_size": "icon-scubierta",
        "banios":       "icon-bano",
        "cocheras":     "icon-cochera",
        "recamaras":    "icon-dormitorio",
    }

    title = " - "
    title_cont = soup.find("h1", class_="title-property")
    if type(title_cont) is Tag:
        title = title_cont.text
    else:
        logger.error("cannot find h1{title-property}")

    id = " - "
    id_tag = soup.find("section", id="reactPublisherCodes")
    if type(id_tag) is Tag:
        a1 = id_tag.find_all("span")
        if len(a1) > 0:
            id = a1[1].next_sibling.text.strip()
        else:
            logger.error("cannot find section{reactPublisherCodes.span}")
    else:
        logger.error("cannot find section{reactPublisherCodes}")

    price = " - "
    tipo = " - "
    price_tag = soup.find("div", class_="price-value")
    if type(price_tag) is Tag:
        a1 = price_tag.find("span")
        if type(a1) is Tag:
            tipo_str = a1.text.split(" ")
            if len(tipo_str) > 0:
                tipo = tipo_str[0]
            a2 = a1.find("span")
            if type(a2) is Tag:
                price = a2.text.strip()
            else:
                logger.error("cannot find {price-value.span.span}")
        else:
            logger.error("cannot find {price-value.span}")
    else:
        logger.error("cannot find {price-value}")

    zone = " - "
    zone_tag = soup.find("section", id="map-section")
    if type(zone_tag) is Tag:
        a1 = zone_tag.find("h4")
        if type(a1) is Tag:
            zone = a1.text
        else:
            logger.error("cannot find section{map-section.h4}")
            return
    else:
        logger.error("cannot find section{map-section}")
        return

    post = {
        "id": sanitaze_str(id),
        "extraction_date": strftime(DATE_FORMAT, gmtime()),
        "message_date": strftime(DATE_FORMAT, gmtime()),
        "title":        sanitaze_str(title),
        "price":        sanitaze_str(price),
        "currency":     "",
        "type":         sanitaze_str(tipo),
        "url":          url,
        "location":     {
            "full":       "",
            "zone":       sanitaze_str(zone),
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
        if key == "cocheras":
            post[key] += " cocheras"

    return post


# Get the all the postings in one search
def get_postings(filters: dict, spin_msg: str):
    last_page = False
    page = 1
    total = 0

    while not last_page:
        logger.debug(f"Page nro {page}")
        data = request.make(LIST_URL, 'POST', json=filters).json()

        # Scrape the data from the JSON
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
    if geo is None:
        print("No se encontro informacion de la geolocalizacion para la propiedad")
        return None
    # markers=19.363745800000000,-99.279810700000000
    query["markers"] = [f"{geo.get('latitude')},{geo.get('longitude')}"]
    query.pop("signature")  # Generamos una nueva par aque google acepte la peticion

    url = url._replace(query=urllib.parse.urlencode(query, doseq=True))
    return url.geturl()


def main(filters: dict, spin_msg):
    sheet = Sheet(logger, 'scraper_mapping.json')
    sheets_headers = sheet.get("Extracciones!A1:Z1")[0]

    row_ads = []  # La lista que se guarda en el google sheets
    for post in get_postings(filters, spin_msg):
        msg = generate_post_message(post, spin_msg)
        post["message"] = msg.replace('\n', '')
        publisher = get_publisher(post, msg)

        if publisher is not None:
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


def upload_images(form_id: str, submission_id: str, urls: list[str], qids: list[str]):
    pool = ThreadPool(processes=20)

    results = []
    nro = 1
    for url, qid in zip(urls, qids):
        logger.debug("Obteniendo imagen: " + url)
        img_data = None
        if "maps" in url:
            res = requests.get(url)
            if res is None or not res.ok:
                logger.error("Imposible obtener la imagen de ubicacion: "+ url)
                logger.error(res.text)
                img_data = None
            else:
                img_data = res.content
        else:
            img_data = jotform.get_img_data(url)

        if img_data is None:
            logger.error("Imposible obtener la imagen: "+ url)
            continue

        r = pool.apply_async(
                jotform.upload_image,
                args=(form_id, submission_id, qid, img_data, f"{submission_id}_{qid}_{nro}", )
            )
        results.append(r)
        nro += 1

    for r in results:
        err = r.get()
        if err is None:
            logger.success("Imagen subida correctamente")
        else:
            logger.error("No fue posible subir la image: " + str(err))


def cotizacion_post(post, form_id, asesor, cliente, id_propuesta):
    map_url = post["map_url"]
    images_urls = post["images_urls"]

    logger.debug("Uploading Cotizacion Form")
    res = jotform.submit_cotizacion_form(logger, form_id, post, asesor, cliente, id_propuesta)
    if res is None:
        logger.error("No fue posible subir la cotizacion a jotform")
        return None

    submission_id = res["content"]["submissionID"]
    images_qids = ["77", "44", "44", "44", "44"]  # TODO: No hardcodear
    upload_images(form_id, submission_id, [map_url] + images_urls, images_qids)

    res = jotform.obtain_pdf(form_id, submission_id)
    if res is None:
        return None

    pdf_path = f"./src/inmuebles24/pdfs/{submission_id}.pdf"
    with open(pdf_path, 'wb') as f:
        f.write(res)

    return pdf_path


# Genera cotizaciones en pdf para los postings en la lista
def cotizacion(asesor: dict, cliente: str, posts: list[dict]):
    form_id = "242244461116044"  # TODO: No hardcodear
    pdfs = []
    results = []
    pool = ThreadPool(processes=8)

    id_cotizacion = ''.join(random.choices(string.ascii_letters, k=10))

    orden = 1
    for post in posts:
        post["orden"] = orden
        r = pool.apply_async(cotizacion_post, args=(post, form_id, asesor, cliente, id_cotizacion, ))
        results.append(r)
        orden += 1

    for r in results:
        pdf_path = r.get()
        if pdf_path is not None:
            pdfs.append(pdf_path)

    str_date = datetime.datetime.today().strftime("%d-%m-%Y")
    file_name = f"Propuesta terrenos {cliente} {id_cotizacion} {str_date}.pdf"
    combine_pdfs(pdfs, f"/app/pdfs/{file_name}")
    return file_name


def posts_from_list(res) -> list[dict]:
    posts_count = 0
    max_posts = 3
    posts = []
    for post in res.get("listPostings", []):
        if posts_count >= max_posts:
            break
        posts_count += 1
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

    # cotizacion(asesor, res)
    url = "https://www.inmuebles24.com/propiedades/clasificado/alclcain-residencia-en-renta-en-colinas-de-san-javier-de-lujo-141920496.html"
    post = get_post_data(url)
    print(json.dumps(post, indent=4))

    # Link REBORA
    # https://wa.me/5213328092850?text=Me%20interesa%20este%20terreno%20(ID: {id_terreno})%20de%20la%20propuesta:%20(ID: {id})
