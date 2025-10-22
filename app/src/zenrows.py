import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry
from typing import Optional, Dict, Any, Callable
import time


class ZenRowsClient:
    def __init__(self,
                 api_key: str,
                 login_method: Optional[Callable] = None,
                 max_retries: int = 3,
                 default_params: Optional[Dict[str, Any]] = None,
                 default_headers: Optional[Dict[str, Any]] = None,
                 default_cookies: Optional[Dict[str, Any]] = None,
                 unauthorized_codes=[401, 403]):

        self.api_key = api_key
        self.login_method = login_method
        self.max_retries = max_retries
        self.base_url = "https://api.zenrows.com/v1/"
        self.unauthorized_codes = unauthorized_codes

        # Parámetros por defecto para todas las requests
        self.default_params = default_params or {}
        self.default_headers = default_headers or {}
        self.default_cookies = default_cookies or {}

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

    def _merge_headers(self, default: Dict[str, Any], override: Dict[str, Any]) -> Dict[str, Any]:
        """Combina headers por defecto con los de override, dando prioridad a override"""
        merged = default.copy()
        merged.update(override)
        return merged

    def _merge_cookies(self, default: Dict[str, Any], override: Dict[str, Any]) -> Dict[str, Any]:
        """Combina cookies por defecto con los de override, dando prioridad a override"""
        merged = default.copy()
        merged.update(override)
        return merged

    def _prepare_request_args(self, target_url: str, **kwargs) -> Dict[str, Any]:
        """Prepara los argumentos para la request, combinando valores por defecto con los específicos"""

        # Parámetros base de ZenRows
        base_params = {
            'apikey': self.api_key,
            'url': target_url
        }

        # Combinar parámetros: default_params + base_params + kwargs['params']
        params_from_kwargs = kwargs.get('params', {})
        merged_params = self._merge_params(self.default_params, base_params)
        final_params = self._merge_params(merged_params, params_from_kwargs)

        # Preparar los kwargs finales
        final_kwargs = kwargs.copy()
        final_kwargs['params'] = final_params

        # Combinar headers si se proporcionan
        if 'headers' in kwargs:
            final_kwargs['headers'] = self._merge_headers(self.default_headers, kwargs['headers'])
        else:
            final_kwargs['headers'] = self.default_headers.copy()

        # Combinar cookies si se proporcionan
        if 'cookies' in kwargs:
            final_kwargs['cookies'] = self._merge_cookies(self.default_cookies, kwargs['cookies'])
        else:
            final_kwargs['cookies'] = self.default_cookies.copy()

        return final_kwargs

    def _make_request(self, method: str, target_url: str, **kwargs) -> requests.Response | None:
        """
        Método interno que maneja la lógica de reintentos y login
        """
        # Preparar los argumentos de la request
        final_kwargs = self._prepare_request_args(target_url, **kwargs)
        zenrows_url = self.base_url.strip('/')

        for attempt in range(self.max_retries):
            try:
                response = self.session.request(method, zenrows_url, **final_kwargs)

                if response.ok:
                    return response

                # Si es un error de autenticación/prohibido y tenemos método de login
                if response.status_code in self.unauthorized_codes and self.login_method:
                    print(f"Request failed with status {response.status_code}, attempt {attempt + 1}/{self.max_retries}")

                    if attempt < self.max_retries - 1:
                        print("Attempting login...")
                        if self.login_method():
                            print("Login successful, retrying...")
                            time.sleep(1)
                            continue

                # Para otros errores, retornar la respuesta
                return response

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

    def update_default_headers(self, **headers):
        """Actualizar headers por defecto"""
        self.default_headers.update(headers)
