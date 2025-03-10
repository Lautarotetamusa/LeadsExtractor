from multiprocessing.context import TimeoutError
from multiprocessing.pool import ThreadPool
import os
import time
from typing import Iterator

from src.property import Property
from src.message import generate_post_message
from src.sheets import Sheet
from src.logger import Logger

SENDER_PHONE = os.getenv("SENDER_PHONE")
assert SENDER_PHONE is not None, "Falta la variable 'SENDER_PHONE'"


class Scraper():
    def __init__(self, name: str):
        self.name = name
        self.sleep_secs: float = 1.0
        self.sender = {
            "email": os.getenv("SENDER_EMAIL"),
            "name":  os.getenv("SENDER_NAME"),
            "phone": os.getenv("SENDER_PHONE"),
        }

        self.logger = Logger(name)
        self.sheet = Sheet(self.logger, 'scraper_mapping.json')
        self.headers = self.sheet.get("Extracciones!A1:Z1")[0]

    def get_posts(self, param: str | dict) -> Iterator[list[dict]]:
        yield []

    def send_message(self, msg: str, post: dict):
        pass

    def view_phone(self, post) -> str:
        return ""

    def get_total_posts(self) -> int:
        return 0

    def publish(self, property: Property):
        return

    def _action(self, ad, spin_msg: str | None) -> list[str]:
        if spin_msg is not None:
            msg = generate_post_message(ad, spin_msg)
            ad["publisher"]["phone"] = self.send_message(msg, ad)
            ad["message"] = msg.replace('\n', '')
        else:
            ad["message"] = ""

        return self.sheet.map_lead(ad, self.headers)

    def main(self, spin_msg: str | None, param: str | dict):
        total_posts = 0
        pool = ThreadPool(processes=20)

        max = 10
        timeout = 30
        max_messages = 1
        occurences = {}

        for page in self.get_posts(param):
            total_posts += len(page)
            results = []
            row_ads = []

            for ad in page:
                print(ad)
                phone = ad["publisher"]["phone"]
                if phone == "" or phone is None:
                    ad["publisher"]["phone"] = self.view_phone(ad)
                    phone = ad["publisher"]["phone"]

                if phone == SENDER_PHONE:
                    self.logger.debug("Propiedad de Rebora encontrada")
                    continue

                id = ad["publisher"]["phone"]
                if id in occurences:
                    occurences[id] += 1
                else:
                    occurences[id] = 0

                if occurences[id] > max_messages:
                    self.logger.warning(f"Maxima cantidad de mensajes enviados para {phone}")
                    continue

                r = pool.apply_async(self._action, args=(ad, spin_msg, ))
                time.sleep(self.sleep_secs)
                results.append(r)

                if len(results) >= max:
                    for r in results:
                        try:
                            row = r.get(timeout)
                        except TimeoutError:
                            self.logger.error("Timeout running action")
                            continue
                        row_ads.append(row)
                    results = []

            # Save the lead in the sheet
            self.sheet.write(row_ads, "Extracciones!A2")
        self.logger.success(f"Se encontraron un total de {total_posts} en la url especificada")

    def test(self, spin_msg: str | None, param: str | dict):
        total_posts = 0

        max_messages = 1
        occurences = {}

        for page in self.get_posts(param):
            total_posts += len(page)
            row_ads = []

            for ad in page:
                phone = ad["publisher"]["phone"]
                if phone == "" or phone is None:
                    ad["publisher"]["phone"] = self.view_phone(ad)
                    phone = ad["publisher"]["phone"]

                if phone == SENDER_PHONE:
                    self.logger.debug("Propiedad de Rebora encontrada")
                    continue

                id = ad["publisher"]["phone"]
                if id in occurences:
                    occurences[id] += 1
                else:
                    occurences[id] = 0

                if occurences[id] > max_messages:
                    self.logger.warning(f"Maxima cantidad de mensajes enviados para {phone}")
                    continue

            self.sheet.write(row_ads, "PruebasExtracciones!A2")
        self.logger.success(f"Se encontraron un total de {total_posts} en la url especificada")
