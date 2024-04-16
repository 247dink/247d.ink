# 247d.ink
URL Shortener

https://firebase.google.com/docs/hosting/cloud-run

```bash
# To configure gcloud:
$ snap install google-cloud-cli --classic
# Build image
$ docker build . --tag us-central1-docker.pkg.dev/d-ink-4bf48/two47d-ink/247d.ink:test
# Authenticate and configure docker auth
$ gcloud auth login
$ gcloud auth configure-docker us-central1-docker.pkg.dev
# Push image
$ docker push us-central1-docker.pkg.dev/d-ink-4bf48/two47d-ink/247d.ink:test
```

```bash
# Set default region
$ gcloud config set run/region us-central1
# Deploy cloud run app
$ gcloud run deploy --project dink-412003 --image us-central1-docker.pkg.dev/d-ink-4bf48/two47d-ink/247d.ink:test
```

```bash
# Deploy hosting
$ firebase deploy --only hosting
```