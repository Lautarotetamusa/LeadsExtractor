import requests

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
