import json
import time
import os
import uuid
import requests
import urllib.parse
from datetime import datetime, date, timedelta
from typing import Iterator

from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

from src.api import download_file
from src.property import Property, OperationType
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

    def unpublish(self, publication_id: str) -> Exception | None:
        unpublish_url = f"{API_URL}/property/{publication_id}"

        payload = {
            "id": publication_id,
            "status": "inactive"
        }

        res = self.request.make(unpublish_url, "PATCH", json=payload)
        if res is None:
            return Exception(f"error unpublishing the property with id {publication_id}")
        if not res.ok:
            return Exception(f"error unpublishing the property with id {publication_id}. err: {res.text}")

    def highlight(self, publication_id: str) -> Exception | None:
        url = f"{API_URL}/featured_property/"
        days_duration = 3 # i dont know if this can be greater 
        payload = {
            "properties": [
                {
                    "end_date": (date.today() + timedelta(days=days_duration)).strftime("%Y-%m-%d"), # "2025-08-05",,
                    "property": publication_id,
                    "start_date": date.today().strftime("%Y-%m-%d"), # "2025-05-05",
                    "status": 1
                }
            ]
        }
        res = self.request.make(url, "POST", json=payload)
        if res is None:
            return Exception(f"error unpublishing the property with id {publication_id}")
        if not res.ok:
            return Exception(f"error unpublishing the property with id {publication_id}. err: {res.text}")

    def get_properties(self, status="published", featured=True) -> Iterator[dict]:
        url = f"{API_URL}/my_property?"
        params = {
            "page": 1,
            "ordering": "-created",
            "status": status,
            "featured": featured
        }

        next = url + urllib.parse.urlencode(params, doseq=True)
        while next is not None:
            res = self.request.make(next, "GET")
            if res is None:
                break

            data = res.json()
            for result in data.get("results"):
                yield result

            next = data.get("next")

    def publish(self, property: Property) -> tuple[Exception, None] | tuple[None, str]:
        prop_id = self.publish_first_step(property)
        if prop_id is None:
            return Exception("cannot obtain the property id"), None
        err = self.add_ubication(prop_id, property)
        if err is not None:
            return err, None

        err = self.upload_images(prop_id, property.images)
        if err is not None:
            return err, None

        return None, str(prop_id)

    # The property its pending, returns the property id,
    def publish_first_step(self, property: Property) -> int | None:
        operation_type_map = {
            "sell": 1,
            "rent": 2
        }
        property_type_map = {
            "house": "18",
            "department": "19"
        }

        # casasyterrenos uses html description, its necessary to do this to see correctly the enters 
        description = property.description.replace("\n", "<br>")

        payload = {
            "bathrooms": property.bathrooms,
            "rooms": property.rooms,
            # "floor_number": 0,
            "sqr_mt_lot": property.m2_total,
            "sqr_mt_construction": property.m2_covered,
            "name": property.title,
            "description": f"<div><!--block-->Ubicacion: {property.ubication.address}<br>{description}!</div>",
            "property_type": property_type_map[str(property.type)],
            "operation_type": [operation_type_map[str(property.operation_type)]],
            # "membership": "434451",
            # "currency_transfer": property.currency.lower(),
            "parking": property.parking_lots,
            "age": property.antiquity,
            "half_bathrooms": property.half_bathrooms,
            "virtual_tour_url": property.virtual_route,
            "youtube_url": property.video_url
        }

        if property.operation_type == OperationType.RENT.value:
            payload["price_rent"] = int(property.price)
            payload["currency_rent"] = property.currency.lower()
        elif property.operation_type == OperationType.SALE.value:
            payload["price_sale"] = int(property.price)
            payload["currency_sale"] = property.currency.lower()
        else:
            self.logger.error(f"unexpected operation type {property.operation_type}")
            return None

        res = self.request.make(API_URL + "/property", "POST", json=payload)
        if res is None:
            return None
        data = res.json()
        return data.get("id")

    def add_ubication(self, prop_id: int, property: Property) -> Exception | None:
        ubication = internal_ubications[property.ubication.address]["internal"]

        payload = {
            "latitude": property.ubication.location.lat,
            "longitude": property.ubication.location.lng,
            **ubication
        }

        res = self.request.make(f"{API_URL}/property/{prop_id}", "PATCH", json=payload)
        if res is None or not res.ok:
            return Exception("error adding ubication data")
        return None

    # 1. Sign the all the in one POST request images
    # 2. Upload the images with one POST request
    #       for each image sending the sign data
    # 3. Update the property, add the all images to the property
    def upload_images(self, property_id: int, images: list[dict[str, str]]):
        sign_url = "https://cytpanel.casasyterrenos.com/api/v1/aws_signature_url"
        upload_url = "https://cyt-media.s3.amazonaws.com/"

        sign_payload = {
            "property": property_id,
            "urls": []
        }
        for image in images:
            img_type = "png" if "png" in image["url"] else "jpeg"
            sign_payload["urls"].append({
                "format": img_type
            })

        res = self.request.make(sign_url, "POST", json=sign_payload)
        if res is None or not res.ok:
            return Exception("cannot sign the images")
        sign_data = res.json()

        i = 0
        uploaded_images = []
        for image in images:
            self.logger.debug(f"downloading {image['url']}")
            img_data = download_file(image["url"])
            if img_data is None:
                return Exception(f"cannot download the image {image['url']}")

            img_type = "png" if "png" in image["url"] else "jpeg"
            fields: dict[str, str] = sign_data[i]["fields"]
            files = []
            for key, value in fields.items():
                files.append(
                    # (field_name, (filename, value))
                    (key, (None, value))
                )
            files.append(
                ('file', (f"image.{img_type}", img_data, f"image/{img_type}"))
            )

            self.logger.debug(f"uploading image {image['url']}")
            # If we pass the Authorization token aws returns an error
            res = requests.post(upload_url, files=files)
            if not res.ok:
                self.logger.error(res.text)
                return Exception(f"cannot upload the image {image['url']}")
            self.logger.success("upload successfully")

            location = res.headers.get("Location")
            if location is None:
                return Exception("cannot get the image location")

            uploaded_images.append({
                # random id, idk if its necessary
                "id": f"images/{img_type}/{i}/{str(uuid.uuid4())}",
                # replace encoded char '/'
                "url": location.replace("%2F", "/")
            })

            i += 1

        res = self.request.make(f"{API_URL}/property/{property_id}", "PATCH", json={
            "images": uploaded_images
        })
        if res is None or not res.ok:
            return Exception("cannot update the property images")

        self.logger.success("property uploaded successfully")
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
