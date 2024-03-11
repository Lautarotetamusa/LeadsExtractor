from flask import Flask, request, jsonify, render_template, send_from_directory
from time import gmtime, strftime
from flask_cors import CORS
import threading
import json

from src.asesor import assign_asesor
import src.infobip as infobip
from src.whatsapp import Whatsapp
from src.logger import Logger
from src.sheets import Sheet
from src.lead import Lead
from src.numbers import parse_number 

app = Flask(__name__)
CORS(app)

logger = Logger("Server")
sheet = Sheet(logger, "mapping.json")
whatsapp = Whatsapp()
headers = sheet.get("A2:Z2")[0]

with open('messages/bienvenida_1.txt') as f:
    bienvenida_1 = f.read()
with open('messages/bienvenida_2.txt') as f:
    bienvenida_2 = f.read()

#Scrapers
from src.inmuebles24.scraper import main as inmuebles24_scraper
from src.lamudi.scraper import main as lamudi_scraper
from src.propiedades_com.scraper import main as propiedades_scraper
from src.casasyterrenos.scraper import main as casasyterrenos_scraper

PORTALS = {
    "casasyterrenos": {
        "scraper": casasyterrenos_scraper
    },
    "propiedades": {
        "scraper": propiedades_scraper
    },
    "inmuebles24": {
        "scraper": inmuebles24_scraper
    },
    "lamudi": {
        "scraper": lamudi_scraper 
    }
}

def ejecutar_script(portal, url_or_filters, spin_msg):
    # Reemplaza esta línea con la lógica de ejecución de tus scripts
    print("Ejecutando", f"python main.py {portal} {url_or_filters} {spin_msg}")
    PORTALS[portal]["scraper"](url_or_filters, spin_msg)

@app.route('/')
def index():
    return render_template('index.html')

@app.route('/webhooks', methods=['GET'])
def webhooks_validate():
    print(request.args)
    #hub_mode = request.args.get('hub.mode')
    hub_challenge = request.args.get('hub.challenge')
    #hub_verify_token = request.args.get('hub.verify_token')
    #TOKEN = 'Lautaro123.'

    return hub_challenge

@app.route('/asesor', methods=['GET'])
def recive_ivr_call():
    args = request.args.to_dict()
    phone = args.get("msidsn", None)
    fuente = args.get("fuente", None)
    #No se puede obtener el nombre desde infobip
    name = args.get("name", None)
    if not phone or not fuente or not name:
        return {
            "error": "Falta algun campo en la peticion desde infobip (phone, fuente, name)"
        }, 400
    logger.debug("msidsn: "+str(phone))

    #Hacemos esto porque infobip nos devuelve un campo msidn con el numero,
    #Pero este campo no esta correcatmente formatedo, por ejemplo para el numero:
    #5493411234567, devolveria 523411234567. Es decir agregando un 52 como si fuese de mexico
    phone = phone[2::] #Removemos el '52'
    lead = Lead()
    fecha = strftime("%d/%m/%Y", gmtime())
    lead.set_args({
        "nombre": phone,
        "telefono": parse_number(logger, phone, code="MX") or phone,
        "fuente": fuente,
        "fecha_lead": fecha,
        "fecha": fecha
    })

    is_new, lead = assign_asesor(lead)
    if is_new: #Lead nuevo
        infobip.create_person(logger, lead)

    whatsapp.send_msg_asesor(lead.asesor['phone'], lead)
    row_lead = sheet.map_lead(lead.__dict__, headers)
    sheet.write([row_lead])
    return lead.asesor

@app.route('/webhooks', methods=['POST'])
def recive_wpp_msg():
    data = request.get_json()

    try:
        if not 'contacts' in data['entry'][0]['changes'][0]['value']:
            #Mensaje que envia el bot o cambio de estado en el mensaje
            #logger.debug(json.dumps(data, indent=4))
            return ''
    except KeyError:
        logger.error("La peticion no es la esperada")
        logger.debug(json.dumps(data, indent=4))
        return ''

    logger.success("Nuevo mensaje del lead recibido!")
    print(json.dumps(data, indent=4))
    value = data['entry'][0]['changes'][0]['value']

    lead = Lead()
    fecha = strftime("%d/%m/%Y", gmtime())
    lead.set_args({
        "telefono": value['contacts'][0]['wa_id'],
        "nombre": value['contacts'][0]['profile']['name'],
        "fuente": "Whatsapp",
        "fecha_lead": fecha,
        "fecha": fecha
    })
    lead.telefono = parse_number(logger, '+'+lead.telefono) or lead.telefono
    lead.link = f"https://web.whatsapp.com/send/?phone={lead.telefono}"

    is_new, lead = assign_asesor(lead)
    if is_new: #Lead nuevo
        whatsapp.send_image(lead.telefono)
        whatsapp.send_message(lead.telefono, bienvenida_1)
        whatsapp.send_message(lead.telefono, bienvenida_2.format(
            asesor_name=lead.asesor['name'], 
            asesor_phone=lead.asesor['phone'])
        )
        whatsapp.send_video(lead.telefono)

        infobip.create_person(logger, lead)
    else: #Lead existente
        whatsapp.send_response(lead.telefono, lead.asesor)
    
    whatsapp.send_msg_asesor(lead.asesor['phone'], lead)
    row_lead = sheet.map_lead(lead.__dict__, headers)
    sheet.write([row_lead])
    return lead.asesor

@app.route('/.well-known/pki-validation/<path:path>')
def serve_static(path):
    return send_from_directory('static', path)

@app.route('/execute', methods=['POST'])
def ejecutar_script_route():
    data = request.get_json()

    portal = data.get('portal')
    spin_msg = data.get('message')
    url_or_filters = data.get('url_or_filters')

    if not all([portal, spin_msg, url_or_filters]):
        return jsonify({'error': 'Se requieren campos portal, message y url_or_filters'}), 400

    # Ejecutar el script en segundo plano usando threading
    thread = threading.Thread(target=ejecutar_script, args=(portal, url_or_filters, spin_msg))
    thread.start()

    return "Proceso en segundo plano iniciado."

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=80, debug=True)
