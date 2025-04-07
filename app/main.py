from src.commands.cotizador import cotizador
from src.commands.portal import portal
from src.commands.scrapers import scraper

import sys

TASKS = {
    "portal": portal,
    "scraper": scraper,
    "cotizador": cotizador
}

def USAGE():
    print("""
        python main.py <TASK> <...TASK ARGS>

        TASKS:
            - portal
            - scraper
            - cotizador
    """)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Task its required")
        USAGE()
        exit(1)

    task = sys.argv[1]

    if task not in TASKS:
        print(f"Task {task} doesnt exists")
        USAGE()
        exit(1)
    
    TASKS[task](sys.argv[2:]) 
