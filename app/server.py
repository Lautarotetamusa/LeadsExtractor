from flask import Flask, request, jsonify, Response, send_from_directory
from flask_cors import CORS
import threading
import os
import uuid

from src.logger import Logger
from src.lead import Lead
import src.jotform as jotform

# Scrapers
from src.inmuebles24.scraper import main as inmuebles24_scraper
import src.inmuebles24.scraper as inmuebles24
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

# Ejecucion de scripts
SCRAPERS = {
    "casasyterrenos": CasasyterrenosScraper,
    "propiedades": PropiedadesScraper,
    "inmuebles24": inmuebles24_scraper,
    "lamudi": LamudiScraper
}

DATE_FORMAT = os.getenv("DATE_FORMAT")
assert DATE_FORMAT is not None, "DATE_FORMAT is not seted"


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

@app.route('/download/<filename>', methods=['GET'])
def download_file(filename):
    # Verificar si el archivo existe en el directorio de carga
    file_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)
    if os.path.exists(file_path):
        return send_from_directory(app.config['UPLOAD_FOLDER'], filename)
    return jsonify({'error': 'Archivo no encontrado'}), 404


@app.route('/cotizacion', methods=['POST'])
def cotizacion():
    data = request.get_json()

    portal = data.get('portal')
    posts  = data.get('posts')
    asesor = data.get('asesor')
    cliente = data.get('cliente')

    if portal not in SCRAPERS:
        return jsonify({'error': f"El portal {portal} no existe"}), 404

    if not all([portal, asesor, posts]):
        return jsonify({'error': 'Se requieren campos portal, "asesor", "posts", "cliente"'}), 400

    # Generar un ID único para la tarea
    task_id = str(uuid.uuid4())
    tasks[task_id] = {
        "status": "in_progress",
        "task_id": task_id
    }

    # Ejecutar el script en segundo plano usando threading
    thread = threading.Thread(target=execute_cotizacion, args=(asesor, posts, cliente, task_id))
    thread.start()

    return jsonify(tasks[task_id])

@app.route('/cotizacion_urls', methods=['POST'])
def cotizacion_with_urls():
    data = request.get_json()

    portal = data.get('portal')
    asesor = data.get('asesor')
    urls = data.get('urls')
    cliente = data.get('cliente')

    if portal not in SCRAPERS:
        return jsonify({'error': f"El portal {portal} no existe"}), 404

    if not all([portal, asesor, urls]):
        return jsonify({'error': 'Se requieren campos portal, "asesor", "urls", "cliente"'}), 400

    # Generar un ID único para la tarea
    task_id = str(uuid.uuid4())
    tasks[task_id] = {
        "status": "in_progress",
        "task_id": task_id
    }

    # Ejecutar el script en segundo plano usando threading
    thread = threading.Thread(target=execute_cotizacion_urls, args=(asesor, urls, cliente, task_id))
    thread.start()

    return jsonify(tasks[task_id])


@app.route('/check-task/<task_id>', methods=['GET'])
def check_task(task_id):
    if task_id not in tasks:
        return jsonify({"error": "Tarea no encontrada"}), 404

    task = tasks.get(task_id)
    if not task:
        return jsonify({"error": "Tarea no encontrada"}), 404

    return jsonify(task)


def execute_cotizacion(asesor, res, cliente, task_id):
    try:
        posts = inmuebles24.posts_from_list(res)
        pdf_file_name = inmuebles24.cotizacion(asesor, cliente, posts)
        tasks[task_id]["status"] = "completed"
        tasks[task_id]["pdf_file_name"] = pdf_file_name
    except Exception as e:
        tasks[task_id]["status"] = "error"
        print(e)
        print(f"Error generando la cotización: {str(e)}")


def execute_cotizacion_urls(asesor, urls, cliente, task_id):
    try:
        posts = []
        for url in urls:
            posts.append(inmuebles24.get_post_data(url))
        pdf_file_name = inmuebles24.cotizacion(asesor, cliente, posts)
        tasks[task_id]["status"] = "completed"
        tasks[task_id]["pdf_file_name"] = pdf_file_name
    except Exception as e:
        tasks[task_id]["status"] = "error"
        print(f"Error generando la cotización: {str(e)}")


def execute_scraper(portal: str, params: str | dict, spin_msg: str):
    assert portal in SCRAPERS, f"El portal {portal} no existe"

    # Reemplaza esta línea con la lógica de ejecución de tus scripts
    print("Ejecutando", f"python main.py {portal} {params} {spin_msg}")
    scraper = SCRAPERS[portal]()
    scraper.main(spin_msg, params)


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=80, debug=True)
