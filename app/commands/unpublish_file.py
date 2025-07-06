# Unpublish all the ids on a txt file
# TODO: add the others portals
# TODO: Combine the unpublish functions

import os
import sys
sys.path[0] = os.getcwd()

from src.inmuebles24.inmuebles24 import Inmuebles24
from src.portal import Portal

if __name__ == "__main__":
    with open("../ids.txt", "r") as file:
        lines = [line.rstrip() for line in file]

    i24 = Inmuebles24()

    ids = []
    i = 0
    for id in lines:
        print(id)

        ids.append(id)

        if len(ids) == 20 or i == len(lines)-1:
            print(ids)
            err = i24.unpublish(ids)
            if err is not None:
                print(err)

            ids = []
        i += 1
