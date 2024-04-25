DOCKER_COMPOSE=docker compose
GOPATH=$(shell go env GOPATH)


build:
	${DOCKER_COMPOSE} build


run:
	${DOCKER_COMPOSE} up


lint:
	${MAKE} -C client lint


mypy:
	${MAKE} -C client mypy


lint-client: mypy lint


lint-server: ${GOPATH}/bin/golangci-lint
	docker run --rm -v $(PWD)/server:/app -w /app golangci/golangci-lint:v1.57.2 golangci-lint run -v