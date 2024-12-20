if __name__ == "__main__":
    import sys
    from dotenv import load_dotenv
    sys.path.append('.')
    load_dotenv()

import bs4
import requests
import json
from time import gmtime, strftime
import os
import re

from src.scraper import Scraper
from src.make_requests import ApiRequest

PROPS_URL = "https://propiedades.com/properties/filtrar"
URL_SEND  = "https://propiedades.com/messages/getContactActivity"#"https://propiedades.com/messages/send"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT != None, "DATE_FORMAT is not seted"

class PropiedadesScraper(Scraper):
    def __init__(self):
        super().__init__("Propiedades")
        self.request = ApiRequest(self.logger, ZENROWS_API_URL, {
            "apikey": os.getenv("ZENROWS_APIKEY"),
            "url": "",
            "js_render": "true",
            "antibot": "true",
            "premium_proxy": "true",
            "proxy_country": "mx",
        })

    def send_message(self, msg: str, post: dict):
        # headers = {
        #     "Content-Type": "multipart/form-data; boundary=---------------------------364683554611898991106638518"
        # }
        # data = {
        #     "lead_type": "2",
        #     "is_ficha": "0",
        #     "id": post["id"],
        #     "ssr": 1,
        #     "referral": "null",
        #     "referenceAddress": "",
        #     "ContactForm[name]": "juan",#self.sender["name"],
        #     "ContactForm[lastname]": "",
        #     "ContactForm[phone]": 3415854221,#self.sender["phone"],
        #     "ContactForm[email]": "juanpozzi@gmail.com",#self.sender["email"],
        #     "ContactForm[acceptTerms]": "1",
        #     "ContactForm[registerUser]": "false",
        #     "ContactForm[lead_source]": "5",
        #     "ContactForm[body]": "Hola!",#msg,
        #     # "lead_source": "5",
        # }

        data = {
                "id": post["id"],
                "is_ficha": "1",
                "contact_with_whatsapp": "undefined"
        }

        print(requests.Request('POST', 'http://httpbin.org/post', files=data).prepare().body.decode('utf8'))

        print(json.dumps(data, indent=4))
        res = self.request.make(URL_SEND, 'POST', json=data)
        # res = requests.post(URL_SEND, files=data, headers=headers)
        if res is None or not res.ok:
            self.logger.error("Error enviando el mensaje")
            return
            #self.logger.error(res.text)
        data = res.json()
        if "success" in data:
            self.logger.success(f"Mensaje enviado con exito")
        else:
            self.logger.error("Error enviando el mensaje")
            self.logger.error(res.text)

        print(res.json())
        return res.json()["track"]["phone_contact"]

    def extract_posts(self, raw: bs4.BeautifulSoup) -> list[dict]:
        next_page_tag = raw.find("script", id="__NEXT_DATA__")
        if next_page_tag == None:
            return []
        next_page = next_page_tag.text
        data = json.loads(next_page)["props"]["pageProps"]["results"]["properties"]

        posts = []
        for post in data:
            posts.append({
                "fuente": "PROPIEDADES.COM",
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
                "publisher": {
                    "name": "",
                    "id": "",
                    "whatsapp": "",
                    "phone": "",
                    "cellPhone": ""
                }
            })
        return posts

    def get_total_pages(self, soup: bs4.BeautifulSoup) -> int:
        ul = soup.find(attrs={"aria-label": "items-pagination"})
        if ul == None:
            self.logger.error("No se encontro el ul the paginacion")
            exit(1)
        a = ul.contents[len(ul.contents)-2].a
        return int(a.text)

    def get_posts(self, param):
        assert type(param) is str
        url = param
        page = 1
        last_page = 1e9
        if not "pagina" in url:
            url += "?pagina={page}"
        else:
            re.sub(r"pagina=(.)", 'pagina={page}', url)

        #url = f"https://propiedades.com/df/casas-venta/recamaras-2?pagina={page}"
        while page <= last_page:
            formatted_url = url.format(page=page)
            self.logger.debug(f"Page: {page} url: {formatted_url} ")
            res = self.request.make(formatted_url, "GET")
    
            if res == None:
                self.logger.error("Response error")
                continue

            soup = bs4.BeautifulSoup(res.text, "html.parser")

            if last_page == 1e9:
                last_page = self.get_total_pages(soup)
                self.logger.debug(f"Total pages: {last_page}")
    
            posts = self.extract_posts(soup)
            print(len(posts))
            page += 1

            yield posts

if __name__ == "__main__":
    url = "https://propiedades.com/guadalajara-centro-guadalajara/residencial-renta"
    url = "https://propiedades.com/zapopan/terrenos-comerciales-venta?pagina=1#remates=2&precio-min=3500000"


    post = {'fuente': 'PROPIEDADES.COM', 'id': 28277332, 'title': 'San Francisco #1228, Col. Los Cajetes C.P. 45234, Zapopan', 'extraction_date': '2024-11-29', 'message_date': '2024-11-29', 'price': 20490000, 'currency': 'MXN', 'type': 'Terreno comercial', 'url': '', 'bedrooms': 0, 'bathrooms': 0, 'building_size': '3169', 'location': {'full': 'San Francisco #1228, Col. Los Cajetes C.P. 45234, Zapopan', 'zone': '', 'city': 'Zapopan', 'province': 'Jalisco'}, 'publisher': {'name': '', 'id': '', 'whatsapp': '', 'phone': '', 'cellPhone': ''}}

    msg = """¡Hola! {nombre}, ¿cómo estás?

He visto que tienes publicaciones de terrenos en {ubicacion} y nosotros construimos casas de lujo que podrían interesar a tu cartera de clientes en esa zona y áreas cercanas. Me gustaría explorar una alianza contigo .

Soy  Gerente Comercial de Rebora Arquitectos. Ofrecemos un 2.5% a la firma de contrato de anticipo (a diferencia de una propiedad terminada, que es 50% al inicio y 50% a la escritura). Para más detalles, por favor visita rebora.com.mx/socios-comerciales/ o contáctame por WhatsApp: 33 2809 2850.

Sabemos que esta alianza puede tener grandes beneficios para ti, tu cliente y nosotros , durante el mes de Julio y Agosto te obsequiamos una bolsa Louis Vuitton (3 modelos a escoger) o un iPhone 15 Pro Max de 1 TB por cada contrato cerrado.

¿Cómo lo ves? ¿Qué día podemos agendar una cita?

Saludos, Gerencia Comercial"""

    scraper = PropiedadesScraper()
    # scraper.main(msg, url)
    scraper.send_message("Hola!", post)
