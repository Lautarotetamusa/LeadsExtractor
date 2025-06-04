from datetime import date, datetime
import json
import os
import sys

sys.path[0] = os.getcwd()

from src.casasyterrenos.casasyterrenos import CasasYTerrenos
from src.propiedades_com.propiedades import Propiedades
from src.lamudi.lamudi import Lamudi
from src.inmuebles24.inmuebles24 import Inmuebles24
from src.portal import Portal
from src.property import Internal

from commands.zones import ZONES

if __name__ == "__main__":
    PORTALS: dict[str, Portal ]= {
        "casasyterrenos": CasasYTerrenos(),
        "propiedades": Propiedades(),
        "inmuebles24": Inmuebles24(),
        "lamudi": Lamudi(),
    }

    args = sys.argv

    if len(args) < 2:
        print("portal its required")
        exit(1)

    portal_name = args[1]
    if portal_name not in PORTALS.keys():
        print(f"Portal {portal_name} doesnt exists")
        exit(1)

    with open("internal_zones.json", "r") as f:
        internal_zones: dict[str, Internal] = json.load(f).get(portal_name, {})

    portal = PORTALS[portal_name]
    rows = []

    for zone in ZONES:
        internal = internal_zones.get(zone)
        if internal is None:
            print(zone, "not internal")
            rows.append('0')
            continue

        count = 0
        for prop in portal.get_properties(query={"internal": internal}, featured=False):
            # 2025-05-28 13:04:44
            if "last_update" in prop:
                last_update = datetime.strptime(prop["last_update"], '%Y-%m-%d %H:%M:%S')
                if abs(datetime.today() - last_update).days > 1:
                    print("propiedad publicada hace mas de 1 dia")
                    break
            count += 1

        rows.append(f"{count}")

        print(zone, "count: ", count)

    with open("validation.txt", "w") as f:
        f.write('\n'.join(rows))
