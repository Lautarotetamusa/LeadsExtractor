from datetime import datetime
import json
import os
from enum import IntEnum
from typing import Generator
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

from src.portal import Mode, Portal
from src.lead import Lead

# mis propiedades: https://propiedades.com/api/v3/property/MyProperties
# En este archivo tenemos todas las propieades previamente extraidas
with open("src/propiedades_com/properties.json") as f:
    PROPS = json.load(f)

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT is not None, "DATE_FORMAT is not seted"
API_URL = "https://ggcmh0sw5f.execute-api.us-east-2.amazonaws.com"


# Lista de estados posibles de un lead
class Status(IntEnum):
    NUEVO = 1
    CONTACTADO = 2
    CALIFICADO = 3
    EN_PROCESO = 4
    CONVERTIDO = 5
    CERRADA = 6


class Propiedades(Portal):
    def __init__(self):
        super().__init__(
            name="propiedades",
            contact_id_field="id",
            send_msg_field="",
            username_env="PROPIEDADES_USERNAME",
            password_env="PROPIEDADES_PASSWORD",
            params_type="headers",
            unauthorized_codes=[401],
            filename=__file__
        )

    def get_leads(self, mode=Mode.ALL) -> Generator:
        first = True
        page = 1
        end = False
        url = f"{API_URL}/prod/get/leads?page={page}&country=MX"

        while (not end) and (first or page is not None):
            res = self.request.make(url)
            if res is None:
                break
            data = res.json()["leads"]

            page = data["page"]["next_page"]
            url = f"{API_URL}/prod/get/leads?page={page}&country=MX"
            total = data["page"]["items"]

            if first:
                self.logger.debug(f"total: {total}")
                first = False

            # Si el modo es NEW Buscamos solamente los leads sin leer
            if mode == Mode.NEW:
                leads = []
                for lead in data["properties"]:
                    if lead["status"] != Status.NUEVO:
                        self.logger.debug("Se encontro un lead ya contactado, paramos")
                        end = True
                        break
                    leads.append(lead)
                yield leads
            else:  # SI el modo es ALL vamos a traernos todos los leads
                yield data["properties"]

    def get_lead_info(self, raw_lead: dict) -> Lead:
        lead = Lead()
        lead.set_args({
            "fuente": self.name,
            "fecha_lead": datetime.strptime(raw_lead["updated_at"], DATE_FORMAT).strftime(DATE_FORMAT),
            "lead_id": str(raw_lead["id"]),
            "nombre": raw_lead.get("name", ""),
            "email": raw_lead.get("email", ""),
            "telefono": raw_lead.get("phone"),
        })

        prop = self.get_lead_property(str(raw_lead["property_id"]))

        # Si no tiene propiedad lo marcamos como cerrado
        if prop is None:
            self.make_contacted(raw_lead)
            return lead

        prop["address"] = raw_lead["address"]
        if prop["titulo"] == "":
            prop["titulo"] = prop["address"]
        lead.propiedad = prop

        return lead

    def get_lead_property(self, property_id: str) -> dict | None:
        if property_id not in PROPS:
            self.logger.error(f"No se encontro la propiedad con id {property_id}")
            return None
        data = PROPS[property_id]

        return {
            "id": str(data["id"]),
            "titulo": data["short_address"],
            "link": data["url"],
            "precio": str(data.get("price", "")),
            "ubicacion": data.get("address", ""),
            "tipo": data.get("type_children_string", ""),
            "municipio": data["municipality"],
            "bedrooms": data.get("bedrooms", ""),
            "bathrooms": data.get("bathrooms", ""),
            "total_area": data.get("size_m2", ""),
            "covered_area": ""
        }

    def _change_status(self, lead: dict, status=Status.CONTACTADO):
        id = lead[self.contact_id_field]
        self.logger.debug(f"Marcando como {status.name} al lead {id}")

        url = f"{API_URL}/prod/leads/status"

        req = {
            "lead_id": id,
            "status": status,
            "country": "MX"
        }

        res = self.request.make(url, "PUT", json=req)
        if (res is None):
            self.logger.error(f"No se pudo marcar al lead como {status.name}")
        else:
            self.logger.success(f"Se marco a lead {id} como {status.name}")

    def make_contacted(self, lead: dict):
        self._change_status(lead, Status.CERRADA)

    def login(self, session="session"):
        login_url = "https://propiedades.com/login"
        options = Options()
        options.add_argument(f"--user-data-dir={session}")
        options.add_argument("--headless")
        # Necesario para correrlo como root dentro del container
        options.add_argument("--no-sandbox")

        driver = webdriver.Chrome(options=options)

        self.logger.debug("Iniciando sesion")
        try:
            driver.get(login_url)

            # Esperar que cargue la pagina
            WebDriverWait(driver, 20).until(EC.visibility_of_element_located((By.XPATH, "//input[@data-gtm='text_field_email']")))

            username_input = driver.find_element(By.XPATH, "//input[@data-gtm='text_field_email']")
            username_input.send_keys(self.username)
            driver.find_element(By.XPATH, "//button[@data-id='login_correo']").click()

            password_input = driver.find_element(By.XPATH, "//input[@data-gtm='text_field_password']")
            password_input.send_keys(self.password)
            driver.find_element(By.XPATH, "//button[@data-id='login_password']").click()

            WebDriverWait(driver, 10).until(EC.visibility_of_element_located((By.XPATH, "//img[@data-testid='user-avatar']")))
            self.logger.success("Sesion iniciada con exito")

            self.logger.debug("Obteniendo access token")
            cookies = driver.get_cookies()
            self.logger.success("Cookies obtenidas con exito:"+str(cookies))
        except Exception as e:
            self.logger.error(str(e))
        finally:
            driver.quit()
