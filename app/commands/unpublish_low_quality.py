# Unpublish properties with poor % (eg: 66% missing images)
# TODO: add the others portals

import os
import sys
import json
sys.path[0] = os.getcwd()

from src.inmuebles24.inmuebles24 import Inmuebles24
from src.portal import Portal

if __name__ == "__main__":
    i24 = Inmuebles24()

    ids = []
    query = {
        "page": 1,
        "searchParameters": "sort:lowerQuality"
    }

    for prop in i24.get_properties(query=query):
        id = prop["postingId"]
        ids.append(id)

        quality_info = i24.get_quality(id)
        print(quality_info)
        print(json.dumps(quality_info, indent=4))

        if len(ids) == 20:
            # err = i24.unpublish(ids)
            # if err is not None:
            #     print(err)

            ids = []
