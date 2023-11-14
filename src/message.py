import spintax
import re

#msg = """Hola [nombre], {¿como estas?|¿Cómo te va?}, {ya estamos revisando tu solicitud|sobre tu solicitud}, la propiedad [titulo] ubicada en [ubicacion] si está disponible, {te comparto el whatsapp de nuestro asesor asignado|con gusto te envio el whatsapp de nuestro asesor}, Brenda, con mucho gusto te comparte la información  —-> 3345898765"""
msg = """Hola que tal [nombre] ¿cómo estás?, muy buenas tardes ¡Bienvenido a Rebora!
Somos una empresa especializada en la venta de residencias 100% personalizadas con un
diseño único, moderno y con enorme calidad en cada material.
La propiedad [titulo] ubicada en [ubicacion] si está disponible, te comparto el whatsapp de Brenda tu asesor asignado, con mucho gusto te comparte la información  –-> 33 1342 0733
Beneficios de una Casa Rebora:
- SEGURIDAD, CONFORT Y PLACER en cada espacio de TÚ casa
- Casas 100% inteligentes y 100% personalizadas
- Plusvalía del 20 - 25% desde el primer año."""

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