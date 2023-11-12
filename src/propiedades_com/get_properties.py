import requests
import json
from zenrows import ZenRowsClient

COOKIES = {
    "__cf_bm": "PYDGMqr4dIhg2ucMHU3c3As5lmVwLfXUkrPOmnutx1g-1699729811-0-ATf5YOqQ8v+pIsk7awJmnAseDSPr3ipZ/D89VR8ekk2cn0UrYKAe6ayFzMKincZ/vYls9XzWuA6ziJIIre57Ro4=",
    "66169bb371957cc42c524c330e24735f": "f70ca332f453d019c8627a0a90c5d88bab74a646s:101:\"bcf2b13593ea44ac04eeba5154a3d802d643b603a:4:{i:0;i:1394158;i:1;s:6:\"Rebora\";i:2;i:2592000;i:3;a:0:{}}\";",
    "cf_clearance": "L4Mkl5oqO2c9UfzDIk00XDwxVvTPaBwV4I29RjTUm7g-1699726338-0-1-49d9c0fb.19fad274.1ed89f1b-160.2.1699726338",
    "G_ENABLED_IDPS": "google",
    "g_state": "{\"i_p\":1699726302265,\"i_l\":1}",
    "GCLB": "CNeluaaHw8WABQ",
    "PHPSESSID": "casvj8afk28t00li381aomvnb1",
    "userToken": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJqdGkiOiJNVFk1T1RjeE9URXdNakl3TnpNMk9UQTRNamM9IiwiaWF0IjoxNjk5NzE5MTAyLCJpc3MiOiJQcm9waWVkYWRlcy5jb20iLCJleHAiOjE3MzEyNTUxMDIsImRhdGFVc2VyIjp7ImlkIjoiMTM5NDE1OCIsIm5hbWUiOiJSZWJvcmEiLCJsYXN0bmFtZSI6IlRvcnJlcyIsImZ1bGxfbmFtZSI6IlJlYm9yYSBUb3JyZXMiLCJlbWFpbCI6InBvcnRhbGVzcmVib3JhbXhAZ21haWwuY29tIiwicGhvbmUiOiIzMzEzNDIwNzMzIiwiZW1haWxfY29udGFjdCI6InZlbnRhcy5yZWJvcmFAZ21haWwuY29tIiwicHJvZmlsZVBpY3R1cmUiOiI5NGEyODg0ZWQxMTFiNmYyZDNhMjQ1ZjE2ODhlNTRmMS5wbmciLCJjcmVhdGVkIjoiMTBcLzExXC8yMCIsIm5vdGlmaWNhdGlvbnMiOjEsInBpY3R1cmVfdXJsIjoiaHR0cHM6XC9cL3Byb3BpZWRhZGVzY29tLnMzLmFtYXpvbmF3cy5jb21cL2ZpbGVzXC9wcm9maWxlc1wvOTRhMjg4NGVkMTExYjZmMmQzYTI0NWYxNjg4ZTU0ZjEucG5nIiwiaXNfYWRtaW4iOiIwIiwicm9sZV9pZCI6IjEiLCJ1c2VyX3R5cGUiOjEsInZlcnNpb25fbG9naW4iOjF9fQ.AjyHN-hfePLbkE1KnFI8u84yWZUXjc5cfBIQZW-xQR0"
}

APIKEY = "f569290d8d48d2d19a288ff2a95cad9b6679ab4d"

client = ZenRowsClient(APIKEY)

params = {
    "js_render": "true",
	#"resolve_captcha": "true",
    "antibot": "true",
    "premium_proxy":"true",
    "proxy_country":"mx",
	"session_id": 1,
    "json_response": 1,
    "wait": 5
}

def post_req(url, data):
	res = client.post(url, params=params, cookies=COOKIES, data=data)
	if (res.status_code < 200 or res.status_code > 299):
		print("no se pudo realizar la request a la url: "+url)
		print("STATUS: "+str(res.status_code))
		print("RESPONSE:", res.text)
		try:
			json.dumps(res.json(), indent=4)
		except Exception as e:
			print(res.text)
		return None
	return res

#Obtener la informacion de la propiedad
def main():
    property_url = f"https://propiedades.com/api/v3/property/MyProperties"
    data = {
        "source": "web",
        "identifier": "11",
        "purpose": "5",
        "highlighted": "0",
        "page": "1"
    }

    total_pages = 2

    with open("src/propiedades.com/properties.json", "r") as f: 
        props = json.load(f)

    while int(data["page"]) < total_pages:
        print("Pagina: ", data["page"])

        data = post_req(property_url, data=data)
        #res = requests.post(property_url, cookies=COOKIES, data=data)
        #if res.status_code != 200:
        #    print(f"ERROR obteniendo la pagina {data['page']}")
        #    print(res.status_code)
        #    print(res.text)
        #    exit(1)
        #data = res.json()
        
        props += data["data"]["properties"]

        print(f"{len(props)} cantidad de propiedades extraidas")

        data["page"] += 1

        total_pages = data["paginate"]["pages"]

    with open("src/propiedades.com/properties.json", "w") as f: 
        json.dump(props, f, indent=4)

if __name__ == "__main__":
    #main()

    with open ("src/propiedades.com/properties.json") as f:
        props = json.load(f)

        print("len: ", len(props))

    indexed_props = {}

    for prop in props:
        indexed_props[prop["id"]] = prop

    with open ("src/propiedades.com/properties_obj.json", "w") as f:
        json.dump(indexed_props, f, indent=4)