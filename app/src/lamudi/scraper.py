if __name__ == "__main__":
    import sys
    from dotenv import load_dotenv
    sys.path.append('.')
    load_dotenv()

import requests
import json
import os

from time import gmtime, strftime
from bs4 import BeautifulSoup

from src.scraper import Scraper
from src.logger import Logger

SITE = "https://www.lamudi.com.mx"
CONTACT_URL = f"{SITE}/adform/api/lead-contact"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
DATE_FORMAT = "%d/%m/%Y"

logger = Logger("scraper lamudi.com")

class LamudiScraper(Scraper):
    def __init__(self):
        super().__init__("Lamudi")
        self.sender = {
            "message": "",
            "pageViewId": "",
            "propertyAdId": "",
            "userEmail": os.getenv("SENDER_EMAIL"),
            "userName":  os.getenv("SENDER_NAME"),
            "userPhone": os.getenv("SENDER_PHONE")
        }

    def send_message(self, msg: str, post: dict):
        property_id = post["id"]
        logger.debug(f"Enviando mensaje al ad {property_id}")

        sender = self.sender.copy()
        sender["message"] = msg
        sender["propertyAdId"] = property_id
        res = requests.post(CONTACT_URL, json=sender)
        
        if res.status_code != 204:
            logger.error(f"Ocurrio un error enviando un mensaje al ad  {property_id}")
            return None

        logger.success("Mensaje enviado con exito")

    def view_phone(self, post):
        property_url = post["url"]
        logger.debug(f"Obteniendo el telefono de la propiedad {property_url}")
        headers = {"User-Agent": "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:98.0) Gecko/20100101 Firefox/98.0"}

        try:
            res = requests.get(property_url, headers=headers)

            soup = BeautifulSoup(res.text, "html.parser")
            phone = soup.find("a", class_="agency__phone").text.strip()
            logger.success(f"Telefono obtenido con exito: {phone}")
        except Exception as e:
            logger.error("Ocurrio un error obteniendo el telefono")
            logger.error(str(e))
            return ""
        return phone

    # Esta funcion recibe el html de una pagina y nos devuelve una lista con todas los ads de esa pagina
    def extract_posts(self, raw: BeautifulSoup):
        ads = raw.find_all("a", attrs={"data-test": "normal-listing"})
        print("ads: ", len(ads))
        posts = []
        for ad in ads:
            location  = ad.find("span", attrs={"data-test": "snippet-content-location"}).text.split(',')
            bedrooms  = ad.find("span", attrs={"data-test": "bedrooms-value"})
            bathromms = ad.find("span", attrs={"data-test": "full-bathrooms-value"})
            area = ad.find("div", attrs={"data-test": "area-value"})
        
            # El primero es el simbolo $, lo ignoramos con _
            price_text = ad.find("div", class_="snippet__content__price").text.strip()
            [_, price, currency] = price_text.split(" ")

            phone = ""
            wpp_btn = ad.find("button", attrs={"data-test": "snippet-whatsapp-button-in-content"})
            if wpp_btn != None:
                phone = wpp_btn.get("value", "").split("phone=")[1].split("&text")[0]

            if phone == self.sender["userPhone"]:
                print("Se encontro una propiedad de Rebora")
                continue  

            post = {
                "fuente": "LAMUDI", 
                "id":           ad.get("data-idanuncio"),
                "title":        ad.find("span", class_="snippet__content__title").get("content"),
                "extraction_date": strftime(DATE_FORMAT, gmtime()),
                "message_date": strftime(DATE_FORMAT, gmtime()),
                "price":        price,
                "currency":     currency,
                "type":         "",
                "url":          SITE + ad.get("href", ""),
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
                    "name": ad.find("span", attrs={"data-test": "agency-name"}).get("text", ""),
                    "id":   ad.get("data-idagencia", ""),
                    "whatsapp": phone,
                    "phone": "",
                    "cellPhone": ""
                }
            }

            posts.append(post)
        
        return posts
    
    def get_posts(self, url):
        headers = {"User-Agent": "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:98.0) Gecko/20100101 Firefox/98.0"}

        is_last: bool = False
        page = 1

        while not is_last:
            self.logger.debug(f"{page}: {url}")
            res = requests.get(url, headers=headers)
            soup = BeautifulSoup(res.text, "html.parser")

            pagination = soup.find("a", id="pagination-next")
            if pagination == None:
                self.logger.error("No se encontro div de paginacion")
                continue

            #Usamos json.loads para parsear el booleano, "false" será False
            is_last = json.loads(pagination.get("data-islast", ""))
            url = pagination.get("href", "")
            page += 1
           
            posts = self.extract_posts(soup) 
            yield posts 

if __name__ == "__main__":
    url = "https://www.lamudi.com.mx/jalisco/guadalajara/casa/for-sale/?minPrice=10000000"
    msg = """Hola! {nombre}, como estás? 
Veo que tienes publicaciones en {ubicacion} y nosotros tenemos casas en Pre-venta para tu cartera de clientes en esa misma zona y zonas cercanas. Me interesa hacer una alianza con {nombre}.

Soy Gerente comercial de Rebora Arquitectos y ofrecemos el 2.5% a la firma de contrato de anticipo (A diferencia de una propiedad terminada que es 50% al inicio y 50% a la escritura). Por favor visita: rebora.com.mx/socios-comerciales/ o dejo mi numero de WhatsApp: 33 2809 2850.

Beneficios de nuestro programa de socios comerciales:

•⁠  ⁠Aumenta en un 50% tus ingresos anuales, al mostrarles a tus clientes la casa que están buscando.
•⁠  ⁠Diversifica tus ingresos con tu cartera actual de clientes.
•⁠  ⁠Cierra tu primera operación tan rápido como en un mes.

•⁠  ⁠Marcelo Michel"""
    scraper = LamudiScraper()
    scraper.main(msg, url)
