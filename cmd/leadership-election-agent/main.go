package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"code.cloudfoundry.org/leadership-election/internal/api"
)

func main() {
	a := addr()
	log.Printf("Starting Leadership Election...")
	defer log.Printf("Closing Leadership Election...")

	log.Printf("listening on %s", a)

	http.HandleFunc("/v1/leader", api.LeaderHandler)
	log.Fatal(http.ListenAndServe(a, nil))
}

func addr() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "localhost:8080"
	}
	return fmt.Sprintf("localhost:%s", port)
}
