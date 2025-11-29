import os
import sys
import json
sys.path[0] = os.getcwd()

from src.casasyterrenos.casasyterrenos import CasasYTerrenos
from src.portal import Portal

if __name__ == "__main__":
    portal = CasasYTerrenos()

    ids = []

    for prop in portal.get_properties(featured=False):
        ids.append(prop["id"])

    print(json.dumps(ids))
