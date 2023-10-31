import requests

auth_token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VyIjp7ImlkIjoiZmFmYjA4NTEtZjlkOC00MGNkLTg0N2QtZDBjNzA2YzE2YjkxIiwicHVibGlzaGVySWQiOiJkODk3Y2RjYS03NjliLTQyNWItYmM5My0yYzNjM2Y4YjU2MjkiLCJhY2NlcHRlZFRlcm1zQW5kQ29uZGl0aW9ucyI6dHJ1ZSwiZW1haWwiOiJyZWJvcmFteEBnbWFpbC5jb20iLCJsYXN0UGFzc3dvcmRNb2RpZmljYXRpb25EYXRlIjoxNjg2NjA5MzQ5fSwiZXhwIjoxNjk5OTIxNDcxfQ.b5TLPFlEzti4XZRDzPYePnHAMO_GXPC1b9rjzO351v0"

cookies = {
    "authToken": auth_token
}

def get_lead_property(lead_id):
    property_url = f"https://api.proppit.com/leads/{lead_id}/properties"

    res = requests.get(property_url, cookies=cookies)
    if (res.status_code != 200):
        print("ERROR: ", res.status_code)
        print(res.text)
        return None

    props = res.json()["data"]
    formatted_props = []
    for p in props:
        pass
        
def get_lead_info(lead):
    lead_info = {
        "nombre": lead["name"],
        "link": f"https://proppit.com/leads/{lead['id']}",
        "telefono_1": lead['phone'],
        "telefono_2": "",
        "email": lead['email'],
        "propiedad": {
            "nombre": data["posting"]["title"],
            "link": "",
            "precio": data["posting"]["price"]["amount"],
            "ubicacion": data["posting"]["address"],
            "tipo": data["posting"]["real_estate_type"]["name"],
        },
        "busquedas": {
            "zonas": [i["name"] for i in busqueda["searched_locations"]["streets"]],
            "tipo": busqueda["lead_info"]["search_type"]["type"],
            "total_area": busqueda["property_features"]["total_area_xm2"]["min"] + " " + busqueda["property_features"]["total_area_xm2"]["max"],
            "covered_area": busqueda["property_features"]["covered_area_xm2"]["min"] + " " + busqueda["property_features"]["covered_area_xm2"]["max"],
            "banios": busqueda["property_features"]["baths"]["min"] + " " + busqueda["property_features"]["baths"]["max"],
            "recamaras": busqueda["property_features"]["bedrooms"]["min"] + " " + busqueda["property_features"]["bedrooms"]["max"],
            "presupuesto": busqueda["lead_info"]["price"]["min"] + " " + busqueda["lead_info"]["price"]["max"],
            "cantidad_anuncios": busqueda["lead_info"]["views"],
            "contactos": busqueda["lead_info"]["contacts"],
            "inicio_busqueda": busqueda["lead_info"]["started_search_days"] 
        }
    }


def get_leads():
    res = requests.get("https://api.proppit.com/leads?_limit=25&_order=desc&_page=1&_sort=lastActivity", cookies=cookies)

    if (res.status_code != 200):
        print("ERROR: ", res.status_code)
        print(res.text)
        return None

    leads = res.json()["data"]["rows"]
    

leads = get_leads()
print(leads)