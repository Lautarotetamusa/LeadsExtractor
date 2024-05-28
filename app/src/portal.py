import json
import os
from typing import Iterator

from src.lead import Lead
from src.make_requests import Request
from src.logger import Logger
from src.message import format_msg
import src.api as api
import src.jotform as jotform

from enum import IntEnum

with open('../messages/bienvenida_1.txt') as f:
    bienvenida_1 = f.read()
with open('../messages/bienvenida_2.txt') as f:
    bienvenida_2 = f.read()
with open('../messages/response_msg.txt') as f:
    response_msg = f.read()

class Mode(IntEnum):
    NEW = 1
    ALL = 2

class Portal():
    def __init__(self,
                 name: str,
                 contact_id_field: str,
                 send_msg_field: str,
                 username_env: str,
                 password_env: str,
                 params_type: str, #headers | cookies
                 filename: str = __file__
        ):

        self.name = name
        self.logger = Logger(name)
        self.request = Request(None, None, self.logger, self.login)

        self.contact_id_field = contact_id_field
        self.send_msg_field = send_msg_field
        self.logger.debug(f"Inicializando {self.name}")

        self.params_file   = os.path.dirname(os.path.realpath(filename)) + "/params.json"
        self.username = os.getenv(username_env)
        self.password = os.getenv(password_env)

        if (not os.path.exists(self.params_file)):
            self.logger.debug("Creando el archivo params.json")
            with open(self.params_file, "a") as f:
                json.dump({}, f)

        with open(self.params_file, "r") as f:
            if params_type == "cookies":
                self.request.cookies = json.load(f)
            elif params_type == "headers":
                self.request.headers = json.load(f)
            else:
                self.logger.error("Incorrect params_type, must be 'cookies' or 'headers'")
                exit(1)
    
    def send_message_condition(self, lead: dict) -> bool:
        return True
    def get_leads(self, mode: Mode) -> Iterator[list[dict]]:
        yield []
    def get_lead_info(self, raw_lead: dict) -> Lead:
        return Lead()
    def send_message(self, id: str,  message: str):
        pass
    def make_contacted(self, id: str):
        pass
    def login(self):
        pass

    def main(self):
        for page in self.get_leads(Mode.NEW):
            for lead_res in page:
                lead = self.get_lead_info(lead_res)

                if lead.telefono == None or lead.telefono == "":
                    self.logger.debug("El lead no tiene telefono, no hacemos nada")
                    self.make_contacted(lead_res[self.contact_id_field])
                    continue

                is_new, lead = api.assign_asesor(self.logger, lead)
                if lead == None:
                    continue

                #Cotizacion
                if is_new:
                    self.logger.debug("Lead con mt2 construccion, generando cotizacion pdf")
                    pdf_url = jotform.new_submission(self.logger, lead) 
                    if pdf_url != None:
                        lead.cotizacion = pdf_url
                    else:
                        self.logger.error("No se pudo obtener la cotizacion en pdf")

                api.new_communication(self.logger, lead)
                lead.print()

                if is_new: #Lead nuevo
                    portal_msg = bienvenida_1 + ' ' + format_msg(lead, bienvenida_2)
                else: #Lead existente
                    portal_msg = format_msg(lead, response_msg)

                #Mensaje del portal
                if self.send_msg_field in lead_res:
                    self.send_message(lead_res[self.send_msg_field], portal_msg)
                lead.message = portal_msg.replace('\n', '')

                self.make_contacted(lead_res[self.contact_id_field])
