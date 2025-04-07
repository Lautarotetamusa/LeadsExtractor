from threading import Thread

from src.casasyterrenos.casasyterrenos import CasasYTerrenos
from src.propiedades_com.propiedades import Propiedades
from src.lamudi.lamudi import Lamudi
from src.inmuebles24.inmuebles24 import Inmuebles24

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

PORTAL_TASKS = [
    "first_run",
    "main",
    "test",
]

def USAGE():
    print("""
        python main.py portal <PORTAL> <TASK>
        PORTALS:
            - casasyterrenos
            - propiedades
            - inmuebles24
            - lamudi
            - all:
                This option run all the portals in parallel threads

        PORTAL TASKS:
            - first_run
            - main
            - test
    """)

def portal(args: list[str]):
    if len(args) < 1:
        USAGE()
        exit(1)

    portal_name = args[0]
    if portal_name not in PORTALS.keys():
        print(f"Portal {portal_name} doesnt exists")
        USAGE()
        exit(1)

    task = args[1]
    if task not in PORTAL_TASKS:
        print(f"Task {task} doesnt exists")
        USAGE()
        exit(1)
    
    portal = PORTALS[portal_name]() 
    getattr(portal, task)()
