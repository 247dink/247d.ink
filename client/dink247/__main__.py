import sys
import argparse
import logging

from dink247 import Client


LOGGER = logging.getLogger()
LOGGER.addHandler(logging.StreamHandler())


def main(args):
    print(Client().create(args.url, ttl=args.ttl))


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('url')
    parser.add_argument('-t', '--ttl', type=int, dest='ttl')
    parser.add_argument('-v', '--verbose', action='store_true', dest='debug')

    args = parser.parse_args()

    if args.debug:
        LOGGER.setLevel(logging.DEBUG)

    main(args)
