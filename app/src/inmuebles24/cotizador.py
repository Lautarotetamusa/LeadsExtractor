if __name__ == "__main__":
    import sys
    from dotenv import load_dotenv
    sys.path.append('.')
    load_dotenv()

from bs4 import BeautifulSoup, Tag

import json
import os
import re
import urllib.parse

from time import gmtime, strftime
from src.logger import Logger
from src.make_requests import ApiRequest
from src.inmuebles24.scraper import extract_post_data

ZENROWS_API_URL = "https://api.zenrows.com/v1/"
DATE_FORMAT = "%d/%m/%Y"

logger = Logger("Cotizador Inmuebles24.com")
request = ApiRequest(logger, ZENROWS_API_URL, {
    "apikey": os.getenv("ZENROWS_APIKEY"),
    "url": "",
    # "js_render": "true",
    "antibot": "true",
    "premium_proxy": "true",
    "proxy_country": "mx",
})

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
        assert type(map_url) is str
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
