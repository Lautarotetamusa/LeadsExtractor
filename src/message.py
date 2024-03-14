from src.lead import Lead
import spintax

def format_msg(lead: Lead, spin_msg: str) -> str:
    spin = spintax.spin(spin_msg)
    format = spin.replace("[","{").replace("]","}")

    return format.format(**{
        "nombre": lead.nombre,
        "ubicacion": lead.propiedad["ubicacion"],
        "titulo": lead.propiedad["titulo"],
        "asesor_name": lead.asesor["name"],
        "asesor_phone": lead.asesor["phone"]
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
