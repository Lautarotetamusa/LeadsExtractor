LEAD_SCHEMA = {
    "fuente": "",
    "fecha_lead": "",
    "id": "",
    "fecha": "",
    "nombre": "",
    "link": "",
    "telefono": "",
    "email": "",
    "propiedad": {
        "id": "",
        "titulo": "",
        "link": "",
        "precio": "",
        "ubicacion": "",
        "tipo": "",
    },
    "busquedas": {
        "zonas": "",
        "tipo": "",
        "presupuesto": "",
        "cantidad_anuncios": "",
        "contactos": "",
        "inicio_busqueda": "",
        "total_area": "",
        "covered_area": "",
        "banios": "",
        "recamaras": "",
    }
}  

if __name__ == "__main__":
    from src.sheets import Gmail, Sheet, set_prop
    from src.logger import Logger
    from src.message import generate_mensage
    import json

    logger = Logger("Gmail first run")
    sheets = Sheet(logger, "mapping.json")
    gmail = Gmail({'email': 'ventas@rebora.com.mx'}, logger)

    headers = sheets.get_dict_headers("A2:W2")
    rows = sheets.get("A3:W2992")

    mappings = []
    for header in headers:
        mappings.append(sheets.mapping[header])

    for row in rows:
        lead = LEAD_SCHEMA.copy()
        for i, col in enumerate(row):
            lead = set_prop(lead, mappings[i], col)
        #print(json.dumps(lead, indent=4))

        if lead['email'] != '':
            message = generate_mensage(lead)
            print(message)
            #gmail.send_message(message, 'subject', lead['email'])
