# Script: get the inmuebles24 internal ids for the required zones
# state: Jalisco - 74
# city:  Zapopan - 779
# city: Tlajomulco de Zúñiga - 790

def USAGE():
    print("""
        python search_internal_zones.py <neighborhoods-file.json> <state-id> <city-id>
        python search_internal_zones.py validate

          Ej:
        python search_internal_zones.py neighborhoods-779.json 74 779
    """)

def search_neighborhood(neighbors: list[dict], key: str) -> dict | None:
    for neighborhood in neighbors:
        # print("|",key.lower(), "|", sep="")
        # print("|", neighborhood["name"].lower(), "|", sep="") 
        key = key.lower().strip()
        name = neighborhood["nombre"].lower().strip()
        if key in name or name in key:
            return {
                "colony": neighborhood["nombre"],
                "id": neighborhood["id"],
            }

    return None

def validate():
    with open("internal_zones.json", "r") as f:
        zones = json.load(f)

    with open("zones.csv", "r") as f:
        rows = csv.DictReader(f)

        not_found = 1
        for row in rows:
            if row["address"] not in zones:
                not_found += 1
                print(row["internal"], "dont found")

        print("not found: ", not_found)

if __name__ == "__main__":
    import json
    import csv
    import sys

    if sys.argv[1] == "validate":
        validate()
        exit(0)

    if len(sys.argv) < 4:
        print("missing required parasm")
        USAGE()
        exit(1)

    neighborhood_file = sys.argv[1]
    state_id = sys.argv[2]
    city_id = sys.argv[3]

    print("state_id: ", state_id)
    print("city_id: ", city_id)

    with open(neighborhood_file, "r") as f:
        neighborhoods = json.load(f)["contenido"]["localidades"]

    with open("internal_zones.json", "r") as f:
        zones = json.load(f)
 
    with open("zones.csv", "r") as f:
        rows = csv.DictReader(f)

        not_found = 1
        for row in rows:
            if row["internal"] == "":
                continue

            internal_ubication = search_neighborhood(neighborhoods, row["internal"])

            if internal_ubication is None:
                not_found += 1
                # print(row["internal"], "not found")
                continue
            print(row["internal"])

            # internal_ubication["state"] = state
            internal_ubication["state_id"] = state_id
            # internal_ubication["city"] = city
            internal_ubication["city_id"] = city_id

            zones[row["address"]] = {
                "zone": row["zone"],
                "internal": internal_ubication,
            }

        print("not found: ", not_found)

    # print(json.dumps(zones, indent=4))
    with open("internal_zones.json", "w") as f:
        json.dump(zones, f)
