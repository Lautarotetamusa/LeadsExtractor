import os
import sys
sys.path[0] = os.getcwd()

import json
from src.inmuebles24.inmuebles24 import Inmuebles24

if __name__ == "__main__":
    portal = Inmuebles24()

    prop_ids: list[str] = []
    properties = []
    for p in portal.get_properties(status="ONLINE"):

        print(p)
        # prop_ids += [str(p["postingId"]) for p in properties]
        #
        # print(json.dumps(prop_ids, indent=4))
