package agent

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/raft"
)

// Agent is a Leadership Election Agent. It determines if the local process
// should act as a leader or not.
type Agent struct {
	log  *log.Logger
	port int
	lis  net.Listener
	m    Metrics

	nodeIndex int
	nodes     []string
}

// New returns a new Agent.
func New(nodeIndex int, nodes []string, opts ...AgentOption) *Agent {
	a := &Agent{
		log:  log.New(ioutil.Discard, "", 0),
		port: 8080,

		nodeIndex: nodeIndex,
		nodes:     nodes,
	}

	for _, o := range opts {
		o(a)
	}

	return a
}

// AgentOption configures an Agent by overriding defaults.
type AgentOption func(*Agent)

// WithLogger returns an AgentOption that configures the logger for the Agent.
// It defaults to a silent logger.
func WithLogger(log *log.Logger) AgentOption {
	return func(a *Agent) {
		a.log = log
	}
}

// WithPort configures the port to bind the HTTP server to. It will always
// bind to localhost. Defaults to 8080.
func WithPort(port int) AgentOption {
	return func(a *Agent) {
		a.port = port
	}
}

// Metrics registers Gauge metrics.
type Metrics interface {
	// NewGauge returns a function to set the value for the given
	// metric.
	NewGauge(name string) func(value float64)
}

// WithMetrics configures the metrics for Agent.
func WithMetrics(m Metrics) AgentOption {
	return func(a *Agent) {
		a.m = m
	}
}

// Start starts the Agent. It does not block.
func (a *Agent) Start() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", a.port))
	if err != nil {
		a.log.Fatalf("failed to listen on localhost:%d", a.port)
	}
	a.lis = lis

	metric := a.m.NewGauge("leadership_status")

	isLeader := a.startRaft()

	go func() {
		for range time.Tick(time.Second) {
			if isLeader() {
				metric(1)
				continue
			}
			metric(0)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/leader", func(w http.ResponseWriter, r *http.Request) {
		if isLeader() {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusLocked)
	})
	srv := &http.Server{
		Handler: mux,
	}

	go func() {
		a.log.Fatal(srv.Serve(lis))
	}()
}

func (a *Agent) startRaft() func() bool {
	localAddr := a.nodes[a.nodeIndex]
	addr, err := net.ResolveTCPAddr("tcp", localAddr)
	if err != nil {
		a.log.Fatalf("failed to resolve address %s: %s", localAddr, err)
	}

	network, err := raft.NewTCPTransportWithLogger(
		localAddr,
		addr,
		100,
		30*time.Second,
		a.log,
	)
	if err != nil {
		a.log.Fatalf("failed to create raft TCP transport: %s", err)
	}

	store := raft.NewInmemStore()
	r, err := raft.NewRaft(
		&raft.Config{
			ProtocolVersion:    raft.ProtocolVersionMax,
			LocalID:            raft.ServerID(localAddr),
			HeartbeatTimeout:   100 * time.Millisecond,
			ElectionTimeout:    1 * time.Second,
			CommitTimeout:      1 * time.Second,
			MaxAppendEntries:   100,
			SnapshotInterval:   time.Second,
			LeaderLeaseTimeout: 100 * time.Millisecond,
		},
		nil,
		store,
		store,
		raft.NewInmemSnapshotStore(),
		network,
	)

	if err != nil {
		a.log.Fatalf("failed to create raft cluster: %s", err)
	}

	var peers []raft.Server
	for _, addr := range a.nodes {
		peers = append(peers, raft.Server{
			ID:      raft.ServerID(addr),
			Address: raft.ServerAddress(addr),
		})
	}

	r.BootstrapCluster(raft.Configuration{Servers: peers})

	return func() bool {
		return r.Leader() == raft.ServerAddress(localAddr)
	}
}

// Addr returns the address the Agent is listening to for HTTP requests (e.g.,
// 127.0.0.1:8080). It is only valid after calling Start().
func (a *Agent) Addr() string {
	return a.lis.Addr().String()
}
