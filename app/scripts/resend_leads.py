import sys
from dotenv import load_dotenv
sys.path.append('.')
load_dotenv()

from src.lead_actions import new_lead_action
import src.api as api
from src.whatsapp import Whatsapp
import src.jotform as jotform

from src.logger import Logger

if __name__ == "__main__":
    logger = Logger("resend leads")
    wpp = Whatsapp(logger)

    date = "05-23-2024"
    leads = api.get_communications(logger, date)
    print(len(leads))

    phones = {}
    count = 0

    for lead in leads:
        if lead.telefono in phones: #Duplicado
            count += 1
            logger.debug("Lead con mt2 construccion, generando cotizacion pdf")
            pdf_url = jotform.new_submission(logger, lead) 
            if pdf_url != None:
                lead.cotizacion = pdf_url
            else:
                logger.error("No se pudo obtener la cotizacion en pdf")
            
            portal_msg = new_lead_action(wpp, lead)
            
            wpp.send_response(lead.telefono, lead.asesor)

        phones[lead.telefono] = True
        #else: #Lead existente
        #    portal_msg = format_msg(lead, response_msg)

    print(count)
