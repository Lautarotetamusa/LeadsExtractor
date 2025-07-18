from datetime import datetime
from enum import Enum
import os
import json
from typing import Any, Iterator
import requests
import concurrent.futures
import urllib.parse

from src.address import extract_street_from_address
from src.api import download_file
from src.property import Internal, PlanType, Property
from src.make_requests import ApiRequest
from src.portal import Mode, Portal
from src.lead import Lead
from src.inmuebles24.amenities import amenities

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT is not None, "DATE_FORMAT is not seted"
SITE_URL = "https://www.inmuebles24.com/"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
ZENROWS_API_KEY = os.getenv("ZENROWS_APIKEY")
PARAMS = {
    "apikey": ZENROWS_API_KEY,
    "url": "",
    # "js_render": "true",
    # "antibot": "true",
    "premium_proxy": "true",
    "proxy_country": "mx",
    # "session_id": 10,
    "custom_headers": "true",
    "original_status": "true",
    # "autoparse": "true"
}

publication_plan_map: dict[str, int] = {
    PlanType.HIGHLIGHTED.value: 102,
    PlanType.SUPER.value: 101
}

def extract_busqueda_info(data: dict | None) -> dict:
    if data is None:
        return {}

    features = data.get("property_features", {})
    lead_info = data.get("lead_info", {})

    busqueda = {
        "zonas": ','.join([i["name"] for i in data["searched_locations"]["streets"]]) if "streets" in data.get("searched_locations", {}) else "",
        "tipo": lead_info.get("search_type", {}).get("type", ""),
        "presupuesto": f"{lead_info.get('price', {}).get('min', '')}, {lead_info.get('price', {}).get('max', '')}",
        "cantidad_anuncios": lead_info.get("views", None),
        "contactos": lead_info.get("contacts", None),
        "inicio_busqueda": lead_info.get("started_search_days", None)
    }

    range_props = {
        "total_area": "total_area_xm2",
        "covered_area": "covered_area_xm2",
        "banios": "baths",
        "recamaras": "bedrooms",
    }
    for prop in range_props:
        try:
            value = features[range_props[prop]]
            busqueda[prop] = str(value["min"]) + ", " + str(value["max"])
        except KeyError:
            busqueda[prop] = ""

    return busqueda

class Status(Enum):
    pending = 1
    contacted = 2
    phone_error = 4

class Inmuebles24(Portal):
    def __init__(self):
        super().__init__(
                name="inmuebles24",
                contact_id_field="contact_publisher_user_id",
                send_msg_field="id",
                username_env="INMUEBLES24_USERNAME",
                password_env="INMUEBLES24_PASSWORD",
                params_type="headers",
                unauthorized_codes=[401],
                filename=__file__
                )
        self.api_req = ApiRequest(self.logger, ZENROWS_API_URL, PARAMS)
        self.request.cookies = {
            "allowCookies": "true",
            "changeInbox": "true",
            "crto_is_user_optout": "false",
            "G_ENABLED_IDPS": "google",
            "hashKey": self.request.headers["hashKey"],
            "hideWelcomeBanner": "true",
            "IDusuario": self.request.headers["idUsuario"],
            "JSESSIONID": "1971D320A1B78F9CDD06DAA4DF456B3C",
            "pasoExitoso": "{'trigger':'Continuar'&'stepId':1&'stepName':'Datos principales-Profesional'}",
            "reputationModalLevelSeen": "true",
            "reputationModalLevelSeen2": "true",
            "reputationTourLevelSeen": "true",
            "sessionId": self.request.headers["sessionId"],
            "showWelcomePanelStatus": "true",
            "tableCreditsOpen": "true",
            "usuarioPublisher": "true",
            "usuarioFormId": self.request.headers["idUsuario"],

            # TODO: dont hardcode this fields
            "usuarioFormApellido": "Residences",
            "usuarioFormEmail": "control.general@rebora.com.mx",
            "usuarioFormNombre": "RBA",
            "usuarioFormTelefono": "523341690109",
            "usuarioIdCompany": "50796870",
            "usuarioLogeado": "control.general@rebora.com.mx",
        }

    def login(self):
        self.logger.debug("Iniciando sesion")
        login_url = f"{SITE_URL}login_login.ajax"

        data = {
            "email": self.username,
            "password": self.password,
            "recordarme": "true",
            "homeSeeker": "true",
            "urlActual": SITE_URL
        }
        # self.request.headers["Content-Type"] = "application/x-www-form-urlencoded;charset=UTF-8"
        self.request.headers["Content-Type"] = "application/json; charset=UTF-8"

        res = self.api_req.make(login_url, "POST", data=data)
        if res is None:
            self.logger.error("Cant login")
            return

        data = res.json()
        contenido = data.get("contenido", {}) 
        if not contenido.get("result", False):
            self.logger.error(str(contenido.get("validation", {}).get("errores")))
            return

        self.request.headers = {
            "sessionId": data["contenido"]["sessionID"],
            "idUsuario": str(data["contenido"]["idUsuario"]),
            "hashKey": data["contenido"]["cookieHash"],
            "x-panel-portal": "24MX",
            "content-type": "application/json;charset=UTF-8",
        }

        with open(self.params_file, "w") as f:
            json.dump(self.request.headers, f, indent=4)
        self.logger.success("Sesion iniciada con exito")

    def get_leads(self, mode=Mode.NEW):
        status = "nondiscarded"  # Este campo lo pasamos siempre
        sort = "unread"  # Los no leidos estarán primeros
        first = True
        offset = 0
        total = 0
        limit = 100
        finish = False

        while first or (not finish and offset < total):
            leads_url = f"{SITE_URL}leads-api/publisher/leads?offset={offset}&limit={limit}&spam=false&status={status}&sort={sort}"
            self.logger.debug(f"GET {leads_url}")
            PARAMS["url"] = leads_url
            res = self.request.make(ZENROWS_API_URL, 'GET', params=PARAMS)
            if res is None:
                break
            offset += limit

            data = res.json()
            self.logger.debug("Response type:" + str(type(data)))
            if isinstance(data, list):
                data = data[0]
            if not isinstance(data, dict):
                self.logger.error("Unexpected response type")
                continue

            if first:
                total = data["paging"]["total"]
                self.logger.debug(f"Total: {total}")
                first = False

            if mode == Mode.NEW:    # Obtenemos todos los leads sin leer de este pagina
                leads = []
                for lead in data["result"]:
                    response_status = lead.get("contact_response_status", {}).get("name")
                    status = lead.get("statuses", [{}])[0]
                    if status == "UNREAD" or response_status == "Pendiente":
                        leads.append(lead)
                if len(leads) == 0: 
                    self.logger.debug("No se encontro ningun lead 'Pendiente' o 'UNREAD', paramos")
                    finish = True
                    break
                yield leads
            else:  # Obtenemos todos los leads
                yield data["result"]

    def get_lead_info(self, raw_lead):
        raw_lead_id = raw_lead["id"]
        contact_id = raw_lead[self.contact_id_field]
        raw_busqueda = self.get_busqueda_info(contact_id)
        busqueda = extract_busqueda_info(raw_busqueda)
        posting = raw_lead.get("posting", {})

        lead = Lead()
        phone = raw_lead.get("phone")
        if phone is None or phone == "":
            phone = raw_lead.get("lead_user", {}).get("phone")

        lead.set_args({
            # TODO: Comprobar que el mensaje haya sido enviado por el lead y no por nosotros
            "message": raw_lead.get("last_message", {}).get("text", ""),
            "lead_id": raw_lead_id,
            "contact_id": contact_id,
            "fuente": self.name,
            "fecha_lead": datetime.strptime(raw_lead["last_lead_date"], "%Y-%m-%dT%H:%M:%S.%f+00:00").strftime(DATE_FORMAT),
            "nombre": raw_lead.get("lead_user", {}).get("name"),
            "link": f"{SITE_URL}panel/interesados/{contact_id}",
            "email": raw_lead.get("lead_user", {}).get("email"),
            "telefono": phone,
        })
        utm_channel = raw_lead.get("actions", [{}])[0].get("type", {}).get("name")
        # Los 3 tipos de mensaje que recibimos en inmuebles24
        channel_map = {
            "CONTACT":  "inbox",
            "WHATSAPP": "whatsapp",
            "VIEW_DATA": "ivr",
        }
        if utm_channel is not None and utm_channel in channel_map:
            lead.set_args({
                "utm": {
                    "utm_channel": channel_map[utm_channel]
                }
            })

        lead.set_busquedas(busqueda)
        posting_id = posting.get("id", None)
        lead.set_propiedad({
            "id": posting.get("id", None),
            "titulo": posting.get("title", ""),
            "link": f"https://www.inmuebles24.com/propiedades/-{posting_id}.html",
            "precio": str(posting.get("price", {}).get("amount", "")),
            "ubicacion": posting.get("address", ""),
            "tipo": posting.get("real_estate_type", {}).get("name"),
            "municipio": posting.get("location", {}).get("parent", {}).get("name", "") # Ciudad
        })
        # Algunos leads pueden ver nuestro telefono sin necesariamente venir por una propiedad
        if lead.propiedad["id"] is None:
            lead.fuente = "viewphone"

        return lead

    #  Get the information about the searchers of the lead
    def get_busqueda_info(self, lead_id) -> dict | None:
        self.logger.debug("Extrayendo la informacion de busqueda del lead: "+lead_id)
        busqueda_url = f"{SITE_URL}leads-api/publisher/contact/{lead_id}/user-profile"

        self.logger.debug(f"GET {busqueda_url}")
        PARAMS["url"] = busqueda_url
        res = self.request.make(ZENROWS_API_URL, 'GET', params=PARAMS)

        if res is None:
            self.logger.error("No se pudo obtener la informacion de busqueda para el lead: "+lead_id)
            return None
        try:
            self.logger.success("Informacion de busqueda extraida con exito")
            return res.json()
        except requests.exceptions.JSONDecodeError:
            self.logger.error(f"El lead {lead_id} no tiene informacion de busqueda")
            return None

    def send_message_condition(self, lead) -> bool:
        # si no hay last_message es porque no le enviaste nd
        is_last_message = "last_message" in lead
        message_to = lead.get("last_message", {}).get("to")
        is_user = message_to == self.request.headers["idUsuario"]
        return not is_last_message or is_user

    # Usa el lead_id
    def send_message(self, id: str,  message: str):
        self.logger.debug(f"Enviando mensaje a lead {id}")
        msg_url = f"{SITE_URL}leads-api/leads/{id}/messages"

        data = {
            "is_comment": False,
            "message": message,
            "message_attachments": []
        }

        params = PARAMS.copy()
        params["url"] = msg_url
        #res = requests.post(ZENROWS_API_URL, params=params, json=data, headers=self.request.headers)
        self.logger.debug(f"POST {msg_url}")
        res = self.request.make(ZENROWS_API_URL, 'POST', params=params, json=data)

        if res is not None and res.status_code >= 200 and res.status_code < 300:
            self.logger.success(f"Mensaje enviado correctamente a lead {id}")
        else:
            self.logger.error(f"Error enviando mensaje al lead {id}")

    def _change_status(self, lead: dict, status: Status):
        contact_id = lead[self.contact_id_field]
        status_url = f"{SITE_URL}leads-api/publisher/contact/status/{contact_id}"

        id = lead["id"]
        params = PARAMS.copy()
        params["url"] = status_url
        params["autoparse"] = False
        data = {
            "lead_id": id,
            "lead_status_id": status.value
        }
        self.logger.debug(f"POST {status_url}")
        res = self.request.make(ZENROWS_API_URL, 'POST', params=params, json=data)

        if res is not None and res.status_code >= 200 and res.status_code < 300:
            self.logger.success(f"Se marco a lead {contact_id} como {status}")
        else:
            if res is not None:
                self.logger.error(res.content)
                self.logger.error(res.status_code)
            self.logger.error(f"Error marcando al lead {contact_id} como {status}")
        PARAMS["autoparse"] = True

    def get_last_prop_id(self) -> None | int: 
        posts = self.get_properties_page(1)
        if len(posts) == 0:
            self.logger.error("cannot get the properties")
            return None

        return posts[0]["postingId"]

    def get_properties_page(self, page=1) -> list[dict]:
        self.logger.debug("getting properties")
        list_url = f"https://www.inmuebles24.com/avisos-api/panel/api/v2/postings?"
        params = {
            "page": page,
            "limit": 20,
            "searchParameters": "sort:createdNewer&onlineFirst=true"
        }

        list_url += urllib.parse.urlencode(params)

        p = PARAMS.copy()
        p["url"] = list_url
        res = self.request.make(ZENROWS_API_URL, "GET", params=p)
        if res is None or not res.ok:
            return []

        return res.json().get("postings", [])

    def get_properties(self, status="ONLINE", featured=False, query={}) -> Iterator[dict]:
        self.logger.debug("getting properties")
        list_url = f"https://www.inmuebles24.com/avisos-api/panel/api/v2/postings?"
        # "page={page}&limit=20&searchParameters=status:{status};sort:createdNewer&onlineFirst=true"
        params = {
            "page": 1,
            "limit": 20,
            "searchParameters": f"status:{status}"
        }
        if "page" in query:
            params["page"] = query["page"]

        if "internal" in query:
            params["searchParameters"] += f";searchCode:{query['internal']['colony']}"

        posts = [1]
        while len(posts) > 0:
            url = list_url + urllib.parse.urlencode(params)

            p = PARAMS.copy()
            p["url"] = url
            self.logger.debug("GET " + str(url))
            res = self.request.make(ZENROWS_API_URL, "GET", params=p)
            if res is None or not res.ok:
                self.logger.error("cannot get the properties")
                break

            if not "page" in query:
                params["page"] += 1

            posts = res.json().get("postings", [])
            if posts is None:
                break
            for post in posts:
                yield post

    def highlight(self, publication_id: str, plan: PlanType) -> Exception | None:
        self.logger.debug("highlighting property plan: " + str(plan))
        url = "https://www.inmuebles24.com/avisos-api/panel/api/v2/posting/publish"

        payload = {
            # "postingAndPublisherId": {
            #     "146374728": 50322706
            # },
            "postingIds": [
                publication_id
            ],
            "publicationPlan": publication_plan_map[str(plan)]
        }

        params = PARAMS.copy()
        params["url"] = url
        res = self.request.make(ZENROWS_API_URL, "PUT", params=params, json=payload)
        if res is None or not res.ok:
            return Exception(f"cannot highlight the property {res.text if res is not None else ''}")

    def unpublish(self, publication_ids: list[str]) -> Exception | None:
        unpublish_url = "https://www.inmuebles24.com/avisos-api/panel/api/v2/posting/suspend"
        archive_url = "https://www.inmuebles24.com/avisos-api/panel/api/v2/posting/archive"

        payload = {
            "finishReasonId": "6", # Operation canceledstr
            "finishReasonText":	None,
            "postings": [id for id in publication_ids]
        }

        params = PARAMS.copy()
        params["url"] = unpublish_url
        res = self.request.make(ZENROWS_API_URL, "PUT", params=params, json=payload)
        if res is None or not res.ok:
            return Exception("cannot unpublish the properties")

        params["url"] = archive_url
        res = self.request.make(ZENROWS_API_URL, "PUT", params=params, json=payload)
        if res is None or not res.ok:
            return Exception("cannot archive the properties")

    def publish(self, property: Property) -> tuple[Exception, None] | tuple[None, str]:
        tipo_propiedad_map = {
            "house": "1",
        }
        currency_map = {
            "MXN": "10"
        }

        payload = {
            "aviso.titulo": property.title,
            "aviso.descripcion": property.description,
            "isDesarrollo-control": "false",
            "aviso.idTipoDePropiedad": tipo_propiedad_map[str(property.type)],
            "aviso.antiguedad": property.antiquity,
            "aviso.habitaciones": property.rooms,
            "aviso.garages": property.parking_lots,
            "aviso.banos": property.bathrooms,
            "aviso.mediosBanos": property.half_bathrooms,
            "monedaVenta": currency_map[property.currency],
            "precioVenta": property.price,
            "aviso.valorCondominio": "",
            "aviso.metrosCubiertos": property.m2_covered,
            "aviso.metrosTotales": property.m2_total,
            "lat": property.ubication.location.lat,
            "lng": property.ubication.location.lng,

            "southwest": "",
            "zoom": "13",
            "aviso.email": "",
            "aviso.idAviso": "",
            "checkContentEnhancerHidden": "",
            "guardarComoBorrador": "1",
            "irAlPaso": ""
        }

        err, id_geoloc = self.get_geolocation(property)
        if err is not None:
            return err, None

        payload = {
            **payload,
            "aviso.idProvincia": property.internal["state_id"],
            "aviso.idCiudad": property.internal["city_id"],
            "aviso.idZona": str(property.internal["colony_id"]),
            # "aviso.idSubZonaCiudad": "",
            "aviso.direccion": extract_street_from_address(property.ubication.address),
            "aviso.codigoPostal": "",
            "direccion.mapa": "exacta", # exacta | zona | mapa
            "aviso.claveInterna": "",
            "idGeoloc": id_geoloc,
        }

        cookies = {
            "allowCookies": "true",
            "changeInbox": "true",
            "crto_is_user_optout": "false",
            "G_ENABLED_IDPS": "google",
            "hashKey": self.request.headers["hashKey"],
            "hideWelcomeBanner": "true",
            "IDusuario": self.request.headers["idUsuario"],
            "pasoExitoso": "{'trigger':'Continuar'&'stepId':1&'stepName':'Datos principales-Profesional'}",
            "reputationModalLevelSeen": "true",
            "reputationModalLevelSeen2": "true",
            "reputationTourLevelSeen": "true",
            "sessionId": self.request.headers["sessionId"],
            "showWelcomePanelStatus": "true",
            "tableCreditsOpen": "true",
            "usuarioPublisher": "true",
            "usuarioFormId": self.request.headers["idUsuario"],

            # TODO: dont hardcode this fields
            "usuarioFormApellido": "Residences",
            "usuarioFormEmail": "control.general@rebora.com.mx",
            "usuarioFormNombre": "RBA",
            "usuarioFormTelefono": "523341690109",
            "usuarioIdCompany": "50796870",
            "usuarioLogeado": "control.general@rebora.com.mx",
        }

        first_step_url = "https://www.inmuebles24.com/publicarPasoDatosPrincipales.bum"
        res = self.api_req.make(first_step_url, "POST", data=payload, cookies=cookies)
        if res is None or not res.ok:
            return Exception("error in creation of the property"), None
        self.logger.success("property succesfully publicated")

        # The request dont give us the id of the property created, then we save the last prop id and updated
        self.last_prop_id = self.get_last_prop_id()

        self.logger.debug("adding amenities")
        amenities_url = f"https://www.inmuebles24.com/publicarPasoCaracteristicas.bum?idaviso={self.last_prop_id}&checkContentEnhancerUrl=true"
        res = self.api_req.make(amenities_url, "POST", data=amenities, cookies=cookies)
        if res is None or not res.ok:
            return Exception("error adding the amenities"), None
        self.logger.success("amenities succesfully added")

        err = self.upload_images(str(self.last_prop_id), property, cookies)
        if err is not None:
            self.logger.error("error uploading images")
            return err, None

        # Publish
        self.logger.debug("publishing property")
        publish_url = f"{SITE_URL}publicarPasoPublicacion.bum?idaviso={self.last_prop_id}&idProducto=101491&idPlanDePublicacion=103"
        res = self.api_req.make(publish_url, "GET", cookies=cookies)
        if res is None or not res.ok:
            return Exception("error publishing the property"), None

        return None, str(self.last_prop_id)

    def save_geolocation(self, prop: Property, internal_name: str) -> tuple[Exception, None] | tuple[None, dict]:
        save_url = "https://www.inmuebles24.com/publicar_guardarGeoloc.ajax"

        payload = {
            "idGeolocXX": "",
            "geoloc.type": "street_address",
            "geoloc.location_type": "RANGE_INTERPOLATED",
            "geoloc.streetaddress": prop.ubication.address,
            # "geoloc.street_number": "",
            # "geoloc.route": "Avenida+Juan+Palomar+y+Arias",
            "geoloc.country": "Mexico",
            "geoloc.administrative_area_level_1": prop.internal["state"],
            "geoloc.administrative_area_level_2": "",
            "geoloc.administrative_area_level_3": "",
            "geoloc.locality": prop.internal["city"],
            "geoloc.sublocality": "",
            "geoloc.neighborhood": "",
            "geoloc.premise": "",
            "geoloc.subpremise": "",
            "geoloc.point_of_interest": "",
            # "geoloc.postal_code": "45110",
            # "geoloc.southwest": "(20.68462221970849,+-103.4283811302915)",
            # "geoloc.northEast": "(20.6873201802915,+-103.4256831697085)",
            "geoloc.lat": prop.ubication.location.lat,
            "geoloc.lng": prop.ubication.location.lng,
            "direccionOriginal": prop.ubication.address + ", " + internal_name
        }

        headers = {"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8"}
        res = self.api_req.make(save_url, "POST", headers=headers, data=payload)
        if res is None or not res.ok:
            return Exception("error saving geolocation"), None

        return None, res.json()

    def get_geolocation(self, prop: Property) -> tuple[Exception, None] | tuple[None, str]:
        self.logger.debug("getting geolocation", prop.internal["colony"])
        url_obtener_loc = "https://www.inmuebles24.com/publicar_obtenerLocalidadAGeoloc.ajax"
        url_geolicalizar = "https://www.inmuebles24.com/publicar_geolocalizar.ajax"

        payload = {
            "idCiudad": str(prop.internal["city_id"]),
            "idZona": str(prop.internal["colony_id"]),
            "idSubZona": "",
            "idCodigoPostal": ""
        }
        res = self.api_req.make(url_obtener_loc, "POST", data=payload)
        if res is None or not res.ok:
            return Exception("error getting geolocation"), None

        data = res.json()
        # ej: Fraccionamiento Las Lomas Club Golf,Zapopan,Jalisco,Mexico
        internal_name = data.get("contenido")
        if internal_name is None:
            return Exception("cannot get geolocation content"), None
        payload = {
            "direccionOriginal": prop.ubication.address + ", " + internal_name
        }

        res = self.api_req.make(url_geolicalizar, "POST", data=payload)
        if res is None or not res.ok:
            return Exception("error getting geolocation"), None

        data = res.json()
        geo_id: str | None = data.get("contenido", {}).get("geolocdefault", {}).get("idgeolocalizacion")
        if geo_id is None:
            self.logger.warning("cannot find geolocation, saving a new one")

            err, data = self.save_geolocation(prop, internal_name)
            if err is not None or data is None:
                return Exception("cannot save geolocation"), None

            geo_id: str | None = data.get("contenido", {}).get("geolocdefault", {}).get("idgeolocalizacion")
            if geo_id is None:
                return Exception("cannot get geo_id from saved geolocation"), None

        return None, geo_id

    # upload each image multipart/form-data to upload_url
    # add image to the property with add_image_url
    # add the video and virtual route too
    def upload_images(self, prop_id: str, prop: Property, cookies):
        upload_url = f"https://www.inmuebles24.com/avisoImageUploader.bum?idAviso={prop_id}"
        add_image_url = f"https://www.inmuebles24.com/publicarPasoMultimedia.bum?idaviso={prop_id}"

        def process_image(image):
            self.logger.debug(f"downloading {image['url']}")
            img_data = download_file(image["url"])
            if img_data is None:
                return Exception(f"cannot download the image {image['url']}")
            self.logger.success("image downloaded successfully")

            img_type = "png" if "png" in image["url"] else "jpeg"
            file_name = f"image.{img_type}"
            files = [
                ('name', (file_name)),
                ('file', (file_name, img_data, f"image/{img_type}")),
            ]

            self.logger.debug("uploading the image")
            params = PARAMS.copy()
            params["url"] = upload_url
            self.request.cookies = cookies
            res = self.request.make(ZENROWS_API_URL, "POST",
                params=params,
                files=files
            )
            if res is None:
                self.logger.error("unknonw error uploading the image", image["url"])
                return None
            if not res.ok:
                self.logger.error("error uploading the image: ", image["url"], "error: ", res.text)
                return None

            file_url = res.text.replace("FILEID:", "")
            self.logger.success(f"image uploaded successfully, url: {file_url}")
            return file_url


        # Necessary for the upload image request, otherwise will fail
        if "content-type" in self.request.headers:
            ct = self.request.headers["content-type"]
            del self.request.headers["content-type"]

        uploaded_images = []
        with concurrent.futures.ThreadPoolExecutor() as executor:
            results = executor.map(process_image, prop.images)
            for result in results:
                if isinstance(result, Exception):
                    return result
                if result is not None:
                    uploaded_images.append(result)

        if len(uploaded_images) < len(prop.images):
            self.logger.error("not all the images are successfully added, publication failed")
            return Exception("not all the images are successfully added")

        if "content-type" in self.request.headers:
            self.request.headers["content-type"] = ct

        # Plano. TODO: que el plano no sea la 4ta foto.
        plano_img_parts = uploaded_images[4].split("/")
        plano_img_id = plano_img_parts[len(plano_img_parts)-1]

        payload: dict[str, Any] = {
            "multimediaaviso.idMultimediaAviso[]": [],
            "idAviso": prop_id,
            "irAlPaso": "",
            "checkContentEnhancerHidden": "",
            "guardarComoBorrador": "1",
            "mutimediaaviso.keyRecorrido360": prop.virtual_route if prop.virtual_route is not None else "",
            "mutimediaaviso.idRecorrido360": "-1",
            "mutimediaaviso.idRecorrido360Multimedia--1": prop.virtual_route if prop.virtual_route is not None else "",
            "mutimediaaviso.idRecorrido360Orden[-1]": "2",

            # Plano. TODO: que el plano no sea la 4ta foto.
            "multimediaaviso.idPlanosMultimedia[]":	"nueva_0",
            "multimediaaviso.orden[nueva_0_plano]":	"1",
            "multimediaaviso.rotacion[nueva_0_plano]": "0",
            "multimediaaviso.urlTempPlano[nueva_0]": plano_img_id
        }

        # add the video url
        if prop.video_url is not None:
            # https://youtu.be/H9RlIXuS1is. key => H9RlIXuS1is
            parts = prop.video_url.split("/")
            video_key = parts[len(parts) - 1]

            payload = {
                **payload,
                "mutimediaaviso.keyvideo": video_key,
                f"mutimediaaviso.idVideoMultimediaOrden[{video_key}]": "1",
            }

        i = 0
        for img_url in uploaded_images:
            # Obtain the file name. ej:
            # file_url: https://storage.googleapis.com/rp-tmp-images/avisos/18/01/46/26/66/66/180x140/temp_aviso_146266666_656c6604-52ef-4b20-8ba5-21b0b9d3cbd6.jpg
            # the file id its: temp_aviso_146266666_656c6604-52ef-4b20-8ba5-21b0b9d3cbd6.jpg
            # split by / and get the last part
            parts = img_url.split("/")
            img_id = parts[len(parts)-1]

            id = f"nueva_{i}"
            payload[f"multimediaaviso.orden[{id}]"] = str(i+1)
            payload[f"multimediaaviso.rotacion[{id}]"] = "0"
            payload[f"multimediaaviso.urlTempImage[{id}]"] = img_id
            payload[f"multimediaaviso.titulo[{id}]"] = ""
            payload["multimediaaviso.idMultimediaAviso[]"].append(id)

            i += 1 

        params = PARAMS.copy()
        params["js_render"] = "true"
        params["antibot"] = "true"
        params["url"] = add_image_url

        max_tries = 3
        t = 1
        error = True

        while error and t <= max_tries:
            res = requests.post(
                ZENROWS_API_URL,
                # "POST",
                params=params, 
                data=payload, 
                cookies=cookies, 
                headers={
                    "sessionId": self.request.headers["sessionId"],
                    "idUsuario": self.request.headers["idUsuario"],
                }
            )
            if res is None or not res.ok:
                self.logger.error("error adding the image to the publication")
                t += 1
                continue

            if "CAPTCHA" in res.text:
                self.logger.error("CAPTCHA found on the image publication response")
                t += 1
                continue
            
            t += 1
            error = False

        if error:
            return Exception("cannot add the images to the publication")

        self.logger.success("images added successfully to the publication")

    def make_contacted(self, lead: dict):
        self._change_status(lead, Status.contacted)

    def make_failed(self, lead: dict):
        self._change_status(lead, Status.phone_error)
