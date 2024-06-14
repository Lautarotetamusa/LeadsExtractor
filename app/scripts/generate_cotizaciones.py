import sys
from dotenv import load_dotenv
sys.path.append('.')
load_dotenv()

from src.logger import Logger
import src.api as api
import src.jotform as jotform

if __name__ == "__main__":
    logger = Logger("generate cotizaciones") 

    date = "01-01-2001"
    leads = api.get_communications(logger, date, True)
    logger.debug(f"generando cotizaciones para {len(leads)} leads")

    for lead in leads:
        logger.debug("Lead con mt2 construccion, generando cotizacion pdf")
        pdf_url = jotform.new_submission(logger, lead) 
        if pdf_url != None:
            lead.cotizacion = pdf_url
        else:
            logger.error("No se pudo obtener la cotizacion en pdf")

        api.new_communication(logger, lead)
