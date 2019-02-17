package serve

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// Imetrics interfaces nessesary metrics
type Imetrics interface {
	IncChecks()
	IncErrors()
}

// StatusMetrics implents metrics interface for work info
type StatusMetrics struct {
	checked int64
	errors  int64
	mu      sync.Mutex
}

// IncChecks adds 1 to checks
func (s *StatusMetrics) IncChecks() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checked++
}

// IncErrors incs errors counter
func (s *StatusMetrics) IncErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors++
}

// Metrics handles metrics counters
type Metrics struct {
	checks prometheus.Counter // Num incoming files
	errors prometheus.Counter // Num errors
}

// NewMetrics creats a new metrics object
func NewMetrics(prefix string) (*Metrics, error) {
	m := &Metrics{}

	m.checks = registerCounter(fmt.Sprintf("%s_checks", prefix), "Number of url timing checks performed")
	m.errors = registerCounter(fmt.Sprintf("%s_errors", prefix), "Number of errors")

	return m, nil
}

// StartMetricsServer starts health and metrics listener on a given port
func StartMetricsServer(logger *log.Entry, port int, gatherer prometheus.Gatherer) net.Listener {

	logger.Infof("Listen for metrics queries on :%d/metrics\n", port)
	addr := fmt.Sprintf(":%d", port)
	router := http.NewServeMux()

	router.Handle("/health", promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
	router.Handle("/metrics", prometheus.Handler())

	server := http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      router,
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("Unable to setup Health endpoint (%s): %s", addr, err)
	}

	go func() {
		logger.Printf("Metrics endpoint is listening on %s", lis.Addr().String())
		logger.Printf("Metrics server closing: %s", server.Serve(lis))
	}()
	return lis
}

func registerCounter(name, help string) prometheus.Counter {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})

	prometheus.MustRegister(counter)
	return counter
}

// IncChecks adds 1 to incomming
func (m *Metrics) IncChecks() {
	m.checks.Inc()
}

// IncErrors incs error counter with one
func (m *Metrics) IncErrors() {
	m.errors.Inc()
}
