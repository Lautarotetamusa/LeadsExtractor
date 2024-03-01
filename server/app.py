from flask import Flask, request, jsonify, render_template, send_from_directory, session
from flask_cors import CORS
import threading
import urllib.parse
import json

from src.logger import Logger
from src.sheets import Sheet
from src.schema import LEAD_SCHEMA
from src.whatsapp import Whatsapp

app = Flask(__name__)
CORS(app)

logger = Logger("Server")
sheet = Sheet(logger, "mapping.json")
whatsapp = Whatsapp()
headers = sheet.get("A2:Z2")[0]

#Scrapers
from src.inmuebles24.scraper import main as inmuebles24_scraper
from src.lamudi.scraper import main as lamudi_scraper
from src.propiedades_com.scraper import main as propiedades_scraper
from src.casasyterrenos.scraper import main as casasyterrenos_scraper
import src.infobip as infobip

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

#Esta ruta hace el round robin del asesor para el flow de infobip
#Si la llamamos con id=0, nos devuelve el asesor 1. luego llamamos al id=1 y nos devuelve el 2.
#De este modo siempre podemos llamar al siguiente asesor desde el flow de infobip
ASESOR_i = 0
def round_robin():
    global ASESOR_i
    rows = sheet.get('Asesores!A2:C25')
    activos = [row for row in rows if row[2] == "Activo"]
    print("Activos: ", activos)
    ASESOR_i += 1
    ASESOR_i %= len(activos)
    
    asesor = {
        "name":  activos[ASESOR_i][0],
        "phone": activos[ASESOR_i][1]
    }
    logger.debug("Asesor asignado: "+asesor["name"])
    return asesor

@app.route('/asesor', methods=['GET'])
def get_asesor():
    args = request.args.to_dict()
    phone = args.get("msidsn", None)
    fuente = args.get("fuente", None)
    if not phone or not fuente:
        return {
            "error": "Falta el campo 'msidsn' o el campo 'fuente' en los args"
        }, 400

    #Hacemos esto porque infobip nos devuelve un campo msidn con el numero,
    #Pero este campo no esta correcatmente formatedo, por ejemplo para el numero:
    #5493411234567, devolveria 523411234567. Es decir agregando un 52 como si fuese de mexico
    phone = phone[2::] #Removemos el '52'
    json_filter = {"#contains": {"contactInformation": {"phone": {"number": phone}}}}
    filtro = urllib.parse.quote(json.dumps(json_filter))
    person = infobip.search_person(logger, filtro)

    #si la persona ya existe buscamos el asesor de esa persona
    if person != None:
        logger.debug("Una persona que ya existe realizo una llamada")
        name =  person.get('customAttributes', {}).get('asesor_name', None)
        phone = person.get('customAttributes', {}).get('asesor_phone', None)
        #Si la persona ya tiene asesor asignado (que debería ser siempre), pero por si acasol lo chequeamos
        if name != None and phone != None:
            return {
                "name":  name,
                "phone": phone
            }

    logger.debug("Una persona nueva realizo una llamada")
    #Si la persona no existe creamos una persona nueva
    lead = LEAD_SCHEMA.copy()
    lead['telefono'] = phone
    lead['fuente'] = fuente
    lead['message'] = ""

    asesor = round_robin()
    lead['asesor_name'] = asesor['name']
    lead['asesor_phone'] = asesor['phone']

    whatsapp.send_message(asesor['phone'], f"""Tienes un nuevo lead asignado. 
Nombre: {lead['nombre']}
Telefono: {lead['telefono']}
Comunicate con este lead lo antes posible.
    """)

    infobip.create_person(logger, lead)
    row_lead = sheet.map_lead(lead, headers)
    sheet.write([row_lead])

    return asesor

@app.route('/webhooks', methods=['GET'])
def webhooks_validate():
    print(request.args)
    hub_mode = request.args.get('hub.mode')
    hub_challenge = request.args.get('hub.challenge')
    hub_verify_token = request.args.get('hub.verify_token')
    TOKEN = 'Lautaro123.'

    return hub_challenge

@app.route('/webhooks', methods=['POST'])
def webhook():
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

    message_type = 'text'
    lead = LEAD_SCHEMA.copy()
    lead['message'] = ''
    lead['fuente'] = "Whatsapp"
    value = data['entry'][0]['changes'][0]['value']
    lead['telefono'] = value['contacts'][0]['wa_id']
    lead['nombre']   = value['contacts'][0]['profile']['name']
    #Por algun motivo el tipo request_welcome no se envia bien. Tarda muchisimo y no se envía bien.
    message_type     = value['messages'][0]['type']
    #message_type = 'request_welcome' if value['messages'][0]['text']['body'] == "Asigname con un asesor" else 'text'

    #Cuando el lead ya se habia comunicado con rebora
    if message_type != 'request_welcome': 
        #whatsapp.send_message(logger, lead['telefono'], "Ya te hemos asignado a un asesor, en unos momentos se comunicará contigo")
        return ''

    #Asignar asesor
    asesor = round_robin()
    lead['asesor_name']   = asesor['name']
    lead['asesor_phone']  = asesor['phone']

    #Enviar bienvida de whatsapp a la persona
    whatsapp.send_image(lead['telefono'])
    whatsapp.send_message(lead['telefono'], """Bienvenido a Rebora 👋, ahorra un 15% al comprar tu casa. 🏡

En Rebora la incertidumbre pasó a convertirse en tranquilidad, transparencia y precios justos.""")
    whatsapp.send_message(lead['telefono'], f"""Quiero invitarte a ver el siguiente video donde explicamos en 1 minuto cómo puedes ahorrar un 15% al comprar tu casa con Rebora. 

{asesor['name']} será tu asesor y se pondrá en contacto contigo en un par de minutos! 🙋‍♂️📲

Telefono del asesor:
+{asesor['phone']}""")

    whatsapp.send_video(lead['telefono'])

    whatsapp.send_message(asesor['phone'], f"""Tienes un nuevo lead asignado. 
Nombre: {lead['nombre']}
Telefono: {lead['telefono']}
Comunicate con este lead lo antes posible.
    """)

    infobip.create_person(logger, lead, valid_number=True)
    row_lead = sheet.map_lead(lead, headers)
    print(row_lead)
    sheet.write([row_lead])
    return ''

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
