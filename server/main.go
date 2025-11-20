package main

import (
	"log"
	"time"

	"github.com/edsalkeld/bbmb/server/metrics"
	"github.com/edsalkeld/bbmb/server/queue"
	"github.com/edsalkeld/bbmb/server/tcp"
)

const (
	TCPPort             = ":9876"
	MetricsPort         = ":9877"
	TimeoutScanInterval = 1 * time.Second
)

func main() {
	log.Println("Starting BBMB server...")

	manager := queue.NewManager()
	collector := metrics.NewCollector(manager)

	manager.SetTimeoutCallback(func(count int) {
		for i := 0; i < count; i++ {
			collector.IncrMessagesTimedOut()
		}
	})

	manager.StartTimeoutScanner(TimeoutScanInterval)

	go collector.StartServer(MetricsPort)

	server := tcp.NewServer(TCPPort, manager, collector)

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
