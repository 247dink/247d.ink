import os
import logging
from datetime import datetime, timedelta, timezone
from typing import Optional, Union

import requests
from requests.exceptions import JSONDecodeError
import jwt


LOGGER = logging.getLogger(__name__)
LOGGER.addHandler(logging.NullHandler())

SHARED_SECRET = os.getenv('DINK247_SHARED_SECRET', None)
SERVICE_URL = os.getenv('DINK247_SERVICE_URL', 'https://247d.ink/')


class Client:
    def __init__(self,
                 secret: Union[str, bytes, None] = SHARED_SECRET,
                 service_url: str = SERVICE_URL
                 ) -> None:
        if isinstance(secret, str):
            secret = secret.encode()
        self.secret = secret
        self.service_url = service_url

    def sign(self, url: str, ttl: int = 0) -> str:
        if self.secret is None:
            raise TypeError('secret missing')
        # NOTE: exp is part of JWT spec, it is the expiration of the token.
        #       ttl is the expiration (in days) of the link.
        payload = {
            "url": url,
            "ttl": ttl,
            "exp": datetime.now(tz=timezone.utc) + timedelta(hours=4),
        }
        return jwt.encode(payload, self.secret, algorithm='HS256')

    def create(self,
               url: str,
               base_url: Optional[str] = None,
               ttl: int = 0
               ) -> str:
        if base_url is None:
            base_url = self.service_url
        base_url = base_url.rstrip("/")
        r = requests.post(
            self.service_url,
            self.sign(url, ttl),
            headers={'Content-Type': 'application/jwt'},
        )
        try:
            id = r.json()['id']
            LOGGER.debug(r.json())

        except JSONDecodeError:
            raise Exception(r.content.decode())

        return f'{base_url}/{id}'
