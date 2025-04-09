import json
import os
import time
from typing import Iterator, Tuple
from multiprocessing.pool import ThreadPool

from src.property import Property
from src.lead import Lead
from src.make_requests import Request
from src.logger import Logger
from src.message import format_msg
import src.api as api

from enum import IntEnum

dirpath = os.path.dirname(__file__)
msgpath = os.path.join(dirpath, '../../messages')

with open(os.path.join(msgpath, "bienvenida_1.txt")) as f:
    bienvenida_1 = f.read()
with open(os.path.join(msgpath, 'bienvenida_2.txt')) as f:
    bienvenida_2 = f.read()
with open(os.path.join(msgpath, 'response_message.txt')) as f:
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
                 params_type: str,  # headers | cookies
                 unauthorized_codes: list[int],
                 filename: str = __file__
        ):

        self.name = name
        self.logger = Logger(name)
        self.request = Request(None, None, self.logger, self.login, unauthorized_codes)

        self.contact_id_field = contact_id_field
        self.send_msg_field = send_msg_field
        self.logger.debug(f"Inicializando {self.name}")

        self.params_file = os.path.dirname(os.path.realpath(filename)) + "/params.json"
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
    def make_contacted(self, lead: dict):
        pass
    def make_failed(self, lead: dict):
        pass
    def publish(self, property: Property) -> tuple[Exception, None] | tuple[None, str]:
        """publish a property in a portal.
        @returns error or property internal portal id
        """
        return None, "not-implemented"
    def unpublish(self, publication_id: str) -> Exception | None:
        """removes a publication from a portal"""
        pass
    def login(self):
        pass

    def default_action(self, lead_res: dict, msg):
        # Marcar como contactado
        if self.contact_id_field in lead_res:
            self.make_contacted(lead_res)
        else:
            self.logger.warning("No se encontro campo para contactar")

        # Mensaje del portal
        if self.send_msg_field in lead_res:
            self.send_message(lead_res[self.send_msg_field], msg)
        else:
            self.logger.warning("No se encontro campo para enviar mensaje")

    def main(self):
        for page in self.get_leads(Mode.NEW):
            for lead_res in page:
                lead = self.get_lead_info(lead_res)

                if lead.telefono is None or lead.telefono == "":
                    self.logger.debug("El lead no tiene telefono, no hacemos nada")
                    self.make_contacted(lead_res)
                    self.make_failed(lead_res)
                    continue

                lead.print()
                is_new, lead = api.new_communication(self.logger, lead)
                if lead is None:
                    # TODO: No hardcodear esto
                    not_valid_msg = """Verificamos tu número de teléfono, y no es válido, para llamar o para enviarte un WhatsApp, por favor envíanos un mensaje a nuestro WhatsApp oficial, o si prefieres llamarnos, te comparto el número de nuestro departamento comercial que está listo, para ayudarte en todo el proceso, bienvenido a tu nueva casa con Rebora"""
                    txt = bienvenida_1 + not_valid_msg 
                    self.default_action(lead_res, txt)
                    self.make_failed(lead_res)
                    continue

                if is_new:
                    portal_msg = bienvenida_1 + ' ' + format_msg(lead, bienvenida_2)
                else:
                    portal_msg = format_msg(lead, response_msg)

                self.default_action(lead_res, portal_msg)


    def first_run(self):
        pool = ThreadPool(processes=20)

        count = 0
        for page in self.get_leads(Mode.ALL):
            results = []
            for lead_res in page:
                r = pool.apply_async(self.get_lead_info, args=(lead_res, ))
                time.sleep(0.4)
                results.append(r)

            for r in results:
                lead = r.get()
                if lead.telefono is None or lead.telefono == "":
                    self.logger.debug("El lead no tiene telefono, no hacemos nada")
                    continue

                is_new, lead = api.new_communication(self.logger, lead)
                self.logger.debug(count)
                count += 1

    # Obtener solamente 10 leads para ver si está funcionando correctamente
    def test(self):
        # self.login()

        pool = ThreadPool(processes=20)

        count = 0
        max = 10
        for page in self.get_leads(Mode.ALL):
            results = []
            for lead_res in page:
                r = pool.apply_async(self.get_lead_info, args=(lead_res, ))
                time.sleep(0.4)
                results.append(r)
                count += 1
                if count >= max:
                    break

            for r in results:
                lead = r.get()
                if lead.telefono is None or lead.telefono == "":
                    self.logger.debug("El lead no tiene telefono, no hacemos nada")
                    continue

                lead.print()
                api.new_communication(self.logger, lead)
                self.logger.debug(count)

            return

    # Testear si la paginacion esta funcionando bien
    def test_page(self):
        count = 0
        max_pages = 4
        for page in self.get_leads(Mode.NEW):
            if count >= max_pages:
                pass
                # break

            count += 1
