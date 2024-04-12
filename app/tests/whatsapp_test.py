import sys
sys.path.append('.')

from src.whatsapp import Whatsapp

if __name__ == "__main__":
    wpp = Whatsapp()
    #wpp.send_message("5493415854220", "Mensaje de prueba")
    #wpp.send_response("5493415854220", {"name": "Diego", "phone": "4444"})
    wpp.send_template("5493415854220", "8_sobreprecio", [], "es_ES")
