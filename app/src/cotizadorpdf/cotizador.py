import time
from jinja2 import Environment, FileSystemLoader
import os
from datetime import date, datetime

from src.cotizadorpdf.time_line import grafico_pagos 

from selenium import webdriver
from selenium.webdriver.chrome.options import Options
import base64

def get_pdf_from_html(path):
    webdriver_options = Options()
    webdriver_options.add_argument("--headless")
    # Necesario para correrlo como root dentro del container
    webdriver_options.add_argument("--no-sandbox")
    driver = webdriver.Chrome(options=webdriver_options)

    driver.get(f"file://{path}")

    driver.execute_script("return window.print()")
    pdf = driver.execute_cdp_cmd("Page.printToPDF", {"printBackground": True})
    pdf_data = base64.b64decode(pdf["data"])

    driver.quit()
    return pdf_data

def formato_miles(valor):
    return "{:,.0f}".format(valor)

def renderizar_html(template_name, contexto):
    env = Environment(
        loader=FileSystemLoader(os.getcwd()),  # Busca en el directorio actual
        comment_start_string='{=', # Porque el template tiene css embebido que usa simbolos iguales que a los comentarios de jinja
        comment_end_string='=}'
    )
    env.filters["formato_miles"] = formato_miles
    # styles = url_for('static', filename='styles.css')
    # env.globals['url_for'] = styles
    template = env.get_template(template_name)
    return template.render(contexto)

def calcular_importe_calidad(pal):
    if(pal == "Premium"):
        return 17500
    return 0

def format(num):
    numero_formateado = "{:,.0f}".format(num).replace(",", ".")
    return numero_formateado

# 📌 Datos dinámicos que queremos insertar en el HTML
def translateContext(cin):
    porcentaje_administracion = cin["elaborado_por"]["porcentaje_administracion"]
    enombre = cin['elaborado_por']['nombre']
    email = cin['elaborado_por']['mail']
    etelefono = cin['elaborado_por']['telefono']
    nombre = cin['datos']['nombre']
    fecha = date.today().strftime("%d/%m/%Y")
    print("amenidades", "amenidades" in cin)
    banios = cin['areas_interiores']['banos']
    recamaras = cin['areas_interiores']['cuartos'] 
    calidad = cin['pagos']['tipo']
    importe_inicial = calcular_importe_calidad(calidad)
    coeficiente_ganancia = (1- (porcentaje_administracion/100))
    importe_calidad = int((importe_inicial + 2000) / coeficiente_ganancia)
    area_sotano = cin['areas_interiores']['sotano']
    area_planta_baja = cin['areas_interiores']['planta_baja']
    area_planta_alta = cin['areas_interiores']['planta_alta']
    area_roof = cin['areas_interiores']['roof']
    area_interior = area_sotano+area_planta_baja+area_planta_alta+area_roof
    valor_interior = area_interior*importe_calidad
    area_rampa = cin['areas_exteriores']['rampa'] 
    area_jardin = cin['areas_exteriores']['jardin']
    area_alberca = cin['areas_exteriores']['alberca'] 
    area_muro_perimetral = cin['areas_exteriores']['muro_perimetral']

    # Niveles
    niveles = 0
    if area_sotano != "" and int(area_sotano) > 0:
        niveles += 1
    if area_planta_baja != "" and int(area_planta_baja) > 0:
        niveles += 1
    if area_planta_alta != "" and int(area_planta_alta) > 0:
        niveles += 1
    if area_roof != "" and int(area_roof) > 0:
        niveles += 1

    valor_rampa = int(cin['valor_exteriores']['rampa']/coeficiente_ganancia) 
    valor_jardin = int(cin['valor_exteriores']['jardin']/coeficiente_ganancia)
    valor_alberca = int(cin['valor_exteriores']['alberca']/coeficiente_ganancia)
    valor_muro_perimetral = int(cin['valor_exteriores']['muro_perimetral']/coeficiente_ganancia)
    area_exterior = area_rampa + area_alberca + area_jardin + area_muro_perimetral
    valor_exterior = area_rampa*valor_rampa + area_alberca*valor_alberca + area_jardin*valor_jardin + area_muro_perimetral*valor_muro_perimetral
    valor_licencia = int(cin['valor_permisos']['licencia'] / coeficiente_ganancia)
    valor_gestorias = int(cin['valor_permisos']['gestorias'] /coeficiente_ganancia)
    valor_topografia = int(cin['valor_permisos']['topografia'] /coeficiente_ganancia)
    valor_mecanica = int(cin['valor_permisos']['mecanica'] / coeficiente_ganancia)
    valor_calculo = int(cin['valor_permisos']['calculo'] / coeficiente_ganancia)
    valor_permisos = (valor_licencia + valor_gestorias + valor_calculo) * area_interior + 5800 * valor_mecanica + valor_topografia
    valor_metro_rebora = 2000
    valor_rebora = valor_metro_rebora * area_interior
    valor_preconstruccion = int(valor_rebora*0.3)
    valor_construccion = int(valor_rebora*0.7)
    is_valor_permisos = cin["elaborado_por"]["is_valor_permisos"]
    valor_interior_m2 = int(valor_interior/area_interior)
    valor_exterior_m2 = int(valor_exterior/area_interior)
    valor_permisos_m2 = int(valor_permisos/area_interior)
    valor_subtotal_m2 = int((valor_interior_m2+valor_exterior_m2+valor_permisos_m2*is_valor_permisos))
    valor_administracion_m2 = int((valor_subtotal_m2*porcentaje_administracion)/100)
    valor_administracion = int(valor_administracion_m2 * area_interior)
    valor_total = valor_exterior+valor_interior+valor_permisos
    valor_total_min = valor_total * 0.95
    valor_total_max = valor_total * 1.05
    valor_total_min_m2 = importe_calidad * 0.95
    valor_total_m2 = importe_calidad
    valor_total_max_m2 = importe_calidad * 1.05
    costo_directo = int((valor_total) * (100/(100+porcentaje_administracion)))
    porcentaje_administracion_rebora_total = int((valor_interior-valor_interior*coeficiente_ganancia)+ valor_exterior-valor_exterior*coeficiente_ganancia + valor_permisos - valor_permisos*coeficiente_ganancia)
    ahorro_materiales = (importe_inicial*area_interior)*0.08
    ahorro_contratista = (importe_inicial*area_interior)*0.1
    costo_sin_ahorro = valor_total + ahorro_contratista + ahorro_materiales
    anticipo = cin['pagos']['inicial']
    porcentaje_previo = cin['pagos']['porcentaje_inicio_obra']
    valor_previo = int((valor_total-anticipo)*porcentaje_previo/100)
    meses = cin['pagos']['meses']
    pagos_mensuales = int((valor_total - valor_previo - anticipo)/meses)
    valor_terreno= cin['elaborado_por']['valor_terreno']
    area_terreno= cin['elaborado_por']['area_terreno']
    plusvalia_terreno_rebora= valor_terreno*0.24
    plusvalia_terreno_otro_despacho= valor_terreno*0.20
    plusvalia_terreno_casa_construida= valor_terreno*0.20
    valor_otro_despacho = int(valor_total * 1.12)
    valor_casa_construida = int(valor_total * 1.18)
    plusvalia = int(valor_total * 0.26)
    plusvalia_otro_despacho = int(valor_otro_despacho * 0.12)
    plusvalia_casa_construida = int(valor_casa_construida * 0.06)
    plusvalia_total_rebora = plusvalia_terreno_rebora + plusvalia
    print("plusvalia_total_rebora", plusvalia_total_rebora)
    plusvalia_total_otro_despacho = plusvalia_terreno_otro_despacho + plusvalia_otro_despacho
    plusvalia_total_casa_construida = plusvalia_terreno_casa_construida + plusvalia_casa_construida
    precio_venta_rebora=int(valor_terreno+valor_total+plusvalia_total_rebora)
    precio_venta_otro_despacho = int(valor_otro_despacho+plusvalia_total_otro_despacho)
    precio_venta_casa_construida=int(valor_casa_construida+plusvalia_casa_construida)
    roi_rebora=int(precio_venta_rebora-valor_total-valor_terreno)
    roi_otro_despacho=int(precio_venta_otro_despacho-valor_otro_despacho)
    roi_casa_construida=int(precio_venta_casa_construida-valor_casa_construida)
    roi_terreno = int(valor_terreno)+int(plusvalia_terreno_rebora)
    contexto = {
        "roi_terreno": roi_terreno,
        "enombre": enombre,
        "etelefono":etelefono,
        "email": email,
        "nombre": nombre,
        "fecha": fecha,
        "calidad": calidad,
        "importe_calidad": importe_calidad,
        "area_sotano": area_sotano,
        "area_planta_baja": area_planta_baja,
        "area_planta_alta": area_planta_alta,
        "area_roof": area_roof,
        "area_interior": area_interior,
        "valor_interior": valor_interior,
        "area_rampa": area_rampa,
        "area_jardin": area_jardin,
        "area_alberca": area_alberca,
        "area_muro_perimetral": area_muro_perimetral,
        "valor_rampa": valor_rampa,
        "valor_jardin": valor_jardin,
        "valor_alberca": valor_alberca,
        "valor_muro_perimetral": valor_muro_perimetral,
        "area_exterior": area_exterior,
        "valor_exterior": valor_exterior,
        "valor_licencia": valor_licencia,
        "valor_gestorias": valor_gestorias,
        "valor_topografia": valor_topografia,
        "valor_mecanica": valor_mecanica,
        "valor_calculo": valor_calculo,
        "valor_permisos": valor_permisos,
        "valor_metro_rebora": valor_metro_rebora,
        "valor_rebora": valor_rebora,
        "valor_interior_m2": valor_interior_m2,
        "valor_exterior_m2": valor_exterior_m2,
        "valor_permisos_m2": valor_permisos_m2,
        "valor_administracion_m2": valor_administracion_m2,
        "valor_subtotal_m2": valor_subtotal_m2,
        "valor_administracion": valor_administracion,
        "porcentaje_administracion": porcentaje_administracion,
        "is_valor_permisos": is_valor_permisos,
        "valor_preconstruccion": valor_preconstruccion,
        "valor_construccion": valor_construccion,
        "valor_total_min": valor_total_min,
        "valor_total": valor_total,
        "valor_total_max": valor_total_max,
        "valor_total_min_m2": valor_total_min_m2,
        "valor_total_m2": valor_total_m2,
        "valor_total_max_m2": valor_total_max_m2,
        "porcentaje_administracion_rebora_total": porcentaje_administracion_rebora_total,
        "costo_directo": costo_directo,
        "ahorro_materiales": ahorro_materiales,
        "ahorro_contratista": ahorro_contratista,
        "costo_sin_ahorro": costo_sin_ahorro,
        "anticipo": anticipo,
        "porcentaje_previo": porcentaje_previo,
        "valor_previo": valor_previo,
        "meses": meses,
        "pagos_mensuales": pagos_mensuales,
        "valor_terreno": valor_terreno,
        "area_terreno": area_terreno,
        "valor_otro_despacho": valor_otro_despacho,
        "valor_casa_construida": valor_casa_construida,
        "plusvalia_terreno_rebora": plusvalia_terreno_rebora,
        "plusvalia_terreno_otro_despacho": plusvalia_terreno_otro_despacho,
        "plusvalia_terreno_casa_construida": plusvalia_terreno_casa_construida,
        "plusvalia": plusvalia,
        "plusvalia_otro_despacho": plusvalia_otro_despacho,
        "plusvalia_casa_construida": plusvalia_casa_construida,
        "plusvalia_total_rebora": plusvalia_total_rebora,
        "plusvalia_total_otro_despacho": plusvalia_total_otro_despacho,
        "plusvalia_total_casa_construida": plusvalia_total_casa_construida,
        "roi_otro_despacho": roi_otro_despacho,
        "roi_casa_construida": roi_casa_construida,
        "roi_rebora": roi_rebora,
        "precio_venta_otro_despacho": precio_venta_otro_despacho,
        "precio_venta_casa_construida": precio_venta_casa_construida,
        "precio_venta_rebora": precio_venta_rebora,
        "banios": banios,
        "recamaras": recamaras,
        "niveles": niveles
    }

    print(area_rampa*valor_rampa)
    print(area_alberca*valor_alberca)
    print(area_jardin*valor_jardin)
    print(area_muro_perimetral*valor_muro_perimetral)

    print("---")
    print(valor_exterior)
    print(area_interior)
    print(importe_calidad)
    print(valor_interior)
    print(valor_permisos)
    print(valor_total)

    return contexto

# 📌 Renderizar HTML con datos dinámicos
def to_pdf(json) -> str:
    """returns timestamp to the file"""
    try:
        contexto = translateContext(json)
        contexto['nombre_grafico_pagos'] = grafico_pagos(contexto)
        contexto['nombre_grafico_pagos'] = "./grafico_pagos.png" ##XD
        # contexto['nombre_grafico_etapas'] = grafico_etapas(contexto)
        template = "/src/cotizadorpdf/cotizacion3.html"
        html_content = renderizar_html(template, contexto)
        timestamp_str = datetime.now().strftime("%Y-%m-%d%H:%M:%S")

        base_path = os.getcwd()

        html_path = os.path.join(base_path, "pdfs", "cotizacion" + timestamp_str +".html")
        with open(html_path, "w") as f:
            f.write(html_content)
        pdf_path = os.path.join(base_path, "pdfs", "cotizacion" + timestamp_str +".pdf")

        print(html_path, pdf_path)
        result = get_pdf_from_html(html_path)
        with open(pdf_path, 'wb') as file:
            file.write(result)

        return timestamp_str
    except Exception as e:
        print(e)
        return "error"
