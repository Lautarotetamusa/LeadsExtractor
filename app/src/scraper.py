
from sheets import Sheet
from src.logger import Logger

class Scraper():
    def __init__(self,
            name: str,
        ):
        self.name = name 

        self.logger  = Logger(name)
        self.sheet   = Sheet(self.logger, 'scraper_mapping.json')
        self.headers = self.sheet.get("Extracciones!A1:Z1")[0]

    def main(filters: dict, spin_msg: str):

        page = 1
        len_posts = 0
        total_posts = 1e9
        session_id = generate_session_id()

        #url = f"https://propiedades.com/df/casas-venta/recamaras-2?pagina={page}"
        while len_posts < total_posts:
            logger.debug(f"Page: {page}")
            res = request.make(URL_PROPS, "POST", json=filters)
            data = res.json()

            if total_posts  == 1e9:
                total_posts = data.get("nbHits", 0)
                logger.debug(f"Total posts: {total_posts}")
            #print(data)
            posts = extract_posts(data.get("hits", []))

            page += 1
            len_posts += len(posts)
            filters["searchConfig"]["page"] = page

            row_ads = []
            for ad in posts:
                publisher = get_publisher_info(ad["canonical"])
                if publisher == None:
                    logger.error(f"Ocurrio un error encontrando la informacion de la propiedad {ad['id']}")
                else:
                    ad["publisher"] = publisher
                msg = generate_post_message(ad, spin_msg)
                send_message(ad["id"], ad["publisher"]["id"], session_id, msg)
                ad["message"] = msg.replace('\n', '')
                #Guardamos los leads como filas para el sheets
                row_ad = sheet.map_lead(ad, sheets_headers)
                row_ads.append(row_ad)

            #Save the lead in the sheet
            sheet.write(row_ads, "Extracciones!A2")
        logger.success(f"Se encontraron un total de {len_posts} en la url especificada")
