import requests

CHAT_WEBHOOK = "https://chat.googleapis.com/v1/spaces/AAAABMZYJoU/messages?key=AIzaSyDdI0hCZtE6vySjMm-WEfRq3CPzqKqqsHI&token=q13-NM92LfUUGqx7SxGg2p7jReO6bNxMmGal47MflyQ"

def send_chat_message(message: str):
    app_message = {
            'cardsV2': [{
                'cardId': 'createCardMessage',
                'card': {
                    "sections": [
                        {
                            "header": "Inmuebles 24",
                            "collapsible": True,
                            "uncollapsibleWidgetsCount": 1,
                            "widgets": [
                                {
                                    "decoratedText": {
                                        "icon": {
                                            "knownIcon": "DESCRIPTION"
                                            },
                                        "topLabel": "12:19:37",
                                        "text": "<font color=\"#FF0000\">El token de acceso expiro</font>"
                                        }
                                    }
                                ]
                            }
                        ]
                    }
                }]
            }

    message_headers = {"Content-Type": "application/json; charset=UTF-8"}
    
    res = requests.post(CHAT_WEBHOOK, headers=message_headers, json=app_message)
    print(res.json())

if __name__ == "__main__":
  send_chat_message("test")
