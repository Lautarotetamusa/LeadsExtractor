
import requests

def download_image(url: str) -> bytes | None:
    res = requests.get(url)
    if not res.ok: return None
    return res.content
