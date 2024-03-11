import src.infobip as infobip
from src.logger import Logger
from src.sheets import Sheet
from src.lead import Lead

logger = Logger("Asesor")
sheet = Sheet(logger, "mapping.json")

ASESOR_i = 0
def next_asesor():
    global ASESOR_i
    rows = sheet.get('Asesores!A2:C25')
    activos = [row for row in rows if row[2] == "Activo"]
    print("Activos: ", activos)
    ASESOR_i += 1
    ASESOR_i %= len(activos)
    
    asesor = {
        "name":  activos[ASESOR_i][0],
        "phone": activos[ASESOR_i][1]
    }
    logger.debug("Asesor asignado: "+str(asesor['name']))
    return asesor

#Si la persona no existe en infobip hacemos round robin para otorgarle un asesor
#Si ya existe devolvemos el asesor que ya tiene
def assign_asesor(lead: Lead) -> tuple[bool, Lead]:
    telefono = lead.telefono.replace('+', '') #Removemos el + para poder buscarlo bien en infobip
    infobip_lead = infobip.search_person(logger, telefono)
    
    if infobip_lead == None:
        logger.debug(f"Un nuevo lead se comunico via: {lead.fuente}")
        asesor = next_asesor()
        lead.set_asesor(asesor)
        is_new = True
    else:
        logger.debug(f"Un lead existente se volvio a comunicar via: {lead.fuente}")
        lead.set_args(infobip_lead.get_no_empty_values()) #Agregamos los datos del lead desde infobip
        lead.set_asesor(infobip_lead.asesor) #Agregamos el asesor
        is_new = False

    return is_new, lead
