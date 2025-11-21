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
        "searchParameters": "sort:lowerQuality"
    }

    for prop in i24.get_properties(query=query):
        id = prop["postingId"]

        quality_info = i24.get_quality(id)
        if quality_info is None:
            print("percentage not found")
            continue
        if not "percentage_correctness" in quality_info:
            print("percentage not found")
            continue

        if int(quality_info.get("percentage_correctness", "100")) < 80:
            print("low quality found")
            # print(json.dumps(quality_info, indent=4))
            ids.append(id)

        print(ids)
