"""
    {
	"formID": "240546574547868",
	"jsExecutionTracker": "build-date-1714562511750=>init-started:1714551684861=>validator-called:1714551684884=>validator-mounted-true:1714551684884=>init-complete:1714551684889=>interval-complete:1714551705922=>onsubmit-fired:1714551749636=>submit-validation-passed:1714551749643",
	"submitSource": "form",
	"buildDate": "1714562511750",
	"q124_nombreCliente": "test",
	"q4_emailCliente": "test@gmail.com",
	"q57_escribaUna": "asesor+test",
	"q123_email123": "asesor@gmail.com",
	"q5_numeroDe[full]": "34+1999-9999",
	"q117_pagoInicial": "2,000,000",
	"q118_pagoInicial118": "25%",
	"q122_aCuantos": "16+meses",
	"q9_escribaUna9": "Premium",
	"q15_cuantosCuartos": "2",
	"q16_cuantosBanos56": "2",
	"q76_cuantosBanos": "0",
	"q78_tamanoDe": "200",
	"q79_tamanoDe79": "200",
	"q77_tamanoDel": "44",
	"q84_tamanoDe84": "0",
	"q85_tamanoDe85": "30",
	"q86_tamanoDe86": "0",
	"q87_tamanoDe87": "50",
	"q88_valorSotano": "0",
	"q92_valorPlanta": "4,500,000",
	"q93_precioPlanta": "4,500,000",
	"q94_valorRoof": "990,000",
	"q95_valorTotal95": "9,990,000",
	"q96_valorRampaestacionamiento": "0",
	"q97_valorJardin": "45,000",
	"q98_valorAlberca": "0",
	"q99_valorMuro": "350,000",
	"q100_valorTotal": "395,000",
	"q101_metrosTotales": "444",
	"q102_valorLicencia": "133,200",
	"q103_valorGestoria": "16,872",
	"q104_valorTopografia": "9,500",
	"q105_valorMecanica": "23,200",
	"q106_escribaUna106": "26,196",
	"q107_valorTotal107": "208,968",
	"q108_valorMonto": "10,593,968",
	"q109_valorPorcentaje": "2,148,492",
	"q110_pagosMensuales": "402,842",
	"q111_plusvaliaRebora": "1,800,975",
	"q112_costoTotal": "11,653,365",
	"q113_costoConstruida": "12,183,063",
	"q114_plusvaliaCompetencia": "1,165,337",
	"q115_plusvaliaContruida": "730,984",
	"q116_valorTipo": "22,500",
	"q119_valorPago": "2,000,000",
	"website": "",
	"simple_spc": "240546574547868-240546574547868",
	"event_id": "1714551684861_240546574547868_MP0dJjM",
	"embedUrl": "https://rebora.com.mx/",
	"timeToSubmit": "20",
	"validatedNewRequiredFieldIDs": "{\"new\":1,\"id_4\":\"te\",\"id_57\":\"as\",\"id_123\":\"as\",\"id_5\":\"34\",\"id_117\":\"2,\",\"id_118\":\"25\",\"id_122\":\"16\",\"id_9\":\"De\",\"id_15\":\"2\",\"id_16\":\"2\",\"id_76\":\"0\",\"id_78\":\"20\",\"id_79\":\"20\",\"id_77\":\"44\",\"id_84\":\"0\",\"id_85\":\"30\",\"id_86\":\"0\",\"id_87\":\"50\"}",
	"visitedPages": "{\"1\":true,\"2\":true,\"3\":true}"
}
"""
#Peticion que envia Rebora cuando se carga un formulario nuevo

import requests
import os
import urllib.parse

from src.logger import Logger
from src.lead import Lead

API_KEY = os.getenv("JOTFORM_API_KEY")
FORM_ID = os.getenv("JOTFORM_FORM_ID")
assert API_KEY != "" and API_KEY != None, "JOTFORM_API_KEY is not in enviroment"
assert FORM_ID != "" and FORM_ID != None, "JOTFORM_FORM_ID is not in enviroment"

API_URL = "https://api.jotform.com"
PDF_URL = "https://www.jotform.com/pdf-submission/{submissionID}"

#Devuelve un link al PDF creado
def new_submission(logger: Logger, lead: Lead) -> None | str:
    logger.fuente += " > Jotform API"
    url = urllib.parse.urljoin(API_URL, f"form/{FORM_ID}/submissions")

    #Valores por defecto
    if lead.busquedas['covered_area'] == "" or lead.busquedas['covered_area'] == None:
        lead.busquedas['covered_area'] = '350, 350'

    if lead.busquedas['banios'] == "" or lead.busquedas['banios'] == None:
        lead.busquedas['banios'] = '5, 5'

    if lead.busquedas['recamaras'] == "" or lead.busquedas['recamaras'] == None:
        lead.busquedas['recamaras'] = '4, 4'

    props = {
        "covered_area": 0,
        "banios": 0 ,
        "recamaras": 0,
    }
    #La data viene como banios: 5, 6. osea entre 5 y 6, calculamos un promedio entre esos dos valores y usamos eso
    for key in props:
        try:
            [min, max] = lead.busquedas[key].split(', ')
            props[key] = int((int(min) + int(max)) / 2)
        except Exception as e:
            logger.error(f"No se pudo obtener el {key}")
            logger.debug(f"busquedas['{key}']: " + str(lead.busquedas[key]))
            logger.error(str(e))
            return None

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
            "submission[15]": props['recamaras'], #q15_cuantosCuartos	"2"
            "submission[16]": props['banios'], #q16_cuantosBanos56	"3"
            "submission[76]": "0",# TAMAÑO DEL SOTANOq76_cuantosBanos	"5"
            "submission[78]": int(props['covered_area'] * 0.45), #Tamaño planta baja q78_tamanoDe	"1"
            "submission[79]": int(props['covered_area'] * 0.45), #Tamaño planta alta q79_tamanoDe79	"7"
            "submission[77]": int(props['covered_area'] * 0.10), # Tamaño del roof garden q77_tamanoDel	"8"
            "submission[84]": "0", # Tamaño rampa de estacionamiento q84_tamanoDe84	"0"
            "submission[85]": "30", # Tamaño jardin exterior q85_tamanoDe85	"0"
            "submission[86]": "0", # Tamaño de alberca q86_tamanoDe86	"0"
            "submission[87]": "50", # Tamaño muro perimetral q87_tamanoDe87	"0"
            "submission[88]": "0", # Valor sotano q88_valorSotano	"112,500"
    }

    res = requests.post(url, data=payload)

    if res.ok:
        logger.success('Solicitud exitosa')
        data = res.json()
        submissionID = data.get('content', {}).get('submissionID', None)
        if submissionID == None:
            return None

        return PDF_URL.format(submissionID=submissionID)
    else:
        logger.error('Error en la solicitud:' + res.text)
        return None

def get_questions_form(logger: Logger, form_id: str = FORM_ID):
    logger.fuente += " > Jotform API"
    url = urllib.parse.urljoin(API_URL, f"form/{form_id}/questions")
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
