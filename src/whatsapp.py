import requests
import json
import os

from src.logger import Logger
from src.lead import Lead

NUMBER_ID = os.getenv("WHATSAPP_NUMBER_ID")
ACCESS_TOKEN = os.getenv("WHATSAPP_ACCESS_TOKEN")

URL = f"https://graph.facebook.com/v17.0/{NUMBER_ID}/messages"
image_link = "https://buffer.com/library/content/images/size/w1200/2023/10/free-images.jpg"
image_id = "1183968386347752" #bienvenida.jpeg, 240 KB
video_id = "3551608448489563" #Bienivenida 12MB

class Whatsapp():
    def __init__(self):
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
        self.send_template(
            to=to,
            name="bienvenida",
            components=[{
                "type": "header",
                "parameters": [{
                    "type": "image",
                    "image": {
                        "id": image_id
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

    def send_msg_asesor(self, to: str, lead: Lead):
        self.send_template(
            to=to,
            name="msg_asesor",
            components=[{
                "type": "body",
                "parameters": [{
                    "type": "text",
                    "text": lead.nombre
                },{
                    "type": "text",
                    "text": '+'+lead.telefono
                }]
            }])

    def send_image(self, to: str):
        payload = {
            "messaging_product": "whatsapp",
            "to": to,
            "type": "image",
            "image": {
                "id" : image_id
            }
        }
        self.send_request(payload)

    def send_video(self, to: str):
        payload = {
            "messaging_product": "whatsapp",
            "to": to,
            "type": "video",
            "video": {
                "id": video_id
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
