import os
import msal
import webbrowser
import urllib.parse
import requests
from dotenv import load_dotenv

load_dotenv()
APP_ID = os.environ.get("MS_OPENGRAPH_APP_ID")
CLIENT_SECRET = os.environ.get("MS_OPENGRAPH_CLIENT_SECRET")
SCOPES = ["Files.ReadWrite.All"]
AUTHORITY_URL = "https://login.microsoftonline.com/common"
TOKEN_FILES = {"access": "./src/onedrive/access_token.txt", "refresh": "./src/onedrive/refresh_token.txt"}


def load_token_from_file(token_type: str) -> str:
    """ Carga access token o refresh token desde archivo """
    try:
        with open(TOKEN_FILES[token_type], "r") as f:
            return f.readline().strip()
    except FileNotFoundError:
        return None

def save_token_to_file(token_type: str, token: str):
    """ Guarda un token en archivo """
    with open(TOKEN_FILES[token_type], "w") as f:
        f.write(token)

def save_access_token(token: str):
    """ Guarda el access token """
    save_token_to_file("access", token)

def save_refresh_token(token: str):
    """ Guarda el refresh token. """
    save_token_to_file("refresh", token)


def procure_new_tokens() -> tuple:
    print("APP_ID:", APP_ID)
    print("CLIENT_SECRET:", CLIENT_SECRET)
    client_instance = msal.ConfidentialClientApplication(APP_ID, client_credential=CLIENT_SECRET, authority=AUTHORITY_URL)
    auth_url = client_instance.get_authorization_request_url(SCOPES)
    webbrowser.open(auth_url)
    
    print("\nPor favor, ingresa la URL completa que aparece después de la autorización (de la redirección del navegador):")
    redirected_url = input("URL: ").strip()

    parsed_url = urllib.parse.urlparse(redirected_url)
    query_params = urllib.parse.parse_qs(parsed_url.query)
    
    authorization_code = query_params.get("code", [None])[0]
    
    if not authorization_code:
        print("No se encontró el código en la URL.")
        exit(1)
    token_response = client_instance.acquire_token_by_authorization_code(authorization_code, SCOPES)
    
    if "access_token" in token_response:
        access_token = token_response["access_token"]
        refresh_token = token_response.get("refresh_token", "")
        save_access_token(access_token)
        save_refresh_token(refresh_token)
        print("Autenticación exitosa.")
        return access_token
    else:
        print("Error en autenticación:", token_response.get("error_description"))
        exit(1)

def refresh_access_token() -> str:
    refresh_token = load_token_from_file("refresh")
    if not refresh_token:
        print("No hay refresh token, solicitando autenticación manual.")
        return procure_new_tokens()

    url = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
    data = {
        "client_id": APP_ID,
        "client_secret": CLIENT_SECRET,
        "scope": "https://graph.microsoft.com/Files.ReadWrite.All",
        "grant_type": "refresh_token",
        "refresh_token": refresh_token,
    }
    
    response = requests.post(url, data=data)
    token_data = response.json()
    
    if "access_token" in token_data:
        new_access_token = token_data["access_token"]
        save_access_token(new_access_token)
        return new_access_token
    else:
        print("No refresh token, generar nuevo")
        return procure_new_tokens()


def get_access_token() -> str:
    """ Devuelve un access token válido"""
    token = load_token_from_file("access")
    if token:
        return token
    return refresh_access_token()


if __name__ == "__main__":
    access_token = get_access_token()
    print("Token obtenido:", access_token[:30] + "...")
