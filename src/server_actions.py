from src.whatsapp import Whatsapp
import src.infobip as infobip
from src.logger import Logger
from src.sheets import Sheet
from src.message import format_msg

logger = Logger("Actions")
sheet = Sheet(logger, "mapping.json")
whatsapp = Whatsapp()
headers = sheet.get("A2:Z2")[0]

with open('messages/bienvenida_1.txt') as f:
    bienvenida_1 = f.read()
with open('messages/bienvenida_2.txt') as f:
    bienvenida_2 = f.read()

def new_lead_action(logger: Logger, lead):
    whatsapp.send_image(lead.telefono)
    whatsapp.send_message(lead.telefono, bienvenida_1)
    whatsapp.send_message(lead.telefono, format_msg(lead, bienvenida_2))
    whatsapp.send_video(lead.telefono)

    infobip.create_person(logger, lead)

def common_lead_action(lead):
    whatsapp.send_msg_asesor(lead.asesor['phone'], lead)
    row_lead = sheet.map_lead(lead.__dict__, headers)
    sheet.write([row_lead])
