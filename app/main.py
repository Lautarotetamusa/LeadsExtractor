from threading import Thread

from src.casasyterrenos.casasyterrenos import CasasYTerrenos
from src.propiedades_com.propiedades import Propiedades
from src.lamudi.lamudi import Lamudi
from src.inmuebles24.inmuebles24 import Inmuebles24

#Scrapers
from src.inmuebles24.scraper import Inmuebles24Scraper
from src.lamudi.scraper import LamudiScraper
from src.propiedades_com.scraper import PropiedadesScraper 
from src.casasyterrenos.scraper import CasasyterrenosScraper

import sys

def run_all():
    threads = []
    for portal_name in PORTALS:
        if portal_name == "all":
            continue
         
        thread = Thread(target=PORTALS[portal_name]["main"], args=())
        thread.start()
        threads.append(thread)

    for thread in threads:
        thread.join()

PORTALS = {
    "casasyterrenos": CasasYTerrenos,
    "propiedades": Propiedades,
    "inmuebles24": Inmuebles24,
    "lamudi": Lamudi,
}

SCRAPES = {
    "propiedades": PropiedadesScraper,
    "casasyterrenos": CasasyterrenosScraper,
    "lamudi": LamudiScraper,
    "inmuebles24": Inmuebles24Scraper
}

TASKS = [
    "first_run",
    "main",
    "scraper",
    "test",
    "test_page"
]

def USAGE():
    print("""
        python main.py <PORTAL> <TASK>
        PORTALS:
            - casasyterrenos
            - propiedades
            - inmuebles24
            - lamudi
            - all:
                This option run all the portals in parallel threads

        TASKS:
            - first_run
            - main
            - scraper <URL> <MESSAGE>
            - test
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
    if task not in TASKS:
        print(f"Task {task} doesnt exists")
        USAGE()
        exit(1)
    
    portal = PORTALS[portal]() 
    getattr(portal, task)(*sys.argv[3:])
