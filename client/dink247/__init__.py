import os
import hmac
import base64
import logging
from hashlib import sha256
from pprint import pprint

import requests
from requests.exceptions import JSONDecodeError


LOGGER = logging.getLogger(__name__)
LOGGER.addHandler(logging.NullHandler())

SHARED_TOKEN = os.getenv('SHARED_TOKEN', None)
SERVICE_URL = os.getenv('SERVICE_URL', 'http://localhost:8080/')


class Client:
    def __init__(self, token=SHARED_TOKEN, service_url=SERVICE_URL):
        try:
            self.token = token.encode()
        except AttributeError:
            self.token = token
        self.service_url = service_url

    def sign(self, url):
        if self.token is None:
            raise Exception("Token was not provided")
        h = hmac.new(self.token, url.encode(), sha256)
        return base64.b64encode(h.digest())

    def create(self, url):
        s = self.sign(url)
        r = requests.post(
            self.service_url,
            {'url': url},
            headers={'X-Signature': s},
        )
        try:
            id = r.json()['id']
            LOGGER.debug(r.json())

        except JSONDecodeError:
            raise Exception(r.content.decode())

        return f'{self.service_url.rstrip("/")}/{id}'
