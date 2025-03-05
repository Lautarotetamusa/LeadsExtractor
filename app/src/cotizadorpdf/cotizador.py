from jinja2 import Environment, FileSystemLoader
import threading
import os
from datetime import date, datetime
from weasyprint import HTML
from flask import url_for
# 📌 Cargar la plantilla HTML y reemplazar variables dinámicas

def formato_miles(valor):
    return "{:,.0f}".format(valor).replace(",", ".")

def renderizar_html(template_name, contexto):
    env = Environment(loader=FileSystemLoader(os.getcwd()))  # Busca en el directorio actual
    env.filters["formato_miles"] = formato_miles
    styles = url_for('static', filename='styles.css')
    env.globals['url_for'] = styles
    template = env.get_template(template_name)
    return template.render(contexto)



def calcular_importe_calidad(pal):
    if(pal == "Premium"):
        return 18000
    if(pal == "Lujo"):
        return 20000
    if(pal == "Alto Lujo"):
        return 22500
    return 0

def format(num):
    numero_formateado = "{:,.0f}".format(num).replace(",", ".")
    return numero_formateado

# 📌 Datos dinámicos que queremos insertar en el HTML
def translateContext(cin):
    enombre = cin['elaborado_por']['nombre']
    email = cin['elaborado_por']['mail']
    etelefono = cin['elaborado_por']['telefono']
    nombre = cin['datos']['nombre']
    fecha = date.today()
    calidad = cin['pagos']['tipo']
    importe_calidad = calcular_importe_calidad(calidad)
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
    valor_rampa = cin['valor_exteriores']['rampa'] 
    valor_jardin = cin['valor_exteriores']['jardin']
    valor_alberca = cin['valor_exteriores']['alberca'] 
    valor_muro_perimetral = cin['valor_exteriores']['muro_perimetral']
    area_exterior = area_rampa + area_alberca + area_jardin + area_muro_perimetral
    valor_exterior = area_rampa*valor_rampa + area_alberca*valor_alberca + area_jardin*valor_jardin + area_muro_perimetral*valor_muro_perimetral
    valor_licencia = cin['valor_permisos']['licencia']
    valor_gestorias = cin['valor_permisos']['gestorias']
    valor_topografia = cin['valor_permisos']['topografia']
    valor_mecanica = cin['valor_permisos']['mecanica']
    valor_calculo = cin['valor_permisos']['calculo']
    valor_permisos = (valor_licencia + valor_gestorias + valor_calculo) * area_interior + 5800 * valor_mecanica + valor_topografia
    valor_rebora = 2000 * area_interior
    valor_preconstruccion = int(valor_rebora*0.3)
    valor_construccion = int(valor_rebora*0.7)
    is_valor_permisos = cin["elaborado_por"]["is_valor_permisos"]
    porcentaje_administracion = cin["elaborado_por"]["porcentaje_administracion"]
    valor_administracion = int((valor_exterior + valor_interior + valor_permisos*is_valor_permisos)*porcentaje_administracion/100)
    valor_total = valor_exterior+valor_interior+valor_permisos+valor_rebora+valor_administracion
    anticipo = cin['pagos']['inicial']
    porcentaje_previo = cin['pagos']['porcentaje_inicio_obra']
    valor_previo = int((valor_total-anticipo)*porcentaje_previo/100)
    meses = cin['pagos']['meses']
    pagos_mensuales = int((valor_total - valor_previo - anticipo)/meses)
    valor_otro_despacho = int(valor_total * 1.12)
    valor_casa_construida = int(valor_total * 1.18)
    plusvalia = int(valor_total * 0.17)
    plusvalia_otro_despacho = int(valor_otro_despacho * 0.10)
    plusvalia_casa_construida = int(valor_casa_construida * 0.06)
    precio_venta_rebora=int(valor_total*1.22)
    precio_venta_otro_despacho = int(valor_otro_despacho*1.15)
    precio_venta_casa_construida=int(valor_casa_construida*1.07)
    roi_rebora=int(precio_venta_rebora-valor_total)
    roi_otro_despacho=int(precio_venta_otro_despacho-valor_otro_despacho)
    roi_casa_construida=int(precio_venta_casa_construida-valor_casa_construida)

    contexto = {
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
        "valor_rebora": valor_rebora,
        "valor_administracion": valor_administracion,
        "porcentaje_administracion": porcentaje_administracion,
        "valor_preconstruccion": valor_preconstruccion,
        "valor_construccion": valor_construccion,
        "valor_total": valor_total,
        "anticipo": anticipo,
        "porcentaje_previo": porcentaje_previo,
        "valor_previo": valor_previo,
        "meses": meses,
        "pagos_mensuales": pagos_mensuales,
        "valor_otro_despacho": valor_otro_despacho,
        "valor_casa_construida": valor_casa_construida,
        "plusvalia": plusvalia,
        "plusvalia_otro_despacho": plusvalia_otro_despacho,
        "plusvalia_casa_construida": plusvalia_casa_construida,
        "roi_otro_despacho": roi_otro_despacho,
        "roi_casa_construida": roi_casa_construida,
        "roi_rebora": roi_rebora,
        "precio_venta_otro_despacho": precio_venta_otro_despacho,
        "precio_venta_casa_construida": precio_venta_casa_construida,
        "precio_venta_rebora": precio_venta_rebora 
    }
    return contexto

# 📌 Renderizar HTML con datos dinámicos

def to_pdf(json):
    try:
        contexto = translateContext(json)
        html_content = renderizar_html("/src/cotizadorpdf/presupuesto2.html", contexto)
        timestamp_str = datetime.now().strftime("%Y-%m-%d%H:%M:%S")
        pdf_filename = os.path.join("pdfs", "cotizacion" + timestamp_str +".pdf")
        HTML(string=html_content, base_url=".").write_pdf(pdf_filename)
        return timestamp_str
    except Exception as e:
        return "error"
