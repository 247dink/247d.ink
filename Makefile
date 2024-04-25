DOCKER_COMPOSE=docker compose


build:
	${DOCKER_COMPOSE} build


run:
	${DOCKER_COMPOSE} up


lint:
	${MAKE} -C client lint


mypy:
	${MAKE} -C client mypy


ci: mypy lint
