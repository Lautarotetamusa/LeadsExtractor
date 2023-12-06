import os
import json
from bs4 import BeautifulSoup
from time import gmtime, strftime

from src.logger import Logger
from src.sheets import Sheet
from src.make_requests import ApiRequest
from src.message import generate_post_message

SITE = "https://www.casasyterrenos.com"
PROPS_URL = f"{SITE}/api/search/"
URL_SEND = ""
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
DATE_FORMAT = "%d/%m/%Y"
SENDER = {
    "email": "ventas.rebora@gmail.com",
    "name": "Brenda Diaz",
    "phone": "3313420733",
}
logger = Logger("scraper casasyterrenos.com")

request = ApiRequest(logger, ZENROWS_API_URL, {
    "apikey": os.getenv("ZENROWS_APIKEY"),
    "url": "",
})

# Esta funcion extrae la data que nos sirve de la api
def extract_posts(data: list[dict]):
    posts = []
    for post in data:
        posts.append({
            "fuente": "CASASYTERRENOS",
            "id": post.get("id", ""),
            "title": post.get("name", ""),
            "extraction_date": strftime(DATE_FORMAT, gmtime()),
            "message_date": strftime(DATE_FORMAT, gmtime()),
            "price": post.get("priceSale", ""),
            "currency": post.get("currency", ""),
            "type": post.get("type", ""),
            "url": SITE + post.get("canonical", ""),
            "bedrooms": post.get("rooms", ""), 
            "bathrooms": post.get("bathrooms", ""),
            "building_size": post.get("construction", ""),
            "parkings": post.get("parkingLots", ""),
            "location": {
                "full": ", ".join([post.get("municipality", ""), post.get("neighborhood", ""), post.get("state", "")]),
                "zone": post.get("neighborhood"),
                "city": post.get("municipality", ""),
                "province": post.get("state", ""), 
            },
            "publisher":{
                "name": "",
                "id": "",
                "whatsapp": "",
                "phone": "",
                "cellPhone": ""
            }
        })
    return posts

def get_publisher_info(property_id):
    logger.debug(f"Obteniendo informacion de la propiedad {property_id}")
    url = f"https://www.casasyterrenos.com/_next/data/9A9l_hFzRC-L7dViWP5UL/propiedad/{property_id}.json?id={property_id}"
    
    res = request.make(url, "GET")
    data = res.json().get("pageProps", {}).get("property", {}).get("seller")
    if not data:
        return None
    publisher = {
        "name": data.get("first_name", "") + data.get("last_name", ""),
        "id": data.get("id", ""),
        "whatsapp": data.get("whatsapp", ""),
        "phone": data.get("phone_number", ""),
        "email": data.get("email", ""),
        "cellPhone": data.get("cel", "") 
    }
    print(publisher)
    return publisher

def main(url: str):
    payload = {
        "filters": {
            "bathrooms": 0,
            "fromConstruction": "0",
            "fromPrice": 0,
            "fromSurface": "0",
            "onlyNewDevelopments": False,
            "parkingLots": 0,
            "propertyType": "Casa",
            "rooms": "0",
            "toConstruction": 1000000000,
            "toPrice": 1000000000,
            "toSurface": 1000000000,
            "transactionType": "venta"
        },
        "orderBy": "Más recientes",
        "searchConfig": {
            "hitsPerPage": 100,
            "page": 1
        },
        "searchObject": {
            "searchBox": [
                19.5927571,
                -98.9604482,
                19.1887101,
                -99.3267771
            ],
            "searchCenter": {
                "latitude": 19.3907336,
                "longitude": -99.14361265
            },
            "searchWidth": 44978.30629654964
        }
    } 

    sheet = Sheet(logger, 'scraper_mapping.json')
    sheets_headers = sheet.get("Extracciones!A1:Z1")[0]

    page = 1
    len_posts = 0
    total_posts = 1e9

    #url = f"https://propiedades.com/df/casas-venta/recamaras-2?pagina={page}"
    while len_posts < total_posts:
        logger.debug(f"Page: {page}")
        res = request.make(PROPS_URL, "POST", json=payload)
        data = res.json()

        if total_posts  == 1e9:
            total_posts = data.get("nbHits", 0)
            logger.debug(f"Total posts: {total_posts}")
        posts = extract_posts(data.get("hits", []))

        page += 1
        len_posts += len(posts)
        payload["searchConfig"]["page"] = page

        row_ads = []
        for ad in posts:
            publisher = get_publisher_info(ad["id"])
            if publisher == None:
                logger.error(f"Ocurrio un error encontrando la informacion de la propiedad {ad['id']}")
            else:
                ad["publisher"] = publisher
            msg = generate_post_message(ad)
            ad["message"] = msg
            #Guardamos los leads como filas para el sheets
            row_ad = sheet.map_lead(ad, sheets_headers)
            row_ads.append(row_ad)

        #Save the lead in the sheet
        sheet.write(row_ads, "Extracciones!A2")
    logger.success(f"Se encontraron un total de {len_posts} en la url especificada")
