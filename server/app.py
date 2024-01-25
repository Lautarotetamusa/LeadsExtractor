from flask import Flask, request, jsonify, render_template, send_from_directory, session
from flask_cors import CORS
import threading

app = Flask(__name__)
CORS(app)
import threading

#Scrapers
from src.inmuebles24.scraper import main as inmuebles24_scraper
from src.lamudi.scraper import main as lamudi_scraper
from src.propiedades_com.scraper import main as propiedades_scraper
from src.casas_y_terrenos.scraper import main as casasyterrenos_scraper

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
@app.route('/asesor', methods=['GET'])
def get_asesor():
    global ASESOR_i
    print(request)
    ASESORES = [
        {
            "id": 0,
            "name": "Brenda Diaz",
            "phone": "+523313420733"
        },
        {
            "id": 1,
            "name": "Juan Sanchez",
            "phone": "+523317186543"
        },
        {
            "id": 2,
            "name": "Aldo Salcido",
            "phone": "+523322563353"
        } 
    ]
    ASESOR_i += 1
    ASESOR_i %= len(ASESORES)
    print("Asesor asignado: ", ASESORES[ASESOR_i])
    return ASESORES[ASESOR_i]

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
