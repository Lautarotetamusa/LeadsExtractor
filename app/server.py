from flask import Flask, request, jsonify, Response, send_from_directory
from time import gmtime, strftime
from flask_cors import CORS
import threading
import os
import json
import phonenumbers

# Portals
from src.lead import Lead
from src.portal import Portal
from src.inmuebles24.inmuebles24 import Inmuebles24
from src.propiedades_com.propiedades import Propiedades
from src.casasyterrenos.casasyterrenos import CasasYTerrenos
from src.lamudi.lamudi import Lamudi

from src.api import new_communication, update_publication
from src.property import Internal, Location, PlanType, Property, Ubication
from src.logger import Logger
import src.tasks as tasks

# Scrapers
from src.inmuebles24.scraper import Inmuebles24Scraper
from src.lamudi.scraper import LamudiScraper
from src.propiedades_com.scraper import PropiedadesScraper
from src.casasyterrenos.scraper import CasasyterrenosScraper

# cotizador
from src.cotizadorpdf.cotizador import to_pdf

app = Flask(__name__, static_folder='static')
CORS(app)

logger = Logger("Server")
# Configuración del directorio para archivos estáticos
app.config['UPLOAD_FOLDER'] = 'pdfs'

# Asegúrate de que el directorio existe
os.makedirs(app.config['UPLOAD_FOLDER'], exist_ok=True)

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT is not None, "DATE_FORMAT is not seted"

# Ejecucion de scripts
SCRAPERS = {
    "casasyterrenos": CasasyterrenosScraper,
    "propiedades": PropiedadesScraper,
    "inmuebles24": Inmuebles24Scraper,
    "lamudi": LamudiScraper
}

PORTALS: dict[str, Portal] = {
    "casasyterrenos": CasasYTerrenos(),
    "propiedades": Propiedades(),
    "inmuebles24": Inmuebles24(),
    "lamudi": Lamudi()
}

with open("internal_zones.json", "r") as f:
    internal_zones: dict[str, dict[str, Internal]] = json.load(f)

thread = threading.Thread(target=tasks.init_task_scheduler)
thread.start()


@app.route('/infobip-ivr', methods=['GET'])
def recive_ivr_call():
    args = request.args.to_dict()
    phone = args.get("msidsn", None)
    fuente = args.get("fuente", None)

    if not phone or not fuente:
        logger.error("Falta algun campo en la peticion")
        return {
            "error": "Falta algun campo en la peticion desde infobip (msidsn, fuente, name)"
        }, 400
    logger.debug("msidsn:", phone)

    lead = Lead()
    fecha = strftime(DATE_FORMAT, gmtime())
    lead.set_args({
        "nombre": phone,
        "fuente": "ivr",
        "fecha_lead": fecha,
        "fecha": fecha
    })

    telefono = parse_number(logger, phone, "MX")
    # Numero mexicano: 5213319466986
    # Numero no mexicano: +525493415854220 
    if not telefono or len(telefono) >= 15:
        if telefono and len(telefono) >= 15:
            logger.warning("the number recived its not mexican", telefono)
        # Si el numero no es mexicano va a llegar con un 52 adelante igual por ejemplo 525493415854220
        phone = phone[2::]  # Removemos el '52'
        telefono = parse_number(logger, "+"+phone)
    lead.telefono = telefono or lead.telefono

    _, lead = new_communication(logger, lead)
    if lead is None:
        return {}, 400

    print(lead.asesor)
    return lead.asesor, 200


@app.route('/generar_cotizacion', methods=['POST'])
def generate_cotization_pdf():
    data = request.get_json()
    print(json.dumps(data, indent=4))
    ret = to_pdf(data)
    if (ret == "error"):
        return Response(ret, status=400)
    return Response(ret, status=200)


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
    thread = threading.Thread(target=execute_scraper, args=(portal, params, spin_msg))
    thread.start()

    return "Proceso en segundo plano iniciado."


def publish(portal: str, property: Property):
    logger.debug("plan: " + str(property.plan))
    # json.dumps(property.__dict__, indent=4)
    err, publication_id = PORTALS[portal].publish(property)
    if err is not None or publication_id is None:
        update_publication(property.id, portal, "failed")
        logger.error("publication failed: " + err.__str__())
        return

    if property.plan != PlanType.SIMPLE.value:
        err = PORTALS[portal].highlight(publication_id, property.plan)
        if err is not None:
            update_publication(property.id, portal, "failed")
            logger.error("publication failed: " + err.__str__())
            return

    update_publication(property.id, portal, "published", publication_id)


def get_internal(portal: str, zone: str) -> Internal | None:
    return internal_zones.get(portal, {}).get(zone)


@app.route('/publish/<portal>', methods=['POST'])
def publish_route(portal: str):
    if portal not in PORTALS:
        logger.warning(f"Portal f{portal} is not valid")
        return jsonify({"error": f"Portal f{portal} is not valid"}), 400

    data = request.get_json()
    internal = get_internal(portal, data.get("zone"))
    if internal is None:
        logger.warning(f"zone {data.get('zone')} does not have internal for {portal}")
        return jsonify({"error": f"zone {data.get('zone')} does not have internal for {portal}"}), 400
    data["internal"] = internal

    property = Property(**data)
    property.ubication = Ubication(**data["ubication"])
    property.ubication.location = Location(**data["ubication"]["location"])

    thread = threading.Thread(target=publish, args=(portal, property))
    thread.start()

    return jsonify({
        "success": True,
        "message": "publishing process has started"
    }), 201


def execute_scraper(portal: str, params: str | dict, spin_msg: str):
    assert portal in SCRAPERS, f"El portal {portal} no existe"

    # Reemplaza esta línea con la lógica de ejecución de tus scripts
    print("Ejecutando", f"python main.py {portal} {params} {spin_msg}")
    scraper = SCRAPERS[portal]()
    scraper.main(spin_msg, params)


@app.route('/download/<filename>', methods=['GET'])
def download_file(filename):
    # Verificar si el archivo existe en el directorio de carga
    file_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)
    if os.path.exists(file_path):
        return send_from_directory(app.config['UPLOAD_FOLDER'], filename)
    return jsonify({'error': 'Archivo no encontrado'}), 404

@app.route('/download/static/<filename>', methods=['GET'])
def download_static_file(filename):
    # Verificar si el archivo existe en el directorio de carga
    file_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)

    static_dir_path = os.path.join(app.config['UPLOAD_FOLDER'], "static")

    if os.path.exists(file_path):
        return send_from_directory(static_dir_path, filename)
    return jsonify({'error': 'Archivo statico no encontrado'}), 404


@app.route('/check-task/<task_id>', methods=['GET'])
def check_task(task_id):
    task = tasks.get_task(task_id)
    if not task:
        return jsonify({"error": "Tarea no encontrada"}), 404

    current_task = task.copy()
    if task["status"] != tasks.Status.in_progress:
        tasks.remove_task(task_id)

    return jsonify(current_task)


def parse_number(logger: Logger, phone: str, code=None) -> str | None:
    logger.debug(f"Parseando numero: {phone}. code: {code}")
    try:
        number = phonenumbers.parse(phone, code)
        parsed_number = phonenumbers.format_number(number, phonenumbers.PhoneNumberFormat.E164)
        logger.success("Numero obtenido: " + parsed_number)
        return parsed_number
    except phonenumbers.NumberParseException:
        logger.error("Error parseando el numero: " + phone)
        return None


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
