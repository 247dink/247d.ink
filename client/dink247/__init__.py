import os
import logging
from datetime import datetime, timedelta, timezone
from pprint import pprint

import requests
from requests.exceptions import JSONDecodeError
import jwt


LOGGER = logging.getLogger(__name__)
LOGGER.addHandler(logging.NullHandler())

SHARED_SECRET = os.getenv('SHARED_SECRET', None)
SERVICE_URL = os.getenv('SERVICE_URL', 'http://localhost:8080/')


class Client:
    def __init__(self, secret=SHARED_SECRET, service_url=SERVICE_URL):
        try:
            self.secret = secret.encode()
        except AttributeError:
            self.secret = secret
        self.service_url = service_url

    def sign(self, url):
        if self.secret is None:
            raise Exception("Secret was not provided")
        return jwt.encode({
            "url": url,
            "exp": datetime.now(tz=timezone.utc) + timedelta(hours=4),
        }, self.secret, algorithm='HS256')

    def create(self, url):
        r = requests.post(
            self.service_url,
            self.sign(url),
            headers={'Content-Type': 'application/jwt'},
        )
        try:
            id = r.json()['id']
            LOGGER.debug(r.json())

        except JSONDecodeError:
            raise Exception(r.content.decode())

        return f'{self.service_url.rstrip("/")}/{id}'
