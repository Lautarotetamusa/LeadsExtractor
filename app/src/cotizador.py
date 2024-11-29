from multiprocessing.pool import ThreadPool
import src.jotform as jotform
from pypdf import PdfWriter
import datetime
import requests
import string
import random

from PIL import Image
from io import BytesIO

from src.logger import Logger
import src.tasks as tasks

# Cotizadores
import src.inmuebles24.cotizador as inmuebles24
import src.easybroker.cotizador as easybroker

logger = Logger("Cotizador")

COTIZADORES = {
    "inmuebles24": inmuebles24,
    "easybroker": easybroker
}

DEFAULT_IMG_SIZE = (360, 266)

def execute_cotizacion(urls: list[str], asesor: dict, cliente: str, task_id: str):
    try:
        posts = []
        pool = ThreadPool(processes=8)
        results = []
        for url in urls:
            cotizador = get_portal_from_url(url)
            if cotizador is None:
                continue
            r = pool.apply_async(cotizador.get_post_data, args=(url, ))
            results.append(r)

        for r in results:
            post = r.get()
            if post is not None:
                posts.append(post)

        pdf_file_name = cotizacion(asesor, cliente, posts)
        tasks.update_status(task_id, tasks.Status.completed)
        tasks.update(task_id, pdf_file_name=pdf_file_name)
    except Exception as e:
        tasks.update_status(task_id, tasks.Status.error)
        tasks.update(task_id, error=str(e))
        print(f"Error generando la cotización: {str(e)}")

def get_portal_from_url(url: str):
    for portal in COTIZADORES:
        if portal in url:
            return COTIZADORES[portal]

    logger.error(f"no se encontro cotizador para la url {url}")
    return None

def combine_pdfs(pdfs: list[str], file_name):
    merger = PdfWriter()

    for pdf in pdfs:
        merger.append(pdf)

    merger.write(file_name)
    merger.close()

#TODO: No hardcodear el tipo de dato JPEG
def resize_image(img_data, target_size):
    try:
        img = Image.open(BytesIO(img_data))
        img = img.resize(target_size)  # Redimensiona la imagen
        output = BytesIO()
        img.save(output, format="JPEG")  # Guarda en el mismo formato que la original
        return output.getvalue()  # Retorna los bytes de la imagen redimensionada
    except Exception as e:
        logger.error(f"Error al redimensionar la imagen: {str(e)}")
        return None

# Subir las imagenes a jotform
def upload_images(form_id: str, submission_id: str, urls: list[str], qids: list[str]):
    pool = ThreadPool(processes=20)

    results = []
    nro = 1
    for url, qid in zip(urls, qids):
        logger.debug("Obteniendo imagen: " + url)
        img_data = None
        if "maps" in url: # Si la imagen es de maps no necesitamos obtenerla con jotform
            res = requests.get(url)
            if res is None or not res.ok:
                logger.error("Imposible obtener la imagen de ubicacion: "+ url)
                logger.error(res.text)
                img_data = None
            else:
                img_data = res.content
        else:
            img_data = jotform.get_img_data(url)

        if img_data is None:
            logger.error("Imposible obtener la imagen: "+ url)
            continue
        
        if "maps" not in url:  # Si la imagen es de maps no hay que hacerle resize
            img_data = resize_image(img_data, DEFAULT_IMG_SIZE)
            if img_data is None: continue

        r = pool.apply_async(
                jotform.upload_image,
                args=(form_id, submission_id, qid, img_data, f"{submission_id}_{qid}_{nro}", )
            )
        results.append(r)
        nro += 1

    for r in results:
        err = r.get()
        if err is None:
            logger.success("Imagen subida correctamente")
        else:
            logger.error("No fue posible subir la image: " + str(err))

# Subir la cotizacion al formulario de jotform
def cotizacion_post(post: dict, form_id: str, asesor: dict, cliente: str, id_propuesta: dict):
    map_url = post["map_url"]
    images_urls = post["images_urls"]

    logger.debug("Uploading Cotizacion Form")
    res = jotform.submit_cotizacion_form(logger, form_id, post, asesor, cliente, id_propuesta)
    if res is None:
        logger.error("No fue posible subir la cotizacion a jotform")
        return None

    submission_id = res["content"]["submissionID"]
    images_qids = ["77", "44", "44", "44", "44"]  # TODO: No hardcodear
    upload_images(form_id, submission_id, [map_url] + images_urls, images_qids)

    res = jotform.obtain_pdf(form_id, submission_id)
    if res is None:
        return None

    pdf_path = f"./cotizaciones/{submission_id}.pdf"
    with open(pdf_path, 'wb') as f:
        f.write(res)

    return pdf_path


# Genera cotizaciones en pdf para los postings en la lista
def cotizacion(asesor: dict, cliente: str, posts: list[dict]):
    form_id = "242244461116044"  # TODO: No hardcodear
    pdfs = []
    results = []
    pool = ThreadPool(processes=8)

    id_cotizacion = ''.join(random.choices(string.ascii_letters, k=10))

    orden = 1
    for post in posts:
        post["orden"] = orden
        r = pool.apply_async(cotizacion_post, args=(post, form_id, asesor, cliente, id_cotizacion, ))
        results.append(r)
        orden += 1

    for r in results:
        pdf_path = r.get()
        if pdf_path is not None:
            pdfs.append(pdf_path)

    str_date = datetime.datetime.today().strftime("%d-%m-%Y")
    file_name = f"Propuesta terrenos {cliente} {id_cotizacion} {str_date}.pdf"
    combine_pdfs(pdfs, f"/app/pdfs/{file_name}")
    return file_name

