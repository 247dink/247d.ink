deps:
	pipenv install --dev


sdist:
	pipenv run python3 setup.py sdist


mypy: deps
	pipenv run mypy .


lint: deps
	pipenv run flake8 .


clean:
	rm -rf dist
