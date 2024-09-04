# pip install requests
import requests

url = 'https://www.inmuebles24.com/leads-api/publisher/contact/status/209874137'
apikey = 'f569290d8d48d2d19a288ff2a95cad9b6679ab4d'
params = {
    'url': url,
    'apikey': apikey,
	'premium_proxy': 'true',
	'proxy_country': 'mx',
	'custom_headers': 'true',
	'autoparse': 'true',
}
headers = {
	'idUsuario': '57036554',
	'x-panel-portal': '24MX',
	'content-type': 'application/json;charset=UTF-8',
	'sessionId': '86a09511-65b6-4605-bc4b-a59f174e9ea8',
	'Cookie': 'g_state={\"i_p\":1726000813426,\"i_l\":4}; hideWelcomeBanner=true; _ga_8XFRKTEF9J=GS1.1.1725401521.121.1.1725402054.60.0.0; _ga=GA1.2.1232673381.1698609867; __rtbh.lid=%7B%22eventType%22%3A%22lid%22%2C%22id%22%3A%22mT2RKa8aWnPnBkTSoKky%22%7D; _ga_F842TPK3EE=GS1.2.1721808563.54.1.1721812127.60.0.0; _hjSessionUser_174024=eyJpZCI6IjhkMDg1YTY3LWM5NTAtNWY2Zi1hMDM3LTg2OTViY2YzZTgyYSIsImNyZWF0ZWQiOjE2OTg2MDk4Njg0ODksImV4aXN0aW5nIjp0cnVlfQ==; _fbp=fb.1.1698609869726.1255128104; mousestats_vi=fc98ac1a434c881e4202; __gads=ID=675995fc19c9ffb7:T=1698775996:RT=1725375129:S=ALNI_MZ-BntMcU8J1M_QDn1lvyYKlRrFzg; __gpi=UID=00000a3be27e374b:T=1698775996:RT=1725375129:S=ALNI_Ma4N5plSE60bXXTWOi2C-fxi8-C-w; __rtbh.uid=%7B%22eventType%22%3A%22uid%22%7D; G_ENABLED_IDPS=google; IDusuario=57036554; __eoi=ID=c1fa39300bf65ec0:T=1724788681:RT=1725375129:S=AA-AfjZiGAvXiOO2nGGNx4bkaPle; _hjMinimizedPolls=987274; _gcl_au=1.1.670376324.1719771146; sessionId=86a09511-65b6-4605-bc4b-a59f174e9ea8; usuarioFormNombre=Rebora; usuarioFormEmail=marketing%40rebora.com.mx; usuarioFormTelefono=523341690109; usuarioFormId=57036554; usuarioFormApellido=Aquitectos; usuarioFormId=57036554; _ga_G2SXGX9QG3=GS1.1.1725401521.37.1.1725401596.59.0.0; showWelcomePanelStatus=true; allowCookies=true; _cfuvid=do0faQ2MP9F6mr6B1uZ1wlPrQ_mtRhHkAV8Z5RRR1t0-1725375150337-0.0.1.1-604800000; JSESSIONID=9A8D1FB00B09F4B60E486152F3D9FCE9; usuarioLogeado=marketing@rebora.com.mx; hashKey=4IVc5it99duNGjgDajLnhm45CaoctMm6xNSlqt3Z2c7+AfyJx5WUsaJsdgqDwYJORGYcHKpInaLLjRUglQ/DrUelQQPM75RfncihfZpg; usuarioPublisher=true; usuarioIdCompany=50796870; reputationModalLevelSeen=true; _hjHasCachedUserAttributes=true; _gid=GA1.2.1700334628.1725364372; email=2024-09-03T18:03:51-03:00; phoneCall=1970-01-01T21:00:00-03:00; whatsapp=1970-01-01T21:00:00-03:00; showFeedbackPublisher=true; __cf_bm=5ABv1uBSFokjTX7F6zXtHez4L9qcXGzC8iAUUHQAByQ-1725412322-1.0.1.1-SDBBABhwMPXK_dF2VqMYv9H4e2CnT_LaEV6nO2GHdPWvCfYBxFu5LL2gvDMfyJeM7tCZiEP9sLL01zgFaWA1PWr6SUCDAdbuyO3axcBtKZY; cf_clearance=1vm1ck_EDzAhJ0BksrKYklxfT.dlRKF4fxPKciRHyMs-1725412323-1.2.1.1-Miw3ono.njxmn07rboI1ZRxqMKGp0mBRILqmeisLQhlTzI0qhFZ45WaeNBM57gfJvX2UbohcoFIausIxNNyDP_Bn__SGF1QYFXBvPqZxhfLizmydw47mONo453w4P0ean7b7TSMDhzAmThS9fX7HogT1gVH5bp4Nl0eHve2BKMDNNAtpuNYFam._Es4IIFY9nk9CYGrgKlDs3JUvGTTP_hr1b65q1naX3q3Xp495yPMmgcPn7uCqavB18.ECnfS7ycj2J46mFz.3q3L8hz3S945xlw7TlUezI7lchha__t8_Fqi.QOkw06EGCpHxnMDJGW0Esm6fTzosI92g43f9fff4yIWUXfYlmeN2Wuni3Qat_K0MYdv9YeiYPMWOwW91; _hjSession_174024=eyJpZCI6Ijk0ZjM0M2Q2LTE1ZDYtNDgyMy04NmNjLTRmYzYxNjMyNWQxOCIsImMiOjE3MjU0MDE1MjQyMTAsInMiOjAsInIiOjAsInNiIjowLCJzciI6MCwic2UiOjAsImZzIjowLCJzcCI6MH0=; otros=1970-01-01T21:00:00-03:00; _dc_gtm_UA-10030475-1=1',
    "User-Agent": "Mozilla/5.0 (X11; Linux x86_64; rv:129.0) Gecko/20100101 Firefox/129.0"
}
data = {
	'lead_id': '465984909',
	'lead_status_id': 2,
}
response = requests.post('https://api.zenrows.com/v1/', params=params, headers=headers, data=data)
print(response.text)
