from __future__ import print_function
import os
from dotenv import load_dotenv
import os.path
import json
from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError

#Estos dos son necesarios para crear y enviar el mensaje de email
from base64 import urlsafe_b64encode
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText

from src.logger import Logger

load_dotenv()

# If modifying these scopes, delete the file token.json.
SCOPES = ['https://www.googleapis.com/auth/spreadsheets', "https://www.googleapis.com/auth/gmail.send"]
SHEET_ID = os.getenv("SHEET_ID")
HOST=os.getenv("HOST")
TOKEN_FILE = 'token.json'
CREDS_FILE = 'credentials.json'

class Google():
    def __init__(self, logger: Logger):
        creds = None
        self.logger = logger

        if os.path.exists(TOKEN_FILE):
            creds = Credentials.from_authorized_user_file(TOKEN_FILE , SCOPES)

        if not creds or not creds.valid:
            if creds and creds.expired and creds.refresh_token:
                creds.refresh(Request())
            else:
                flow = InstalledAppFlow.from_client_secrets_file(CREDS_FILE, SCOPES)
                #Tuve que editar el código de esta función,
                #Donde se generar el redirect_uri, cambiar http por https y sacar el puerto {}:{} -> {}
                creds = flow.run_local_server(
                        bind_addr="0.0.0.0",
                        host=HOST,
                        port=80,
                        open_browser=False,
                        redirect_uri_trailing_slash=False
                )

            with open(TOKEN_FILE, 'w') as token:
                token.write(creds.to_json())

        self.creds = creds

    def _connect_service(self, name: str, version: str):
        try:
            return build(name, version, credentials=self.creds)
        except HttpError as err:
            self.logger.error(f"Error conectandose al servicio {name}")
            self.logger.error(str(err))
            exit(1)

class Gmail(Google):
    def __init__(self, sender: dict, logger: Logger):
        super().__init__(logger)
        self.sender_email = sender["email"]
        self.service = super()._connect_service('gmail', 'v1')

    def create_message(self, to, subject, body_html, attachment):
        message = MIMEMultipart()
        message['to'] = to
        message['from'] = self.sender_email
        message['subject'] = subject

        # Agregar cuerpo HTML
        message.attach(MIMEText(body_html, 'html'))

        message.attach(attachment)

        raw_message = urlsafe_b64encode(message.as_bytes()).decode()
        return {'raw': raw_message}

    def send_message(self, text: str, subject: str, reciver: str, attachment):
        self.logger.debug(f"Enviando mensaje al correo: {reciver}")
        try:
            message = self.create_message(reciver, subject, text, attachment) 

            response = (
                    self.service.users()
                    .messages()
                    .send(userId="me", body=message)
                    .execute()
            )
            #self.logger.success(f'Email enviando con exito, Id: {response["id"]}')
            # para no enviar mensaje al google chat cada vez
            self.logger.debug(f'Email enviando con exito, Id: {response["id"]}')
        except HttpError as error:
            self.logger.error(f"Error enviando el email: {error}")
            response = None
        return response

class Sheet(Google):
    def __init__(self, logger: Logger, mapping_file):
        super().__init__(logger)
        self.id = SHEET_ID
        self.sheet = super()._connect_service('sheets', 'v4').spreadsheets()

        if not os.path.exists(mapping_file):
            self.logger.error(f"El archivo {mapping_file} no existe")
            exit(1)
        with open(mapping_file, "r") as f:
            self.mapping = json.load(f)
        
    def get(self, range):
        result = self.sheet.values().get(spreadsheetId=self.id, range=range).execute()
        return result.get('values', [])

    #Extrae la posicion en la que se encuentran cada uno de los headers del archivo
    def get_dict_headers(self, header_range="A2:Z2"):
        list_headers = self.get(header_range)[0]
        headers = {}
        for element in list_headers:
            headers[element] = list_headers.index(element)
        return headers

    def write(self, data, range_sheet="BD-General"):
        body = {'values': data}

        try:
            self.sheet.values().append(
                spreadsheetId=self.id,
                range=range_sheet,
                valueInputOption="RAW",
                body=body
            ).execute()
            self.logger.success("Se escribio el google sheets correctamente")
        except Exception as e:
            self.logger.error("Ocurrio un error guardando en el sheets: " + str(e))

    # Convertir un objecto lead a algo que se pueda escribir en el sheets.
    def map_lead(self, lead: dict, headers: dict):
        assert(lead != None)

        lead_row = ["" for _ in range(len(headers))]

        i = 0
        for header in headers:
            route = self.mapping[header]
            lead_row[i] = get_prop(lead, route, self.logger)
            i += 1

        return lead_row

# route: propiedad.link -> lead['propiedad']['link']
def get_prop(obj: dict, route: str, logger: Logger):
    split = route.split('.')
    if len(split) == 1:
        if split[0] == "":
            return ""
        if split[0] not in obj:
            logger.error(f"La propiedad {split[0]} no se encuentra en el lead")
            return ""
        return obj[split[0]]
    
    return get_prop(obj[split[0]], '.'.join(split[1:]), logger)

def set_prop(obj: dict, route: str, value: str) -> dict:
    split = route.split('.')
    if len(split) == 1:
        if split[0] == "":
            return obj
        obj[split[0]] = value
        return obj
    
    #obj[split[0]] = {}
    obj[split[0]] = set_prop(obj[split[0]], '.'.join(split[1:]), value)
    return obj
