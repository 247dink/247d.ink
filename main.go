package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

//	"github.com/teris-io/shortid"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("helloword: received a request")
	target := os.Getenv("TARGET")
	if target == "" {
		target = "World"
	}
	fmt.Fprintf(w, "Hello %s!\n", target)
}

func main() {
	log.Print("helloworld: starting server...")

	http.HandleFunc("/", handler)

	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port := fmt.Sprintf(":%s", portStr)

	log.Printf("helloworld: listening on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}