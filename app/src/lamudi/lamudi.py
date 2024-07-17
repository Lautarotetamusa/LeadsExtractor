from time import gmtime, strftime
from datetime import datetime
import json
import os
from typing import Iterator
import uuid

import requests

from src.numbers import parse_number
from src.portal import Mode, Portal
from src.lead import Lead

API_URL = "https://api.proppit.com"
DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT != None, "DATE_FORMAT is not seted"


class Lamudi(Portal):
    def __init__(self):
        super().__init__(
            name="lamudi",
            contact_id_field="id",
            send_msg_field="id",
            username_env="LAMUDI_USERNAME",
            password_env="LAMUDI_PASSWORD",
            params_type="cookies",
            filename=__file__
        )
    def login(self):
        self.logger.debug("Iniciando sesion")
        login_url = f"{API_URL}/login"

        data = {
            "email": self.username,
            "password": self.password
        }
        res = requests.post(login_url, json=data)
        if res == None:
            return
        
        self.request.cookies = {
            "authToken": res.cookies["authToken"]
        }
        with open(self.params_file, "w") as f:
            json.dump(self.request.cookies, f, indent=4)
        self.logger.success("Sesion iniciada con exito")

    def get_leads(self, mode=Mode.NEW) -> Iterator[list[dict]]:
        max = -1
        status = "new lead"
        page = 1
        limit = 25
        cant = 0

        if mode == Mode.NEW:
            url = f"{API_URL}/leads?_limit={limit}&_order=desc&_page={page}&_sort=lastActivity&status={status}"
        else:
            url = f"{API_URL}/leads?_limit={limit}&_order=desc&_page={page}&_sort=lastActivity"
        self.logger.debug(f"Extrayendo leads")

        res = self.request.make(url, 'GET')
        if res == None:
           exit(1) 
        data = res.json()["data"]

        total = data["totalFilteredRows"]
        self.logger.debug(f"total: {total}")

        while (cant < total and (cant < max or max == -1)):
            self.logger.debug(f"len: {cant}")
            yield data["rows"]
            cant += len(data["rows"])
            page += 1

            if mode == Mode.NEW:
                url = f"{API_URL}/leads?_limit={limit}&_order=desc&_page={page}&_sort=lastActivity&status={status}"
            else:
                url = f"{API_URL}/leads?_limit={limit}&_order=desc&_page={page}&_sort=lastActivity"
            self.logger.debug(url)

            res = self.request.make(url, 'GET')
            if res == None:
                break
            data = res.json()["data"]

        self.logger.success(f"Se encontraron {cant} nuevos Leads")

    def get_lead_property(self, lead_id):
        property_url = f"{API_URL}/leads/{lead_id}/properties"
        res = self.request.make(property_url)
        if res == None:
            return [{
                "titulo": "",
                "id": "",
                "link": "",
                "precio": "",
                "ubicacion": "",
                "tipo": "",
                "municipio": ""
            }]
        props = res.json().get("data")

        formatted_props = []
        for p in props:
            formatted_props.append({
                "titulo": p["title"],
                "id": p['id'],
                "link": f"https://www.lamudi.com.mx/detalle/{p['id']}",
                "precio": str(p["price"]["amount"]),
                "ubicacion": p["address"], #Direccion completa
                "tipo": p["propertyType"],
                "municipio": p["geoLevels"][0]["name"] if len(p["geoLevels"]) > 0 else "" #Solamente el municipio, lo usamos para generar el mensaje
            })

        return formatted_props

    def get_lead_info(self, raw_lead: dict) -> Lead:
        lead = Lead()
        lead.set_args({
            "lead_id": raw_lead["id"],
            "fuente": self.name,
            "fecha_lead": datetime.strptime(raw_lead["lastActivity"], "%Y-%m-%dT%H:%M:%SZ").strftime(DATE_FORMAT),
            "nombre": raw_lead["name"],
            "link": f"https://proppit.com/leads/{raw_lead['id']}",
            "email": raw_lead['email'],
            "propiedad": self.get_lead_property(raw_lead["id"])[0],
        })

        telefono = parse_number(self.logger, raw_lead.get("phone", ""), None)
        if not telefono:
            telefono = parse_number(self.logger, raw_lead.get("phone", ""), "MX")
        lead.telefono = telefono or lead.telefono
        return lead

    def send_message(self, id, message):
        self.logger.debug(f"Enviando mensaje a lead {id}")
        #Tenemos que pasarle un id del mensaje, lo generamos nosotros con uuid()
        msg_id = str(uuid.uuid4())
        msg_url = f"{API_URL}/leads/{id}/notes/{msg_id}"

        data = {
            "id": msg_id,
            "message": message
        }
        self.request.make(msg_url, 'POST', json=data)
        self.logger.success(f"Mensaje enviado correctamente a lead {id}")

    def make_contacted(self, id):
        self.logger.debug(f"Marcando como contactacto a lead {id}")
        read_url = f"{API_URL}/leads/{id}/status"

        data = {
            "status": "contacted"
        }
        self.request.make(read_url, 'PUT', json=data)
        self.logger.success(f"Se contacto correctamente a lead {id}")
