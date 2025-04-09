import os
import webbrowser

from src.cotizadorpdf.cotizador import to_pdf

def cotizador(args: list[str]):
    timestamp = to_pdf(test_data)

    if (timestamp == "error"):
        print("error creating the cotizacion")
        exit(1)

    base_path = os.getcwd()
    html_path = os.path.join(base_path, "pdfs", f"cotizacion{timestamp}.html")
    webbrowser.open(f"file://{html_path}")

test_data = {
    "elaborado_por": {
        "nombre": "Diego Torres",
        "telefono": "341 946-6986",
        "mail": "diego.torres@rebora.com.mx",
        "porcentaje_administracion": 21,
        "is_valor_permisos": True,
        "valor_terreno": 10000000,
        "area_terreno": 500
    },
    "datos": {
        "nombre": "Juan Alonso"
    },
    "pagos": {
        "inicial": 1500000,
        "porcentaje_inicio_obra": 25,
        "meses": 18,
        "tipo": "Premium"
    },
    "areas_interiores": {
        "banos": 3,
        "cuartos": 4,
        "sotano": 50,
        "planta_baja": 200,
        "planta_alta": 200,
        "roof": 50
    },
    "areas_exteriores": {
        "rampa": 50,
        "jardin": 20,
        "alberca": 100,
        "muro_perimetral": 100
    },
    "valor_exteriores": {
        "rampa": 7000,
        "jardin": 28500,
        "alberca": 3500,
        "muro_perimetral": 1500
    },
    "valor_permisos": {
        "licencia": 300,
        "gestorias": 38,
        "topografia": 9500,
        "mecanica": 4,
        "calculo": 59
    },
}
