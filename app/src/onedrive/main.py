import requests
import base64
from src.onedrive.onedrive_authorization_utils import get_access_token

def download_file(shared_link) -> bytes | None:
    access_token = get_access_token()
    
    headers = {
        "Authorization": f"Bearer {access_token}",
        "Content-Type": "application/json"
    }

    endpoint = "https://graph.microsoft.com/v1.0/shares/"

    shared_link_encoded = base64.b64encode(shared_link.encode()).decode()
    sharing_token = "u!" + shared_link_encoded.replace("=", "").replace("/", "_").replace("+", "-")

    url = f"{endpoint}{sharing_token}/driveItem"
    response = requests.get(url, headers=headers)

    if not response.ok:
        return None

    item_info = response.json()
    if not "@microsoft.graph.downloadUrl" in item_info:
        return None

    download_url = item_info["@microsoft.graph.downloadUrl"]

    response = requests.get(download_url, stream=True)

    if response.ok:
        return response.content
