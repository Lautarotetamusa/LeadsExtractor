from src.inmuebles24.scraper import Inmuebles24Scraper
from src.lamudi.scraper import LamudiScraper
from src.propiedades_com.scraper import PropiedadesScraper 
from src.casasyterrenos.scraper import CasasyterrenosScraper

SCRAPERS = {
    "propiedades": PropiedadesScraper,
    "casasyterrenos": CasasyterrenosScraper,
    "lamudi": LamudiScraper,
    "inmuebles24": Inmuebles24Scraper
}

def USAGE():
    print("""
        python main.py scraper <SCRAPER> <SPIN_MSG> <FILTER_OR_URL>
        SCRAPERS:
            - casasyterrenos
            - propiedades
            - inmuebles24
            - lamudi
    """)

def scraper(args: list[str]): 
    if len(args) < 3:
        print("Missing required params")
        USAGE()
        exit(1)

    scraper = args[0]
    if scraper not in SCRAPERS:
        print(f"{scraper} is not a valid scraper")
        USAGE()
        exit(1)

    s = SCRAPERS[scraper]()
    getattr(s, "main")(args[1:])
