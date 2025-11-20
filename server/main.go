package main

import (
	"log"
	"time"

	"github.com/edsalkeld/bbmb/server/queue"
	"github.com/edsalkeld/bbmb/server/tcp"
)

const (
	TCPPort = ":9876"
	TimeoutScanInterval = 1 * time.Second
)

func main() {
	log.Println("Starting BBMB server...")

	manager := queue.NewManager()
	manager.StartTimeoutScanner(TimeoutScanInterval)

	server := tcp.NewServer(TCPPort, manager)

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
