from setuptools import setup


setup(
    name="dink247",
    version="1.0.0",
    description="URL shortener",
    author="Ben Timby",
    author_email="btimby@247dink.com",
    packages=["dink247"],
    package_data={"dink247": ["py.typed"]},
    install_requires=[
        'requests',
        'pyjwt',
    ],
)
