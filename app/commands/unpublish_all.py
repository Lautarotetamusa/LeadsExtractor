import os
import sys
sys.path[0] = os.getcwd()

from src.casasyterrenos.casasyterrenos import CasasYTerrenos
from src.propiedades_com.propiedades import Propiedades
from src.lamudi.lamudi import Lamudi
from src.inmuebles24.inmuebles24 import Inmuebles24
from src.portal import Portal

if __name__ == "__main__":
    PORTALS: dict[str, Portal ]= {
        "casasyterrenos": CasasYTerrenos(),
        "propiedades": Propiedades(),
        "inmuebles24": Inmuebles24(),
        "lamudi": Lamudi(),
    }

    args = sys.argv

    if len(args) < 2:
        print("portal its required")
        exit(1)

    portal_name = args[1]
    if portal_name not in PORTALS.keys():
        print(f"Portal {portal_name} doesnt exists")
        exit(1)

    # id_key = "postingId"
    id_key = "id"

    portal = PORTALS[portal_name]
    ids = []
    for prop in portal.get_properties(featured=False):
        id = prop.get(id_key)
        print(id)
        if id is None:
            continue

        ids.append(id)

        if len(ids) == 8:
            err = portal.unpublish(ids)
            if err is not None:
                print(err)

            ids = []
