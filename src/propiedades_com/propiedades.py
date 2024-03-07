from time import gmtime, strftime
from datetime import datetime
import json
from enum import IntEnum
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

from src.portal import Portal
from src.lead import Lead

#mis propiedades: https://propiedades.com/api/v3/property/MyProperties
#En este archivo tenemos todas las propieades previamente extraidas
with open("src/propiedades_com/properties.json") as f:
    PROPS = json.load(f)

DATE_FORMAT = "%d/%m/%Y"
API_URL = "https://ggcmh0sw5f.execute-api.us-east-2.amazonaws.com"

#Lista de estados posibles de un lead
#Los tomamos de la pagina
class Status(IntEnum):
    CONTACTADO = 2,
    CALIFICADO = 3,
    EN_PROCESO = 4,
    CONVERTIDO = 5,
    CERRADA = 6

def main():
    scraper = Propiedades()
    scraper.main()

class Propiedades(Portal):
    def __init__(self):
        super().__init__(
            name = "Propiedades",
            contact_id_field="id",
            send_msg_field="",
            username_env="PROPIEDADES_USERNAME",
            password_env="PROPIEDADES_PASSWORD",
            params_type="headers",
            filename=__file__
        )

    def get_leads(self) -> list[dict]:
        first = True
        page = 1
        end = False
        url = f"{API_URL}/prod/get/leads?page={page}&country=MX"

        leads = []
        while (not end) and (first == True or page != None):
            res = self.request.make(url)
            if res == None:
                return leads
            data = res.json()["leads"]
            
            page = data["page"]["next_page"]
            url = f"{API_URL}/prod/get/leads?page={page}&country=MX"
            total = data["page"]["items"]
            self.logger.debug(f"total: {total}")

            for lead in data["properties"]:
                if lead["status"] == Status.CONTACTADO:
                    self.logger.debug("Se encontro un lead ya conctactado, paramos")
                    end = True #Cuando encontramos un lead conctado paramos
                    break
                leads.append(lead)

        return leads

    def get_lead_info(self, raw_lead: dict) -> Lead:
        prop = self.get_lead_property(str(raw_lead["property_id"]))
        prop["address"] = raw_lead["address"]
        if prop["titulo"] == "": prop["titulo"] = prop["address"]

        lead = Lead()
        lead.set_args({
            "fuente": self.name,
            "fecha_lead": datetime.strptime(raw_lead["updated_at"], '%Y-%m-%d').strftime(DATE_FORMAT),
            "id": raw_lead["id"],
            "fecha": strftime(DATE_FORMAT, gmtime()),
            "nombre": raw_lead["name"],
            "telefono": raw_lead["phone"],
            "email": raw_lead["email"],
            "propiedad": prop,
        })
        return lead

    def get_lead_property(self, property_id: str):
        if property_id not in PROPS:
            self.logger.error(f"No se encontro la propiedad con id {property_id}")
            return {
                "id": "",
                "titulo": "",
                "link": "",
                "precio": "",
                "ubicacion": "",
                "tipo": "",
                "municipio": ""
            }
        data = PROPS[property_id]

        return {
            "id": data["id"],
            "titulo": data["short_address"],
            "link": data["url"],
            "precio": data["price"],
            "ubicacion": data["address"],
            "tipo": data["type_children_string"],
            "municipio": data["municipality"]
        }

    def make_contacted(self, id: str):
        status: Status = Status.CONTACTADO
        self.logger.debug(f"Marcando como {status.name} al lead {id}")
        
        url = f"{API_URL}/prod/leads/status"

        req = {
            "lead_id": id,
            "status": status,
            "country": "MX"
        }

        res = self.request.make(url, "PUT", json=req)
        if (res == None):
            self.logger.error(f"No se pudo marcar al lead como {status.name}")
        else:
            self.logger.success(f"Se marco a lead {id} como {status.name}")

    def login(self, session="propiedades_com/session"):
        login_url = "https://propiedades.com/login"
        options = Options()
        options.add_argument(f"--user-data-dir={session}") #Session
        #options.add_argument(f"--headless") #Session
        options.add_argument("--no-sandbox") # Necesario para correrlo como root dentro del container

        driver = webdriver.Chrome(options=options)
        access_token = None

        self.logger.debug("Iniciando sesion")
        try:
            driver.get(login_url)

            #Esperar que cargue la pagina
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
            self.logger.error("Ocurrio un error generando el access token")
            self.logger.error(str(e))
        finally:
            driver.quit()

if __name__ == "__main__":
    main()
