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

cookies= {
		"_lamudi_user_id": "16270d81-76c0-4775-bc1a-ac49cfbcd47f",
		"aws-waf-token": "c729cfde-8e86-49de-a715-eee34ca23761:FAoAoJ0ZT35BAAAA:uOKwizHBdPw8tMHbg1/DpZuNFxgWtLpTB1rdQv/EM0dcTAkIYL62CT/XosVxl98NCmaZZcrOGGgCILuZpF5PjcftSgw5+1PuFn8q9B/SxRRI98YqJVNJRk9f0B6QLVqB9M1trz/fvYMXhCQRz9PvQYliZSc2yr/Rzo0DGttUvUC64GHfVP9ZxHZHuCU4UDEZRUFB/txOaxXyH3iSmAyKBMeq1qm8BWyGOrhl4WzosWOm1iAborcU",
		"g_state": "{\"i_p\":1722525017033,\"i_l\":4}",
		"Origin": "1",
		"page": "1",
		"t_or": "1",
		"t_pvid": "9b64e864-34e2-4f42-9e40-9782340c1fd2"
	}

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
        res = requests.post(CONTACT_URL, json=sender, cookies=cookies)

        if res.status_code != 204:
            logger.error(f"Ocurrio un error enviando un mensaje al ad  {property_id}")
            return None

        logger.success(f"Mensaje enviado con exito id: {property_id}")

    def view_phone(self, post):
        property_url = post["url"]
        logger.debug(f"Obteniendo el telefono de la propiedad {property_url}")
        headers = {"User-Agent": "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:98.0) Gecko/20100101 Firefox/98.0"}

        try:
            res = requests.get(property_url, headers=headers)
            if res.status_code != 200:
                logger.error("Fallo en la peticion: "+str(res.status_code))

            soup = BeautifulSoup(res.text, "html.parser")
            phone_element = soup.find("a", attrs={"data-test": "agency-name"})
            if phone_element is None:
                logger.error("No se encontro el elemento class='agency__phone'")
                return ""

            phone = phone_element.get("href").strip().split("tel:")[1]

            logger.success(f"Telefono obtenido con exito: {phone}")
        except Exception as e:
            logger.error("Ocurrio un error obteniendo el telefono")
            logger.error(str(e))
            return ""
        return phone

    # Esta funcion recibe el html de una pagina y nos devuelve una lista con todas los ads de esa pagina
    def extract_posts(self, raw: BeautifulSoup):
        ads = raw.find_all("div", attrs={"data-test": "normal-listing"})
        print("ads: ", len(ads))
        posts = []
        for ad in ads:
            location = ad.find("span", attrs={"data-test": "snippet-content-location"}).text.split(',')
            bedrooms = ad.find("span", attrs={"data-test": "bedrooms-value"})
            bathromms = ad.find("span", attrs={"data-test": "full-bathrooms-value"})
            area = ad.find("div", attrs={"data-test": "area-value"})

            price_text = ad.find("div", class_="snippet__content__price").text.strip()
            # [0:3] y [0:2] Para no tener errores de "too many values to unpack"
            splitted = price_text.split(" ")
            if len(splitted) > 2:
                # $ 6,948,000 MXN
                [_, price, currency] = price_text.split(" ")[0:3]
            elif len(splitted) > 1:
                # USD 5,574,631
                [currency, price] = price_text.split(" ")[0:2]
            else:
                currency = ""
                price = ""

            phone = ""
            wpp_btn = ad.find("button", attrs={"data-test": "snippet-whatsapp-button-in-content"})
            if wpp_btn is not None:
                phone = wpp_btn.get("value", "").split("phone=")[1].split("&text")[0]

            href = ad.find("a").get("href")
            if href is None:
                logger.error("no se encontro la url de la propiedad")
                continue

            name = ""
            name_span = ad.find("span", attrs={"data-test": "agency-name"})
            if name_span is None:
                logger.error("no se encontro el span agency-name")
                continue
            name = name_span.text.strip()            
            if name == "":
                logger.error("no se encontro el nombre")
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
                "url":          SITE + href,
                "bedrooms": bedrooms.next_sibling.strip() if bedrooms is not None else "",
                "bathrooms": bathromms.next_sibling.strip() if bathromms is not None else "",
                "building_size": area.next_sibling.strip() if area is not None else "",
                "location": {
                    "full": ','.join(location),
                    "zone": "",
                    "city": location[0],
                    "province": location[1] if len(location) >= 2 else "",
                },
                "publisher": {
                    "name": name,
                    "id":   ad.get("data-idagencia", ""),
                    "whatsapp": phone,
                    "phone": phone,
                    "cellPhone": phone
                }
            }

            posts.append(post)

        return posts

    def get_posts(self, param):
        assert type(param) is str
        url = param
        headers = {"User-Agent": "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:98.0) Gecko/20100101 Firefox/98.0"}

        is_last: bool = False
        page = 1

        while not is_last:
            self.logger.debug(f"{page}: {url}")
            res = requests.get(url, headers=headers)
            soup = BeautifulSoup(res.text, "html.parser")

            if res.status_code != 200:
                self.logger.error("Error en la peticion: "+str(res.status_code))
                page += 1
                continue

            pagination = soup.find("a", id="pagination-next")
            if pagination is None:
                self.logger.error("No se encontro div de paginacion")
                page += 1
                continue

            # Usamos json.loads para parsear el booleano, "false" será False
            is_last = json.loads(pagination.get("data-islast", ""))
            url = pagination.get("href", "")
            page += 1

            yield self.extract_posts(soup)


if __name__ == "__main__":
    url = "https://www.lamudi.com.mx/jalisco/guadalajara/for-sale/?propertyTypeGroups=terreno%2Ccasa%2Cdepartamento?page=17"

    msg = """¡Hola! {nombre}, ¿cómo estás?

He visto que tienes publicaciones de terrenos en {ubicacion} y nosotros construimos casas de lujo que podrían interesar a tu cartera de clientes en esa zona y áreas cercanas. Me gustaría explorar una alianza contigo .

Soy  Gerente Comercial de Rebora Arquitectos. Ofrecemos un 2.5% a la firma de contrato de anticipo (a diferencia de una propiedad terminada, que es 50% al inicio y 50% a la escritura). Para más detalles, por favor visita rebora.com.mx/socios-comerciales/ o contáctame por WhatsApp: 33 2809 2850.

Sabemos que esta alianza puede tener grandes beneficios para ti, tu cliente y nosotros , durante el mes de Julio y Agosto te obsequiamos una bolsa Louis Vuitton (3 modelos a escoger) o un iPhone 15 Pro Max de 1 TB por cada contrato cerrado.

¿Cómo lo ves? ¿Qué día podemos agendar una cita?

Saludos, Gerencia Comercial"""

    scraper = LamudiScraper()
    scraper.test(msg, url)
