from datetime import datetime
import json
import os
from enum import IntEnum
from typing import Generator

import requests
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

from src.onedrive.main import download_file, token
from src.property import OperationType, Property, PropertyType
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

    def publish(self, property: Property) -> Exception | None:
        operation_type_map = {
            OperationType.SALE.value: "1",
            OperationType.RENT.value: "2"
        }

        property_type_map = {
            PropertyType.APARTMENT.value: "1",
            PropertyType.HOUSE.value: "2",
        }

        payload = {
            "from_funnel": "true",
            "source": "token",
            "property[status]": "1",

            "property[type]": property_type_map.get(str(property.type), "2"),
            "property[type_children]": property_type_map.get(str(property.type), "2"),
            "property[purpose]": operation_type_map.get(str(property.operation_type), "1"),
            "property[description]": property.description,
            # "property[features][gardens]": "false",
            "property[features][bedrooms]": str(property.rooms),
            "property[features][floor]": "1",
            "property[features][bathrooms]": str(property.bathrooms),
            "property[features][property_old]": str(property.antiquity),
            "property[features][size_house]": str(property.m2_covered),
            "property[features][size_ground]": str(property.m2_total),
            "property[features][garden_size]": 0,
            "property[services][none]": "1",
            "property[price][sale_price]": str(property.price),
            "property[price][currency]": property.currency.upper(),
            "property[address][sepomex_id]": "56001",
            "property[address][lat]": property.ubication.location.lat,
            "property[address][lng]": property.ubication.location.lng,
            "property[address][streetListData][0][id]": "-1",
            "property[address][streetListData][0][name]": "Otro",
            "property[address][externalNumListData][0][id]": "-1",
            "property[address][externalNumListData][0][name]": "Otro",
            "property[address][check_location]": "true",

            # Address properties
            "property[address][colony_id]": "56001",
            "property[address][state_id]": "14",
            "property[address][state]": "Jalisco",
            "property[address][city_id]": "533",
            "property[address][city]": "Guadalajara",
            "property[address][colony]": "Moderna",
            "property[address][zipcode]": "44190",
            "property[address][street]": "Avenida+8+de+Julio+823A",
            "property[address][num_ext]": "823A",
            "property[address][num_int]": "",
        }

        cookies = {
            "userToken": self.request.headers["Authorization"].replace("Bearer ", "")
        }
        
        params = {
            "apikey": os.getenv("ZENROWS_APIKEY"),
            "url": "https://propiedades.com/api/v3/property/property",
            # "js_render": "true",
            "antibot": "true",
            "premium_proxy": "true",
            "proxy_country": "mx",
            # "session_id": 10,
            "custom_headers": "true",
            "original_status": "true",
            "autoparse": "true"
        }
        res = requests.post("https://api.zenrows.com/v1/", 
            params=params, 
            data=payload,
            cookies=cookies,
            headers=self.request.headers
        )

        if res is None or not res.ok: return Exception("error creating the property: "+res.text)

        property_id = res.json().get("data", {}).get("id_property")
        if property_id is None: return Exception("cannot get the property id")
        self.logger.success("property published successfully id:" + str(property_id))

        # upload_images changes the x-api-key Header, we need to restore it to the original value
        prev_api_key = self.request.headers["x-api-key"]
        err = self.upload_images(property_id, property.images)
        if err != None: return err
        self.request.headers["x-api-key"] = prev_api_key
    
        # Update the property status to "active": "1"
        params["url"] = "https://propiedades.com/api/v3/property/status"
        update_status_payload = {
            "properties_id": [property_id],
            "status": "1",
            "isNewPublish": True,
            "source": "token",
            "requestFromFunnel": True
        }
        res = requests.post("https://api.zenrows.com/v1/", 
            params=params, 
            json=update_status_payload, 
            cookies=cookies, 
            headers=self.request.headers
        )

        if res is None or not res.ok: return Exception("error creating the property: "+res.text)
        return None

    # The upload process of the images its complex, requires multiple steps.
    # 1. Sign the images with amazon s3 bucket
    #   post request to the sign_url with the sign_payload data
    # 2. Upload the images to the s3 bucket
    #   a multipart/form-data request toe the upload_url with the data retrieved in the step 1
    # 3. Update the property placeholder of the image
    #   The placeholder contains the position of each image
    #   its a single POST requests to the 76nst.../prod/update_placeholders with the placeholder_payload.
    #   This step give us the image_id of each image.
    # 4. Confirm the upload
    #   one POST request per image to ujj../prod/confirm_upload_s3
    def upload_images(self, property_id: str, images: list[dict[str, str]]) -> Exception | None:
        sign_url = "https://ujj28ojnla.execute-api.us-east-2.amazonaws.com/prod/property_sign_s3"
        update_placeholder_url = "https://76nst7dli4.execute-api.us-east-2.amazonaws.com/prod/upload_placeholders?country=MX"
        upload_url = "https://propiedadescom.s3.amazonaws.com/"
        confirm_url = "https://ujj28ojnla.execute-api.us-east-2.amazonaws.com/prod/confirm_upload_s3"

        ## 1. Sign all the images to upload to the s3 bucket
        jpeg_count = len(list(filter(lambda i : "jpg" in i["url"] or "jpeg" in i["url"], images)))
        sign_payload = {
            "jpeg": jpeg_count,
            # if no type present in the url, assume .png
            "png": len(images) - jpeg_count,
            "property_id": property_id,
            "country": "MX"
        }
        self.logger.debug("signing images")

        self.request.headers["x-api-key"] = "caTvsPv5aC7HYeXSraZPRaIzcNguO4sH9w9iUmWa"
        res = self.request.make(sign_url, "POST", json=sign_payload)
        if res is None: return Exception("cannot sign the image")
        if not res.ok: return Exception("cannot sign the image: "+res.json())
        sign_data = res.json()

        # necessary to update the placeholder of the images
        placeholder_payload = {
            "images": []
        }

        ## 2. Upload the images
        # Index for each type of image
        index = {
            "png": 0,
            "jpeg": 0,
        }
        i = 0 # global index
        for image in images:
            self.logger.debug(f"downloading {image['url']}")
            img_data = download_file(token, image["url"])
            if img_data is None:
                return Exception(f"cannot download the image {image['url']}")

            img_type = "jpeg" if "jpeg" in image["url"] or "jpg" in image["url"] else "png"
            img_sign = sign_data[img_type][index[img_type]]
            fields = img_sign["response_s3"]["fields"]

            placeholder_payload["images"].append({
                "property_id": property_id,
                "file_name": img_sign["name_image"],
                "position": i # the first image in the property list its going to be the principal
            })

            files = [
                ('Content-Type', (None, f"image/{img_type}")),
                ('Cache-Control', (None, "max-age=2592000")),
                ('key', (None, fields["key"])),
                ('AWSAccessKeyId', (None, fields["AWSAccessKeyId"])),
                ('policy', (None, fields["policy"])),
                ('signature', (None, fields["signature"])),
                ('file', (img_sign["name_image"], img_data, "multipart/form-data")),
            ]
            self.logger.debug(f"uploading {image['url']}")
            res = requests.post(upload_url, files=files)
            if not res.ok: return Exception("cannot upload the image: "+res.text)
            self.logger.success("image uploaded successfully")

            index[img_type] += 1
            i += 1

        ## 3. Update placeholder
        self.logger.debug("update placeholders")
        # This url has another x-api-key
        self.request.headers["x-api-key"] = "XpwAO6cXk889DbUOXB2tU7GWwRbyPIWE9ZAWEfL2"
        res = self.request.make(update_placeholder_url, "POST", json=placeholder_payload)
        if res is None or not res.ok:
            return Exception("cannot update the image placeholder")
        self.logger.success("placeholders updated successfully")

        placeholders = res.json().get("placeholders", [])
        if len(placeholders) == 0: return Exception("cannot get the placeholders")

        ## 4. Confirm the upload
        for placeholder in placeholders:
            placeholder["description"] = ""
            placeholder["country"] = "MX"

            self.logger.debug(f"confirmed {placeholder['image_id']}")
            # This url has another x-api-key
            self.request.headers["x-api-key"] = "caTvsPv5aC7HYeXSraZPRaIzcNguO4sH9w9iUmWa"
            res = self.request.make(confirm_url, "POST", json=placeholder)
            if res is None or not res.ok:
                return Exception(f"cannot confirm the image {placeholder['image_id']}")

            print(placeholder["image_id"])
            print(placeholder["image_original_url"])

            self.logger.success(f"image {placeholder['image_id']} confirmed successfully")


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
