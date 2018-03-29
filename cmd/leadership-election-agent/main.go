package main

import (
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "net/http/pprof"

	envstruct "code.cloudfoundry.org/go-envstruct"
	"code.cloudfoundry.org/leadership-election/app/agent"
)

func main() {
	log.Printf("Starting Leadership Election...")
	defer log.Printf("Closing Leadership Election...")

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	envstruct.WriteReport(&cfg)

	em := expvar.NewMap("LeadershipElectionAgent")

	a := agent.New(
		cfg.NodeIndex,
		cfg.NodeAddrs,
		agent.WithLogger(log.New(os.Stderr, "", log.LstdFlags)),
		agent.WithMetrics(&metrics{em}),
		agent.WithPort(int(cfg.Port)),
	)

	a.Start()

	// health endpoints (pprof and expvar)
	log.Printf("Health: %s", http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.HealthPort), nil))
}

type metrics struct {
	m Map
}

// Map stores the desired metrics.
type Map interface {
	// Add adds a new metric to the map.
	Add(key string, delta int64)

	// AddFloat adds a new metric to the map.
	AddFloat(key string, delta float64)

	// Get gets a Var from the Map.
	Get(key string) expvar.Var
}

// New returns a new Metrics.
func New(m Map) *metrics {
	return &metrics{
		m: m,
	}
}

// NewGauge returns a func to be used to set the value of a gauge metric.
func (m *metrics) NewGauge(name string) func(value float64) {
	if m.m == nil {
		return func(_ float64) {}
	}

	m.m.AddFloat(name, 0)
	f := m.m.Get(name).(*expvar.Float)

	return f.Set
}
