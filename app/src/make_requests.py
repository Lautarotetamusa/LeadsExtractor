import requests
from requests.models import Response

# Codigos de respuesta cuando no esta mas autorizado (expiro token cookies etc)
METHODS = {
    'GET': requests.get,
    'POST': requests.post,
    'PUT': requests.put,
    'PATCH': requests.patch,
    'DELETE': requests.delete
}

class ApiRequest():
    def __init__(self, logger, api_url, api_params):

        self.logger = logger
        self.api_url = api_url
        self.api_params = api_params

    def make(self, url, method='GET', **kwargs) -> None | requests.Response:
        assert method in METHODS, f"Método HTTP no permitido: {method}"

        max_tries = 3
        tries = 0

        while (tries <= max_tries):
            params = self.api_params.copy()
            params["url"] = url

            try:
                res = METHODS[method](self.api_url, params=params, **kwargs)
                self.logger.debug(f"{method} {url}")
            except requests.exceptions.ConnectionError as e:
                self.logger.error("Ocurrio un error en la peticion")
                self.logger.error(str(e))
                tries += 1
                continue

            tries += 1
            if not res.ok:
                self.logger.error(f"Error request to: {url}")
                self.logger.error(res.status_code)
                self.logger.error(res.text)
            else:
                return res

        self.logger.error(f"Se intento una peticion mas de {max_tries} veces")
        return None


class Request():
    def __init__(self, cookies, headers, logger, login_method, codes: list[int]):
        self.cookies: dict[str, str] = cookies
        self.headers: dict[str, str] = headers
        self.logger = logger
        self.login_method = login_method
        self.unauthorized_codes = codes

    # Esta funcion realiza una peticion HTTP através de la libreria requests
    # Cuando el status_code es == 401 significa que el token de acceso a la API expiro, por
    # lo que deberemos llamar a la funcion login()
    def make(self, url, method='GET', **kwargs) -> Response | None:
        assert method in METHODS, f"Método HTTP no permitido: {method}"

        ok = False
        max_tries = 3
        tries = 0

        while (tries <= max_tries) and not ok:
            try:
                res: Response = METHODS[method](url, cookies=self.cookies, headers=self.headers, **kwargs)
                ok = res.ok
                self.logger.debug(f"{method} {url}")
            except:
                ok = False

            if ok:
                return res

            # Verifica el código de estado de la respuesta
            if res.status_code in self.unauthorized_codes:
                self.logger.error("El token de acceso expiro status="+str(res.status_code))
                self.logger.debug(res.text)
                self.login_method()
            else:
                self.logger.error(f"Error request to: {url}")
                self.logger.error(res.status_code)
                self.logger.error(res.text)

            tries += 1

        return None
