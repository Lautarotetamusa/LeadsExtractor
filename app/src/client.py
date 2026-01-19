import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry
from typing import Optional, Dict, Any, Callable
import time


class Client:
    def __init__(self,
                 login_method: Optional[Callable] = None,
                 max_retries: int = 3,
                 default_params: Optional[Dict[str, Any]] = None,
                 unauthorized_codes=[401, 403]):

        self.login_method = login_method
        self.max_retries = max_retries
        self.unauthorized_codes = unauthorized_codes

        # Parámetros por defecto para todas las requests
        self.default_params = default_params or None

        # Configurar la sesión con retries
        self.session = requests.Session()

        # Configurar retries para errores de conexión
        retry_strategy = Retry(
                total=3,
                backoff_factor=0.5,
                status_forcelist=[500, 502, 503, 504],
                )
        adapter = HTTPAdapter(max_retries=retry_strategy)
        self.session.mount("http://", adapter)
        self.session.mount("https://", adapter)

    def _merge_params(self, default: Dict[str, Any], override: Dict[str, Any]) -> Dict[str, Any]:
        """Combina parámetros por defecto con los de override, dando prioridad a override"""
        merged = default.copy()
        merged.update(override)
        return merged

    def _prepare_request_args(self, **kwargs) -> Dict[str, Any]:
        """Prepara los argumentos para la request, combinando valores por defecto con los específicos"""

        final_kwargs = kwargs

        if 'params' in kwargs:
            if self.default_params is not None:
                final_kwargs['params'] = self._merge_params(self.default_params, kwargs['params'])
        elif self.default_params is not None:
            final_kwargs['params'] = self.default_params

        return final_kwargs

    def _make_request(self, method: str, target_url: str, **kwargs) -> requests.Response | None:
        """
        Método interno que maneja la lógica de reintentos y login
        """
        # Preparar los argumentos de la request
        login_req = False
        if 'login_req' in kwargs and kwargs['login_req']:
            login_req = True
            del kwargs['login_req']

        final_kwargs = self._prepare_request_args(**kwargs)

        for attempt in range(self.max_retries):
            try:
                # ----- DEBUG ------
                # import json
                # print(method, target_url)
                # print(json.dumps(final_kwargs, indent=4))

                if login_req:
                    # No usamos la session para no inferir esos headers y cookies
                    response = requests.request(method, target_url, **final_kwargs)
                else:
                    response = self.session.request(method, target_url, **final_kwargs)

                if response.ok:
                    return response
                else:
                    print(f"Request failed with status {response.status_code}, attempt {attempt + 1}/{self.max_retries}")

                # Si es un error de autenticación/prohibido y tenemos método de login
                if response.status_code in self.unauthorized_codes and self.login_method:
                    if attempt < self.max_retries - 1:
                        if login_req:
                            print('ignoring login method')
                            time.sleep(1)
                            continue

                        print("Attempting login...")
                        if self.login_method():
                            print("Login successful, retrying...")
                            time.sleep(1)
                            continue

            except requests.exceptions.RequestException as e:
                print(f"Request exception: {e}, attempt {attempt + 1}/{self.max_retries}")

                if attempt < self.max_retries - 1:
                    time.sleep(1)
                    continue
                return None

    def get(self, url: str, **kwargs) -> requests.Response | None:
        """Realizar una petición GET a través de ZenRows"""
        return self._make_request('GET', url, **kwargs)

    def post(self, url: str, **kwargs) -> requests.Response | None:
        """Realizar una petición POST a través de ZenRows"""
        return self._make_request('POST', url, **kwargs)

    def put(self, url: str, **kwargs) -> requests.Response | None:
        """Realizar una petición PUT a través de ZenRows"""
        return self._make_request('PUT', url, **kwargs)

    def delete(self, url: str, **kwargs) -> requests.Response | None:
        """Realizar una petición DELETE a través de ZenRows"""
        return self._make_request('DELETE', url, **kwargs)

    def patch(self, url: str, **kwargs) -> requests.Response | None:
        """Realizar una petición PATCH a través de ZenRows"""
        return self._make_request('PATCH', url, **kwargs)

    def head(self, url: str, **kwargs) -> requests.Response | None:
        """Realizar una petición HEAD a través de ZenRows"""
        return self._make_request('HEAD', url, **kwargs)

    def options(self, url: str, **kwargs) -> requests.Response | None:
        """Realizar una petición OPTIONS a través de ZenRows"""
        return self._make_request('OPTIONS', url, **kwargs)

    def request(self, method: str, url: str, **kwargs) -> requests.Response | None:
        """Método genérico para cualquier verbo HTTP"""
        return self._make_request(method, url, **kwargs)

    def update_default_params(self, **params):
        """Actualizar parámetros por defecto"""
        self.default_params.update(params)
