from datetime import datetime
import json
import os
from enum import IntEnum
from typing import Generator, Iterator
import urllib.parse
import concurrent.futures

import requests
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

from src.address import extract_street_from_address
from src.api import download_file
from src.propiedades_com.amenities import amenities
from src.property import OperationType, PlanType, Property, PropertyType, Ubication
from src.portal import Mode, Portal
from src.lead import Lead

# mis propiedades: https://propiedades.com/api/v3/property/MyProperties
# En este archivo tenemos todas las propieades previamente extraidas
with open("src/propiedades_com/properties.json") as f:
    PROPS = json.load(f)

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT is not None, "DATE_FORMAT is not seted"
API_URL = "https://ggcmh0sw5f.execute-api.us-east-2.amazonaws.com"

MAIN_API = "https://propiedades.com/api/v3"

ZENROWS_PARAMS = {
    "apikey": os.getenv("ZENROWS_APIKEY"),
    "url": "",
    # "js_render": "true",
    "antibot": "true",
    "premium_proxy": "true",
    "proxy_country": "mx",
    # "session_id": 10,
    "custom_headers": "true",
    "original_status": "true",
    "autoparse": "true"
}

# Lista de estados posibles de un lead
class Status(IntEnum):
    NUEVO = 1
    CONTACTADO = 2
    CALIFICADO = 3
    EN_PROCESO = 4
    CONVERTIDO = 5
    CERRADA = 6

class PropertyStatus(IntEnum):
    PAUSADO = 1
    ARCHIVADO = 9
    VENDIDO = 12

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
                    if lead["status"] == Status.CERRADA:
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


    def get_location(self, ubication: Ubication) -> dict | None: 
        prediction_url = f"{MAIN_API}/property/sepomexid?address={ubication.address}&lat={ubication.location.lat}&lon={ubication.location.lng}"

        cookies = {
            "userToken": self.request.headers["Authorization"].replace("Bearer ", "")
        }
        
        params = ZENROWS_PARAMS.copy()
        params["url"] = prediction_url

        res = requests.get("https://api.zenrows.com/v1/", 
            params=params, 
            cookies=cookies,
            headers=self.request.headers
        )

        data = res.json().get("data")
        if data is None or not "location" in data: 
            self.logger.error("cannot get address internal location: " + str(res.json()))
            return None

        return data["location"]

    def get_properties(self, status="", featured=False, query={}) -> Iterator[dict]:
        url = "https://obgd6o986k.execute-api.us-east-2.amazonaws.com/prod/get_properties_admin?"
        api_key = "VAryo1IvOF7YMMbBU51SW8LbgBXiMYNS7ssvSyKS" 

        params = {
            "page": 1,
            "country": "MX",
            "identifier": 1,
            "purpose": 3,
            "order": "update_desc",
            "highlighted": featured
        }

        if "internal" in query:
            params["search"] = query["internal"]["colony"]

        prev_api_key = self.request.headers["x-api-key"]
        while params["page"] is not None:
            next = url + urllib.parse.urlencode(params, doseq=True)
            self.request.headers["x-api-key"] = api_key
            res = self.request.make(next, "GET")
            if res is None or not res.ok: 
                self.request.headers["x-api-key"] = prev_api_key
                break

            data = res.json()

            params["page"] = data.get("paginate", {}).get("next_page")

            for property in data.get("properties", []):
                yield property

        self.request.headers["x-api-key"] = prev_api_key

    # For the moment its not necessary 
    def highlight(self, publication_id: str, plan: PlanType) -> Exception | None:
        return super().highlight(publication_id, plan)

    def change_prop_status(self, publication_id: str, status: PropertyStatus) -> Exception | None:
        # Update the property status to "active": "1"
        params = ZENROWS_PARAMS.copy()
        params["url"] = f"{MAIN_API}/property/status"
        cookies = {
            "userToken": self.request.headers["Authorization"].replace("Bearer ", "")
        }
        files = [
            ('status', (None, status.value)),
            ('properties_id[0]', (None, publication_id))
        ]
        res = requests.post("https://api.zenrows.com/v1/", 
            params=params, 
            files=files,
            cookies=cookies, 
            headers=self.request.headers
        )

        if res is None or not res.ok: 
            return Exception("error updating status"+res.text)
        print(res.json())

    def unpublish(self, publication_id: str) -> Exception | None:
        return self.change_prop_status(publication_id, PropertyStatus.VENDIDO)

    def publish(self, property: Property) -> tuple[Exception, None] | tuple[None, str]:
        operation_type_map = {
            OperationType.SALE.value: "1",
            OperationType.RENT.value: "2"
        }

        property_type_map = {
            PropertyType.APARTMENT.value: "1",
            PropertyType.HOUSE.value: "2",
        }

        # TODO: If the prop doesnt have itnernal zone, searche it
        # else: # Otherwise get the address via API
        #     self.logger.debug("getting the location data")
        #     location_data = self.get_location(property.ubication)
        #     if location_data is None:
        #         return Exception("cannot get location data"), None
        #     self.logger.success("location data obtained successfully")

        ubication = {
            "property[address][lat]": property.ubication.location.lat,
            "property[address][lng]": property.ubication.location.lng,
            "property[address][num_ext]": "",
            "property[address][num_int]": "",
            "property[address][state]": property.internal["state"],
            "property[address][state_id]": property.internal["state_id"],
            "property[address][city]": property.internal["city"], # Municipality
            "property[address][city_id]": property.internal["city_id"], # Municipalty
            "property[address][colony]": property.internal["colony"],
            "property[address][colony_id]": property.internal["colony_id"],
            "property[address][sepomex_id]": property.internal["colony_id"],
            "property[address][zipcode]": property.internal["zipcode"],
            "property[address][street]": extract_street_from_address(property.ubication.address),
            "property[address][check_location]": "true",
        }

        payload = {
            # Defaults
            "from_funnel": "true",
            "source": "token",
            "id": "0",

            "property[price][sale_price]": str(property.price),
            "property[price][currency]": property.currency.upper(),
            "property[description]": property.description,

            "property[type]": "1", # Its always "1" for house and aparment, type_children indicates that.
            "property[type_children]": property_type_map.get(str(property.type), "2"),
            "property[purpose]": operation_type_map.get(str(property.operation_type), "1"),

            # Features
            "property[features][bedrooms]": str(property.rooms) if property.bathrooms is not None else "0",
            "property[features][bathrooms]": str(property.bathrooms) if property.bathrooms is not None else "0",
            "property[features][bathrooms_half]": str(property.half_bathrooms) if property.half_bathrooms is not None else "0",
            "property[features][parking_num]": str(property.parking_lots) if property.parking_lots is not None else "0",
            "property[features][floor]": "3", # Default 3 floors if the option exists. TODO: dont hardcode this
            "property[features][size_ground]": str(property.m2_total),
            "property[features][ground_unit]": "1",
            "property[features][size_house]": str(property.m2_covered),
            "property[features][size_house_unit]": "1",
            "property[features][gardens]": "false",
            # Its impossible to set the property_new in only one request, if needed must create the prop and later edit and add the property_new field
            "property[features][property_old]": str(max(property.antiquity, 1)), # If the antiquity its 0, must send "1" and set property_new=true
            "property[features][know_property_old]": "1",

            **amenities,
            **ubication,

            # Status
            "property[status]": "17", # Borrador
        }

        cookies = {
            "userToken": self.request.headers["Authorization"].replace("Bearer ", "")
        }
        
        params = ZENROWS_PARAMS.copy()
        params["url"] = f"{MAIN_API}/property/property"
        res = requests.post("https://api.zenrows.com/v1/", 
            params=params, 
            cookies=cookies,
            headers={
                **self.request.headers,
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            data=payload
        )

        if res is None or not res.ok: 
            return Exception("error creating the property: "+res.text), None
        response = res.json()
        success = response.get("success") 
        if not success: 
            error = response.get("errors", {}).get("message", "")
            return Exception(f"error creating the property: {error}"), None

        property_id = response.get("data", {}).get("id_property")
        if property_id is None: 
            return Exception("cannot get the property id"), None
        self.logger.success("property published successfully id:" + str(property_id))

        # upload_images changes the x-api-key Header, we need to restore it to the original value
        prev_api_key = self.request.headers["x-api-key"]
        err = self.upload_images(property_id, property.images)
        if err != None: return err, None
        self.request.headers["x-api-key"] = prev_api_key
    
        # Update the property status to "active": "1"
        params["url"] = f"{MAIN_API}/property/status"
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

        if res is None or not res.ok: 
            return Exception("error creating the property: "+res.text), None
        return None, property_id



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

        max_fail_uploads = 4 

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
        fail_uploads = 0 # cant of images that fails the uploading process

        # Configura el número máximo de hilos (ajusta según necesidad)
        MAX_WORKERS = 5

        # Precomputación de metadatos para evitar condiciones de carrera
        precomputed = []
        index_counter = {'jpeg': 0, 'png': 0}
        for idx, image in enumerate(images):
            img_type = "jpeg" if "jpeg" in image["url"] or "jpg" in image["url"] else "png"
            
            # Verificar suficiencia de firmas
            if index_counter[img_type] >= len(sign_data[img_type]):
                precomputed.append({
                    'error': f"No hay suficientes firmas para tipo {img_type}",
                    'position': idx
                })
            else:
                img_sign = sign_data[img_type][index_counter[img_type]]
                index_counter[img_type] += 1
                fields = img_sign["response_s3"]["fields"]
                precomputed.append({
                    'url': image['url'],
                    'img_type': img_type,
                    'img_sign': img_sign,
                    'fields': fields,
                    'position': idx
                })

        # Función para procesar cada imagen
        def process_image(item):
            if 'error' in item:
                return False, None, item['error'], False
            
            # Descargar imagen
            img_data = download_file(item["url"])
            if img_data is None:
                return False, None, f"No se pudo descargar {item['url']}", True
            
            # Preparar datos para subida
            files = [
                ('Content-Type', (None, f"image/{item['img_type']}")),
                ('Cache-Control', (None, "max-age=2592000")),
                ('key', (None, item['fields']['key'])),
                ('AWSAccessKeyId', (None, item['fields']['AWSAccessKeyId'])),
                ('policy', (None, item['fields']['policy'])),
                ('signature', (None, item['fields']['signature'])),
                ('file', (item['img_sign']['name_image'], img_data, "multipart/form-data")),
            ]
            
            # Subir imagen
            res = requests.post(upload_url, files=files)
            if not res.ok:
                return False, None, f"Error subiendo {item['url']}: {res.text}", False
            
            # Éxito: retornar metadatos para el payload
            return True, {
                "property_id": property_id,
                "file_name": item['img_sign']['name_image'],
                "position": item['position']
            }, None, False

        # Ejecución paralela
        fail_uploads = 0
        critical_failure = None
        placeholder_entries = []

        with concurrent.futures.ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
            futures = {executor.submit(process_image, item): item for item in precomputed}
            
            for future in concurrent.futures.as_completed(futures):
                # Verificar si debemos detener el procesamiento
                if critical_failure is not None or fail_uploads > max_fail_uploads:
                    break
                    
                item = futures[future]
                try:
                    success, entry, error_msg, is_download_failure = future.result()
                    
                    if success:
                        self.logger.success(f"Imagen subida: {item.get('url', 'N/A')}")
                        placeholder_entries.append(entry)
                    else:
                        self.logger.error(error_msg)
                        fail_uploads += 1
                        
                        # Manejar errores críticos
                        if is_download_failure:
                            critical_failure = error_msg
                            
                except Exception as e:
                    fail_uploads += 1
                    self.logger.error(f"Error inesperado: {str(e)}")
                
                # Verificar límites de errores después de cada tarea
                if critical_failure is not None or fail_uploads > max_fail_uploads:
                    # Cancelar tareas pendientes
                    for f in futures:
                        if not f.done():
                            f.cancel()
                    break

        # Gestionar resultados finales
        if critical_failure is not None:
            return Exception(critical_failure)
            
        if fail_uploads > max_fail_uploads:
            return Exception(f"{fail_uploads} imágenes fallaron, cancelando publicación")

        # Agregar entradas exitosas al payload (ordenadas por posición original)
        placeholder_entries.sort(key=lambda x: x['position'])
        placeholder_payload["images"].extend(placeholder_entries)

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
        #
        # def sign_image(placeholder):
        #     placeholder["description"] = ""
        #     placeholder["country"] = "MX"
        #
        #     self.logger.debug(f"confirmed {placeholder['image_id']}")
        #     # This url has another x-api-key
        #     self.request.headers["x-api-key"] = "caTvsPv5aC7HYeXSraZPRaIzcNguO4sH9w9iUmWa"
        #     res = self.request.make(confirm_url, "POST", json=placeholder)
        #     if res is None or not res.ok:
        #         return Exception(f"cannot confirm the image {placeholder['image_id']}")
        #
        #     self.logger.success(f"image {placeholder['image_id']} confirmed successfully")

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

# Función para procesar cada imagen
def process_image(item):
    if 'error' in item:
        return False, None, item['error'], False
    
    # Descargar imagen
    img_data = download_file(item["url"])
    if img_data is None:
        return False, None, f"No se pudo descargar {item['url']}", True
    
    # Preparar datos para subida
    files = [
        ('Content-Type', (None, f"image/{item['img_type']}")),
        ('Cache-Control', (None, "max-age=2592000")),
        ('key', (None, item['fields']['key'])),
        ('AWSAccessKeyId', (None, item['fields']['AWSAccessKeyId'])),
        ('policy', (None, item['fields']['policy'])),
        ('signature', (None, item['fields']['signature'])),
        ('file', (item['img_sign']['name_image'], img_data, "multipart/form-data")),
    ]
    
    # Subir imagen
    res = requests.post(upload_url, files=files)
    if not res.ok:
        return False, None, f"Error subiendo {item['url']}: {res.text}", False
    
    # Éxito: retornar metadatos para el payload
    return True, {
        "property_id": property_id,
        "file_name": item['img_sign']['name_image'],
        "position": item['position']
    }, None, False
