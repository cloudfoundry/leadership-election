package main

import (
	"log"
	"net/http"

	"code.cloudfoundry.org/leadership-election/api"
)

func main() {
	http.HandleFunc("/leader", api.LeaderHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
