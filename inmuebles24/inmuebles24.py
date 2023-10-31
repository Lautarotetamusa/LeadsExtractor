import requests

def get_req(url):
	site_url = "http://www.inmuebles24.com/"

	cookies ={
		"__cf_bm": "BE9PAOj.RDyc1azKak1DtHAkGvSa.ZJPWELH7ws0PI8-1698619476-0-AWJ8DwPIF9Y+3AsDNpSlHIs/JTNMPffXqwnsGSluThjLK1ZG5FOF9JzuXN/dzKFNAe/g63bTIB0kt2R5A+EQ3OXYJSyxvSTYLtp7XYf9g2nT",
		"cf_clearance": "za9dVMWY0XVeHHhQdFyv60WmEOlY_ogtjaeAa8pt2BA-1698436283-0-1-115e9ea5.ba12bc.465fc145-160.2.1698436283",
		"g_state": "{\"i_p\":1699030291139,\"i_l\":3}",
		"hashKey": "4IVc5it99duNGjgDajK2gWBqUKofpMi62d6jqt/V2c7+AfyJw5aUsaJtdgqD1oJORGYcG6pBna3NhUrJvs1aQz7aL2pcBFXZQygv",
		"hideWelcomeBanner": "true",
		"JSESSIONID": "76F9E9D2456D619424F0595CA4873A5A",
		"sessionId": "0a998bcf-d151-4211-8557-24f640491048",
		"usuarioFormApellido": "Aquitectos",
		"usuarioFormEmail": "ventas.rebora@gmail.com",
		"usuarioFormId": "57036554",
		"usuarioFormNombre": "Rebora",
		"usuarioFormTelefono": "3313420733",
		"usuarioIdCompany": "50796870",
		"usuarioLogeado": "ventas.rebora@gmail.com",
		"usuarioPublisher": "true"
	}
    
	headers = [
		{
			"name": "Accept",
			"value": "application/json, text/plain, */*"
		},
		{
			"name": "Accept-Encoding",
			"value": "gzip, deflate, br"
		},
		{
			"name": "Accept-Language",
			"value": "en-US,en;q=0.5"
		},
		{
			"name": "Connection",
			"value": "keep-alive"
		},
		{
			"name": "Host",
			"value": "www.inmuebles24.com"
		},
		{
			"name": "Referer",
			"value": "https://www.inmuebles24.com/panel/interesados"
		},
		{
			"name": "Sec-Fetch-Dest",
			"value": "empty"
		},
		{
			"name": "Sec-Fetch-Mode",
			"value": "cors"
		},
		{
			"name": "Sec-Fetch-Site",
			"value": "same-origin"
		},
		{
			"name": "sessionId",
			"value": "0a998bcf-d151-4211-8557-24f640491048"
		},
		{
			"name": "TE",
			"value": "trailers"
		},
		{
			"name": "User-Agent",
			"value": "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/111.0"
		},
		{
			"name": "x-panel-portal",
			"value": "24MX"
		}
	]

	obj_headers = {h["name"]: h["value"] for h in headers}
	print(obj_headers)

	res = requests.get(site_url+url, headers=obj_headers, cookies=cookies, verify=False)
	if res.status_code != 200:
		print("Status code is not 200: ", res.status_code)
		#print(res.text)
		return None

	print("Good response:", res.status_code)
	print(res.text)
	return res.json()

def get_lead_info(lead_id: str) -> object:
    data_url = f"leads-api/publisher/contact/{lead_id}"
    busqueda_url = f"leads-api/publisher/contact/{lead_id}/user-profile"

    data = get_req(data_url)
    busqueda = get_req(busqueda_url)

    if data == None or busqueda == None:
        return None

    lead_info = {
        "nombre": data["lead_user"]["name"],
        "link": f"https://www.inmuebles24.com/panel/interesados/{lead_id}",
        "telefono_1": data["phone_list"][0],
        "telefono_2": data["phone_list"][1],
        "email": data["lead_user"]["email"],
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

    return lead_info

#get_lead_info("182835939")
res = get_req("https://www.inmuebles24.com/avisos-api/userInfo")
print(res)

#get_by_selenium()