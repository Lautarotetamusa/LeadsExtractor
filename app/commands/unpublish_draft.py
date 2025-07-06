# Unpublish Drafted properties in inmuebles24
# TODO: add the others portals

import os
import sys
sys.path[0] = os.getcwd()

from src.inmuebles24.inmuebles24 import Inmuebles24
from src.portal import Portal

if __name__ == "__main__":
    i24 = Inmuebles24()

    ids = []
    for prop in i24.get_properties(status="DRAFT", query={"page":1}):
        id = prop["postingId"]
        print(id)

        ids.append(id)

        if len(ids) == 20:
            err = i24.unpublish(ids)
            if err is not None:
                print(err)

            ids = []
