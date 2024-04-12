import requests
import json
import os

from time import gmtime, strftime
from bs4 import BeautifulSoup

from src.logger import Logger
from src.sheets import Sheet
from src.message import generate_post_message

SITE = "https://www.lamudi.com.mx"
CONTACT_URL = f"{SITE}/adform/api/lead-contact"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
DATE_FORMAT = "%d/%m/%Y"
SENDER = {
    "message": "",
    "pageViewId": "",
    "propertyAdId": "",
    "userEmail": os.getenv("SENDER_EMAIL"),
    "userName":  os.getenv("SENDER_NAME"),
    "userPhone": os.getenv("SENDER_PHONE")
}

logger = Logger("scraper lamudi.com")

def send_message(property_id: str, msg: str):
    logger.debug(f"Enviando mensaje al ad {property_id}")

    sender = SENDER.copy()
    sender["message"] = msg
    sender["propertyAdId"] = property_id
    res = requests.post(CONTACT_URL, json=sender)
    
    if res.status_code != 204:
        logger.error(f"Ocurrio un error enviando un mensaje al ad  {property_id}")
        return None

    logger.success("Mensaje enviado con exito")

def view_phone(property_url: str):
    logger.debug(f"Obteniendo el telefono de la propiedad {property_url}")
    headers = {"User-Agent": "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:98.0) Gecko/20100101 Firefox/98.0"}

    try:
        res = requests.get(property_url, headers=headers)

        soup = BeautifulSoup(res.text, "html.parser")
        phone = soup.find("div", class_="phone-number").text.strip()
        logger.success(f"Telefono obtenido con exito: {phone}")
    except Exception as e:
        logger.error("Ocurrio un error obteniendo el telefono")
        logger.error(str(e))
        return ""
    return phone

#Esta funcion recibe el html de una pagina y nos devuelve una lista con todas los ads de esa pagina
def scrape_list_page(soup: BeautifulSoup):
    ads = soup.find_all("div",class_ = "listing listing-card item")
    posts = []
    for ad in ads:
        location = ad.get("data-location").split(',')
        bedrooms = ad.find("div", attrs={"data-test": "bedrooms-value"})
        bathromms = ad.find("div", attrs={"data-test": "full-bathrooms-value"})
        area = ad.find("div", attrs={"data-test": "area-value"})
        post = {
            "fuente": "LAMUDI", 
            "id":           ad.get("data-idanuncio"),
            "title":        ad.find("div", class_="listing-card__title").get("content"),
            "extraction_date": strftime(DATE_FORMAT, gmtime()),
            "message_date": strftime(DATE_FORMAT, gmtime()),
            "price":        ad.get("data-price"),
            "currency":     ad.get("data-currency") ,
            "type":         "",
            "url":          SITE + ad.get("data-href"),
            "bedrooms": bedrooms.next_sibling.strip() if bedrooms != None else "",
            "bathrooms": bathromms.next_sibling.strip() if bathromms != None else "",
            "building_size": area.next_sibling.strip() if area != None else "",
            "location": {
                "full": ','.join(location),
                "zone": "",
                "city": location[0],
                "province": location[1] if len(location) >= 2 else "", 
            },
            "publisher":    {
                "name": ad.get("data-nombreagencia"),
                "id": ad.get("data-idagencia"),
                "whatsapp": "",
                "phone": "",
                "cellPhone": ""
            }
        }

        posts.append(post)
    
    return posts

def main(url: str, spin_msg: str):
    sheet = Sheet(logger, 'scraper_mapping.json')
    sheets_headers = sheet.get("Extracciones!A1:Z1")[0]

    headers = {"User-Agent": "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:98.0) Gecko/20100101 Firefox/98.0"}

    is_last: bool = False
    page = 1
    total = 0

    while not is_last:
        logger.debug(f"{page}: {url}")
        res = requests.get(url, headers=headers)
        soup = BeautifulSoup(res.text, "html.parser")

        pagination = soup.find("a", id="pagination-next")
        if pagination == None:
            logger.error("No se encontro div de paginacion")
            exit(1)
        #Usamos json.loads para parsear el booleano, "false" será False
        is_last = json.loads(pagination.get("data-islast"))
        url = pagination.get("href", "")
        page += 1
        
        ads = scrape_list_page(soup)
        row_ads = []
        total += len(ads)
        for ad in ads:
            msg = generate_post_message(ad, spin_msg)
            ad["message"] = msg.replace('\n', '')
            logger.debug(msg)
            send_message(ad["id"], msg)
            ad["phone"] = view_phone(ad["url"])
            
            #Guardamos los leads como filas para el sheets
            row_ad = sheet.map_lead(ad, sheets_headers)
            row_ads.append(row_ad)

        #Save the lead in the sheet
        sheet.write(row_ads, "Extracciones!A2")

    logger.success(f"Se encontraron {total} ads para la url")

if __name__ == "__main__":
    url = "https://www.lamudi.com.mx/terreno/for-sale/bedrooms:3/?sorting=newest"
    main(url, "hola")
