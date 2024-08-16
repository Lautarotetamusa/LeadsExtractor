import requests

img_url = "https://maps.google.com/maps/api/staticmap?center=19.434220199999998612838680855929851531982421875%2C-99.145658499999996138285496272146701812744140625&zoom=16&key=AIzaSyBfTOxuYEmbL50_1P4KU8GI8toKT539agI&size=780x456&sensor=true&scale=2&channel=rplis-i24&markers=19.4342202%2C-99.1456585"

img_data = requests.get(img_url).content

url = "https://www.jotform.com/API/sheets/242244461116044/form/242244461116044/submission/5994143230122633259/files?from=sheets"

payload = {'submissionID': '5994143230122633259', 'qid': '4'}
files=[
  ('uploadedFile[]',('Untitled3.jpg', img_data ,'image/jpeg'))
]

headers = {
  'Origin': 'https://www.jotform.com',
  'Referer': 'https://www.jotform.com/tables/242244461116044',
  'Cookie': '_gj=f61b1aa461bf9b4097b50d92eeaedf345f17fd4a; hblid=idwBHlnuFAtmmILH0v5z60SC12AOKoCK; olfsk=olfsk9475756384112493; FORM_last_folder=allForms; g_state={"i_p":1724073586229,"i_l":3}; SHEET_last_folder=sheets; guest=guest_e5687c44184f0a37; JOTFORM_SESSION=707d4c49-983f-d638-37a4-9c170007; userReferer=https%3A%2F%2Fwww.jotform.com%2F; JF_SESSION_USERNAME=Diego_torres; last_edited_v4_form=242244461116044; DOCUMENT_last_folder=documents; limitAlignment=left_alt; wcsid=coitE2SsUMU2TjBT0v5z60S1CBCOKK1A; _oklv=1723653966572%2CcoitE2SsUMU2TjBT0v5z60S1CBCOKK1A; navLang=en-US; _okdetect=%7B%22token%22%3A%2217236539461720%22%2C%22proto%22%3A%22about%3A%22%2C%22host%22%3A%22%22%7D; _okbk=cd4%3Dtrue%2Cvi5%3D0%2Cvi4%3D1723653946530%2Cvi3%3Dactive%2Cvi2%3Dfalse%2Cvi1%3Dfalse%2Ccd8%3Dchat%2Ccd6%3D0%2Ccd5%3Daway%2Ccd3%3Dfalse%2Ccd2%3D0%2Ccd1%3D0%2C; _ok=4728-686-10-5570; JF_SESSION_USERNAME=Diego_torres; guest=guest_e70346448f994c09; userReferer=https%3A%2F%2Fwww.jotform.com%2Ftables%2F242244461116044'
}

response = requests.request("POST", url, headers=headers, data=payload, files=files)
print(response.text)
