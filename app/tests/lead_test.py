import sys
sys.path.append('..')

from src.lead import Lead

if __name__ == "__main__":
    lead = Lead()

    lead.set_args({
        "nombre": "Lauti",
        "telefono": "3415582213"
    })
    
    values = {k: v for k, v in lead.__dict__.items() if v != "" and type(v) is not dict}
    print(values)
