package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/edsalkeld/bbmb/go-client/client"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	serverAddr := flag.String("server", "localhost:9876", "Server address")

	var queueName, content, guid string
	var timeoutSeconds, waitSeconds int

	switch command {
	case "ensure-queue":
		fs := flag.NewFlagSet("ensure-queue", flag.ExitOnError)
		fs.StringVar(&queueName, "queue", "", "Queue name (required)")
		fs.StringVar(serverAddr, "server", "localhost:9876", "Server address")
		fs.Parse(os.Args[2:])

		if queueName == "" {
			log.Fatal("--queue is required")
		}

		c := client.NewClient(*serverAddr)
		if err := c.Connect(); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer c.Close()

		if err := c.EnsureQueue(queueName); err != nil {
			log.Fatalf("Failed to ensure queue: %v", err)
		}

		fmt.Printf("Queue '%s' ensured\n", queueName)

	case "add":
		fs := flag.NewFlagSet("add", flag.ExitOnError)
		fs.StringVar(&queueName, "queue", "", "Queue name (required)")
		fs.StringVar(&content, "content", "", "Message content (required)")
		fs.StringVar(serverAddr, "server", "localhost:9876", "Server address")
		fs.Parse(os.Args[2:])

		if queueName == "" || content == "" {
			log.Fatal("--queue and --content are required")
		}

		c := client.NewClient(*serverAddr)
		if err := c.Connect(); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer c.Close()

		guid, err := c.AddMessage(queueName, content)
		if err != nil {
			log.Fatalf("Failed to add message: %v", err)
		}

		fmt.Printf("Message added with GUID: %s\n", guid)

	case "pickup":
		fs := flag.NewFlagSet("pickup", flag.ExitOnError)
		fs.StringVar(&queueName, "queue", "", "Queue name (required)")
		timeoutStr := fs.String("timeout", "30", "Timeout in seconds")
		waitStr := fs.String("wait", "0", "Long-poll wait in seconds")
		fs.StringVar(serverAddr, "server", "localhost:9876", "Server address")
		fs.Parse(os.Args[2:])

		if queueName == "" {
			log.Fatal("--queue is required")
		}

		var err error
		timeoutSeconds, err = strconv.Atoi(*timeoutStr)
		if err != nil {
			log.Fatalf("Invalid timeout value: %v", err)
		}
		waitSeconds, err = strconv.Atoi(*waitStr)
		if err != nil {
			log.Fatalf("Invalid wait value: %v", err)
		}

		c := client.NewClient(*serverAddr)
		if err := c.Connect(); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer c.Close()

		msg, err := c.PickupMessage(queueName, timeoutSeconds, waitSeconds)
		if err == client.ErrQueueEmpty {
			fmt.Println("Queue is empty")
			os.Exit(0)
		}
		if err != nil {
			log.Fatalf("Failed to pickup message: %v", err)
		}

		fmt.Printf("GUID: %s\n", msg.GUID)
		fmt.Printf("Content: %s\n", msg.Content)
		fmt.Printf("Checksum: %s\n", msg.Checksum)

	case "consume":
		fs := flag.NewFlagSet("consume", flag.ExitOnError)
		fs.StringVar(&queueName, "queue", "", "Queue name (required)")
		timeoutStr := fs.String("timeout", "30", "Timeout in seconds")
		waitStr := fs.String("wait", "0", "Long-poll wait in seconds")
		fs.StringVar(serverAddr, "server", "localhost:9876", "Server address")
		fs.Parse(os.Args[2:])

		if queueName == "" {
			log.Fatal("--queue is required")
		}

		var err error
		timeoutSeconds, err = strconv.Atoi(*timeoutStr)
		if err != nil {
			log.Fatalf("Invalid timeout value: %v", err)
		}
		waitSeconds, err = strconv.Atoi(*waitStr)
		if err != nil {
			log.Fatalf("Invalid wait value: %v", err)
		}

		c := client.NewClient(*serverAddr)
		if err := c.Connect(); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer c.Close()

		msg, err := c.PickupMessage(queueName, timeoutSeconds, waitSeconds)
		if err == client.ErrQueueEmpty {
			fmt.Println("Queue is empty")
			os.Exit(0)
		}
		if err != nil {
			log.Fatalf("Failed to pickup message: %v", err)
		}

		fmt.Printf("GUID: %s\n", msg.GUID)
		fmt.Printf("Content: %s\n", msg.Content)
		fmt.Printf("Checksum: %s\n", msg.Checksum)

		if err := c.DeleteMessage(queueName, msg.GUID); err != nil {
			log.Fatalf("Failed to delete message: %v", err)
		}

		fmt.Println("Message consumed (deleted)")

	case "delete":
		fs := flag.NewFlagSet("delete", flag.ExitOnError)
		fs.StringVar(&queueName, "queue", "", "Queue name (required)")
		fs.StringVar(&guid, "guid", "", "Message GUID (required)")
		fs.StringVar(serverAddr, "server", "localhost:9876", "Server address")
		fs.Parse(os.Args[2:])

		if queueName == "" || guid == "" {
			log.Fatal("--queue and --guid are required")
		}

		c := client.NewClient(*serverAddr)
		if err := c.Connect(); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer c.Close()

		if err := c.DeleteMessage(queueName, guid); err != nil {
			log.Fatalf("Failed to delete message: %v", err)
		}

		fmt.Printf("Message %s deleted from queue '%s'\n", guid, queueName)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: bbmb-client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  ensure-queue  --queue=<name> [--server=<addr>]")
	fmt.Println("  add           --queue=<name> --content=<text> [--server=<addr>]")
	fmt.Println("  pickup        --queue=<name> [--timeout=<seconds>] [--wait=<seconds>] [--server=<addr>]")
	fmt.Println("  consume       --queue=<name> [--timeout=<seconds>] [--wait=<seconds>] [--server=<addr>]  (pickup + delete)")
	fmt.Println("  delete        --queue=<name> --guid=<id> [--server=<addr>]")
	fmt.Println()
	fmt.Println("Default server: localhost:9876")
}
