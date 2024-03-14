from email.mime.application import MIMEApplication
import json
import os

from src.asesor import assign_asesor
from src.lead import Lead
from src.make_requests import Request
from src.logger import Logger
from src.message import format_msg
from src.sheets import Gmail, Sheet
from src.whatsapp import Whatsapp
import src.infobip as infobip

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

        with open('messages/gmail.html', 'r') as f:
            self.gmail_spin = f.read()
        with open('messages/gmail_subject.html', 'r') as f:
            self.gmail_subject = f.read()
        with open('messages/mensaje_portals.txt', 'r') as f:
            self.msg_spin = f.read()
        with open('messages/bienvenida_1.txt') as f:
            self.bienvenida_1 = f.read()
        with open('messages/bienvenida_2.txt') as f:
            self.bienvenida_2 = f.read()

        # Adjuntar archivo PDF
        with open('messages/attachment.pdf', 'rb') as pdf_file:
            self.attachment = MIMEApplication(pdf_file.read(), _subtype="pdf")
            self.attachment.add_header('Content-Disposition', 'attachment', 
                  filename='Bienvenido a Rebora! Seguridad, Confort y Placer - Casas de gran diseño y alta calidad.pdf'
            )
    
    def send_message_condition(self, lead: dict) -> bool:
        return True
    def get_leads(self) -> list[dict]:
        return []
    def get_lead_info(self, raw_lead: dict) -> Lead:
        return Lead()
    def send_message(self, id: str,  message: str):
        pass
    def make_contacted(self, id: str):
        pass
    def login(self):
        pass

    def main(self):
        leads = self.get_leads()

        leads_info = []
        for lead_res in leads:
            lead = self.get_lead_info(lead_res)

            is_new, lead = assign_asesor(lead)
            if is_new: #Lead nuevo
                wpp.send_image(lead.telefono)
                wpp.send_message(lead.telefono, self.bienvenida_1)
                wpp.send_message(lead.telefono, self.bienvenida_2.format(
                    asesor_name=lead.asesor['name'], 
                    asesor_phone=lead.asesor['phone'])
                )
                wpp.send_video(lead.telefono)

                infobip.create_person(self.logger, lead)
            else: #Lead existente
                wpp.send_response(lead.telefono, lead.asesor)
            
            wpp.send_msg_asesor(lead.asesor['phone'], lead)

            if self.send_message_condition(lead_res):
                if lead.email != None and lead.email != '':
                    if lead.propiedad["ubicacion"] == "":
                        lead.propiedad["ubicacion"] = "que consultaste"
                    else:
                        lead.propiedad["ubicacion"] = "ubicada en " + lead.propiedad["ubicacion"]

                    gmail_msg = format_msg(lead, self.gmail_spin)
                    subject = format_msg(lead, self.gmail_subject)
                    self.gmail.send_message(gmail_msg, subject, lead.email, self.attachment)

                msg = format_msg(lead, self.msg_spin)
                lead.message = msg.replace('\n', '')

                if self.send_msg_field in lead_res:
                    self.send_message(lead_res[self.send_msg_field], msg)
            else:
                self.logger.debug(f"Ya le hemos enviado un mensaje al lead {lead.nombre}, lo salteamos")
                lead.message = ''

            self.make_contacted(lead_res[self.contact_id_field])

            leads_info.append(lead)
            row_lead = self.sheet.map_lead(lead.__dict__, self.headers)
            self.sheet.write([row_lead])
