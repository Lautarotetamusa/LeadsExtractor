import sys

sys.path.append('.')

from src.logger import Logger
import src.api as api
import src.jotform as jotform

if __name__ == "__main__":
    logger = Logger("generate cotizaciones") 

    date = "01-01-2001"
    leads = api.get_communications(logger, date)

    #Cotizacion
    if is_new:
        self.logger.debug("Lead con mt2 construccion, generando cotizacion pdf")
        pdf_url = jotform.new_submission(self.logger, lead) 
        if pdf_url != None:
            lead.cotizacion = pdf_url
        else:
            self.logger.error("No se pudo obtener la cotizacion en pdf")
