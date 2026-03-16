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

# Pick up a message (30 second visibility timeout, non-blocking)
./bbmb-client pickup --queue=myqueue --timeout=30

# Long-poll for up to 5 seconds when queue is empty
./bbmb-client pickup --queue=myqueue --timeout=30 --wait=5

# Delete a message
./bbmb-client delete --queue=myqueue --guid=<message-id>
```

### Using the Python Client

```bash
cd python-client
python -m venv .venv
. .venv/bin/activate
python -m pip install --upgrade pip
python -m pip install -e .

# Ensure queue exists
python cli.py ensure-queue --queue=myqueue

# Add a message
python cli.py add --queue=myqueue --content="Hello from Python!"

# Pick up a message (non-blocking)
python cli.py pickup --queue=myqueue --timeout=30

# Long-poll for up to 5 seconds when queue is empty
python cli.py pickup --queue=myqueue --timeout=30 --wait=5

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
msg, err := c.PickupMessage("myqueue", 30)      // wait defaults to 0
msg, err := c.PickupMessage("myqueue", 30, 5)   // wait up to 5 seconds
err = c.DeleteMessage("myqueue", guid)
```

**Python:**
```python
from bbmb_client import Client

with Client("localhost:9876") as client:
    guid = client.add_message("myqueue", "Hello, World!")
    msg = client.pickup_message("myqueue", timeout_seconds=30)  # wait defaults to 0
    msg = client.pickup_message("myqueue", timeout_seconds=30, wait_seconds=5)
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
3. **PICKUP_MESSAGE (0x03)** - Get message from queue with visibility timeout and optional long-poll wait
4. **DELETE_MESSAGE (0x04)** - Delete message by GUID

### Pickup Timeouts

Pickup now has two independent timeout dimensions:

- `timeout_seconds`: message visibility timeout after successful pickup. If not deleted before this timeout, the message becomes available again.
- `wait_seconds`: long-poll wait timeout while queue is empty. If no message arrives before this timeout, pickup returns `empty`.

Protocol compatibility:
- Legacy pickup payloads (`queue_name + timeout_seconds`) remain supported and are interpreted as `wait_seconds=0`.
- Extended pickup payloads (`queue_name + timeout_seconds + wait_seconds`) enable long polling.

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
- `bbmb_pickup_waits_total` - Total long-poll pickup requests
- `bbmb_pickup_wait_duration_seconds_total` - Total wall time spent waiting in long-poll pickup
- `bbmb_empty_after_wait_total` - Total long-poll pickup requests that timed out empty
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
- **Long-poll pickup support** with backward-compatible wire format (`wait=0` remains non-blocking)
- **SHA256 checksums** validated by server

## Run With systemd

Use the included unit file at `deploy/systemd/bbmb.service`.

```bash
# Build and install the server binary
cd server
go build -o bbmb-server .
sudo install -m 0755 bbmb-server /usr/local/bin/bbmb-server

# Create runtime user and state directory
sudo useradd --system --no-create-home --shell /usr/sbin/nologin bbmb || true
sudo mkdir -p /var/lib/bbmb /etc/bbmb
sudo chown -R bbmb:bbmb /var/lib/bbmb

# Install unit and optional env file
sudo install -m 0644 ../deploy/systemd/bbmb.service /etc/systemd/system/bbmb.service
sudo install -m 0644 ../deploy/systemd/bbmb.env.example /etc/bbmb/bbmb.env

# Enable + start
sudo systemctl daemon-reload
sudo systemctl enable --now bbmb

# Verify
sudo systemctl status bbmb
journalctl -u bbmb -f
```

The service binds to:
- TCP `:9876` for broker traffic
- HTTP `:9877` for Prometheus metrics

## Testing

```bash
# Run server tests
cd server
go test -v ./...
go test -v -coverprofile=coverage.out ./protocol ./queue ./metrics
go tool cover -func=coverage.out

# Run Go client tests
cd go-client
go test -v ./...

# Run Python client checks
cd python-client
python -m venv .venv
. .venv/bin/activate
python -m pip install -e .
python -m pip install black mypy ruff
black --check .
mypy bbmb_client --ignore-missing-imports
python -m unittest discover -s tests -p "test_*.py"
ruff check .

# Format check
gofmt -s -l .
go vet ./...
```

## CI/CD

GitHub Actions workflows run on every PR:
- Server: build, test, coverage, format check, vet
- Go client: build, test, format check, vet
- Python client: multi-version test, black, mypy, ruff

Pushes to `main` also trigger a server release workflow that:
- Runs the server test suite
- Builds `bbmb-server-linux-amd64`
- Uploads the binary, checksum, and tarball as workflow artifacts
- Publishes a GitHub release tagged `v<run-number>` with those assets attached

## License

See specification.md for project details.
