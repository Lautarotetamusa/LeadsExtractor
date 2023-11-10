from __future__ import print_function

import os.path
import json

from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError

from logger import Logger

import os
from dotenv import load_dotenv
load_dotenv()

# If modifying these scopes, delete the file token.json.
SCOPES = ['https://www.googleapis.com/auth/spreadsheets']

SHEET_ID = os.getenv("SHEET_ID")

class Sheet():
    def __init__(self, logger: Logger):
        self.id = SHEET_ID
        self.logger = logger

        creds = None
        # The file token.json stores the user's access and refresh tokens, and is
        # created automatically when the authorization flow completes for the first
        # time.
        if os.path.exists('token.json'):
            creds = Credentials.from_authorized_user_file('token.json', SCOPES)
        # If there are no (valid) credentials available, let the user log in.
        if not creds or not creds.valid:
            if creds and creds.expired and creds.refresh_token:
                creds.refresh(Request())
            else:
                flow = InstalledAppFlow.from_client_secrets_file('../credentials.json', SCOPES)
                creds = flow.run_local_server(port=0)
            # Save the credentials for the next run
            with open('token.json', 'w') as token:
                token.write(creds.to_json())

        try:
            service = build('sheets', 'v4', credentials=creds)

            # Call the Sheets API
            self.sheet = service.spreadsheets()
        except HttpError as err:
            self.logger.error("error conectandose al sheet")
            self.logger.error(str(err))
        
    def get(self, range):
        result = self.sheet.values().get(spreadsheetId=self.id, range=range).execute()
        values = result.get('values', [])
        return values

    #Extrae la posicion en la que se encuentran cada uno de los headers del archivo
    def get_dict_headers(self):
        list_headers = self.get("A2:Z2")[0]
        headers = {}
        for element in list_headers:
            headers[element] = list_headers.index(element)
        return headers

    def write(self, data):
        body = {'values': data}
        self.logger.debug(body)
        result = self.sheet.values().append(
            spreadsheetId=self.id,
            range="BD-General!A3",
            valueInputOption="RAW", body=body
        ).execute()
        self.logger.success("Se escribio el google sheets correctamente")
        self.logger.debug(result)

    # Convertir un objecto lead a algo que se pueda escribir en el sheets.
    def map_lead(self, lead: object, headers: object):
        if not os.path.exists('mapping.json'):
            self.logger.error(f"El archivo mapping.json no existe")
            exit(1)

        with open('mapping.json', "r") as f:
            mapping = json.load(f)

        lead_row = ["" for i in range(len(headers))]

        i = 0
        for header in headers:
            route = mapping[header]
            lead_row[i] = get_prop(lead, route, self.logger)
            i += 1

        return lead_row

# route: propiedad.link -> lead['propiedad']['link']
def get_prop(obj: object, route: str, logger: Logger):
    split = route.split('.')
    if len(split) == 1:
        if split[0] == "":
            return ""
        if split[0] not in obj:
            logger.error(f"La propiedad {split[0]} no se encuentra en el lead")
            return ""
        return obj[split[0]]
    
    return get_prop(obj[split[0]], '.'.join(split[1:]), logger)

def main():
    sheet = Sheet()
    headers = sheet.get("A2:Z2")[0]
    #print(headers)
    #lead = {"fec{'fuente': 'Inmuebles24', 'fecha': '2023-11-06', 'nombre': 'Christian', 'link': 'https://www.inmuebles24.com/panel/interesados/455756085', 'telefono': '3', 'email': 'Michrisca_1705@hotmail.com', 'propiedad': {'titulo': 'Casa en venta en Puerta Del Roble. Hermosa, segura y confortable.', 'link': '', 'precio': 12125000, 'ubicacion': 'Av. Juan Palomar y Arias 930', 'tipo': 'Casa'}, 'busquedas': {'zonas': ['avenida  juan palomar'], 'tipo': 'Casa', 'presupuesto': '12125000 12125000', 'cantidad_anuncios': 573, 'contactos': 2, 'inicio_busqueda': 59, 'total_area': '643 643', 'covered_area': '450 450', 'banios': '5 5', 'recamaras': '5 5'}}
    
    #row_lead = map_lead(lead, headers)
    #sheet.write([row_lead])
    #print(row_lead)

if __name__ == '__main__':
    main()