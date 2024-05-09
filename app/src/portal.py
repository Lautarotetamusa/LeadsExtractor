from email.mime.application import MIMEApplication
import json
import time
import os
from typing import Iterator

from src.lead import Lead
from src.make_requests import Request
from src.logger import Logger
from src.message import format_msg
from src.sheets import Gmail, Sheet 
from src.whatsapp import Whatsapp
import src.api as api
import src.jotform as jotform

from enum import IntEnum

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
        self.wpp    = Whatsapp(self.logger)
        self.gmail  = Gmail({
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
            'cotizacion_1': 'plantilla_cotizacion_1.txt',
            'cotizacion_2': 'plantilla_cotizacion_2.txt',
        }
        self.gmail_spin = ""
        self.gmail_subject = ""
        self.response_msg = ""
        self.bienvenida_1 = ""
        self.bienvenida_2 = ""
        self.cotizacion_1 = ""
        self.cotizacion_2 = ""

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

                is_new, lead = api.new_communication(self.logger, lead)
                if lead == None:
                    continue
                lead.print()

                if is_new: #Lead nuevo
                    bienvenida_2 = format_msg(lead, self.bienvenida_2)

                    if lead.busquedas['covered_area'] == "" or lead.busquedas['covered_area'] == None:
                        cotizacion_msj = self.cotizacion_2
                    else:
                        cotizacion_msj = self.cotizacion_1

                    #Cotizacion
                    self.logger.debug("Lead con mt2 construccion, generando cotizacion pdf")
                    pdf_url = jotform.new_submission(self.logger, lead) 
                    if pdf_url != None:
                        lead.cotizacion = pdf_url
                        self.wpp.send_document(lead.telefono, pdf_url, 
                            filename=f"Cotizacion para {lead.nombre}",
                            caption=cotizacion_msj
                        )
                    else:
                        self.logger.error("No se pudo obtener la cotizacion en pdf")

                    self.wpp.send_image(lead.telefono)
                    self.wpp.send_message(lead.telefono, self.bienvenida_1)
                    self.wpp.send_message(lead.telefono, bienvenida_2)
                    self.wpp.send_video(lead.telefono)

                    portal_msg = self.bienvenida_1 + '\n' + bienvenida_2
                else: #Lead existente
                    portal_msg = format_msg(lead, self.response_msg)

                    self.wpp.send_response(lead.telefono, lead.asesor)

                self.wpp.send_msg_asesor(lead.asesor['phone'], lead, is_new)

                #Mensaje del portal
                if self.send_msg_field in lead_res:
                    self.send_message(lead_res[self.send_msg_field], portal_msg)
                lead.message = portal_msg.replace('\n', '')

                if lead.email != None and lead.email != '':
                    if lead.propiedad["ubicacion"] == "":
                        lead.propiedad["ubicacion"] = "que consultaste"
                    else:
                        lead.propiedad["ubicacion"] = "ubicada en " + lead.propiedad["ubicacion"]

                    #gmail_msg = format_msg(lead, self.gmail_spin)
                    #subject = format_msg(lead, self.gmail_subject)
                    #self.gmail.send_message(gmail_msg, subject, lead.email, self.attachment)

                self.make_contacted(lead_res[self.contact_id_field])

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
                lead.print()
                if lead.telefono == None or lead.telefono == "":
                    self.logger.debug("El lead no tiene telefono, no hacemos nada")
                    continue

                row_leads.append(self.sheet.map_lead(lead.__dict__, self.headers))
                leads.append(lead)  
                self.logger.debug("cargando a la api")
                is_new, lead = api.new_communication(self.logger, lead)
                self.logger.debug("IS NEW: "+str(is_new))

            return
