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
URL_SEND  = "https://propiedades.com/messages/send"
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
        data = {
            "lead_type": "1",
            "ContactForm[name]": self.sender["name"],
            "ContactForm[lastname]": "",
            "ContactForm[phone]": self.sender["phone"],
            "ContactForm[email]": self.sender["email"],
            "ContactForm[acceptTerms]": "1",
            "ContactForm[registerUser]": "false",
            "ContactForm[lead_source]": "1",
            "ContactForm[body]": msg,
            "is_ficha": "1",
            "id": post["id"],
            "lead_source": "1",
        }

        res = requests.post(URL_SEND, json=data)
        if not res.ok:
            self.logger.error("Error enviando el mensaje")
            self.logger.error(res.text)
        data = res.json()
        if "success" in data:
            self.logger.success(f"Mensaje enviado con exito")
        else:
            self.logger.error("Error enviando el mensaje")
            self.logger.error(res.text)

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

    def get_posts(self, url: str):
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
            page += 1

            yield posts

if __name__ == "__main__":
    msg = "hola"
    url = "https://propiedades.com/guadalajara-centro-guadalajara/residencial-renta"

    scraper = PropiedadesScraper()
    scraper.main(msg, url)
