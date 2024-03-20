from threading import Thread

import src.casasyterrenos.casasyterrenos as casasyterrenos
import src.propiedades_com.propiedades as propiedades
import src.lamudi.lamudi as lamudi
import src.inmuebles24.inmuebles24 as inmuebles24

#Scrapers
from src.inmuebles24.scraper import main as inmuebles24_scraper
from src.lamudi.scraper import main as lamudi_scraper
from src.propiedades_com.scraper import main as propiedades_scraper
from src.casasyterrenos.scraper import main as casasyterrenos_scraper

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
    "all":{
        "main": run_all
    },
    "casasyterrenos": {
        "first_run": casasyterrenos.first_run,
        "main": casasyterrenos.main,
        "scraper": casasyterrenos_scraper
    },
    "propiedades": {
        "first_run": propiedades.first_run,
        "main": propiedades.main,
        "scraper": propiedades_scraper
    },
    "inmuebles24": {
        "first_run": inmuebles24.first_run,
        "main": inmuebles24.main,
        "scraper": inmuebles24_scraper
    },
    "lamudi": {
        "first_run": lamudi.first_run,
        "main": lamudi.main,
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
            - all:
                This option run all the portals in parallel threads

        TASKS:
            - first_run
            - main
            - scraper <URL> <MESSAGE>
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
