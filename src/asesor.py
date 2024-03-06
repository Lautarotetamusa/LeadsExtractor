from src.sheets import Sheet
from src.logger import Logger

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
