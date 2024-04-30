import requests
import os
import urllib.parse

from src.logger import Logger
from src.lead import Lead

API_KEY = os.getenv("JOTFORM_API_KEY")
FORM_ID = os.getenv("JOTFORM_FORM_ID")
assert API_KEY != "" and API_KEY != None, "JOTFORM_API_KEY is not in enviroment"
assert FORM_ID != "" and FORM_ID != None, "JOTFORM_FORM_ID is not in enviroment"

api_url = "https://api.jotform.com"

def new_submission(logger: Logger, lead: Lead) -> None | dict:
    logger.fuente += " > Jotform API"
    url = urllib.parse.urljoin(api_url, f"form/{FORM_ID}/submissions")
    payload = {
            'apiKey': API_KEY,
            "submission[124]": lead.nombre, #q124_nombreCliente	"test"
            "submission[4]": lead.email, #q4_emailCliente	"test@gmail.com"
            "submission[57]": lead.asesor['name'], #q57_escribaUna	"test"
            #"submission[123]": lead.asesor['telefono'], #lead.q123_email123	"test@gmail.com"
            "submission[5]": lead.asesor['phone'], #q5_numeroDe[full]	"00+0000-0000"
            #"submission[116]": "22,500", #Valor tipo casa
            "submission[117]": "2,500,000", #q117_pagoInicial	"1,500,000"
            "submission[119]": "2,500,000", #valor pago inicial q117_pagoInicial	"1,500,000"
            "submission[118]": "25%", #q118_pagoInicial118	"20%"
            "submission[122]": "16 meses", #q122_aCuantos	"18+meses"
            "submission[9]": "Premium", #q9_escribaUna9	"Premium"
            "submission[15]": lead.busquedas['recamaras'], #q15_cuantosCuartos	"2"
            "submission[16]": lead.busquedas['banios'], #q16_cuantosBanos56	"3"
            "submission[76]": "0",# TAMAÑO DEL SOTANOq76_cuantosBanos	"5"
            "submission[78]": lead.busquedas['covered_area'] * 0.45, #Tamaño planta baja q78_tamanoDe	"1"
            "submission[79]": lead.busquedas['covered_area'] * 0.45, #Tamaño planta alta q79_tamanoDe79	"7"
            "submission[77]": lead.busquedas['covered_area'] * 0.10, # Tamaño del roof garden q77_tamanoDel	"8"
            "submission[84]": "0", # Tamaño rampa de estacionamiento q84_tamanoDe84	"0"
            "submission[85]": "30", # Tamaño jardin exterior q85_tamanoDe85	"0"
            "submission[86]": "0", # Tamaño de alberca q86_tamanoDe86	"0"
            "submission[87]": "50", # Tamaño muro perimetral q87_tamanoDe87	"0"
            "submission[88]": "0", # Valor sotano q88_valorSotano	"112,500"
    }

    """
    q57_escribaUna	"test"
q80_email80	"test@gmail.com"
q81_numeroDe81	"3415854220"
q14_escribaUna14	"No"
q9_escribaUna9	"Premium"
q20_tamanoDe20	"300"
q75_cuantosCuartos75	"2"
q76_cuantosBanos	"2"
q17_tienesTerreno	"No"
q18_quieresSotano	"Sí"
q43_escribaUna43	"6,750,000"
q48_costoCompetencia	"7,425,000"
q44_costoArquitectos	"7,762,500"
q46_primerPago	"1,687,500"
q47_costoMensualidades	"316,406"
q49_pagoInicial	"1,856,250"
q50_pagoInicial50	"7,762,500"
q51_roiRebora	"1,147,500"
q52_roiCompetencia	"742,500"
q53_roiArquitectos	"776,250"
    """

    res = requests.post(url, data=payload)

    if res.status_code == 200:
        logger.success('Solicitud exitosa')
        data = res.json()
        print(data)
        return data
    else:
        logger.error('Error en la solicitud:' + res.text)
        return None

def get_questions_form(logger: Logger, form_id: str = FORM_ID):
    logger.fuente += " > Jotform API"
    url = urllib.parse.urljoin(api_url, f"form/{form_id}/questions")
    payload = {
        'apiKey': API_KEY
    }

    res = requests.get(url, params=payload)

    if res.ok:
        logger.success('Solicitud exitosa')
        data = res.json()
        print(data)
        return data
    else:
        logger.error('Error en la solicitud:' + res.text)
        return None
