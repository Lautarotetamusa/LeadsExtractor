import json
import time
import os
from time import gmtime, strftime
from datetime import datetime
from typing import Iterator
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

from src.numbers import parse_number
from src.portal import Mode, Portal
from src.lead import Lead

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT != None, "DATE_FORMAT is not seted"

API_URL = "https://cytpanel.casasyterrenos.com/api/v1"
CLIENT_ID = "4je1v2kfou9e9plpv6vf0vmnll"

def main():
    scraper = CasasYTerrenos()
    scraper.main()
def first_run():
    scraper = CasasYTerrenos()
    scraper.first_run()

class CasasYTerrenos(Portal):
    def __init__(self):
        super().__init__(
            name = "Casas y terrenos",
            contact_id_field = "id",
            send_msg_field = "",
            username_env = "CASASYTERRENOS_USERNAME",
            password_env = "CASASYTERRENOS_PASSWORD",
            params_type = "headers",
            filename=__file__
        )

    def get_leads(self, mode=Mode.NEW) -> Iterator[list[dict]]:
        page = 1

        if mode == Mode.NEW:
            status = "1" #Filtramos solamente los leads nuevos
            url = f"{API_URL}/list_contact/?page={page}&status={status}"
        else:
            url = f"{API_URL}/list_contact/?page={page}"

        while url != None:
            res = self.request.make(url)
            if res == None:
                break
            data = res.json()

            url = data["next"]
            self.logger.debug(f"total: {data['count']}")

            yield data["results"]

    def make_contacted(self, id: str):
        self.logger.debug(f"Marcando como contactacto a lead {id}")
        url = f"{API_URL}/contact/{id}"

        data = {
            "id": id,
            "status": 2, # 2 -> Contactado por correo
            "status_description": "Correo"
        }
        res = self.request.make(url, 'PATCH', data=data)
        
        self.logger.success(f"Se contacto correctamente a lead {id}")

    def get_lead_info(self, raw_lead: dict) -> Lead:
        lead = Lead()
        lead.set_args({
            "fuente": self.name,
            "fecha_lead": datetime.strptime(raw_lead["created"], '%d-%m-%Y %H:%M:%S').strftime(DATE_FORMAT),
            "asesor_name": "",
            "id":  raw_lead["id"],
            "fecha": strftime(DATE_FORMAT, gmtime()),
            "nombre":  raw_lead["name"],
            "link": "",
            "email":  raw_lead["email"],
        })
        lead.set_propiedad({
            "id":  raw_lead["property_id"],
            "titulo":  raw_lead["property_title"],
            "link": f"https://www.casasyterrenos.com/propiedad/{raw_lead['property_id']}",
            "tipo":  raw_lead["property_type"],
        })

        telefono = parse_number(self.logger, raw_lead.get("phone", ""), None)
        if not telefono:
            telefono = parse_number(self.logger, raw_lead.get("phone", ""), "MX")
        lead.telefono = telefono or lead.telefono
        return lead

    def login(self, session="session"):
        login_url = "https://panel-pro.casasyterrenos.com/login"
        panel_url = "https://panel-pro.casasyterrenos.com/contacts"
        options = Options()
        options.add_argument(f"--user-data-dir={session}") #Session
        options.add_argument(f"--headless") #Session
        options.add_argument("--no-sandbox") # Necesario para correrlo como root dentro del container

        driver = webdriver.Chrome(options=options)
        access_token = None

        self.logger.debug("Generando un nuevo token de acceso")
        try:
            driver.get(panel_url)

            if (driver.current_url == login_url):
                self.logger.debug("La sesion no esta iniciada, iniciando..")
                driver.get(login_url)

                #Esperar que cargue la pagina
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
            if access_token != None:
                self.logger.success("Access token obtenido con exito")
                self.logger.success(access_token)
            else:
                self.logger.error("No se pudo obtener el access_token")

            #Esperamos para que el token sea correcto
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
