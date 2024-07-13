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
from src.lamudi.scraper import LamudiScraper
from src.propiedades_com.scraper import PropiedadesScraper
from src.casasyterrenos.scraper import CasasyterrenosScraper

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

    _, lead = api.new_communication(logger, lead)

    return {}, 200


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
SCRAPERS = {
    "casasyterrenos": CasasyterrenosScraper,
    "propiedades": PropiedadesScraper,
    "inmuebles24": inmuebles24_scraper,
    "lamudi": LamudiScraper
}

@app.route('/execute', methods=['POST'])
def ejecutar_script_route():
    data = request.get_json()

    portal = data.get('portal')
    spin_msg = data.get('message')
    params = data.get('url_or_filters')

    if portal not in SCRAPERS:
        return jsonify({'error': f"El portal {portal} no existe"}), 404

    if not all([portal, spin_msg, params]):
        return jsonify({'error': 'Se requieren campos portal, message y url_or_filters'}), 400

    # Ejecutar el script en segundo plano usando threading
    thread = threading.Thread(target=ejecutar_script, args=(portal, params, spin_msg))
    thread.start()

    return "Proceso en segundo plano iniciado."


def ejecutar_script(portal: str, params: str | dict, spin_msg: str):
    assert portal in SCRAPERS, f"El portal {portal} no existe"

    # Reemplaza esta línea con la lógica de ejecución de tus scripts
    print("Ejecutando", f"python main.py {portal} {params} {spin_msg}")
    scraper = SCRAPERS[portal]()
    scraper.main(spin_msg, params)


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=80, debug=True)
