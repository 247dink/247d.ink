import sys

from dink247 import Client


def main():
    url = sys.argv[1]
    print(Client().create(url))


if __name__ == '__main__':
    main()
