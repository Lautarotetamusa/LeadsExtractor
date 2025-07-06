import os
import sys
import json
sys.path[0] = os.getcwd()

import webbrowser
from src.cotizadorpdf.cotizador import to_pdf


def cotizador(args: list[str]):
    with open("./src/cotizadorpdf/example.json", "r") as f:
        test_data = json.load(f)
    if len(args) > 0 and args[0] == '--extras':
        test_data["extras"] = []
    # print(json.dumps(test_data, indent=4))

    timestamp = to_pdf(test_data)
    if (timestamp == "error"):
        print("error creating the cotizacion")
        exit(1)

    base_path = os.getcwd()
    html_path = os.path.join(base_path, "pdfs", f"cotizacion{timestamp}.html")
    webbrowser.open(f"file://{html_path}")


if __name__ == "__main__":
    cotizador(sys.argv[1:])
