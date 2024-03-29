import requests
from requests.models import Response

class ApiRequest():
    def __init__(self, logger, api_url, api_params):
        self.logger = logger
        self.api_url = api_url
        self.api_params = api_params

    # Esta funcion realiza una peticion HTTP através de la libreria requests
    def make(self, url, method='GET', **kwargs):
        max_tries = 3
        tries = 0
        while (tries <= max_tries):
            # Define los métodos HTTP permitidos
            allowed_methods = {
                'GET': requests.get,
                'POST': requests.post,
                'PUT': requests.put,
                'PATCH': requests.patch
            }

            if method not in allowed_methods:
                self.logger.error(f"Método HTTP no permitido: {method}")
                raise(requests.HTTPError)

            # Realiza la solicitud con el método especificado y los parámetros adicionales
            params = self.api_params.copy()
            params["url"] = url
            #print(kwargs["json"])
            #print(params)
            try:
                res = allowed_methods[method](self.api_url, params=params, **kwargs)
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
        #self.logger.debug(res.text)
        #self.logger.debug(res.status_code)

class Request():
    def __init__(self, cookies, headers, logger, login_method):
        self.cookies: dict[str, str] = cookies
        self.headers: dict[str, str] = headers
        self.logger = logger
        self.login_method = login_method

    # Esta funcion realiza una peticion HTTP através de la libreria requests
    # Cuando el status_code es == 401 significa que el token de acceso a la API expiro, por
    # lo que deberemos llamar a la funcion login()
    def make(self, url, method='GET', **kwargs) -> Response | None:
        status_code = 401
        max_tries = 3
        tries = 0
        while (tries <= max_tries) and (status_code == 401 or status_code == 422):
            # Define los métodos HTTP permitidos
            allowed_methods = {
                'GET': requests.get,
                'POST': requests.post,
                'PUT': requests.put,
                'PATCH': requests.patch
            }

            if method not in allowed_methods:
                self.logger.error(f"Método HTTP no permitido: {method}")
                raise(requests.HTTPError)

            # Realiza la solicitud con el método especificado y los parámetros adicionales
            res: Response = allowed_methods[method](url, cookies=self.cookies, headers=self.headers, **kwargs)
            #print(res.text)
            self.logger.debug(f"{method} {url}")

            status_code = res.status_code

            # Verifica el código de estado de la respuesta
            if res.status_code >= 400 and res.status_code < 500:
                self.logger.error("El token de acceso expiro")
                self.logger.debug(res.text)
                self.login_method()
            elif 200 <= res.status_code < 300:
                return res
            else:
                self.logger.error(f"Error request to: {url}")
                self.logger.error(res.status_code)
                self.logger.error(res.text)
                raise(requests.HTTPError)

            tries += 1
