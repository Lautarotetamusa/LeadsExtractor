import os
from time import gmtime, strftime

from src.logger import Logger
from src.sheets import Sheet
from src.make_requests import ApiRequest
from src.message import generate_post_message

SITE = "https://www.casasyterrenos.com"
URL_SEND = f"{SITE}/api/amp/register"
URL_INFO = "https://www.casasyterrenos.com/_next/data/9A9l_hFzRC-L7dViWP5UL/propiedad/{property_id}.json?id={property_id}"
URL_PROPS = f"{SITE}/api/search/"
URL_SESSION = f"{SITE}/api/amp/auth"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"

DATE_FORMAT = "%d/%m/%Y"
SENDER = {
    "email": os.getenv("SENDER_EMAIL"),
    "name":  os.getenv("SENDER_NAME"),
    "phone": os.getenv("SENDER_PHONE"),
    "contact_place": "casasyterrenos.com",
    "isDev": False,
    "isProp": True,
    "isProto": False,
    "message": "",
    "property": ""
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
    url = URL_INFO.format(property_id=property_id)
    
    res = request.make(url, "GET")
    data = res.json().get("pageProps", {}).get("property", {}).get("seller")
    if not data:
        return None
    publisher = {
        "name": data.get("first_name", "") + data.get("last_name", ""),
        "id": data.get("client", ""),
        "whatsapp": data.get("whatsapp", ""),
        "phone": data.get("phone_number", ""),
        "email": data.get("email", ""),
        "cellPhone": data.get("cel", "") 
    }
    print(publisher)
    return publisher

def generate_session_id() -> str:
    logger.success("Generando nuevo session id")
    data = {
        "domain": SITE,
        "referrer": ""
    }
    
    res = request.make(URL_SESSION, "POST", json=data).json()
    session_id = res.get("uuid")
    if not session_id:
        logger.error("Error generando el session id, saliendo")
        exit(1)

    logger.success("Session id generado con exito")
    logger.success("session_id: "+session_id)
    return session_id

def send_message(property_id: int, client_id: int, session_id: str, message: str):
    data = SENDER.copy()
    data = {
        "client_id": client_id,
        "distinct_id": f"prp_{property_id}",
        "event_type": "submit",
        "event_value": "contact_form",
        "meta": {
            "client_id": client_id,
            "contact_place": "casasyterrenos.com",
            "email": "ventas.rebora@gmail.com",
            "isDev": False,
            "isProp": True,
            "isProto": False,
            "message": message,
            "name": "Brenda Diaz",
            "phone": "3313420733",
            "property": int(property_id),
            "type": "property"
        },
        "referrer": "",
        "session_uuid": session_id
    }    
    res = request.make(URL_SEND, "POST", json=data)
    if res.status_code != 200:
        logger.error("Error enviando el mensaje")
        logger.error(res.text)
        return None

    logger.success(f"Mensaje enviado con exito a la propiedad {property_id}")
    logger.debug(res.text)

def main(filters: dict, spin_msg: str):
    sheet = Sheet(logger, 'scraper_mapping.json')
    sheets_headers = sheet.get("Extracciones!A1:Z1")[0]

    page = 1
    len_posts = 0
    total_posts = 1e9
    session_id = generate_session_id()

    #url = f"https://propiedades.com/df/casas-venta/recamaras-2?pagina={page}"
    while len_posts < total_posts:
        logger.debug(f"Page: {page}")
        res = request.make(URL_PROPS, "POST", json=filters)
        data = res.json()

        if total_posts  == 1e9:
            total_posts = data.get("nbHits", 0)
            logger.debug(f"Total posts: {total_posts}")
        #print(data)
        posts = extract_posts(data.get("hits", []))

        page += 1
        len_posts += len(posts)
        filters["searchConfig"]["page"] = page

        row_ads = []
        for ad in posts:
            publisher = get_publisher_info(ad["id"])
            if publisher == None:
                logger.error(f"Ocurrio un error encontrando la informacion de la propiedad {ad['id']}")
            else:
                ad["publisher"] = publisher
            msg = generate_post_message(ad, spin_msg)
            send_message(ad["id"], ad["publisher"]["id"], session_id, msg)
            ad["message"] = msg
            #Guardamos los leads como filas para el sheets
            row_ad = sheet.map_lead(ad, sheets_headers)
            row_ads.append(row_ad)

        #Save the lead in the sheet
        sheet.write(row_ads, "Extracciones!A2")
    logger.success(f"Se encontraron un total de {len_posts} en la url especificada")
