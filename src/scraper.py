from email.mime.application import MIMEApplication
import os
from src.make_requests import Request
from src.logger import Logger
from src.message import generate_mensage
from src.sheets import Gmail, Sheet
import src.infobip as infobip

class Scraper():
    def __init__(self, name: str, contact_id_field: str, send_msg_field: str):
        self.name = name
        self.logger = Logger(name)
        self.gmail = Gmail({
            "email": os.getenv("EMAIL_CONTACT"),
        }, self.logger)

        self.contact_id_field = contact_id_field
        self.send_msg_field = send_msg_field
        self.setup()

    def setup(self):
        self.logger.debug(f"Inicializando {self.name}")
        self.sheet = Sheet(self.logger, "mapping.json")
        self.headers = self.sheet.get("A2:Z2")[0]

        with open('messages/gmail.html', 'r') as f:
            self.gmail_spin = f.read()
        with open('messages/gmail_subject.html', 'r') as f:
            self.gmail_subject = f.read()

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
    def get_lead_info(self, lead: dict) -> dict:
        return {}
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
            print(lead_res)
            lead = self.get_lead_info(lead_res)

            if self.send_message_condition(lead):
                if lead['email'] != '':
                    if lead["propiedad"]["ubicacion"] == "":
                        lead["propiedad"]["ubicacion"] = "que consultaste"
                    else:
                        lead["propiedad"]["ubicacion"] = "ubicada en " + lead["propiedad"]["ubicacion"]

                    gmail_msg = generate_mensage(lead, self.gmail_spin)
                    subject = generate_mensage(lead, self.gmail_subject)
                    self.gmail.send_message(gmail_msg, subject, lead["email"], self.attachment)

                    infobip.create_person(self.logger, lead)

                msg = generate_mensage(lead)
                lead["message"] = msg.replace('\n', '')
                self.send_message(lead_res[self.send_msg_field], msg)
            else:
                self.logger.debug(f"Ya le hemos enviado un mensaje al lead {lead['nombre']}, lo salteamos")
                lead['message'] = ''

            self.make_contacted(lead_res[self.contact_id_field])

            leads_info.append(lead)
            row_lead = self.sheet.map_lead(lead, self.headers)
            self.sheet.write([row_lead])
