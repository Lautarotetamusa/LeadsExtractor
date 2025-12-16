from datetime import datetime
import json
import os
import re
from enum import IntEnum
from typing import Generator, Iterator
import urllib.parse
import concurrent.futures
from src.client import Client

import requests

from src.address import extract_street_from_address
from src.api import download_file
from src.propiedades_com.amenities import amenities
from src.property import OperationType, PlanType, Property, PropertyType, Ubication
from src.portal import Mode, Portal
from src.lead import Lead

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT is not None, "DATE_FORMAT is not seted"

MAIN_API = "https://propiedades.com/api/v3"

cancel_url = "https://propiedades.com/api/checkout/cancel"
highlight_url = "https://propiedades.com/api/slots/manage"
prediction_url = f"{MAIN_API}/property/sepomexid"
change_status_url = f"{MAIN_API}/property/status"
create_prop_url = f"{MAIN_API}/property/property"
login_url = f"{MAIN_API}/auth/login"
upload_url = "https://propiedadescom.s3.amazonaws.com/"

pattern = r"https://([a-zA-Z0-9]+)\.execute-api"
url_api_key_mapping = {
    "ujj28ojnla": "caTvsPv5aC7HYeXSraZPRaIzcNguO4sH9w9iUmWa",
    "76nst7dli4": "XpwAO6cXk889DbUOXB2tU7GWwRbyPIWE9ZAWEfL2",
    "ggcmh0sw5f": "ylbSWNPGu1yvyCclycC23wMA12nuPRy76rMvAto0",
    "obgd6o986k": "VAryo1IvOF7YMMbBU51SW8LbgBXiMYNS7ssvSyKS"
}

ujj28ojnla = "https://ujj28ojnla.execute-api.us-east-2.amazonaws.com/prod"
url_76nst7dli4 = "https://76nst7dli4.execute-api.us-east-2.amazonaws.com/prod"
ggcmh0sw5f = "https://ggcmh0sw5f.execute-api.us-east-2.amazonaws.com/prod"
obgd6o986k = "https://obgd6o986k.execute-api.us-east-2.amazonaws.com/prod"


sign_url = f"{ujj28ojnla}/property_sign_s3"
confirm_url = f"{ujj28ojnla}/confirm_upload_s3"

update_placeholder_url = f"{url_76nst7dli4}/upload_placeholders?country=MX"

get_leads_url = f"{ggcmh0sw5f}/get/leads"+"?page={page}&country=MX"
change_lead_status_url = f"{ggcmh0sw5f}/leads/status"

get_slots_url = f"{obgd6o986k}/get_packages_slot_user?country=MX"
get_properties_url = f"{obgd6o986k}/get_properties_admin?"


operation_type_map = {
    OperationType.SALE.value: "1",
    OperationType.RENT.value: "2"
}

property_type_map = {
    PropertyType.APARTMENT.value: "1",
    PropertyType.HOUSE.value: "2",
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


class PropiedadesClient(Client):
    """
    Cada url del tipo <name>.amazon tiene un 'x-api-key' asociada.
    Para que la peticion funcione se debe agregar este encabezado dependiente de la url
    """

    def _make_request(self, method: str, target_url: str, **kwargs) -> requests.Response | None:
        key = None
        match = re.search(pattern, target_url)
        if match:
            for name, k in url_api_key_mapping.items():
                if match.group(1) == name:
                    key = k

        if key is not None:
            if 'headers' not in kwargs:
                kwargs['headers'] = {
                    'x-api-key': key
                }
            else:
                kwargs['headers']['x-api-key'] = key

        return super()._make_request(method, target_url, **kwargs)


class Propiedades(Portal):
    def __init__(self):
        super().__init__(
            name="propiedades",
            contact_id_field="id",
            send_msg_field="",
            username_env="PROPIEDADES_USERNAME",
            password_env="PROPIEDADES_PASSWORD",
            filename=__file__
        )

        self.client = PropiedadesClient(self.login, unauthorized_codes=[401])
        self.load_session_params()

    def load_session_params(self):
        self.client.session.headers.update(self.params.get("headers", {}))
        self.client.session.cookies.update(self.params.get("cookies", {}))

    def get_leads(self, mode=Mode.ALL) -> Generator:
        first = True
        page = 1
        end = False

        while (not end) and (first or page is not None):
            url = get_leads_url.format(page=page)
            res = self.client.get(url)
            if res is None:
                break
            data = res.json()["leads"]

            page = data["page"]["next_page"]
            total = data["page"]["items"]

            if first:
                self.logger.debug(f"total: {total}")
                first = False

            # Si el modo es NEW Buscamos solamente los leads sin leer
            if mode == Mode.NEW:
                leads = []
                for lead in data["properties"]:
                    print(lead["id"], lead["status"])
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

        lead.propiedad = prop

        return lead

    def get_lead_property(self, property_id: str) -> dict | None:
        query = {
            "page": 1,
            "country": "MX",
            "identifier": 4,
            "purpose": 3,
            "search": property_id,
            "origin": 1
        }
        props = list(self.get_properties(query=query))

        if len(props) == 0:
            self.logger.error(f"No se encontro la propiedad con id {property_id}")
            return {"id": property_id} # Devolvemos solamente el property_id
        data = props[0]

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

    def change_lead_status(self, lead: dict, status=Status.CONTACTADO):
        id = lead[self.contact_id_field]
        self.logger.debug(f"Marcando como {status.name} al lead {id}")

        req = {
            "lead_id": id,
            "status": status,
            "country": "MX"
        }

        res = self.client.put(change_lead_status_url, json=req)
        if res is None:
            self.logger.error(f"No se pudo marcar al lead como {status.name}")
        else:
            self.logger.success(f"Se marco a lead {id} como {status.name}")

    def get_slots(self) -> Exception | list[object]:
        res = self.client.get(get_slots_url)

        if res is None or not res.ok:
            return Exception("cannot get available slots")

        data = res.json()

        return data.get("slots", [])

    def make_contacted(self, lead: dict):
        self.change_lead_status(lead, Status.CERRADA)

    def get_properties(self, status="", featured=False, query={}) -> Iterator[dict]:
        params = {
            "page": 1,
            "country": "MX",
            "identifier": 1,
            "purpose": 3,
            "order": "update_desc",
            "highlighted": featured
        }

        if query != {}:
            params = query

        if "internal" in query:
            params["search"] = query["internal"]["colony"]

        while params["page"] is not None:
            next = get_properties_url + urllib.parse.urlencode(params, doseq=True)
            res = self.client.get(next)
            if res is None or not res.ok:
                break

            data = res.json()

            params["page"] = data.get("paginate", {}).get("next_page")

            for property in data.get("properties", []):
                yield property

    def unhighlight(self, publication_id: str) -> Exception | None:
        self.logger.debug("unhighlighting", publication_id)
        data = {
            "properties_ids[]": str(publication_id)
        }
        res = self.client.post(cancel_url, data=data)

        if not res.ok:
            return Exception("cannot unhighlight the property", res.text)

        if not res.json().get("success", False):
            return Exception("cannot unhighlight the property: "+res.text)

        self.logger.success("unhighlighted succesffully", publication_id)

    # For the moment its not necessary
    def highlight(self, publication_id: str, plan: PlanType) -> Exception | None:
        self.logger.debug("highlighting publication", publication_id)
        slots = self.get_slots()
        if slots is Exception:
            return slots

        if len(slots) == 0:
            return Exception("there are not slots left")
        slot = slots[0]

        data = [{
            'property_id': publication_id,
            'slot_id': slot.get("slot_id")
        }]

        res = self.client.post(highlight_url, data=json.dumps(data))

        if res is None or not res.ok:
            return Exception("cannot highlight the property: "+res.text)

        if not res.json().get("success", False):
            return Exception("cannot highlight the property: "+res.text)

        return None

    def change_prop_status(self, publication_id: str, status: PropertyStatus) -> Exception | None:
        files = [
            ('status', (None, status.value)),
            ('properties_id[0]', (None, publication_id))
        ]
        res = self.client.post(change_status_url, files=files)

        if res is None or not res.ok:
            return Exception("error updating status"+res.text)

    # TODO: por qué hay dos funciones xd
    def update_property_status(self, property_id) -> Exception | None:
        # Update the property status to "active": "1"
        update_status_payload = {
            "properties_id": [property_id],
            "status": "1",
            "isNewPublish": True,
            "source": "token",
            "requestFromFunnel": True
        }
        res = self.client.post(change_status_url, json=update_status_payload)

        if res is None:
            return Exception("error creating the property")
        if not res.ok:
            return Exception("error creating the property: "+res.text)
        return None

    def unpublish(self, publication_id: str) -> Exception | None:
        """
        Propiedades dont let us change the status of more than 10 properties per day
        return self.change_prop_status(publication_id, PropertyStatus.VENDIDO)
        """
        return self.unhighlight(publication_id)

    def publish(self, property: Property) -> tuple[Exception, None] | tuple[None, str]:
        # TODO: If the prop doesnt have itnernal zone, searche it
        # else: # Otherwise get the address via API
        #     self.logger.debug("getting the location data")
        #     location_data = selfcasasyterrenos.get_location(property.ubication)
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
            "property[address][city]": property.internal["city"],  # Municipality
            "property[address][city_id]": property.internal["city_id"],  # Municipalty
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

            "property[type]": "1",  # Its always "1" for house and aparment, type_children indicates that.
            "property[type_children]": property_type_map.get(str(property.type), "2"),
            "property[purpose]": operation_type_map.get(str(property.operation_type), "1"),

            # Features
            "property[features][bedrooms]": str(property.rooms) if property.bathrooms is not None else "0",
            "property[features][bathrooms]": str(property.bathrooms) if property.bathrooms is not None else "0",
            "property[features][bathrooms_half]": str(property.half_bathrooms) if property.half_bathrooms is not None else "0",
            "property[features][parking_num]": str(property.parking_lots) if property.parking_lots is not None else "0",
            # Default 3 floors if the option exists. TODO: dont hardcode this
            "property[features][floor]": "3",
            "property[features][size_ground]": str(property.m2_total),
            "property[features][ground_unit]": "1",
            "property[features][size_house]": str(property.m2_covered),
            "property[features][size_house_unit]": "1",
            "property[features][gardens]": "false",
            # Its impossible to set the property_new in only one request, if needed must create the prop and later edit and add the property_new field
            # If the antiquity its 0, must send "1" and set property_new=true
            "property[features][property_old]": str(max(property.antiquity, 1)),
            "property[features][know_property_old]": "1",

            **amenities,
            **ubication,

            # Status
            "property[status]": "17", # Borrador
        }

        res = self.client.post(create_prop_url, data=payload, headers={
            'Content-Type': 'application/x-www-form-urlencoded'
        })

        if res is None:
            return Exception("error creating the property: "), None
        if not res.ok:
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

        err = self.upload_images(property_id, property.images)
        if err is not None:
            return err, None

        err = self.update_property_status(property_id)
        if err is not None:
            return err, None

        return None, property_id

    # The upload process of the images its complex, requoires multiple steps.
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
        max_fail_uploads = 4

        # 1. Sign all the images to upload to the s3 bucket
        jpeg_count = len(list(filter(lambda i : "jpg" in i["url"] or "jpeg" in i["url"], images)))
        sign_payload = {
            "jpeg": jpeg_count,
            # if no type present in the url, assume .png
            "png": len(images) - jpeg_count,
            "property_id": property_id,
            "country": "MX"
        }
        self.logger.debug("signing images")

        res = self.client.post(sign_url, json=sign_payload)
        if res is None:
            return Exception("cannot sign the image")
        if not res.ok:
            return Exception("cannot sign the image: "+res.json())
        sign_data = res.json()

        # necessary to update the placeholder of the images
        placeholder_payload = {
            "images": []
        }

        # 2. Upload the images
        # Index for each type of image
        index = {
            "png": 0,
            "jpeg": 0,
        }
        i = 0  # global index
        fail_uploads = 0  # cant of images that fails the uploading process

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

        # 3. Update placeholder
        self.logger.debug("update placeholders")
        res = self.client.post(update_placeholder_url, json=placeholder_payload)
        if res is None or not res.ok:
            return Exception("cannot update the image placeholder")
        self.logger.success("placeholders updated successfully")

        placeholders = res.json().get("placeholders", [])
        if len(placeholders) == 0:
            return Exception("cannot get the placeholders")

        # 4. Confirm the upload
        for placeholder in placeholders:
            placeholder["description"] = ""
            placeholder["country"] = "MX"

            self.logger.debug(f"confirmed {placeholder['image_id']}")
            res = self.client.post(confirm_url, json=placeholder)
            if res is None or not res.ok:
                return Exception(f"cannot confirm the image {placeholder['image_id']}")

            self.logger.success(f"image {placeholder['image_id']} confirmed successfully")

    def login(self, session="session"):
        self.logger.debug("Iniciando sesion")
        data = {
            "email": self.username,
            "password": self.password,
            "ssr": "1"
        }

        headers = {
            "Content-Type": "application/x-www-form-urlencoded",
        }

        res = requests.post(login_url, data=data, headers=headers)
        if not res.ok:
            raise Exception("cannot login")

        js = res.json()
        if not js.get("success"):
            raise Exception("error on login")

        data = js.get("data")
        if not data:
            raise Exception("error on login")

        self.update_params({
            "headers": {
                "Authorization": f"Bearer {data.get('token')}",
            },
            "cookies": {
                "userToken": data.get('token'),
                "PHPSESSID": res.cookies.get("PHPSESSID")
            }
        })
        self.load_session_params()
