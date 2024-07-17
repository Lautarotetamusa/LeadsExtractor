from time import gmtime, strftime
from datetime import datetime
import os
import json
import requests

from src.portal import Mode, Portal
from src.lead import Lead
from src.numbers import parse_number 

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT != None, "DATE_FORMAT is not seted"
SITE_URL = "https://www.inmuebles24.com/"
ZENROWS_API_URL = "https://api.zenrows.com/v1/"
PARAMS = {
        #"resolve_captcha": "true",
        "apikey": os.getenv("ZENROWS_APIKEY"),
        "url": "",
        #"js_render": "true",
        "antibot": "true",
        "premium_proxy": "true",
        "proxy_country": "mx",
        #"session_id": 10,
        "custom_headers": "true",
        "original_status": "true",
        "autoparse": "true"
        }

def extract_busqueda_info(data: dict | None) -> dict:
    if data == None:
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

class Inmuebles24(Portal):
    def __init__(self):
        super().__init__(
                name="inmuebles24",
                contact_id_field="contact_publisher_user_id",
                send_msg_field="id",
                username_env="INMUEBLES24_USERNAME",
                password_env="INMUEBLES24_PASSWORD",
                params_type="headers",
                filename=__file__
                )

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
        self.request.headers["Content-Type"] = "application/x-www-form-urlencoded;charset=UTF-8"

        self.logger.debug(f"POST {login_url}")
        params = PARAMS.copy()
        params["url"] = login_url
        res = requests.post(ZENROWS_API_URL, params=params, data=data)
        if not res.ok:
            self.logger.error("status="+str(res.status_code)+" res="+str(res.text))
            return
        data = res.json()

        self.request.headers = {
                "sessionId": data["contenido"]["sessionID"],
                "x-panel-portal": "24MX",
                "content-type": "application/json;charset=UTF-8",
                "idUsuario": str(data["contenido"]["idUsuario"])
                }

        print("headers", self.request.headers)
        with open(self.params_file, "w") as f:
            json.dump(self.request.headers, f, indent=4)
        self.logger.success("Sesion iniciada con exito")

    def get_leads(self, mode=Mode.NEW):
        status = "nondiscarded" #Este campo lo pasamos siempre
        sort = "unread" #Los no leidos estarán primeros
        first = True
        offset = 0
        total = 0
        limit = 20
        finish = False

        while first or (not finish and offset < total):
            leads_url = f"{SITE_URL}leads-api/publisher/leads?offset={offset}&limit={limit}&spam=false&status={status}&sort={sort}"
            self.logger.debug(f"GET {leads_url}")
            PARAMS["url"] = leads_url
            res = self.request.make(ZENROWS_API_URL, 'GET', params=PARAMS)
            if res == None:
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

            if mode == Mode.NEW: #Obtenemos todos los leads sin leer de este pagina
                leads = []
                for lead in data["result"]:
                    if lead["statuses"][0] == "READ": #Como los leads estan ordenandos, al encontrar uno con estado READ. paramos
                        self.logger.debug("Se encontro un lead con status READ, deteniendo")
                        finish = True
                        break
                    leads.append(lead) 
                yield leads 
            else: #Obtenemos todos los leads
                yield data["result"]

    def get_lead_info(self, raw_lead):
        raw_lead_id = raw_lead["id"]
        contact_id = raw_lead[self.contact_id_field]
        raw_busqueda = self.get_busqueda_info(contact_id)
        busqueda = extract_busqueda_info(raw_busqueda)
        posting = raw_lead.get("posting", {})

        lead = Lead()
        lead.set_args({
            "lead_id": raw_lead_id,
            "contact_id": contact_id, 
            "fuente": self.name,
            "fecha_lead": datetime.strptime(raw_lead["last_lead_date"], "%Y-%m-%dT%H:%M:%S.%f+00:00").strftime(DATE_FORMAT),
            "nombre": raw_lead.get("lead_user", {}).get("name"),
            "link": f"{SITE_URL}panel/interesados/{contact_id}",
            "email": raw_lead.get("lead_user", {}).get("email"),
            })
        lead.set_busquedas(busqueda)
        lead.set_propiedad({
            "id": posting.get("id", None),
            "titulo": posting.get("title", ""),
            "precio": str(posting.get("price", {}).get("amount", "")),
            "ubicacion": posting.get("address", ""),
            "tipo": posting.get("real_estate_type", {}).get("name"),
            "municipio": posting.get("location", {}).get("parent", {}).get("name", "") #Ciudad
            })
        #Algunos leads pueden ver nuestro telefono sin necesariamente venir por una propiedad
        if lead.propiedad["id"] == None:
            lead.fuente = "viewphone"

        telefono = parse_number(self.logger, raw_lead.get("phone", ""), "MX")
        if not telefono:
            telefono = parse_number(self.logger, raw_lead.get("phone", ""))
        lead.telefono = telefono or lead.telefono

        return lead

    # Get the information about the searchers of the lead
    def get_busqueda_info(self, lead_id) -> dict | None:
        self.logger.debug("Extrayendo la informacion de busqueda del lead: "+lead_id)
        busqueda_url = f"{SITE_URL}leads-api/publisher/contact/{lead_id}/user-profile"

        self.logger.debug(f"GET {busqueda_url}")
        PARAMS["url"] = busqueda_url
        res = self.request.make(ZENROWS_API_URL, 'GET', params=PARAMS)

        if res == None:
            self.logger.error("No se pudo obtener la informacion de busqueda para el lead: "+lead_id)
            return None
        try:
            self.logger.success("Informacion de busqueda extraida con exito")
            return res.json()
        except requests.exceptions.JSONDecodeError:
            self.logger.error(f"El lead {lead_id} no tiene informacion de busqueda")
            return None

    def send_message_condition(self, lead) -> bool:
        return "last_message" not in lead or (lead.get("last_message", {}).get("to") == self.request.headers["idUsuario"])

    def send_message(self, id: str,  message: str):
        self.logger.debug(f"Enviando mensaje a lead {id}")
        msg_url = f"{SITE_URL}leads-api/leads/{id}/messages"

        print(message)
        data = {
                "is_comment": False,
                "message": message,
                "message_attachments": []
                }

        PARAMS["url"] = msg_url
        res = self.request.make(ZENROWS_API_URL, 'POST', params=PARAMS, json=data)

        if res != None and res.status_code >= 200 and res.status_code < 300:
            self.logger.success(f"Mensaje enviado correctamente a lead {id}")
        else:
            self.logger.error(f"Error enviando mensaje al lead {id}")

    def make_contacted(self, id: str):
        status = "READ"
        status_url = f"{SITE_URL}leads-api/leads/status/{status}?=&contact_publisher_user_id={id}"

        PARAMS["url"] = status_url
        PARAMS["autoparse"] = False
        res = self.request.make(ZENROWS_API_URL, 'POST', params=PARAMS)

        if res != None and res.status_code >= 200 and res.status_code < 300:
            self.logger.success(f"Se marco a lead {id} como {status}")
        else:
            if res != None:
                self.logger.error(res.content)
                self.logger.error(res.status_code)
            self.logger.error(f"Error marcando al lead {id} como {status}")
        PARAMS["autoparse"] = True

if __name__ == "__main__":
    inmuebles24 = Inmuebles24()
    inmuebles24.main()
