import requests

updated_since = "2024-11-11T00:00:00Z"
updated_until = "2024-11-12T00:00:00Z"
format_url = f"https://api.pipedrive.com/api/v2/deals?updated_since={updated_since}&updated_until={updated_until}"
format_url += "&cursor={cursor}"

headers = {
    "api_token": "e6f1603d24d49769c9521c9b8a42d6107ca740aa",
    "x-api-token": "e6f1603d24d49769c9521c9b8a42d6107ca740aa"
}

cursor = ""
while cursor is not None:
    url = format_url.format(cursor=cursor)
    print(url)
    res = requests.get(url, headers=headers)
    if not res.ok:
        print(res.json())
        exit(1)

    data = res.json()
    cursor = data["additional_data"]["next_cursor"] 
    print(len(data["data"]))
