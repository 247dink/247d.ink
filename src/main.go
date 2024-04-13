package main

import (
	"os"
	"fmt"
	"log"
	"net/url"
	"net/http"
	"encoding/json"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"

	"github.com/247dink/247d.ink/link"
)

func setup_sentry() *sentryhttp.Handler {
	sentry_dsn := os.Getenv("SENTRY_DSN")
	if sentry_dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: sentry_dsn,
			TracesSampleRate: 1.0,
			Debug: true,
		}); err != nil {
			fmt.Printf("sentry.Init failed: %s\n", err)
		}
	}

	sentryHandler := sentryhttp.New(sentryhttp.Options{})
	return sentryHandler
}

func main() {
	log.Print("247d.ink: starting server...")

	server, err := link.NewServer()
	defer link.Client.Close()
	if err != nil {
		log.Fatalln(err.Error())
	}

	mux := http.NewServeMux()
	sentryHandler := setup_sentry()
	defaultUrl := os.Getenv("DEFAULT_REDIRECT")

	mux.HandleFunc("GET /{id...}", sentryHandler.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			if defaultUrl == "" {
				http.NotFound(w, r)
			} else {
				http.Redirect(w, r, defaultUrl, http.StatusSeeOther)
			}
			return
		}

		log.Printf("GET Request received. id: %s", id)
		obj := server.Get(id, r)
		if obj == nil {
			if defaultUrl == "" {
				http.NotFound(w, r)
			} else {
				http.Redirect(w, r, defaultUrl, http.StatusSeeOther)
			}
			return
		}

		http.Redirect(w, r, obj.Url, http.StatusMovedPermanently)
	}))

	mux.HandleFunc("POST /", sentryHandler.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("POST Request received.")
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		arg := r.FormValue("url")
		uri, err := url.ParseRequestURI(arg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		signature := r.Header.Get("X-Signature")
		log.Printf("Signature: %s", signature)
		if !server.CheckSignature(uri.String(), signature) {
			http.Error(w, "Bad or missing signature", http.StatusUnauthorized)
			return
		}

		obj := server.Save(uri.String(), r)
		if obj == nil {
			http.Error(w, "Could not save url", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(obj); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}
	hostStr := os.Getenv("HOST")
	if hostStr == "" {
		hostStr = "localhost"
	}
	bind := fmt.Sprintf("%s:%s", hostStr, portStr)

	log.Printf("247d.ink: listening on %s", bind)
	log.Fatal(http.ListenAndServe(bind, mux))
}