package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/edsalkeld/bbmb/server/metrics"
	"github.com/edsalkeld/bbmb/server/queue"
	"github.com/edsalkeld/bbmb/server/tcp"
)

const (
	DefaultTCPPort      = 9876
	DefaultMetricsPort  = 9877
	TimeoutScanInterval = 1 * time.Second
)

type config struct {
	TCPAddress     string
	MetricsAddress string
}

func main() {
	cfg, err := parseConfig(flag.CommandLine, osArgs())
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Println("Starting BBMB server...")

	manager := queue.NewManager()
	collector := metrics.NewCollector(manager)

	manager.SetTimeoutCallback(func(count int) {
		for i := 0; i < count; i++ {
			collector.IncrMessagesTimedOut()
		}
	})

	manager.StartTimeoutScanner(TimeoutScanInterval)

	go collector.StartServer(cfg.MetricsAddress)

	server := tcp.NewServer(cfg.TCPAddress, manager, collector)

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func parseConfig(fs *flag.FlagSet, args []string) (config, error) {
	fs.SetOutput(io.Discard)

	port := fs.Int("port", DefaultTCPPort, "TCP port for the BBMB broker")
	metricsPort := fs.Int("metrics-port", DefaultMetricsPort, "HTTP port for the Prometheus metrics endpoint")

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}

	if err := validatePort(*port, "port"); err != nil {
		return config{}, err
	}
	if err := validatePort(*metricsPort, "metrics-port"); err != nil {
		return config{}, err
	}

	return config{
		TCPAddress:     addressForPort(*port),
		MetricsAddress: addressForPort(*metricsPort),
	}, nil
}

func validatePort(port int, name string) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", name)
	}
	return nil
}

func addressForPort(port int) string {
	return fmt.Sprintf(":%d", port)
}

func osArgs() []string {
	return os.Args[1:]
}
