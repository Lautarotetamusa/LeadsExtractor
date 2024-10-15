if __name__ == "__main__":
    import sys
    from dotenv import load_dotenv
    sys.path.append('.')
    load_dotenv()

import os
from time import gmtime, strftime
from bs4 import BeautifulSoup, Tag 

from make_requests import ApiRequest
from logger import Logger

logger = Logger("EasyBroker")
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
DATE_FORMAT = "%d/%m/%Y"
request = ApiRequest(logger, ZENROWS_API_URL, {
    "apikey": os.getenv("ZENROWS_APIKEY"),
    "url": "",
})

def safe_find(soup: BeautifulSoup, name: str, **kwargs) -> str:
    default = " - "
    container = soup.find(name, **kwargs)
    if type(container) is Tag:
        return container.text.strip()
    else:
        logger.error(f"cannot find {name}{kwargs}")
    return default

def set_attrs(soup: BeautifulSoup, post):
    attrs = {
        "antiguedad":   "fa-calendar",
        "size":         "",
        "building_size": "fa-cube",
        "banios":       "fa-bath",
        "cocheras":     "fa-car",
        "recamaras":    "fa-bed",
    }
    for key in attrs:
        container = soup.find("i", class_=attrs[key])
        if type(container) is not Tag:
            continue

        if container.nextSibling is None:
            continue

        post[key] = container.nextSibling.text.strip()

def get_post_data(url: str) -> dict | None:
    res = request.make(url, "GET")
    if res is None:
        return None
    if not res.ok:
        logger.error(res.text)
        return None

    logger.debug("pagina obtenida con exito" + str(res.status_code))
    soup = BeautifulSoup(res.text, "html.parser")

    price = safe_find(soup, "div", class_="digits")
    title = safe_find(soup, "h1", class_="property-title")
    location = safe_find(soup, "h2", class_="location")
    tipo = safe_find(soup, "div", class_="operation-type")

    post = {
        "id": " - ",
        "extraction_date": strftime(DATE_FORMAT, gmtime()),
        "message_date": strftime(DATE_FORMAT, gmtime()),
        "title":        title,
        "price":        price,
        "currency":     "MXN",
        "type":         tipo,
        "url":          url,
        "location":     {
            "zone":       location,
        },
        "images_urls": extract_images(soup),
        "map_url": extract_ubication_image(soup)
    }
    set_attrs(soup, post)
    print(post)

# TODO: re-escalar las imagenes al tamaño cuadrado
def extract_images(soup: BeautifulSoup):
    images = []
    pictures_divs = soup.find_all("div", class_="picture")
    for div in pictures_divs:
        if div == None:
            continue

        img = div.find("img")
        if type(img) is not Tag:
            logger.error("cannot find 'img' tag in div{class_: picture}")
            continue

        img_src = img.get("src")
        if img_src is None:
            logger.error("cannot find 'src' property in image")
            continue

        images.append(img_src)
    return images

def extract_ubication_image(soup: BeautifulSoup) -> str | None:
    link_a = soup.find("a", class_="copy-clipboard button --link")
    if type(link_a) is not Tag:
        return None

    url = link_a.get("data-clipboard-text")
    if url is None: 
        return None
    return str(url)

if __name__ == "__main__":
    url = "https://www.easybroker.com/mx/listings/departamento-en-venta-en-chapalita-zapopan-chapalita"
    get_post_data(url)
