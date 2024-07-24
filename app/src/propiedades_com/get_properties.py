import os
import requests
import json

import sys
sys.path.append('.')
from src.logger import Logger

URL = "https://obgd6o986k.execute-api.us-east-2.amazonaws.com/prod/get_properties_admin?page={page}&country=MX&identifier=4&purpose=3&order=update_desc"

logger = Logger("propiedades")

headers = {
    "x-api-key": "VAryo1IvOF7YMMbBU51SW8LbgBXiMYNS7ssvSyKS",
    "Authorization": "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJqdGkiOiJNVGN5TVRneU16QXpNRFkxT0RjMk5UYzBOZz09IiwiaWF0IjoxNzIxODIzMDMwLCJpc3MiOiJQcm9waWVkYWRlcy5jb20iLCJleHAiOjE3NTMzNTkwMzAsImRhdGFVc2VyIjp7ImlkIjoiMTM5NDE1OCIsIm5hbWUiOiJSZWJvcmEiLCJsYXN0bmFtZSI6IkFycXVpdGVjdG9zIiwiZnVsbF9uYW1lIjoiUmVib3JhIEFycXVpdGVjdG9zIiwiZW1haWwiOiJtYXJrZXRpbmdAcmVib3JhLmNvbS5teCIsInBob25lIjoiMzM0MTY5MDEwOSIsImVtYWlsX2NvbnRhY3QiOiJ2ZW50YXMucmVib3JhQGdtYWlsLmNvbSIsInByb2ZpbGVQaWN0dXJlIjoiOTRhMjg4NGVkMTExYjZmMmQzYTI0NWYxNjg4ZTU0ZjEucG5nIiwiY3JlYXRlZCI6IjEwXC8xMVwvMjAiLCJub3RpZmljYXRpb25zIjoxLCJwaWN0dXJlX3VybCI6Imh0dHBzOlwvXC9wcm9waWVkYWRlc2NvbS5zMy5hbWF6b25hd3MuY29tXC9maWxlc1wvcHJvZmlsZXNcLzk0YTI4ODRlZDExMWI2ZjJkM2EyNDVmMTY4OGU1NGYxLnBuZyIsImlzX2FkbWluIjoiMCIsInJvbGVfaWQiOiIxIiwidXNlcl90eXBlIjoxLCJ2ZXJzaW9uX2xvZ2luIjoxfX0.a7A8n37StXPLnksWbqAzkaBUoPSfUx0Etj3az3SHe_k"
}

def get_properties():
    page = 1
    props = {}

    while page is not None:
        res = requests.get(URL.format(page=page), headers=headers)
        if not res.ok:
            logger.error("Error getting page: "+str(res.status_code))
            logger.error(res.text)
            continue

        data = res.json()
        page = data.get("paginate", {}).get("next_page")
        for prop in data.get("properties", []):
            props[prop["id"]] = prop
        logger.success(f"page {page}: {len(props)} de propiedades en total")

    
    props_file = os.path.dirname(os.path.realpath(__file__)) + "/properties.json"
    logger.debug(f"Dumping props into {props_file}")
    with open(props_file, "w") as f:
        json.dump(props, f)

if __name__ == "__main__":
    get_properties()
