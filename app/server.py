from flask import Flask, request, jsonify, Response, send_from_directory
from flask_cors import CORS
import threading
import os

# Portals
from src.portal import Portal
from src.inmuebles24.inmuebles24 import Inmuebles24
from src.propiedades_com.propiedades import Propiedades
from src.casasyterrenos.casasyterrenos import CasasYTerrenos
from src.lamudi.lamudi import Lamudi

from src.api import update_publication
from src.property import Location, Property, Ubication
from src.logger import Logger
from src.lead import Lead
import src.jotform as jotform
from src.cotizador import execute_cotizacion
import src.tasks as tasks
from src.numbers import parse_number

# Scrapers
from src.inmuebles24.scraper import Inmuebles24Scraper
from src.lamudi.scraper import LamudiScraper
from src.propiedades_com.scraper import PropiedadesScraper
from src.casasyterrenos.scraper import CasasyterrenosScraper

#cotizador
from src.cotizadorpdf.cotizador import to_pdf

app = Flask(__name__, static_folder='static')
CORS(app)

logger = Logger("Server")
# Configuración del directorio para archivos estáticos
app.config['UPLOAD_FOLDER'] = 'pdfs'

# Asegúrate de que el directorio existe
os.makedirs(app.config['UPLOAD_FOLDER'], exist_ok=True)

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

thread = threading.Thread(target=tasks.init_task_scheduler)
thread.start()

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

@app.route('/generar_cotizacion', methods=['POST'])
def generate_cotization_pdf():
    data = request.get_json()
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
    err = PORTALS[portal].publish(property)
    if err == None:
        update_publication(property.id, portal, "published")
    else:
        logger.error("publication failed: " + err.__str__())
        update_publication(property.id, portal, "failed")

@app.route('/publish/<portal>', methods=['POST'])
def publish_route(portal: str):
    if portal not in PORTALS: 
        return jsonify({"error": f"Portal f{portal} is not valid"}), 400

    data = request.get_json()
    property = Property(**data)
    property.ubication = Ubication(**data["ubication"])
    property.ubication.location = Location(**data["ubication"]["location"])

    thread = threading.Thread(target=publish, args=(portal, property))
    thread.start()

    return jsonify({"success": True, "message": "publishing process has started"}), 201

@app.route('/cotizacion', methods=['POST'])
def cotizacion():
    data = request.get_json()

    asesor = data.get("asesor")
    phone_str = asesor.get("phone", "")
    
    # Lo parseamos sin codigo porque se debe agregar en el campo para que ande con whatsapp
    phone = parse_number(logger, phone_str, None)
    if phone is None:
        return jsonify({'error': f"El numero {phone_str} no es valido"}), 400

    urls = data.get('urls')
    cliente = data.get('cliente')

    if not all([asesor, urls]):
        return jsonify({'error': 'Se requieren campos portal, "asesor", "urls", "cliente"'}), 400

    # Generar un ID único para la tarea
    task_id = tasks.new_task() 

    # Ejecutar el script en segundo plano usando threading
    thread = threading.Thread(target=execute_cotizacion, args=(urls, asesor, cliente, task_id))
    thread.start()

    return jsonify(tasks.get_task(task_id))

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


@app.route('/check-task/<task_id>', methods=['GET'])
def check_task(task_id):
    task = tasks.get_task(task_id)
    if not task:
        return jsonify({"error": "Tarea no encontrada"}), 404

    current_task = task.copy()
    if task["status"] != tasks.Status.in_progress:
        tasks.remove_task(task_id)

    return jsonify(current_task)


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080, debug=True)
