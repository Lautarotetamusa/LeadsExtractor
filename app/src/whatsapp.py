import requests
import json
import os

from src.logger import Logger
from src.lead import Lead

NUMBER_ID = os.getenv("WHATSAPP_NUMBER_ID")
ACCESS_TOKEN = os.getenv("WHATSAPP_ACCESS_TOKEN")
IMAGE_ID = os.getenv("WHATSAPP_IMAGE_ID")
VIDEO_ID = os.getenv("WHATSAPP_VIDEO_ID")

assert ACCESS_TOKEN != "" and ACCESS_TOKEN != None, "ACCESS_TOKEN is not in enviroment"
assert NUMBER_ID != "" and NUMBER_ID != None, "NUMBER_ID is not in enviroment"
assert IMAGE_ID != "" and IMAGE_ID != None, "IMAGE_ID is not in enviroment"
assert VIDEO_ID != "" and VIDEO_ID != None, "VIDEO_ID is not in enviroment"

URL = f"https://graph.facebook.com/v17.0/{NUMBER_ID}/messages"
image_link = "https://buffer.com/library/content/images/size/w1200/2023/10/free-images.jpg"

class Whatsapp():
    def __init__(self, logger: Logger | None = None):
        if logger:
            self.logger = Logger(logger.fuente + " > Whatsapp API")
        else:
            self.logger = Logger("Whatsapp API")

        self.headers = {
            'Authorization': f'Bearer {ACCESS_TOKEN}',
            'Content-Type': 'application/json'
        }
    
    def send_request(self, payload, **args):
        res = requests.post(URL, data=json.dumps(payload), headers=self.headers, **args)
        if not res.ok:
            self.logger.error("Error enviando el mensaje")
            self.logger.error(res.text)
            return

        self.logger.success("Mensaje enviado con exito")
        self.logger.debug(res.text)

    def send_message(self, to: str, message: str):
        assert to != None, "El numero de telefono receptor es None"

        payload = {
            "messaging_product": "whatsapp",
            "recipient_type": "individual",
            "to": to,
            "type": "text",
            "text": {
                "preview_url": False,
                "body": message
            }
        }
        return self.send_request(payload)
    
    def send_template(self, to: str, name: str, components, language="es_MX"):
        assert to != None, "El numero de telefono receptor es None"
        print(components)

        payload = {
            "messaging_product": "whatsapp",
            "recipient_type": "individual",
            "to": to,
            "type": "template",
            "template": { 
                "name": name, 
                "language": {
                    "code": language
                },
                "components": components
            }
        }
        return self.send_request(payload)

    def send_bienvenida(self, to: str, asesor: dict[str, str]):
        assert to != None, "El numero de telefono receptor es None"
        assert 'name' in asesor and asesor['name'] != "" and asesor['name'] != None, "El asesor no tiene nombre"
        assert 'phone' in asesor and asesor['phone'] != "" and asesor['phone'] != None, "El asesor no tiene telefono"

        self.logger.debug("Enviando bienvenida lead "+to)
        self.send_template(
            to=to,
            name="bienvenida",
            components=[{
                "type": "header",
                "parameters": [{
                    "type": "image",
                    "image": {
                        "id": IMAGE_ID
                    }
                }]
            },{
                "type": "body",
                "parameters": [{
                    "type": "text",
                    "text": asesor['name']
                },{
                    "type": "text",
                    "text": '+'+asesor['phone']
                }]
            }])

    def send_response(self, to: str, asesor: dict[str, str]):
        assert to != None, "El numero de telefono receptor es None"
        assert 'name' in asesor and asesor['name'] != "" and asesor['name'] != None, "El asesor no tiene nombre"

        self.logger.debug("Enviando respuesta al lead "+to)
        self.send_template(
            to=to,
            name="2do_mensaje_bienvenida",
            components=[{
                "type": "body",
                "parameters": [{
                    "type": "text",
                    "text": asesor['name']
                }]
            }])

    def send_msg_asesor(self, to: str, lead: Lead, is_new: bool=True):
        assert to != None, "El numero de telefono es None"
        assert lead.telefono != "", "El telefono del lead esta vacio"
        assert lead.fecha_lead != "", "La fecha del lead esta vacia"
        assert lead.fuente != "", "La fuente del lead esta vacia"

        if is_new:
            template = "msg_asesor_2"
        else:
            template = "msg_asesor_duplicado"

        self.logger.debug("Enviando mensaje al asesor "+to)
        self.send_template(
            to=to,
            name=template,
            components=[{
                "type": "body",
                "parameters": [
                    {
                        "type": "text",
                        "text": lead.nombre if lead.nombre != "" else " - "
                    },{
                        "type": "text",
                        "text": lead.fuente
                    },{
                        "type": "text",
                        "text": '+'+lead.telefono
                    },{
                        "type": "text",
                        "text": lead.email if lead.email != "" else " - "
                    },{
                        "type": "text",
                        "text": lead.fecha_lead
                    },{
                        "type": "text",
                        "text": lead.link if lead.link != "" else " - "
                    },{
                        "type": "text",
                        "text": lead.propiedad["titulo"] if lead.propiedad["titulo"] != "" else " - "
                    },{
                        "type": "text",
                        "text": lead.propiedad["precio"] if lead.propiedad["precio"] != "" else " - "
                    },{
                        "type": "text",
                        "text": lead.propiedad["ubicacion"] if lead.propiedad["ubicacion"] != "" else " - "
                    },{
                        "type": "text",
                        "text": lead.propiedad["link"] if lead.propiedad["link"] != "" else " - "
                    }
                ]
            }])

    def send_image(self, to: str):
        assert to != None, "El numero de telefono receptor es None"

        self.logger.debug("Enviando imagen a "+to)
        payload = {
            "messaging_product": "whatsapp",
            "to": to,
            "type": "image",
            "image": {
                "id" : IMAGE_ID
            }
        }
        self.send_request(payload)

    def send_document(self, to: str, url: str, filename: str, caption: str):
        assert to != None, "El numero de telefono receptor es None"

        self.logger.debug("Enviando documento a "+to)
        payload = {
            "messaging_product": "whatsapp",
            "to": to,
            "type": "document",
            "document": {
                "link" : url,
                "caption": caption,
                "filename": filename
            }
        }
        self.send_request(payload)

    def send_video(self, to: str):
        assert to != None, "El numero de telefono receptor es None"

        self.logger.debug("Enviando video a "+to)
        payload = {
            "messaging_product": "whatsapp",
            "to": to,
            "type": "video",
            "video": {
                "id": VIDEO_ID
            }
        }
        self.send_request(payload)

if __name__ == "__main__":
    nro_diego = "5213319466986"
    nro_mio   =  "5493415854220"
    nro = nro_mio

    whatsapp = Whatsapp()
    whatsapp.send_video(nro)
    #send_static_template('', nro)
    #send_message('', nro, "otro mensaje distinto")
    #send_bienvenida('', nro, {
    #    "name": "Brenda",
    #    "phone": "12345"
    #})

# <Link ajdkajdksjak> underlayColor={} activeOpacity={}>
#s/<Link \(.*>\) /<Link to="\1" \2 underlayColor={} activeOpacity={} >
