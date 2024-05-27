import sys
from dotenv import load_dotenv
sys.path.append('.')
load_dotenv()

from src.lead_actions import new_lead_action
import src.api as api
from src.whatsapp import Whatsapp
import src.jotform as jotform
from src.message import format_msg

from src.logger import Logger

if __name__ == "__main__":
    logger = Logger("resend leads")
    wpp = Whatsapp(logger)

    date = "05-06-2024"
    leads = api.get_communications(logger, date)
    print(len(leads))

    for lead in leads:
        is_new = lead.__dict__["is_new"]
        if is_new:
            logger.debug("Lead con mt2 construccion, generando cotizacion pdf")
            pdf_url = jotform.new_submission(logger, lead) 
            if pdf_url != None:
                lead.cotizacion = pdf_url
            else:
                logger.error("No se pudo obtener la cotizacion en pdf")

        lead.print()

        if is_new: #Lead nuevo
            portal_msg = new_lead_action(wpp, lead)
        #else: #Lead existente
        #    portal_msg = format_msg(lead, response_msg)

            wpp.send_response(lead.telefono, lead.asesor)
