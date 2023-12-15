import requests
from dotenv import load_dotenv
import os

load_dotenv()
CHAT_WEBHOOK = os.getenv("CHAT_WEBHOOK")
if not CHAT_WEBHOOK:
    print("ERROR: falta CHAT_WEBHOOK en el enviroment")
    exit(1)

def send_message(app_message: dict):
    message_headers = {"Content-Type": "application/json; charset=UTF-8"}
    
    res = requests.post(CHAT_WEBHOOK, headers=message_headers, json=app_message)
    if res.status_code != 200:
        print("Error escribiendo en google chat")

def generate_message(msg: str, html_color: str, time: str, fuente: str):
    return {
        'cardsV2': [{
            'cardId': 'createCardMessage',
            'card': {
                "sections": [
                    {
                        "header": fuente,
                        "collapsible": True,
                        "uncollapsibleWidgetsCount": 1,
                        "widgets": [
                            {
                                "decoratedText": {
                                    "icon": {
                                        "knownIcon": "DESCRIPTION"
                                        },
                                    "topLabel": time,
                                    "text": f"<font color=\"{html_color}\">{msg}</font>"
                                    }
                                }
                            ]
                        }
                    ]
                }
            }]
        }

if __name__ == "__main__":
    app_message = {
      "text": f"<font color=\"#FF0000\">prueba texto rojo</font>",
      "formattedText": f"<font color=\"#FF0000\">prueba texto rojo</font>"
    }

    send_message(app_message)
