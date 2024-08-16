from flask import Flask, request, jsonify, Response, send_from_directory
from time import gmtime, strftime
from flask_cors import CORS
import threading
import os
import uuid

from src.logger import Logger
from src.lead import Lead
from src.numbers import parse_number
import src.api as api
import src.jotform as jotform

# Scrapers
from src.inmuebles24.scraper import main as inmuebles24_scraper, cotizacion as inmuebles24_cotizacion
from src.lamudi.scraper import LamudiScraper
from src.propiedades_com.scraper import PropiedadesScraper
from src.casasyterrenos.scraper import CasasyterrenosScraper

app = Flask(__name__)
CORS(app)

logger = Logger("Server")

tasks = {}
# Configuración del directorio para archivos estáticos
app.config['UPLOAD_FOLDER'] = 'pdfs'

# Asegúrate de que el directorio existe
os.makedirs(app.config['UPLOAD_FOLDER'], exist_ok=True)

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

    return lead.asesor, 200


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

@app.route('/download/<filename>', methods=['GET'])
def download_file(filename):
    # Verificar si el archivo existe en el directorio de carga
    file_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)
    if os.path.exists(file_path):
        return send_from_directory(app.config['UPLOAD_FOLDER'], filename)
    return jsonify({'error': 'Archivo no encontrado'}), 404

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

@app.route('/cotizacion', methods=['POST'])
def cotizacion():
    data = request.get_json()

    portal = data.get('portal')
    posts = data.get('posts')
    asesor = data.get('asesor')

    if portal not in SCRAPERS:
        return jsonify({'error': f"El portal {portal} no existe"}), 404

    if not all([portal, asesor, posts]):
        return jsonify({'error': 'Se requieren campos portal, message y url_or_filters'}), 400

    # Generar un ID único para la tarea
    task_id = str(uuid.uuid4())
    tasks[task_id] = 'in_progress'
    print(tasks)

    # Ejecutar el script en segundo plano usando threading
    thread = threading.Thread(target=ejecutar_cotizacion, args=(asesor, posts, task_id))
    thread.start()

    return jsonify({'taskId': task_id})

@app.route('/check-cotizacion-status/<task_id>', methods=['GET'])
def check_cotizacion_status(task_id):
    print(tasks)
    if task_id not in tasks:
        return jsonify({'error': 'Tarea no encontrada'}), 404

    status = tasks.get(task_id)
    if not status:
        return jsonify({'error': 'Tarea no encontrada'}), 404
    return jsonify({'status': status, 'pdf_url': 'download/result.pdf'})

def ejecutar_cotizacion(asesor, posts, task_id):
    try:
        inmuebles24_cotizacion(asesor, posts)
        tasks[task_id] = 'completed'
    except Exception as e:
        tasks[task_id] = 'error'
        print(f"Error generando la cotización: {str(e)}")

def ejecutar_script(portal: str, params: str | dict, spin_msg: str):
    assert portal in SCRAPERS, f"El portal {portal} no existe"

    # Reemplaza esta línea con la lógica de ejecución de tus scripts
    print("Ejecutando", f"python main.py {portal} {params} {spin_msg}")
    scraper = SCRAPERS[portal]()
    scraper.main(spin_msg, params)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=80, debug=True)
