# Script: get the casasyterrenos internal ids for the required zones

import re

def USAGE():
    print("""
        python search_internal_zones.py <neighborhoods-file.json> <state> <state-id> <city> <city-id>

          Ej:
        python search_internal_zones.py colonies.json Jalisco 14 Zapopan 534
    """)

def search_neighborhood(neighbors: list[dict], key: str) -> dict | None:
    for neighborhood in neighbors:
        # print("|",key.lower(), "|", sep="")
        # print("|", neighborhood["name"].lower(), "|", sep="") 
        key = key.lower().strip()
        name = neighborhood["name"].lower().strip()
        if key in name or name in key:
            return {
                "colony": neighborhood["name"],
                "id": neighborhood["sepomexId"],
                "zipcode": neighborhood["zipcode"]
            }

    return None

if __name__ == "__main__":
    import json
    import csv
    import sys

    if len(sys.argv) < 6:
        print("missing required parasm")
        USAGE()
        exit(1)

    neighborhood_file = sys.argv[1]
    state = sys.argv[2]
    state_id = sys.argv[3]
    city = sys.argv[4]
    city_id = sys.argv[5]

    with open(neighborhood_file, "r") as f:
        neighborhoods = json.load(f)

    zones = {}
    with open("zones.csv", "r") as f:
        rows = csv.DictReader(f)

        for row in rows:
            if row["internal"] == "":
                continue

            internal_ubication = search_neighborhood(neighborhoods, row["internal"])

            if internal_ubication is None:
                print(row["internal"], "not found")
                continue

            internal_ubication["state"] = state
            internal_ubication["state_id"] = state_id
            internal_ubication["city"] = city
            internal_ubication["city_id"] = city_id

            zones[row["address"]] = {
                "zone": row["zone"],
                "internal": internal_ubication,
            }

    # print(json.dumps(zones, indent=4))
    with open("internal_zones.json", "w") as f:
        json.dump(zones, f)
