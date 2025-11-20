# BBMB - Basic Bare-bones Message Broker

A lightweight, high-performance TCP-based message broker with FIFO queues, written in Go.

## Features

- **Simple FIFO queues** with atomic pickup operations
- **TCP-based protocol** for efficient binary communication
- **Message timeout & requeue** - messages automatically return to queue if not deleted
- **SHA256 checksum validation** for data integrity
- **Prometheus metrics** for observability
- **Client libraries** in Go and Python with CLI tools

## Quick Start

### Running the Server

```bash
cd server
go build -o bbmb-server .
./bbmb-server
```

The server listens on:
- **TCP port 9876** - message broker protocol
- **HTTP port 9877** - Prometheus metrics endpoint

### Using the Go Client

```bash
cd go-client
go build -o bbmb-client ./cmd

# Ensure queue exists
./bbmb-client ensure-queue --queue=myqueue

# Add a message
./bbmb-client add --queue=myqueue --content="Hello, World!"

# Pick up a message (30 second timeout)
./bbmb-client pickup --queue=myqueue --timeout=30

# Delete a message
./bbmb-client delete --queue=myqueue --guid=<message-id>
```

### Using the Python Client

```bash
cd python-client
pip install -e .

# Ensure queue exists
python cli.py ensure-queue --queue=myqueue

# Add a message
python cli.py add --queue=myqueue --content="Hello from Python!"

# Pick up a message
python cli.py pickup --queue=myqueue --timeout=30

# Delete a message
python cli.py delete --queue=myqueue --guid=<message-id>
```

### Using as a Library

**Go:**
```go
import "github.com/edsalkeld/bbmb/go-client/client"

c := client.NewClient("localhost:9876")
if err := c.Connect(); err != nil {
    log.Fatal(err)
}
defer c.Close()

guid, err := c.AddMessage("myqueue", "Hello, World!")
msg, err := c.PickupMessage("myqueue", 30)
err = c.DeleteMessage("myqueue", guid)
```

**Python:**
```python
from bbmb_client import Client

with Client("localhost:9876") as client:
    guid = client.add_message("myqueue", "Hello, World!")
    msg = client.pickup_message("myqueue", timeout_seconds=30)
    client.delete_message("myqueue", guid)
```

## Protocol

BBMB uses a simple binary protocol over TCP:

```
[4 bytes: total length][1 byte: command type][payload]
```

### Operations

1. **ENSURE_QUEUE (0x01)** - Create queue if it doesn't exist
2. **ADD_MESSAGE (0x02)** - Add message to queue (returns GUID)
3. **PICKUP_MESSAGE (0x03)** - Get message from queue with timeout
4. **DELETE_MESSAGE (0x04)** - Delete message by GUID

### Message Lifecycle

1. Client adds message with content and SHA256 checksum
2. Server validates checksum and assigns GUID
3. Message enters queue in "available" state
4. Client picks up message, enters "picked up" state with timeout
5. Client deletes message OR timeout expires and message returns to queue

## Metrics

Access Prometheus metrics at `http://localhost:9877/metrics`

**Global metrics:**
- `bbmb_uptime_seconds` - Server uptime
- `bbmb_memory_*` - Memory usage statistics
- `bbmb_queues_total` - Number of queues
- `bbmb_messages_added_total` - Total messages added
- `bbmb_messages_picked_up_total` - Total messages picked up
- `bbmb_messages_deleted_total` - Total messages deleted
- `bbmb_messages_timed_out_total` - Total messages timed out
- `bbmb_active_connections` - Current active connections

**Per-queue metrics:**
- `bbmb_queue_messages_total{queue="name"}` - Total messages in queue
- `bbmb_queue_messages_available{queue="name"}` - Available messages

## Architecture

```
server/
├── queue/          # Queue and message data structures
├── protocol/       # Binary protocol codec
├── tcp/            # TCP server and handlers
├── metrics/        # Prometheus metrics
└── main.go         # Entry point

go-client/
├── client/         # Go client library
└── cmd/            # CLI wrapper

python-client/
├── bbmb_client/    # Python client library
├── cli.py          # CLI wrapper
└── setup.py        # Package setup
```

## Design Decisions

- **1MB message size limit** to prevent memory exhaustion
- **In-memory only** (no persistence yet, architecture supports future addition)
- **FIFO ordering** maintained even after timeout/requeue
- **Atomic pickup** - messages locked while picked up
- **Non-blocking pickup** - returns immediately if queue empty
- **SHA256 checksums** validated by server

## Testing

```bash
# Run server tests
cd server
go test -v -coverprofile=coverage.out ./...

# Run Go client tests
cd go-client
go test -v ./...

# Format check
gofmt -s -l .
go vet ./...
```

## CI/CD

GitHub Actions workflows run on every PR:
- Server: build, test, coverage, format check, vet
- Go client: build, test, format check, vet
- Python client: multi-version test, black, mypy, ruff

## License

See specification.md for project details.
