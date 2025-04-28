import os
import sys
sys.path[0] = os.getcwd()

import json
from src.inmuebles24.inmuebles24 import Inmuebles24

if __name__ == "__main__": 
    portal = Inmuebles24()

    prop_ids: list[str] = []
    properties = []
    page = 1
    while True:
        properties = portal.get_properties(status="ONLINE", page=page)
        if properties is None:
            break

        prop_ids += [str(p["postingId"]) for p in properties]
        page += 1

    print(json.dumps(prop_ids, indent=4))

    err = portal.unpublish_multiple(prop_ids)
    if err is not None:
        print(err)

    print("publications unpublished succesfully")
