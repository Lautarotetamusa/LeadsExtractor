import requests
import json
import os

NUMBER_ID = os.getenv("WHATSAPP_NUMBER_ID")
ACCESS_TOKEN = os.getenv("WHATSAPP_ACCESS_TOKEN")

ACCESS_TOKEN="EAAbDrjrkX5IBOZCUNjSxUXBmqnCFfOk9HHc40a3gC2W9nBZAruAdbUTAvZCV3vmX19QZB6ByfHS5VVuQXTqoSFaKS6a2TcEVKLTQYgfONCPObof6pSRfK6KIxclPyE0kdnkVZB6LS67gEcuf5TskzK0a9KOnPGcCAeHJZCA6T3Ic5NhJz2WOgQpghpzOlRLkVxrD5W97Lz8nmYc5bgK3ukivQyZAusZD"
NUMBER_ID="218861484652252"

URL = f"https://graph.facebook.com/v19.0/{NUMBER_ID}/messages"

def send_message(to: str):
    payload = {
        "messaging_product": "whatsapp",
        #"recipient_type": "individual",
        "to": to,
        "type": "template",
        "template": { 
            "name": "bienvenida_rebora", 
            "language": {
                "code": "es_MX"
            }
        }
    }
    headers = {
        'Authorization': f'Bearer {ACCESS_TOKEN}',
        'Content-Type': 'application/json'
    }
    res = requests.post(URL, data=json.dumps(payload), headers=headers)
    if not res.ok:
        print("Error enviando el mensaje")
        print(res.text)
        return

    print("Mensaje enviado con exito")

send_message("54341155854220")
