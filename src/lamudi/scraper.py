import requests
import json

from time import gmtime, strftime
from bs4 import BeautifulSoup

from src.logger import Logger
from src.sheets import Sheet
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
            "id":           ad.get("data-idanuncio"),
            "title":        ad.find("div", class_="listing-card__title").get("content"),
            "extraction_date": strftime(DATE_FORMAT, gmtime()),
            "message_date": strftime(DATE_FORMAT, gmtime()),
            "price":        ad.get("data-price"),
            "currency":     ad.get("data-currency") ,
            "type":         "",
            "url":          site + ad.get("data-href"),
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

def main(url: str):
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
            msg = generate_post_message(ad)
            ad["message"] = msg
            logger.debug(msg)
            send_message(ad["id"], msg)
            
            #Guardamos los leads como filas para el sheets
            row_ad = sheet.map_lead(ad, sheets_headers)
            row_ads.append(row_ad)

        #Save the lead in the sheet
        sheet.write(row_ads, "Extracciones!A2")

    logger.success(f"Se encontraron {total} ads para la url")

if __name__ == "__main__":
    url = "https://www.lamudi.com.mx/terreno/for-sale/bedrooms:3/?sorting=newest"
    main(url)
