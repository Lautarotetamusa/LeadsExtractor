from src.lead import Lead
import spintax
import re

def get_fields(msg):
    #Match all the fields names
    #ej: fields([title]kdlsmieid["nombre"]dskaodw["edad"]]) => ["title", "nombre", "edad"]
    pat = r'(?<=\[).+?(?=\])'
    return re.findall(pat, msg)

def generate_mensage(lead: Lead, spin_msg: str | None=None):
    if not spin_msg:
        spin_msg = """Hola que tal [nombre] ¿cómo estás?, muy buenas tardes ¡Bienvenido a Rebora!
        Somos una empresa especializada en la venta de residencias 100% personalizadas con un
        diseño único, moderno y con enorme calidad en cada material.
        La propiedad [titulo] ubicada en [ubicacion] si está disponible, te comparto el whatsapp de Brenda tu asesor asignado, con mucho gusto te comparte la información  –-> 33 1342 0733
        Beneficios de una Casa Rebora:
        - SEGURIDAD, CONFORT Y PLACER en cada espacio de TÚ casa
        - Casas 100% inteligentes y 100% personalizadas
        - Plusvalía del 20 - 25% desde el primer año."""
    spin = spintax.spin(spin_msg)
    format = spin.replace("[","{").replace("]","}")

    ubicacion = lead.propiedad["ubicacion"]
    #if "municipio" in lead.propiedad"]:
    #    ubicacion = lead.propiedad"]["municipio"]
    return format.format(**{
        "nombre": lead.nombre,
        "ubicacion": ubicacion,
        "titulo": lead.propiedad["titulo"]
    })

def generate_post_message(post: dict, spin_msg: str):
    # msg = """[nombre], {¿como estas?|¿Cómo te va?}, {Estoy viendo tu propiedad y me interesa saber si sigue disponible|sigue disponible tu propiedad?}, es la propiedad [titulo] ubicada en [ubicacion] , {te comparto mi whatsapp | te paso mi num y whatsapp}, {Julian|Roberto|Raul}, {Espero la información Gracias|Espero respuesta, queda pendiente, gracias}  —-> {3345879056|3332458909|3323430985}"""
    spin = spintax.spin(spin_msg)
    format = spin.replace("[","{").replace("]","}")

    return format.format(**{
        "nombre": post["publisher"]["name"],
        "ubicacion": post["location"]["city"],
        "titulo": post["title"]
    })
