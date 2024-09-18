import sys

sys.path.append('.')
from src.inmuebles24.inmuebles24 import Inmuebles24

if __name__ == "__main__":
    inmuebles24 = Inmuebles24()

    inmuebles24.send_message("466554403", "Mensaje de prueba")
