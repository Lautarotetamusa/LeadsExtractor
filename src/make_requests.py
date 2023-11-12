import requests

class Request():
    def __init__(self, cookies, headers, logger, login_method):
        self.cookies = cookies
        self.headers = headers
        self.logger = logger
        self.login_method = login_method

    # Esta funcion realiza una peticion HTTP através de la libreria requests
    # Cuando el status_code es == 401 significa que el token de acceso a la API expiro, por
    # lo que deberemos llamar a la funcion login()
    def make(self, url, method='GET', **kwargs):
        status_code = 401
        while status_code == 401:
            # Define los métodos HTTP permitidos
            allowed_methods = {
                'GET': requests.get,
                'POST': requests.post,
                'PUT': requests.put
            }

            if method not in allowed_methods:
                self.logger.error(f"Método HTTP no permitido: {method}")
                raise(requests.HTTPError)

            # Realiza la solicitud con el método especificado y los parámetros adicionales
            res = allowed_methods[method](url, cookies=self.cookies, headers=self.headers, **kwargs)
            self.logger.debug(f"{method} {url}")

            status_code = res.status_code

            # Verifica el código de estado de la respuesta
            if res.status_code == 401:
                self.logger.error("El token de acceso expiro")
                self.login_method()
            elif 200 <= res.status_code < 300:
                return res
            else:
                self.logger.error(f"Error request to: {url}")
                self.logger.error(res.status_code)
                self.logger.error(res.text)
                raise(requests.HTTPError)