from datetime import datetime
from dateutil.relativedelta import relativedelta
from flask import Flask, request, jsonify, render_template, send_from_directory
from time import gmtime, strftime
from flask_cors import CORS
import threading
import json
import os

from src.server_actions import common_lead_action, new_lead_action
from src.asesor import assign_asesor, next_asesor
from src.whatsapp import Whatsapp
from src.logger import Logger
from src.lead import Lead
from src.numbers import parse_number 

app = Flask(__name__)
CORS(app)

logger = Logger("Server")
whatsapp = Whatsapp()

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT != None, "DATE_FORMAT is not seted"

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

@app.route('/roundrobin', methods=['GET'])
def round_robin():
    asesor = next_asesor()  
    return asesor

@app.route('/asesor', methods=['GET'])
def recive_ivr_call():
    args = request.args.to_dict()
    phone = args.get("msidsn", None)
    fuente = args.get("fuente", None)

    if not phone or not fuente:
        logger.error("Falta algun campo en la peticion")
        return {
            "error": "Falta algun campo en la peticion desde infobip (msidsn, fuente, name)"
        }, 400
    logger.debug("msidsn: "+str(phone))

    lead = Lead()
    fecha = strftime(DATE_FORMAT, gmtime())
    lead.set_args({
        "nombre": phone,
        "fuente": "IVRcall",
        "fecha_lead": fecha,
        "fecha": fecha
    })
    telefono = parse_number(logger, phone, "MX")
    if not telefono:
        #Si el numero no es mexicano va a llegar con un 52 adelante igual por ejemplo 525493415854220
        phone = phone[2::] #Removemos el '52'
        telefono = parse_number(logger, "+"+phone)
    lead.telefono = telefono or lead.telefono

    is_new, lead = assign_asesor(lead)
    lead.validate()
    if is_new: #Lead nuevo
        new_lead_action(logger, lead)
    else: #Lead existente
        whatsapp.send_response(lead.telefono, lead.asesor)

    common_lead_action(lead)
    return lead.asesor

@app.route('/webhooks', methods=['POST'])
def recive_wpp_msg():
    data = request.get_json()
    value = data['entry'][0]['changes'][0]['value']

    try:
        if not 'contacts' in value:
            #Mensaje que envia el bot o cambio de estado en el mensaje
            #logger.debug(json.dumps(data, indent=4))
            return ''
    except KeyError:
        logger.error("La peticion no es la esperada")
        logger.debug(json.dumps(data, indent=4))
        return ''

    logger.success("Nuevo mensaje del lead recibido!")
    print(json.dumps(data, indent=4))
    msg_type = value.get('messages', [{}])[0].get('type', None)

    lead = Lead()
    fecha = strftime(DATE_FORMAT, gmtime())
    lead.set_args({
        "telefono": value['contacts'][0]['wa_id'],
        "nombre": value['contacts'][0]['profile']['name'],
        "fuente": "Whatsapp",
        "fecha_lead": fecha,
        "fecha": fecha
    })
    lead.telefono = parse_number(logger, '+'+lead.telefono) or lead.telefono
    lead.link = f"https://web.whatsapp.com/send/?phone={lead.telefono}"

    save = False
    is_new, lead = assign_asesor(lead)
    lead.validate()
    if is_new: #Lead nuevo
        new_lead_action(logger, lead)
        save = True
    else: #Lead existente
        assert lead.fecha_lead != "" and lead.fecha_lead != None, f"El lead {lead.telefono} no tiene fecha de lead"

        months=3
        treshold = datetime.now().date() - relativedelta(month=months)
        fecha_lead = datetime.strptime(lead.fecha_lead, "%Y-%m-%d").date()
        if fecha_lead <= treshold: #Los leads mas viejos que 3 meses
            if msg_type == "request_welcome": #El lead abrio la conversacion
                logger.debug("Nuevo request_welcome en lead ya asignado hace mas de 3 meses, lo salteamos")
            else:
                whatsapp.send_response(lead.telefono, lead.asesor)
                save = True
        else:
            whatsapp.send_response(lead.telefono, lead.asesor)
            save = True

    if save:
        common_lead_action(lead)
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
