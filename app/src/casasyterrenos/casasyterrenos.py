import json
import time
import os
from datetime import datetime
from typing import Iterator

import requests
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

from src.property import Property
from src.portal import Mode, Portal
from src.lead import Lead

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT is not None, "DATE_FORMAT is not seted"

PROPERTY_URL = "https://www.casasyterrenos.com/propiedad"

API_URL = "https://cytpanel.casasyterrenos.com/api/v1"
MEDIA_URL = "https://cyt-media.s3.amazonaws.com/"
CLIENT_ID = "4je1v2kfou9e9plpv6vf0vmnll"


class CasasYTerrenos(Portal):
    def __init__(self):
        super().__init__(
            name="casasyterrenos",
            contact_id_field="id",
            send_msg_field="",
            username_env="CASASYTERRENOS_USERNAME",
            password_env="CASASYTERRENOS_PASSWORD",
            params_type="headers",
            unauthorized_codes=[401],
            filename=__file__
        )

    def get_leads(self, mode=Mode.NEW) -> Iterator[list[dict]]:
        page = 1

        if mode == Mode.NEW:
            status = "1"  # Filtramos solamente los leads nuevos
            url = f"{API_URL}/list_contact/?page={page}&status={status}"
        else:
            url = f"{API_URL}/list_contact/?page={page}"

        while url is not None:
            res = self.request.make(url)
            if res is None:
                break
            data = res.json()

            url = data["next"]
            self.logger.debug(f"total: {data['count']}")

            yield data["results"]

    def make_contacted(self, lead: dict):
        id = lead[self.contact_id_field]
        self.logger.debug(f"Marcando como contactacto a lead {id}")
        url = f"{API_URL}/contact/{id}"

        data = {
            "id": id,
            "status": 2,  # 2 -> Contactado por correo
            "status_description": "Correo"
        }
        self.request.make(url, 'PATCH', data=data)

        self.logger.success(f"Se contacto correctamente a lead {id}")

    def get_lead_info(self, raw_lead: dict) -> Lead:
        lead = Lead()
        lead.set_args({
            "message": raw_lead.get("message", ""),
            "fuente": self.name,
            "fecha_lead": datetime.strptime(raw_lead["created"], '%d-%m-%Y %H:%M:%S').strftime(DATE_FORMAT),
            "lead_id":  str(raw_lead["id"]),
            "nombre":  raw_lead.get("name"),
            "link": "",
            "email":  raw_lead.get("email"),
            "telefono": raw_lead.get("phone", "")
        })
        lead.set_propiedad({
            "id":  str(raw_lead.get("property_id", "")),
            "titulo":  raw_lead.get("property_title", ""),
            "link": f"{PROPERTY_URL}/{raw_lead['property_id']}",
            "tipo":  raw_lead.get("property_type", ""),
        })

        return lead

    def publish(self, property: Property) -> Exception | None:
        # prop_id = self.publish_first_step(property)
        prop_id = 3754373
        if prop_id == None: return Exception("cannot obtain the property id")

        self.add_ubication(prop_id, property) 

    # The property its pending, returns the property id,
    def publish_first_step(self, property: Property) -> int | None:
        operation_type_map = {
            "sale": 1,
            "rent": 2
        }
        property_type_map = {
            "house": "18",
            "department": "19"
        }

        payload = {
            "bathrooms": property.bathrooms,
            "rooms": property.rooms,
            # "floor_number": 0,
            "sqr_mt_lot": property.m2_covered,
            "sqr_mt_construction": property.m2_total,
            "name": property.title,
            "description": f"<div><!--block-->{property.description}!</div>",
            "price_sale": property.price,
            "property_type": property_type_map[property.type.__str__()],
            "operation_type": [ operation_type_map[property.operation_type.__str__()] ],
            # "membership": "434451",
            "currency_sale": property.currency.lower(),
            "currency_rent": property.currency.lower(),
            "currency_transfer": property.currency.lower(),
            # "parking": 0,
            # "age": 0,
            # "half_bathrooms": 0,
            # "levels": 0,
            # "floors": 0,
            # "sqr_mt_office": 0,
            # "sqr_mt_cellar": 0,
            # "sqr_mt_front": 0,
            # "sqr_mt_long": 0,
            # "sqr_mt_terrace": 0,
            # "sqr_mt_vault": 0
        }

        res = self.request.make(API_URL + "/property", "POST", json=payload)
        if res == None: return None
        data = res.json()
        return data.get("id")

    def add_ubication(self, prop_id: int, property: Property) -> Exception | None:
        payload = {
            "state": "14",
            "latitude": 20.6900164,
            "longitude": -103.4733259,
            # "colony": 834012,
            # "municipality": 2995,
            "exterior_number": "246",
            "interior_number": "",
            # "street": "Marsella"
        }

        res = self.request.make(f"{API_URL}/property/{prop_id}", "PATCH", json=payload)
        if res == None: return Exception("error adding ubication data")
        print(res.text)
        print(res.status_code)
        # data = res.json()
        # return data.get("id")
        return None


    def login(self, session="session"):
        login_url = "https://panel-pro.casasyterrenos.com/login"
        panel_url = "https://panel-pro.casasyterrenos.com/contacts"
        options = Options()
        options.add_argument(f"--user-data-dir={session}")
        options.add_argument("--headless")
        # Necesario para correrlo como root dentro del container
        options.add_argument("--no-sandbox")

        driver = webdriver.Chrome(options=options)
        access_token = None

        self.logger.debug("Generando un nuevo token de acceso")
        try:
            driver.get(panel_url)

            if (driver.current_url == login_url):
                self.logger.debug("La sesion no esta iniciada, iniciando..")
                driver.get(login_url)

                # Esperar que cargue la pagina
                WebDriverWait(driver, 10).until(EC.visibility_of_element_located((By.XPATH, "//input[@name='username']")))

                username_input = driver.find_element(By.XPATH, "//input[@name='username']")
                password_input = driver.find_element(By.XPATH, "//input[@name='password']")
                send_btn = driver.find_element(By.XPATH, "//button[@data-splitbee-event='log-in']")

                username_input.send_keys(self.username)
                password_input.send_keys(self.password)
                send_btn.click()

                WebDriverWait(driver, 10).until(EC.visibility_of_element_located((By.XPATH, "/html/body/div/div/div/div[1]/div/nav/a")))
                self.logger.success("Sesion iniciada con exito")

            self.logger.debug("Obteniendo access token")
            access_token = driver.execute_script(f"return window.localStorage.getItem('CognitoIdentityServiceProvider.{CLIENT_ID}.{self.username}.idToken');")
            if access_token is not None:
                self.logger.success("Access token obtenido con exito")
                self.logger.success(access_token)
            else:
                self.logger.error("No se pudo obtener el access_token")

            # Esperamos para que el token sea correcto
            driver.implicitly_wait(10)
            time.sleep(10)
        except Exception as e:
            self.logger.error("Ocurrio un error generando el access token")
            self.logger.error(str(e))
        finally:
            driver.quit()

        self.request.headers = {
            "Authorization": f"Bearer {access_token}"
        }
        with open(self.params_file, "w") as f:
            json.dump(self.request.headers, f, indent=4)


if __name__ == "__main__":
    scraper = CasasYTerrenos()
    scraper.main()
