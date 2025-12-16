from typing import Optional, Dict, Any, Callable
from src.client import Client
import requests

ZENROWS_API_URL = "https://api.zenrows.com/v1/"


class ZenRowsClient(Client):
    def __init__(self,
                 api_key: str,
                 login_method: Optional[Callable] = None,
                 max_retries: int = 3,
                 default_params: Optional[Dict[str, Any]] = None,
                 unauthorized_codes=[401, 403]):

        self.api_key = api_key
        self.default_params = default_params
        super().__init__(login_method, max_retries, default_params, unauthorized_codes)

    def _make_request(self, method: str, target_url: str, **kwargs) -> requests.Response | None:
        kwargs['params'] = {
            'apikey': self.api_key,
            'url': target_url
        }

        return super()._make_request(method, ZENROWS_API_URL, **kwargs)
