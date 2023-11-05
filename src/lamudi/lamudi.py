import sys
import os
current = os.path.dirname(os.path.realpath(__file__))
parent = os.path.dirname(current)
sys.path.append(parent)

import requests
import json
from src.message import generate_mensage

api_url = "https://api.proppit.com"

auth_token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VyIjp7ImlkIjoiZmFmYjA4NTEtZjlkOC00MGNkLTg0N2QtZDBjNzA2YzE2YjkxIiwicHVibGlzaGVySWQiOiJkODk3Y2RjYS03NjliLTQyNWItYmM5My0yYzNjM2Y4YjU2MjkiLCJhY2NlcHRlZFRlcm1zQW5kQ29uZGl0aW9ucyI6dHJ1ZSwiZW1haWwiOiJyZWJvcmFteEBnbWFpbC5jb20iLCJsYXN0UGFzc3dvcmRNb2RpZmljYXRpb25EYXRlIjoxNjg2NjA5MzQ5fSwiZXhwIjoxNjk5OTIxNDcxfQ.b5TLPFlEzti4XZRDzPYePnHAMO_GXPC1b9rjzO351v0"

cookies = {
    "authToken": auth_token
}

def send_message(lead_id, msg):
    msg_url = f"{api_url}/leads/{lead_id}/emails"

    data = {
        "leadId": lead_id,
        "message": msg
    }

    res = requests.post(msg_url, cookies=cookies, json=data)
    if (res.status_code != 201):
        print("ERROR: ", res.status_code)
        print(res.text)
        return None

    print(lead_id, "Mensaje enviado correctamente")

def get_lead_property(lead_id):
    property_url = f"{api_url}/leads/{lead_id}/properties"

    res = requests.get(property_url, cookies=cookies)
    if (res.status_code != 200):
        print("ERROR: ", res.status_code)
        print(res.text)
        return None

    print(res.cookies)

    props = res.json()["data"]
    formatted_props = []
    for p in props:
        formatted_props.append({
            "titulo": p["title"],
            "id": p['id'],
            "link": f"https://www.lamudi.com.mx/detalle/{p['id']}",
            "precio": p["price"]["amount"],
            "ubicacion": p["address"], #Direccion completa
            "tipo": p["propertyType"],
            "municipio": p["geoLevels"][0]["name"] if len(p["geoLevels"]) > 0 else "" #Solamente el municipio, lo usamos para generar el mensaje
        })

    return formatted_props
        
def get_lead_info(lead):
    lead_info = {
        "id": lead["id"],
        "fuente": "Lamudi",
        "nombre": lead["name"],
        "link": f"{api_url}/leads/{lead['id']}",
        "telefono_1": lead['phone'],
        "telefono_2": "",
        "email": lead['email'],
        "propiedad": get_lead_property(lead["id"])[0],
        #"busquedas": {
        #    "zonas": [i["name"] for i in busqueda["searched_locations"]["streets"]],
        #    "tipo": busqueda["lead_info"]["search_type"]["type"],
        #    "total_area": busqueda["property_features"]["total_area_xm2"]["min"] + " " + busqueda["property_features"]["total_area_xm2"]["max"],
        #    "covered_area": busqueda["property_features"]["covered_area_xm2"]["min"] + " " + busqueda["property_features"]["covered_area_xm2"]["max"],
        #    "banios": busqueda["property_features"]["baths"]["min"] + " " + busqueda["property_features"]["baths"]["max"],
        #    "recamaras": busqueda["property_features"]["bedrooms"]["min"] + " " + busqueda["property_features"]["bedrooms"]["max"],
        #    "presupuesto": busqueda["lead_info"]["price"]["min"] + " " + busqueda["lead_info"]["price"]["max"],
        #    "cantidad_anuncios": busqueda["lead_info"]["views"],
        #    "contactos": busqueda["lead_info"]["contacts"],
        #    "inicio_busqueda": busqueda["lead_info"]["started_search_days"] 
        #}
    }
    return lead_info

def get_leads(max=25):
    status = "new lead" #Filtramos solamente los leads nuevos
    page = 1
    limit = 25
    url = f"{api_url}/leads?_limit={limit}&_order=desc&_page={page}&_sort=lastActivity&status={status}"
    print(url)
    res = requests.get(url, cookies=cookies)

    if (res.status_code != 200):
        print("ERROR: ", res.status_code)
        print(res.text)
        return None
    data = res.json()["data"]

    total = data["totalFilteredRows"]
    print("total: ", total)

    leads = []
    while (len(leads) < total and (len(leads) < max or max == -1)):
        print("len: ", len(leads))
        leads += data["rows"]
        page += 1

        url = f"{api_url}/leads?_limit={limit}&_order=desc&_page={page}&_sort=lastActivity&status={status}"
        print(url)
        res = requests.get(url, cookies=cookies)

        if (res.status_code != 200):
            print("ERROR: ", res.status_code)
            print(res.text)
            return None
        data = res.json()["data"]

    print("Se encontraron", len(leads), "nuevos Leads")
    return leads

def main():
    leads = get_leads()

    leads_info = []
    for lead_res in leads:
        lead = get_lead_info(lead_res)

        msg = generate_mensage(lead)
        send_message(lead["id"], msg)

        leads_info.append(lead)

    with open('lamudi.json', 'r') as f:
        json.dump(leads_info, f)

if __name__ == "__main__":
    leads = get_leads()
    lead = get_lead_info(leads[0])

    print(json.dumps(lead, indent=4))

    msg = generate_mensage(lead)
    print(msg)

    send_message(lead["id"], msg)