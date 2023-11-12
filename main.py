from src.casas_y_terrenos.casas_y_terrenos import main as casas_y_terrenos
from src.propiedades_com.propiedades import main as propiedades
from src.lamudi.lamudi import main as lamudi

import sys

PORTALS = {
    "casasyterrenos": casas_y_terrenos,
    "propiedades": propiedades,
    "lamudi": lamudi
}

def USAGE():
    print("""
        python main.py <PORTAL>
        PORTALS:
            - casasyterrenos
            - propiedades
            - lamudi
    """)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        USAGE()
        exit(1)
    
    if sys.argv[1] not in PORTALS:
        USAGE()
        exit(1)

    PORTALS[sys.argv[1]]()