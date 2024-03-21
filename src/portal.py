from email.mime.application import MIMEApplication
import json
import time
import os
from threading import Thread
from typing import Generator, Iterator

from src.asesor import assign_asesor
from src.lead import Lead
from src.make_requests import Request
from src.logger import Logger
from src.message import format_msg
from src.sheets import Gmail, Sheet
from src.whatsapp import Whatsapp
import src.infobip as infobip

from enum import IntEnum

class Mode(IntEnum):
    NEW = 1
    ALL = 2

wpp = Whatsapp()

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
        self.gmail = Gmail({
            "email": os.getenv("EMAIL_CONTACT"),
        }, self.logger)
        self.request = Request(None, None, self.logger, self.login)

        self.contact_id_field = contact_id_field
        self.send_msg_field = send_msg_field
        self.setup()

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

    def setup(self):
        self.logger.debug(f"Inicializando {self.name}")
        self.sheet = Sheet(self.logger, "mapping.json")
        self.headers = self.sheet.get("A2:Z2")[0]

        resources = {
            'gmail_spin': 'gmail.html',
            'gmail_subject': 'gmail_subject.html',
            'response_msg': 'response_message.txt',
            'bienvenida_1': 'bienvenida_1.txt',
            'bienvenida_2': 'bienvenida_2.txt',
        }
        self.gmail_spin = ""
        self.gmail_subject = ""
        self.response_msg = ""
        self.bienvenida_1 = ""
        self.bienvenida_2 = ""

        for resource in resources:
            with open(f"messages/{resources[resource]}", 'r') as f:
                self.__dict__[resource] = f.read()
            assert resource in self.__dict__ and self.__dict__[resource] != "", f"{resource} no se cargo correctamente"

        # Adjuntar archivo PDF
        with open('messages/attachment.pdf', 'rb') as pdf_file:
            self.attachment = MIMEApplication(pdf_file.read(), _subtype="pdf")
            self.attachment.add_header('Content-Disposition', 'attachment', 
                  filename='Bienvenido a Rebora! Seguridad, Confort y Placer - Casas de gran diseño y alta calidad.pdf'
            )
    
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

                is_new, lead = assign_asesor(lead)
                lead.validate()
                if is_new: #Lead nuevo
                    bienvenida_2 = self.bienvenida_2.format(
                        asesor_name=lead.asesor['name'], 
                        asesor_phone=lead.asesor['phone']
                    )

                    wpp.send_image(lead.telefono)
                    wpp.send_message(lead.telefono, self.bienvenida_1)
                    wpp.send_message(lead.telefono, bienvenida_2)
                    wpp.send_video(lead.telefono)

                    portal_msg = self.bienvenida_1 + '\n' + bienvenida_2

                    infobip.create_person(self.logger, lead)
                else: #Lead existente
                    portal_msg = format_msg(lead, self.response_msg)

                    wpp.send_response(lead.telefono, lead.asesor)
                
                wpp.send_msg_asesor(lead.asesor['phone'], lead)

                #Mensaje del portal
                if self.send_msg_field in lead_res:
                    self.send_message(lead_res[self.send_msg_field], portal_msg)
                lead.message = portal_msg.replace('\n', '')

                if lead.email != None and lead.email != '':
                    if lead.propiedad["ubicacion"] == "":
                        lead.propiedad["ubicacion"] = "que consultaste"
                    else:
                        lead.propiedad["ubicacion"] = "ubicada en " + lead.propiedad["ubicacion"]

                    gmail_msg = format_msg(lead, self.gmail_spin)
                    subject = format_msg(lead, self.gmail_subject)
                    self.gmail.send_message(gmail_msg, subject, lead.email, self.attachment)

                self.make_contacted(lead_res[self.contact_id_field])

                row_lead = self.sheet.map_lead(lead.__dict__, self.headers)
                self.sheet.write([row_lead])

    def first_run(self):
        from multiprocessing.pool import ThreadPool
        pool = ThreadPool(processes=20)

        for page in self.get_leads(Mode.ALL):
            row_leads = []
            leads: list[Lead] = []
            results = []
            for lead_res in page:
                r = pool.apply_async(self.get_lead_info, args=(lead_res, ))
                time.sleep(0.5)
                results.append(r)
        
            for r in results:
                lead = r.get()
                if lead.telefono == None or lead.telefono == "":
                    self.logger.debug("El lead no tiene telefono, no hacemos nada")
                    continue

                row_leads.append(self.sheet.map_lead(lead.__dict__, self.headers))
                leads.append(lead)  

            results = infobip.add_fecha_lead(self.logger, leads)
            self.logger.debug(results)
            if type(results) is bool:
                self.logger.debug("Ocurrio un error vamos a crear las personas")
                #TODO: Crear todas las personas
            else:
                new_phones = []
                for result in results:
                    if 'errors' in result:
                        phone = result.get('query', {}).get('identifier')
                        new_phones.append(phone)
                        self.logger.debug(f"El lead con telefono no existe: {phone}, cargando")
                        #infobip.create_persons(self.logger, leads)
                    else:
                        self.logger.debug("Todas las personas se actualizaron correctamente")
                for phone in new_phones:
                    for lead in leads:
                        if lead.telefono[1:] == phone:
                            infobip.create_person(self.logger, lead)
                            break

            self.sheet.write(row_leads, range_sheet="2da-Corrida")
