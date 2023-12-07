import bs4
import requests
import json
from time import gmtime, strftime
import os
import re

from src.logger import Logger
from src.sheets import Sheet
from src.make_requests import ApiRequest
from src.message import generate_post_message

PROPS_URL = "https://propiedades.com/properties/filtrar"
URL_SEND  = "https://propiedades.com/messages/send"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
DATE_FORMAT = "%d/%m/%Y"
SENDER = {
    "email": os.getenv("SENDER_EMAIL"),
    "name":  os.getenv("SENDER_NAME"),
    "phone": os.getenv("SENDER_PHONE"),
}
logger = Logger("scraper propiedades.com")
request = ApiRequest(logger, ZENROWS_API_URL, {
    "apikey": os.getenv("ZENROWS_APIKEY"),
    "url": "",
    "js_render": "true",
    "antibot": "true",
    "premium_proxy": "true",
    "proxy_country": "mx",
})


#Get the list of properties in all pages of a search
#Send the messages to all properties
#(Payload for the search, msg to send, proxy list, number of pages to scrape)
def get_properties(payload: dict, msg="", npages=0):
    properties = []

    page = payload["page"]
    totalpages = page + npages
    firstpage = (npages == 0) #If npages == 0, get the total of pages, else only scrap this number of pages and not get it

    while payload["page"] <= totalpages:
        logger.debug("Page nro: "+str(payload["page"]))
        res = requests.post(PROPS_URL, payload)
        print(res.text)
        print(res.status_code)
        data = res.json()
        ads = data["markers"]

        print(len(ads), "ads found")

        if firstpage:
            totalpages = data["paginate"]["pages"]
            print("Total pages: ", totalpages)
            firstpage = False

        for id in ads:
            property = {
            	"id": id,
            	"title":	ads[id]["alt_thumbnail"],
            	"adress": 	ads[id]["full_address"],
            	"city": 	ads[id]["sepomex"]["city"],
            	"price": 	ads[id]["price_format"],
            	"purpose": 	ads[id]["purpose"],
            	"type": 	ads[id]["type_str"],
            	"bathrooms":ads[id]["bathrooms"],
            	"bedrooms": ads[id]["bedrooms"],
            	"terreno": 	ads[id]["size_ground"],
            	"piso": 	ads[id]["ground_unit"],
            	"url": 		ads[id]["property_url"],
            	"highlighted": ads[id]["highlighted"]
            }

            msg = generate_post_message(property)
            msgres = send_msg(msg, SENDER, property)
            print(msgres)
            properties.append(property)

        payload["page"] += 1

    return properties

#Send message to a publication
def send_msg(msg: str, sender: dict, post: dict):
    data = {
    	"lead_type": "1",
    	"ContactForm[name]": sender["name"],
    	"ContactForm[lastname]": "",
    	"ContactForm[phone]": sender["phone"],
    	"ContactForm[email]": sender["email"],
    	"ContactForm[acceptTerms]": "1",
    	"ContactForm[registerUser]": "false",
    	"ContactForm[lead_source]": "1",
    	"ContactForm[body]": msg,
    	"is_ficha": "1",
    	"id": post["id"],
    	"lead_source": "1",
    }

    res = requests.post(URL_SEND, json=data).json()
    if "success" in res:
        logger.success(f"Mensaje enviado con exito")
    else:
        logger.error("Error enviando el mensaje")
        logger.error(res.text)
    print()

    return res

# "https://ujj28ojnla.execute-api.us-east-2.amazonaws.com/prod/property_listing"
# con ZenRows no funca nunca
# "https://propiedades.com/properties/filtrar"
# Con ZenRows aveces funciona, PremiumProxies activado
# No se le pueden pasar los parametros de busqueda
# Con selenium no anda
# llamar directamente a la url de los productos
# Funciona con zenrows, props dentro de __NEXT_DATA, PremiumProxies

# Esta funcion hace una peticion directamente a la url de las propiedades
# Ej: https://propiedades.com/df/casas-venta/recamaras-2?pagina=1
def get_posts_in_page(soup: bs4.BeautifulSoup):
    next_page = soup.find("script", id="__NEXT_DATA__").text
    data = json.loads(next_page)["props"]["pageProps"]["results"]["properties"]

    posts = []
    for post in data:
        posts.append({
            "id": post.get("id", ""),
            "title": post.get("full_address", ""),
            "extraction_date": strftime(DATE_FORMAT, gmtime()),
            "message_date": strftime(DATE_FORMAT, gmtime()),
            "price": post.get("price_real", ""),
            "currency": post.get("currency", ""),
            "type": post.get("type_str", ""),
            "url": post.get("url_propery", ""),
            "bedrooms": post.get("bedrooms", ""), 
            "bathrooms": post.get("bathrooms", ""),
            "building_size": post.get("size_m2", ""),
            "location": {
                "full": post.get("full_address", ""),
                "zone": "",
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

# Recibe el soup de la pagina y devuelve la cantidad total de paginas
def extract_total_pages(soup: bs4.BeautifulSoup) -> int:
    ul = soup.find(attrs={"aria-label": "items-pagination"})
    if ul == None:
        logger.error("No se encontro el ul the paginacion")
        exit(1)
    a = ul.contents[len(ul.contents)-2].a
    return int(a.text)

def main(url: str):
    sheet = Sheet(logger, 'scraper_mapping.json')
    sheets_headers = sheet.get("Extracciones!A1:Z1")[0]

    page = 1
    last_page = 1e9
    if not "pagina" in url:
        url += "?pagina={page}"
    else:
        re.sub(r"pagina=(.)", 'pagina={page}', url)

    #url = f"https://propiedades.com/df/casas-venta/recamaras-2?pagina={page}"
    while page <= last_page:
        formatted_url = url.format(page=page)
        logger.debug(f"Page: {page} url: {formatted_url} ")
        res = request.make(formatted_url, "GET")
        soup = bs4.BeautifulSoup(res.text, "html.parser")

        if last_page == 1e9:
            last_page = extract_total_pages(soup)
            logger.debug(f"Total pages: {last_page}")
        posts = get_posts_in_page(soup)

        row_ads = []
        for ad in posts:
            ad["message"] = ""
            #Guardamos los leads como filas para el sheets
            row_ad = sheet.map_lead(ad, sheets_headers)
            row_ads.append(row_ad)

        #Save the lead in the sheet
        sheet.write(row_ads, "Extracciones!A2")

        page += 1
