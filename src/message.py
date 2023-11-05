import spintax
import re

msg = """Hola [nombre], {¿como estas?|¿Cómo te va?}, {ya estamos revisando tu solicitud|sobre tu solicitud}, la propiedad [titulo] ubicada en [ubicacion] si está disponible, {te comparto el whatsapp de nuestro asesor asignado|con gusto te envio el whatsapp de nuestro asesor}, {Brenda|Diego|Rafa}, con mucho gusto te comparte la información  —-> {3319466986|3345898765|335698786}"""

def get_fields(msg):
    #Match all the fields names
    #ej: fields([title]kdlsmieid["nombre"]dskaodw["edad"]]) => ["title", "nombre", "edad"]
    pat = r'(?<=\[).+?(?=\])'
    return re.findall(pat, msg)

def generate_mensage(lead):
    spin = spintax.spin(msg)

    format = spin.replace("[","{").replace("]","}")

    print("format: ", format)

    return format.format(**{
        "nombre": lead["nombre"],
        "ubicacion": lead["propiedad"]["municipio"],
        "titulo": lead["propiedad"]["titulo"]
    })