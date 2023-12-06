from src.casas_y_terrenos.casas_y_terrenos import main as casas_y_terrenos
from src.propiedades_com.propiedades import main as propiedades
from src.lamudi.lamudi import main as lamudi_main, first_run as lamudi_first_run
from src.inmuebles24.inmuebles24 import main as inmuebles24

#Scrapers
from src.inmuebles24.scraper import main as inmuebles24_scraper
from src.lamudi.scraper import main as lamudi_scraper
from src.propiedades_com.scraper import main as propiedades_scraper
from src.casas_y_terrenos.scraper import main as casasyterrenos_scraper

import sys

PORTALS = {
    "casasyterrenos": {
        "first_run": "",
        "main": casas_y_terrenos,
        "scraper": casasyterrenos_scraper
    },
    "propiedades": {
        "first_run": "",
        "main": propiedades,
        "scraper": propiedades_scraper
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
            - scraper <URL>
    """)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        USAGE()
        exit(1)
    portal = sys.argv[1]
    if portal not in PORTALS:
        print(f"Portal {portal} doesnt exists")
        USAGE()
        exit(1)
    task = "main"
    if len(sys.argv) >= 3:
        task = sys.argv[2]
    if task == "scraper" and len(sys.argv) < 4:
        print("Task scraper needs a URL parameter") 
        USAGE()
        exit(1)
    if task not in PORTALS[portal]:
        print(f"Task {task} doesnt exists")
        USAGE()
        exit(1)
    
    print(portal, task, *sys.argv[3:])
    PORTALS[portal][task](*sys.argv[3:])
