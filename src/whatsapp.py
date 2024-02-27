import requests
import json
import os

from src.logger import Logger

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
            if self.logger:
                self.logger.error("Error enviando el mensaje")
                self.logger.error(res.text)
            else:
                print("error enviando el mensaje")
                print(res.text)
            return

        if self.logger:
            self.logger.success("Mensaje enviado con exito")
        else:
            print("Mensaje enviado con exito")
            print(res.text)

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

    def send_bienvenida(self, to: str, asesor: dict[str, str]):
        payload = {
            "messaging_product": "whatsapp",
            "recipient_type": "individual",
            "to": to,
            "type": "template",
            "template": { 
                "name": "bienvenida", 
                "language": {
                    "code": "es_MX"
                },
                "components": [{
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
                }]
            }
        }
        self.send_request(payload)

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

    whatsapp = Whatsapp(None)
    whatsapp.send_video(nro)
    #send_static_template('', nro)
    #send_message('', nro, "otro mensaje distinto")
    #send_bienvenida('', nro, {
    #    "name": "Brenda",
    #    "phone": "12345"
    #})

# <Link ajdkajdksjak> underlayColor={} activeOpacity={}>
#s/<Link \(.*>\) /<Link to="\1" \2 underlayColor={} activeOpacity={} >
