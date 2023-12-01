import requests
import json

from datetime import datetime
from bs4 import BeautifulSoup

from src.logger import Logger
from src.message import generate_post_message

site = "https://www.inmuebles24.com"
VIEW_URL = "https://www.inmuebles24.com/rp-api/leads/view"
LIST_URL = "https://www.inmuebles24.com/rplis-api/postings"
CONTACT_URL = "https://www.lamudi.com.mx/adform/api/lead-contact"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
DATE_FORMAT = "%d/%m/%Y"
SENDER = {
	"message": "",
	"pageViewId": "",
	"propertyAdId": "",
	"userEmail": "ventas.rebora@gmail.com",
	"userName": "Brenda Diaz",
	"userPhone": "3313420733"
}

logger = Logger("lamudi.com")

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

#Esta funcion recibe el html de una pagina y nos devuelve una lista con todas los ads de esa pagina
def scrape_list_page(soup: BeautifulSoup):
    ads = soup.find_all("div",class_ = "listing listing-card item")
    ad_list = []
    for ad in ads:
        bedrooms = ad.find("div", attrs={"data-test": "bedrooms-value"})
        bathromms = ad.find("div", attrs={"data-test": "full-bathrooms-value"})
        area = ad.find("div", attrs={"data-test": "area-value"})
        ad_structure = {}
       
        ad_structure["id"] = ad.get("data-idanuncio")
        ad_structure["currency"] = ad.get("data-currency") 
        ad_structure["title"] = ad.find("div", class_="listing-card__title").get("content")
        ad_structure["price"] = ad.get("data-price")
        ad_structure["url"] = ad.get("data-href")
        ad_structure["location"] = ad.get("data-location")
        ad_structure["bedrooms"] = bedrooms.next_sibling.strip() if bedrooms != None else ""
        ad_structure["bathrooms"] = bathromms.next_sibling.strip() if bathromms != None else ""
        ad_structure["building_size"] = area.next_sibling.strip() if area != None else ""
        agent = {
            "name": ad.get("data-nombreagencia"),
            "id": ad.get("data-idagencia")
        }
        ad_structure["agent"] = agent

        ad_list.append(ad_structure)
    
    return ad_list

def main():
    url = "https://www.lamudi.com.mx/terreno/for-sale/bedrooms:3/?sorting=newest"
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
        total += len(ads)
        for ad in ads:
            msg = generate_post_message(ad)
            logger.debug(msg)
            send_message(ad["id"], msg)

    logger.success(f"Se encontraron {total} ads para la url")

if __name__ == "__main__":
    main()
