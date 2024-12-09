if __name__ == "__main__":
    import sys
    from dotenv import load_dotenv
    sys.path.append('.')
    load_dotenv()

import json
import string
import random
import os

from time import gmtime, strftime

from src.scraper import Scraper
from src.logger import Logger
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
    # "js_render": "true",
    # "antibot": "true",
    "premium_proxy": "true",
    "proxy_country": "mx",
    "original_status": True
})

class Inmuebles24Scraper(Scraper):
    def __init__(self):
        self.name = "inmuebles24"

    def get_posts(self, param):
        assert type(param) is dict
        filters = param
        last_page = False
        page = 1
        total = 0

        while not last_page:
            logger.debug(f"Page nro {page}")
            print(json.dumps(filters, indent=4)) 
            res = request.make(LIST_URL, 'POST', json=filters)
            # res = requests.post(LIST_URL, json=filters)
            if res is None: break
            data = res.json()

            posts = []
            for post_data in data.get("listPostings", []):
                posts.append(extract_post_data(post_data))
            if len(posts) == 0:
                logger.error("No se encontro ningun post")
                logger.error(data)
                last_page = True

            total += len(posts)
            yield posts
            page += 1
            filters["pagina"] = page
            last_page = data["paging"]["lastPage"]

        logger.success(f"Se encontraron {total} ads para la url")

    def send_message(self, msg: str, post: dict):
        self.get_publisher(post, msg)
        pass

    def view_phone(self, post) -> str:
        self.get_publisher(post)
        return ""

    def get_publisher(self, post: dict, msg=""):
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

if __name__ == "__main__":
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
        "city": "778",
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
        "moneda": "",
        "multipleRets": "",
        "outside": "",
        "pagina": 1,
        "places": "",
        "polygonApplied": None,
        "preciomax": None,
        "preciomin": None,
        "preTipoDeOperacion": "",
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
        "tipoDeOperacion": "2",
        "tipoDePropiedad": "2",
        "valueZone": None,
        "withoutguarantor": None,
        "zone": None
    }
    msg = "Hola!"

    scraper = Inmuebles24Scraper()
    scraper.main(msg, filters)
