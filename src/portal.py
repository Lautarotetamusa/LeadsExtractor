import os

from message import generate_mensage
import src.infobip as infobip
from src.logger import Logger
from src.sheets import Sheet, Gmail

class Portal():
    logger: Logger
    attachemnt_name: str = 'Bienvenido a Rebora! Seguridad, Confort y Placer - Casas de gran diseño y alta calidad.pdf'

    def __init__(self, logger: Logger):
        self.logger = logger
    
        self.sheet = Sheet(logger, "mapping.json")
        self.gmail = Gmail({
            "email": os.getenv("EMAIL_CONTACT"),
            }, logger)
        self.headers = self.sheet.get("A2:Z2")[0]

        self._load_gmail_content()
        logger.debug(f"Extrayendo leads")
    
    def _load_gmail_content(self):
        with open('messages/gmail.html', 'r') as f:
            self.gmail_spin = f.read()
        with open('messages/gmail_subject.html', 'r') as f:
            self.gmail_subject = f.read()

        # Adjuntar archivo PDF
        with open('messages/attachment.pdf', 'rb') as pdf_file:
            self.attachment = MIMEApplication(pdf_file.read(), _subtype="pdf")
            self.attachment.add_header('Content-Disposition', 'attachment', filename=self.attachemnt_name)

    def extract_lead_info(self):
        pass

    def send_message(self):
        pass

    def make_contacted(self, lead_id, **kwargs):
        pass

    def main(self):
        lead = self.extract_lead_info(raw_lead)
        self.logger.debug(lead)

        lead["message"] = generate_mensage(lead).replace('\n', '')
        if lead['email'] != '':
            if lead["propiedad"]["ubicacion"] == "":
                lead["propiedad"]["ubicacion"] = "que consultaste"
            else:
                lead["propiedad"]["ubicacion"] = "ubicada en " + lead["propiedad"]["ubicacion"]

            gmail_msg = generate_mensage(lead, self.gmail_spin)
            subject = generate_mensage(lead, self.gmail_subject)
            self.gmail.send_message(gmail_msg, subject, lead["email"], self.attachment)
            infobip.create_person(self.logger, lead)
        self.make_contacted(lead["id"])

        row_lead = self.sheet.map_lead(lead, headers)
        leads.append(row_lead)

        sheet.write(leads)

        logger.debug(f"len: {len(leads)}")

def main():

    status = "1" #Filtramos solamente los leads nuevos
    first = True
    page = 1
    url = f"{API_URL}/list_contact/?page={page}&status={status}"

    leads = []
    while first == True or url != None:
        data = request.make(url).json()

        url = data["next"]
        total = data["count"]
        logger.debug(f"total: {total}")

        for raw_lead in data["results"]:
            lead = extract_lead_info(raw_lead)
            logger.debug(lead)

            lead["message"] = generate_mensage(lead).replace('\n', '')
            if lead['email'] != '':
                if lead["propiedad"]["ubicacion"] == "":
                    lead["propiedad"]["ubicacion"] = "que consultaste"
                else:
                    lead["propiedad"]["ubicacion"] = "ubicada en " + lead["propiedad"]["ubicacion"]

                gmail_msg = generate_mensage(lead, gmail_spin)
                subject = generate_mensage(lead, gmail_subject)
                gmail.send_message(gmail_msg, subject, lead["email"], attachment)
                infobip.create_person(logger, lead)
            make_contacted(lead["id"])

            row_lead = sheet.map_lead(lead, headers)
            leads.append(row_lead)

        sheet.write(leads)

        logger.debug(f"len: {len(leads)}")
        first = False

    logger.success(f"Se encontraron {len(leads)} nuevos Leads")
