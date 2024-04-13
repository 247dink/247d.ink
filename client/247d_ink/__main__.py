import os
import sys
import hmac
import base64
from pprint import pprint
from hashlib import sha256

import requests


SHARED_TOKEN = os.getenv('SHARED_TOKEN', '').encode()


def sign(url):
    if SHARED_TOKEN is None:
        raise Exception("env variable SHARED_TOKEN is undefined")
    h = hmac.new(SHARED_TOKEN, url.encode(), sha256)
    return base64.b64encode(h.digest())


def main():
    url = sys.argv[1]
    s = sign(url)
    print("Signature: %s" % s.decode())
    r = requests.post('http://localhost:8080', {'url': url}, headers={'X-Signature': s})
    try:
        pprint(r.json())

    except Exception:
        print("ERROR:", r.content.decode())


if __name__ == '__main__':
    main()
