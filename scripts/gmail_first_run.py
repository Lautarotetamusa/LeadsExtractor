import json

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
    from email.mime.application import MIMEApplication
    from src.sheets import Gmail, Sheet, set_prop
    from src.logger import Logger
    from src.message import generate_mensage
    import time
    import sys

    logger = Logger("Gmail first run")
    sheets = Sheet(logger, "mapping.json")
    gmail = Gmail({'email': 'ventas@rebora.com.mx'}, logger)

    headers = sheets.get_dict_headers("A2:W2")
    rows = sheets.get("A3:W2992")

    with open('messages/gmail.html', 'r') as f:
        gmail_spin = f.read()
    with open('messages/gmail_subject.html', 'r') as f:
        gmail_subject = f.read()
    with open('sended.json', 'r') as f:
        sended = json.load(f)
    with open('last_sended.txt', 'r') as f:
        bottom_limit = int(f.read())
    top_limit = int(sys.argv[1])
    #Agregamos un top limit para poder controlar cuantos faltan en el día, por ejemplo si ya corrimos 300 pondriamos 900

    # Adjuntar archivo PDF
    with open('messages/attachment.pdf', 'rb') as pdf_file:
        attachment = MIMEApplication(pdf_file.read(), _subtype="pdf")
        attachment.add_header('Content-Disposition', 'attachment', 
              filename='Bienvenido a Rebora! Seguridad, Confort y Placer - Casas de gran diseño y alta calidad.pdf'
        )

    mappings = []
    for header in headers:
        mappings.append(sheets.mapping[header])

    print("bottom_limit:", bottom_limit)
    print("top_limit:", top_limit)
    enviados = 0
    fila = bottom_limit
    for row in rows[bottom_limit:]:
        lead = LEAD_SCHEMA.copy()
        for i, col in enumerate(row):
            lead = set_prop(lead, mappings[i], col)
        fila += 1

        if lead['email'] != '':
            if lead['email'] in sended:
                print(f"Ya le enviamos un email a {lead['email']}, continuamos")
                continue

            if lead["propiedad"]["ubicacion"] == "":
                lead["propiedad"]["ubicacion"] = "que consultaste"
            else:
                lead["propiedad"]["ubicacion"] = "ubicada en " + lead["propiedad"]["ubicacion"]

            message = generate_mensage(lead, gmail_spin)
            subject = generate_mensage(lead, gmail_subject)
            gmail.send_message(message, subject, lead["email"], attachment)

            sended[lead['email']] = True
            with open('sended.json', 'w') as f:
                json.dump(sended, f)
                
            enviados += 1
            print("se enviaron: ", enviados)
            with open('last_sended.txt', 'w') as f:
                f.write(str(fila))

            time.sleep(30)

        #Para no pasarnos del limite diaria de gmail
        if enviados >= top_limit:
            break
