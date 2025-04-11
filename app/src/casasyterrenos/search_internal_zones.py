# Script: get the casasyterrenos internal ids for the required zones

import re

def USAGE():
    print("""
        python search_internal_zones.py <neighborhoods-file.json> <state-id>
    """)

def search_neighborhood(neighbors: list[dict], key: str) -> dict | None:
    for neighborhood in neighbors:
        if neighborhood["name"].lower() == key.lower():
            return {
                "colony": neighborhood["id"],
                "municipality": neighborhood["municipality"]
            }

    return None

def extract_street_from_address(address: str) -> str:
    parts = [part.strip() for part in address.split(',')]
    
    # Find the first part that contains a 5-digit postal code
    postal_code_index = None
    for i, part in enumerate(parts):
        if re.search(r'\b\d{5}\b', part):
            postal_code_index = i
            break
    
    # If no postal code found, return the entire address
    if postal_code_index is None:
        return ', '.join(parts)
    
    # Extract parts before the postal code
    street_parts = parts[:postal_code_index]
    
    # Find the last part in street_parts that contains a number (street number)
    last_number_index = None
    for i, part in enumerate(street_parts):
        if re.search(r'\d', part):
            last_number_index = i
    
    # Slice street_parts up to and including the last part with a number
    if last_number_index is not None:
        street_parts = street_parts[:last_number_index + 1]
    
    return ', '.join(street_parts)

if __name__ == "__main__":
    import json
    import csv
    import sys

    if len(sys.argv) < 3:
        print("missing required parasm")
        USAGE()
        exit(1)

    neighborhood_file = sys.argv[1]
    state_id = sys.argv[2]

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

            internal_ubication["state"] = state_id
            internal_ubication["street"] = extract_street_from_address(row["address"])
            internal_ubication["exterior_number"] = ""
            internal_ubication["internal_number"] = ""

            zones[row["address"]] = {
                "zone": row["zone"],
                "internal": internal_ubication,
            }

    print(json.dumps(zones, indent=4))
    with open("internal_zones.json", "w") as f:
        json.dump(zones, f)
