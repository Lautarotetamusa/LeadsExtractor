import os
import sys
import json
sys.path[0] = os.getcwd()

from src.propiedades_com.propiedades import Propiedades
from src.portal import Portal

if __name__ == "__main__":
    portal = Propiedades()

    ids = []

    slots = portal.get_slots()

    for prop in portal.get_properties(featured=True):
        ids.append(prop["id"])

        err = portal.unhighlight(prop["id"])
        if err is not None:
            print(err)

    print(json.dumps(ids))
    print(len(ids))
