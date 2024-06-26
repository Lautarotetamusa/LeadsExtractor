from flask import Flask, request, jsonify, Response
from time import gmtime, strftime
from flask_cors import CORS
import threading
import os

from src.logger import Logger
from src.lead import Lead
from src.numbers import parse_number
import src.api as api
import src.jotform as jotform

# Scrapers
from src.inmuebles24.scraper import main as inmuebles24_scraper
from src.lamudi.scraper import main as lamudi_scraper
from src.propiedades_com.scraper import main as propiedades_scraper
from src.casasyterrenos.scraper import main as casasyterrenos_scraper

app = Flask(__name__)
CORS(app)

logger = Logger("Server")

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT is not None, "DATE_FORMAT is not seted"


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
        "fuente": "ivr",
        "fecha_lead": fecha,
        "fecha": fecha
    })

    telefono = parse_number(logger, phone, "MX")
    if not telefono:
        # Si el numero no es mexicano va a llegar con un 52 adelante igual por ejemplo 525493415854220
        phone = phone[2::]  # Removemos el '52'
        telefono = parse_number(logger, "+"+phone)
    lead.telefono = telefono or lead.telefono

    is_new, lead = api.new_communication(logger, lead)

    return lead.asesor


@app.route('/jotform', methods=['POST'])
def generate_pdf():
    lead = Lead()
    data = request.get_json()
    lead_data = data["data"]
    lead.set_args(lead_data)

    pdf_url, err = jotform.new_submission(logger, lead)
    if err is not None:
        logger.error("No se pudo obtener la cotizacion en pdf")
        return Response(err, status=400)

    return Response(pdf_url, status=200)


# Ejecucion de scripts
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


def ejecutar_script(portal, url_or_filters, spin_msg):
    # Reemplaza esta línea con la lógica de ejecución de tus scripts
    print("Ejecutando", f"python main.py {portal} {url_or_filters} {spin_msg}")
    PORTALS[portal]["scraper"](url_or_filters, spin_msg)


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=80, debug=True)
