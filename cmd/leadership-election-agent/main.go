package main

import (
	"expvar"
	"fmt"
	"log"
	"net/http"

	_ "net/http/pprof"

	envstruct "code.cloudfoundry.org/go-envstruct"
	"code.cloudfoundry.org/leadership-election/internal/api"
)

func main() {
	log.Printf("Starting Leadership Election...")
	defer log.Printf("Closing Leadership Election...")

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	envstruct.WriteReport(&cfg)

	expvar.NewMap("LeadershipElectionAgent").AddFloat("leadership_status", 1)

	go func() {
		http.HandleFunc("/v1/leader", api.LeaderHandler)
		log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.Port), nil))
	}()

	// health endpoints (pprof and expvar)
	log.Printf("Health: %s", http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.HealthPort), nil))
}
