import json
import requests
import http.client

ZENROWS_API_URL = "https://api.zenrows.com/v1/"
ZENROWS_API_KEY = "f569290d8d48d2d19a288ff2a95cad9b6679ab4d" 

session = requests.Session()

url_get = "https://www.inmuebles24.com/publicarPasoDatosPrincipales.bum"

cookies = {
    "changeInbox": "true",
    "crto_is_user_optout": "false",
    "crto_mapped_user_id": "P_pg8F9lYnVlZnhkMER6U0w3S3UlMkZGMVcyS0dYMVUlMkJEd3BNS2RMUlZOMERMRWhpSSUzRA",
    "cto_bundle": "CgcyAl9KTmt0RFpmdWd5SVNFcEtjcThGb25QS1lCSlRGcmZZYzN4QSUyQlRxSTdlYXB5dDVFejN2RGFqWm5yRVZGUERYa1EybzJjMW1lMCUyRmtFSnpRRUhuJTJGc3hYdzB0VG1lQ0lZblloMnVFJTJCY2hiODFNSjJ6RFY4Yk0zRjIzUFp6VnlpT1MxY2ElMkJpUDNDNWFMaTk1Rzk0bHFFcG5nJTNEJTNE",
    "G_ENABLED_IDPS": "google",
    "gtm_upi": "4acbf60eb01da91b0b174cb7e71f597c3256d51442697fcde697911ca0f7b861",
    "hashKey": "4IVc5it99duNGjgDajLnhm45CaoNo8S62tCsqt/e2c72AfyJx5CUsaJtcQqDwYJORGYcHapBna3NhUpHWn3kTrlS9trawnalsrPR",
    "hideWelcomeBanner": "true",
    "IDusuario": "57036554",
    "JSESSIONID": "B676855F36B08A48DF6D89A26E6A299F",
    "mousestats_vi": "fc98ac1a434c881e4202",
    "owneremail": "sergio.esqueda@inmuebles24.com",
    "ownerIdempresa": "50796870",
    "ownername": "Sergio Esqueda",
    "pasoExitoso": "{'trigger':'Boton del paso'&'stepId':1&'stepName':'Datos principales-Profesional'}",
    "reputationModalLevelSeen": "true",
    "reputationModalLevelSeen2": "true",
    "reputationTourLevelSeen": "true",
    "sessionId": "4bb40dbf-1816-4cdd-9a1a-0eb011e9b71b",
    "showWelcomePanelStatus": "true",
    "tableCreditsOpen": "false",
    "usuarioFormApellido": "Residences",
    "usuarioFormEmail": "control.general@rebora.com.mx",
    "usuarioFormId": "57036554",
    "usuarioFormNombre": "RBA",
    "usuarioFormTelefono": "523341690109",
    "usuarioIdCompany": "50796870",
    "usuarioLogeado": "control.general@rebora.com.mx",
    "usuarioPublisher": "true"
}

PARAMS = {
    "url": url_get,
    "apikey": ZENROWS_API_KEY,
    # "js_render": "true",
    "antibot": "true",
    "premium_proxy": "true",
    "proxy_country": "mx",
    "custom_headers": "true",
    # "session_id": "2",
    "original_status": "true",
    # "autoparse": "true",
    # "screenshot": "true"
}
url_post = "https://www.inmuebles24.com/publicarPasoDatosPrincipales.bum"

def upload_images(prop_id: str):
    upload_url = f"https://www.inmuebles24.com/avisoImageUploader.bum?idAviso={prop_id}"

    with open("image.png", "rb") as f:
        img_data = f.read()

    files = [
        ('name', ("image.png")),
        ('file', ("image.png", img_data, f"image/png")),
    ]

    print(len(img_data))

    headers = {
        "sessionId": "1e2fa26b-2b9a-4297-bc7d-09f45280fe22",
        "idUsuario": "57036554",
    }

    PARAMS["url"] = upload_url
    res = requests.post(
        ZENROWS_API_URL,
        params=PARAMS,
        files=files,
        cookies=cookies,
        headers=headers
    )

    if not res.ok:
        print(res.text)
        print("error uploading the image")
        return

    file_id = res.text.replace("FILEID:", "")
    print(file_id)

prop_id = "146279428"
add_image_url = f"https://www.inmuebles24.com/publicarPasoMultimedia.bum?idaviso={prop_id}"
uploaded_images = [
    "https://storage.googleapis.com/rp-tmp-images/avisos/18/01/46/27/94/28/180x140/temp_aviso_146279428_19abf147-64a9-46fe-a5d0-f09a02dc865b.jpg",
    "https://storage.googleapis.com/rp-tmp-images/avisos/18/01/46/27/94/28/180x140/temp_aviso_146279428_b251e206-9a37-4fdd-8336-4a9d024e49d5.jpg",
    "https://storage.googleapis.com/rp-tmp-images/avisos/18/01/46/27/94/28/180x140/temp_aviso_146279428_7e40dfee-c88a-4ddd-b699-b4c0b7d23009.jpg",
    "https://storage.googleapis.com/rp-tmp-images/avisos/18/01/46/27/94/28/180x140/temp_aviso_146279428_0502bbf0-ed87-4da6-aa79-b580d573a8db.jpg",
    "https://storage.googleapis.com/rp-tmp-images/avisos/18/01/46/27/94/28/180x140/temp_aviso_146279428_327a8d22-2292-49cd-8209-621a9363ca08.jpg"
]
payload = {
    "multimediaaviso.idMultimediaAviso[]": [],
    "idAviso": prop_id,
    "irAlPaso": "",
    "checkContentEnhancerHidden": "",
    "guardarComoBorrador": "1",
    "mutimediaaviso.keyRecorrido360": "",
    "mutimediaaviso.idRecorrido360": "-1",
    "mutimediaaviso.idRecorrido360Multimedia--1": ""
}
i = 0
for img_url in uploaded_images:
    parts = img_url.split("/")
    img_id = parts[len(parts)-1]

    id = f"nueva_{i}"
    payload[f"multimediaaviso.orden[{id}]"] = str(i+1)
    payload[f"multimediaaviso.rotacion[{id}]"] = "0"
    payload[f"multimediaaviso.urlTempImage[{id}]"] = img_id
    payload[f"multimediaaviso.titulo[{id}]"] = ""
    payload["multimediaaviso.idMultimediaAviso[]"].append(id)

    i += 1 

print(json.dumps(payload, indent=4))


headers = {
    "sessionId": "1e2fa26b-2b9a-4297-bc7d-09f45280fe22",
    "idUsuario": "57036554",
}

cookies = {
		"__cf_bm": "Q_CwlogebfGIOhs4SlKBH6NEwN3HWoWXWMxn7WvYxb8-1745524733-1.0.1.1-_RIlaWL1R4zvKS75HO773o66g7LIPLyujePA1YaXU5x1ynGbIYK4sAM7ViGAkGQiScYviD8PIRcNnL.lM.7n_ruFu5zDkJ0Kt1R7v8ynahHt3XlwLa7hX_1oPXdUU58E",
		"__eoi": "ID=2704e4dae7a16247:T=1740840824:RT=1745524143:S=AA-AfjaxN22Otx8r48YIXZahuVHV",
		"__gads": "ID=b0222fc75dd6049e:T=1732733731:RT=1745524143:S=ALNI_MbrlZ7f4JbJbpscK2l80pYXN9vmbQ",
		"__gpi": "UID=00000fa10b36ff45:T=1732733731:RT=1745524143:S=ALNI_MZ7o18CBCIuVVAToBXE_BF96clQSA",
		"__rtbh.lid": "{\"eventType\":\"lid\",\"id\":\"mT2RKa8aWnPnBkTSoKky\",\"expiryDate\":\"2026-04-24T19:59:06.662Z\"}",
		"__rtbh.uid": "{\"eventType\":\"uid\",\"expiryDate\":\"2026-04-23T10:59:59.299Z\"}",
		"_cfuvid": "WzmR0p456pCrywFgdbX.j0FGiHBBFmeUGNVhQZtnA4s-1745416819430-0.0.1.1-604800000",
		"_fbp": "fb.1.1698609869726.1255128104",
		"_ga": "GA1.1.1232673381.1698609867",
		"_ga_8XFRKTEF9J": "GS1.1.1745523547.186.1.1745524760.60.0.0",
		"_ga_F842TPK3EE": "GS1.2.1721808563.54.1.1721812127.60.0.0",
		"_ga_G2SXGX9QG3": "GS1.1.1745523547.101.1.1745524745.15.0.0",
		"_gcl_au": "1.1.231706276.1739461413",
		"_gid": "GA1.2.186623425.1745357690",
		"_hjHasCachedUserAttributes": "true",
		"_hjMinimizedPolls": "987274",
		"_hjSession_174024": "eyJpZCI6ImU0YzcxZWYxLTNkZjktNGVkYy1hN2U0LWVkMWVjOWJjM2IzYSIsImMiOjE3NDU1MjM1NTExMjUsInMiOjAsInIiOjAsInNiIjowLCJzciI6MCwic2UiOjAsImZzIjowLCJzcCI6MX0=",
		"_hjSessionUser_174024": "eyJpZCI6IjhkMDg1YTY3LWM5NTAtNWY2Zi1hMDM3LTg2OTViY2YzZTgyYSIsImNyZWF0ZWQiOjE2OTg2MDk4Njg0ODksImV4aXN0aW5nIjp0cnVlfQ==",
		"_tt_enable_cookie": "1",
		"_ttp": "01JNTQA21732GYDGPGERQVHJ5M_.tt.1",
		"_uetsid": "9abcae901fc111f0aef3092325dd27ec",
		"_uetvid": "3f1efd60fc0e11efa0d177f0d40a3aa5",
		"allowCookies": "true",
		"cf_clearance": "ZhTXPjMtD_KUyQJqENf_zNe1L0kNKNe_5FQO9s56_nk-1745524496-1.2.1.1-slSoiKyKWHObJyDr4meomdsDj3Sdo.r7nunqJaiDbxbkPLaRnI5pKfONvJSc_ZaIY2HAxYRPnK2IJuEIIgNUQRP5C3jdILcEbK4OUkHl88vy_5uUcsNsmU5mMvXIUiQjlMd6aI0ED.PaQa1vOiGXmsBC8tSHXavd_wBHojIGUPBxAWNffXxWX9chyUD7SpCEdOLxOpcvsX0d8Im2IM.4AeRHp_0Wuh09VfYsNyirmllGaP7UI8FMwpAvNDaeARhONoqYeoVa3Ol_BPohzIxrBH3Owjx..eE6qaSmbaTRi.T8c7GislNXGGH0is.FxDLeNom2s5DQhia0.4vm.j_8pqd4svxCmcZslpF7zGGwNl_GYkpJkFWFOYX_YOi8miPv",
		"changeInbox": "true",
		"crto_is_user_optout": "false",
		"crto_mapped_user_id": "oSHmp19lYnVlZnhkMER6U0w3S3UlMkZGMVcyS0E2dHR0Q1M0dXRIY1NpJTJGJTJCTTFCc29zJTNE",
		"cto_bundle": "69R_nF9KTmt0RFpmdWd5SVNFcEtjcThGb25PbVBJUHRpMUN1NmJiNlhMbDZCaGxwa0JHVndmYkNFTzhrZXo5bkFsdVJjQnVjN1pWWDFETUVwbTJ1THBLTFlHSWtUbXJ0akQ1akxmUHhWRjZOQnRIeE42VjhNTU9yOElDbzBoYjN4Q1BKekJUR2ZwVkxOalVmU0VRZ1NwRCUyRll0dyUzRCUzRA",
		"email": "2025-04-23T20:50:53-03:00",
		"G_ENABLED_IDPS": "google",
		"gtm_upi": "4acbf60eb01da91b0b174cb7e71f597c3256d51442697fcde697911ca0f7b861",
		"hashKey": "4IVc5it99duNGjgDajLnhm45CaoYsNm62tCsqt/Z2c7zAfyJwpuUsaJudgqDwYJORGYcHapBna3NhUqFC4OuAkRNSabk9RD15xxO",
		"hideWelcomeBanner": "true",
		"idUltimoAvisoVisto": "146279327",
		"IDusuario": "57036554",
		"JSESSIONID": "C5B555C41E2F910606C6000E9F3D3429",
		"mousestats_vi": "fc98ac1a434c881e4202",
		"owneremail": "sergio.esqueda@inmuebles24.com",
		"ownerIdempresa": "50796870",
		"ownername": "Sergio Esqueda",
		"pasoExitoso": "{'trigger':'Guardar como borrador'&'stepId':3&'stepName':'Multimedia-Profesional'}",
		"phoneCall": "2025-04-23T19:05:05-03:00",
		"reputationModalLevelSeen": "true",
		"reputationModalLevelSeen2": "true",
		"reputationTourLevelSeen": "true",
		"sessionId": "4bb40dbf-1816-4cdd-9a1a-0eb011e9b71b",
		"showWelcomePanelStatus": "true",
		"tableCreditsOpen": "false",
		"ttcsid": "1745523551368::VJgePHhXmjz5Q6QZauC2.9.1745524746865",
		"ttcsid_CUUATCJC77U6JIU4VTF0": "1745523551368::h2Ox5PM8jBl4fQu2MBU3.9.1745524768031",
		"usuarioFormApellido": "Residences",
		"usuarioFormEmail": "control.general@rebora.com.mx",
		"usuarioFormId": "57036554",
		"usuarioFormNombre": "RBA",
		"usuarioFormTelefono": "523341690109",
		"usuarioIdCompany": "50796870",
		"usuarioLogeado": "control.general@rebora.com.mx",
		"usuarioPublisher": "true",
		"whatsapp": "2025-04-24T02:26:56-03:00"
	}

# PARAMS["url"] = "https://www.inmuebles24.com/publicarPasoMultimedia.bum?idaviso=146279327&checkContentEnhancerUrl=true" 
# res = session.get(
#     ZENROWS_API_URL,
#     params=PARAMS,
#     headers=headers,
#     cookies=cookies,
# )
# print(res.text)
# print(res.status_code)

PARAMS["url"] = add_image_url
print(json.dumps(PARAMS, indent=4))
res = session.post(
    ZENROWS_API_URL,
    params=PARAMS, 
    data=payload, 
    cookies=cookies, 
    headers={
        **headers,
        # "Referer": "https://www.inmuebles24.com/publicarPasoMultimedia.bum?idaviso=146279327&checkContentEnhancerUrl=true&checkContentEnhancerUrl=true",
        # "User-Agent": "Mozilla/5.0 (X11; Linux x86_64; rv:130.0) Gecko/20100101 Firefox/130.0",
        # "Content-Type": "application/x-www-form-urlencoded"
    }
)

print(res.text)
print(res.status_code)
