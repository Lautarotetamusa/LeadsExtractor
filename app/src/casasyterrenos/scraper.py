# 
# La URL sale de las request, no es la url de busquedas
#

if __name__ == "__main__":
    import sys
    from dotenv import load_dotenv
    sys.path.append('.')
    load_dotenv()

import os
import re
from time import gmtime, strftime

import requests

from src.scraper import Scraper
from src.make_requests import ApiRequest

SITE = "https://www.casasyterrenos.com"
URL_SEND = "https://cytpanel.casasyterrenos.com/api/v1/public/contact"
URL_INFO = "https://www.casasyterrenos.com/_next/data/7b7AOhXs2MXfNPVCJedJS/es{canonical}.json"
URL_SESSION = f"{SITE}/api/amp/auth"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
URL_PROPS = f"https://www.casasyterrenos.com/_next/data/NmK5TM6UtMKwPf9rnMbvp/es/buscar/buscar/jalisco/zapopan/casas/venta.json?desde=10000000&hasta=100000000%2B&utm_source=results_page&page=2&slug=buscar&slug=jalisco&slug=zapopan&slug=casas&slug=venta"

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
    "property": "",
    "utm_source": "results_page"
}

class CasasyterrenosScraper(Scraper):
    def __init__(self):
        super().__init__("Casasyterrenos")

        self.session_id = ""
        self.sleep_secs = 0.1
        self.request = ApiRequest(self.logger, ZENROWS_API_URL, {
            "apikey": os.getenv("ZENROWS_APIKEY"),
            "url": "",
        })

    def generate_session_id(self):
        self.logger.success("Generando nuevo session id")
        payload = {
            "domain": SITE,
            "referrer": ""
        }
        
        res = self.request.make(URL_SESSION, "POST", json=payload)
        if res == None:
            self.logger.error("Error generando el session id, saliendo")
            exit(1)
        data = res.json()

        session_id = data.get("uuid")
        if not session_id:
            self.logger.error("Error generando el session id, saliendo")
            exit(1)

        self.logger.success("Session id generado con exito")
        self.logger.success("session_id: "+session_id)
        self.session_id = session_id
    
    def extract_posts(self, raw: list[dict]):
        posts = []

        for post in raw:
            prop = post.get("broker", {})

            args = {
                "rooms": "", 
                "bathrooms": "",
                "construction": "", 
                "parkingLots": "",
            }
            for arg in args:
                v = post.get(arg, "")
                if isinstance(v, dict):
                    args[arg] = str(v.get("from", "")) + ", " + str(v.get("to", ""))
                else:
                    args[arg] = v

            posts.append({
                "fuente": "CASASYTERRENOS",
                "id": post.get("id"),
                "title": post.get("name", ""),
                "extraction_date": strftime(DATE_FORMAT, gmtime()),
                "message_date": strftime(DATE_FORMAT, gmtime()),
                "price": post.get("priceSale", ""),
                "currency": post.get("currency", ""),
                "type": post.get("type", ""),
                "url": SITE + post.get("canonical", ""),
                "bedrooms": args["rooms"], 
                "bathrooms": args["bathrooms"],
                "building_size": args["construction"],
                "parkings": args["parkingLots"],
                "location": {
                    "full": ", ".join([post.get("municipality", ""), post.get("neighborhood", ""), post.get("state", "")]),
                    "zone": post.get("neighborhood"),
                    "city": post.get("municipality", ""),
                    "province": post.get("state", ""), 
                },
                "publisher": {
                    "name": prop.get("name", ""), 
                    "id": "",
                    "whatsapp": prop.get("whatsapp", ""),
                    "phone": prop.get("phone", ""),
                    "cellPhone": prop.get("phone", "")
                }
            })
        return posts

    def send_message(self, msg: str, post):
        self.logger.debug(f"Enviando mensaje a propiedad con {post['id']}")

        property_id = post["id"]
        data = SENDER.copy()
        data["message"] = msg
        data["property"] = int(post["id"])

        res = requests.post(URL_SEND, json=data)
        if not res:
            return None

        if not res.ok:
            self.logger.error("Error enviando el mensaje")
            self.logger.error(res.text)
            return None

        self.logger.success(f"Mensaje enviado con exito a la propiedad {property_id}")

    def get_posts(self, param):
        assert type(param) is str
        base_url = param
        page = 1
        len_posts = 0
        total_posts = 1e9

        s = re.search(r'page=(\d+)', base_url)
        if s is not None:
            page = int(s.group(1))

        base_url = re.sub(r'page=\d+', "page={page}", base_url)
        if not "&page={page}" in base_url:
            base_url += "&page={page}"
        print(base_url)

        while len_posts < total_posts:
            url = base_url.format(page=page)
            self.logger.debug(f"Page: {page} Url: {url}")
            res = requests.get(url)
            if not res.ok: 
                self.logger.error("Error geting page")
                self.logger.error(res.text)
                continue

            try:
                raw = res.json()
            except requests.exceptions.JSONDecodeError:
                self.logger.error("cannot decode json response")
                self.logger.error(res.text)
                continue

            data = raw.get("pageProps", {}).get("initialState")
            if data is None:
                self.logger.debug(res.text)
                return

            if total_posts == 1e9:
                total_posts = data.get("propertyData", {}).get("estimatedTotalHits", 0)
                self.logger.debug(f"Total posts: {total_posts}")

            posts = self.extract_posts(data.get("propertyData", {}).get("properties", []))
            if len(posts) <= 0: return

            page += 1
            len_posts += len(posts)
            yield posts

if __name__ == "__main__":
    url = "https://www.casasyterrenos.com/_next/data/3p3n18wV11NSHARkGU6tj/es/buscar/jalisco/guadalajara/casas-y-departamentos-y-terrenos/venta.json?desde=0&amp%3Bhasta=1000000000&amp%3Butm_source=results_page&slug=jalisco%2Cguadalajara%2Ccasas-y-departamentos-y-terrenos%2Cventa&page=10"

    scraper = CasasyterrenosScraper()
    msg = """Hola! {nombre}, como estás? 
Veo que tienes publicaciones en {ubicacion} y nosotros tenemos casas en Pre-venta para tu cartera de clientes en esa misma zona y zonas cercanas. Me interesa hacer una alianza con {nombre}.

Soy Gerente comercial de Rebora Arquitectos y ofrecemos el 2.5% a la firma de contrato de anticipo (A diferencia de una propiedad terminada que es 50% al inicio y 50% a la escritura). Por favor visita: rebora.com.mx/socios-comerciales/ o dejo mi numero de WhatsApp: 33 2809 2850.

Beneficios de nuestro programa de socios comerciales:

•⁠  ⁠Aumenta en un 50% tus ingresos anuales, al mostrarles a tus clientes la casa que están buscando.
•⁠  ⁠Diversifica tus ingresos con tu cartera actual de clientes.
•⁠  ⁠Cierra tu primera operación tan rápido como en un mes.

•⁠  ⁠Marcelo Michel"""
    scraper.test(msg, url)
