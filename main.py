from src.casas_y_terrenos.casas_y_terrenos import main as casas_y_terrenos
from src.propiedades_com.propiedades import main as propiedades
from src.lamudi.lamudi import main as lamudi_main, first_run as lamudi_first_run
from src.inmuebles24.inmuebles24 import main as inmuebles24

#Scrapers
from src.inmuebles24.scraper import main as inmuebles24_scraper
from src.lamudi.scraper import main as lamudi_scraper

import sys

PORTALS = {
    "casasyterrenos": {
        "first_run": "",
        "main": casas_y_terrenos
    },
    "propiedades": {
        "first_run": "",
        "main": propiedades
    },
    "inmuebles24": {
        "first_run": "",
        "main": inmuebles24,
        "scraper": inmuebles24_scraper
    },
    "lamudi": {
        "first_run": lamudi_first_run,
        "main": lamudi_main,
        "scraper": lamudi_scraper 
    }
}

def USAGE():
    print("""
        python main.py <PORTAL> <TASK>
        PORTALS:
            - casasyterrenos
            - propiedades
            - inmuebles24
            - lamudi

        TASKS:
            - first_run
            - main
            - scrape
    """)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        USAGE()
        exit(1)
    portal = sys.argv[1]
    if portal not in PORTALS:
        USAGE()
        exit(1)
    
    task = "main"
    if len(sys.argv) == 3:
        task = sys.argv[2]
    if task not in PORTALS[portal]:
        USAGE()
        exit(1)

    PORTALS[portal][task]()
