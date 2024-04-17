package main

import (
	"os"
	"fmt"
	"log"
	"strconv"
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/golang-jwt/jwt/v5"

	"github.com/247dink/247d.ink/link"
)

var sentryHandler *sentryhttp.Handler = nil
var defaultUrl string
var secret string
var address string

type JWTClaims struct {
	Url string `json:"url"`
	jwt.RegisteredClaims
}

type Status struct {
	DB string `json:"db"`
}

func init() {
	sentry_dsn := os.Getenv("SENTRY_DSN")
	if sentry_dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: sentry_dsn,
			TracesSampleRate: 1.0,
			Debug: true,
		}); err != nil {
			fmt.Printf("sentry.Init failed: %s\n", err)
		}

		sentryHandler = sentryhttp.New(sentryhttp.Options{})
	}

	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}
	hostStr := os.Getenv("HOST")
	if hostStr == "" {
		hostStr = "0.0.0.0"
	}
	address = fmt.Sprintf("%s:%s", hostStr, portStr)

	defaultUrl = os.Getenv("DEFAULT_REDIRECT")
	secret = os.Getenv("DINK247_SHARED_SECRET")
}

func makeHandler(f http.HandlerFunc) http.HandlerFunc {
	if sentryHandler == nil {
		return f
	}
	return sentryHandler.HandleFunc(f)
}

func main() {
	log.Print("247d.ink: starting server...")

	defer link.Client.Close()

	server, err := link.NewServer()
	if err != nil {
		log.Fatalln(err.Error())
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/", makeHandler(func(w http.ResponseWriter, r *http.Request) {
		obj := &Status{}

		err := server.Check(r)

		var status int
		if err == nil {
			obj.DB = "OK"
			status = http.StatusOK
		} else {
			obj.DB = err.Error()
			status = http.StatusInternalServerError
		}

		w.WriteHeader(status)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(obj); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	mux.HandleFunc("GET /{id...}", makeHandler(func(w http.ResponseWriter, r *http.Request) {
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
		obj, err := server.Get(id, r)
		if err != nil {
			if defaultUrl == "" {
				http.NotFound(w, r)
			} else {
				http.Redirect(w, r, defaultUrl, http.StatusSeeOther)
			}
			return
		}

		http.Redirect(w, r, obj.Url, http.StatusMovedPermanently)
	}))

	mux.HandleFunc("POST /", makeHandler(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("POST Request received.")

		var ttl int = 0

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Could not read request", http.StatusBadRequest)
			return
		}

		token, err := jwt.ParseWithClaims(string(body), &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			http.Error(w, "Could not parse token", http.StatusBadRequest)
			return
		}

		uri, err := url.ParseRequestURI(claims.Url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ttlStr := r.Header.Get("X-TTL")
		if ttlStr != "" {
			ttl, _ = strconv.Atoi(ttlStr)
		}

		obj, err := server.Save(uri.String(), ttl, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(obj); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	log.Printf("247d.ink: listening on %s", address)
	log.Fatal(http.ListenAndServe(address, mux))
}