# BBMB (Basic Bare-bones Message Broker) Implementation Plan

## Project Overview
A TCP-based message broker with FIFO queues, written in Go with client implementations in Go and Python.

## Architecture

### Core Components

1. **Message Broker Server (Go)**
   - TCP server listening on hardcoded port (e.g., 9876)
   - In-memory queue manager with FIFO semantics
   - Message timeout/requeue mechanism
   - Prometheus metrics endpoint

2. **Client Libraries**
   - Go client library + CLI wrapper
   - Python client library + CLI wrapper

## Implementation Phases

### Phase 1: Core Server Infrastructure ✅

#### 1.1 Project Setup ✅
- [x] Initialize Go module (`bbmb-server`)
- [x] Set up project structure:
  ```
  bbmb/
  ├── server/
  │   ├── main.go
  │   ├── queue/
  │   │   ├── queue.go (queue implementation)
  │   │   ├── message.go (message struct)
  │   │   └── manager.go (queue manager)
  │   ├── protocol/
  │   │   ├── codec.go (encode/decode messages)
  │   │   └── commands.go (command types)
  │   ├── tcp/
  │   │   └── server.go (TCP server)
  │   └── metrics/
  │       └── prometheus.go (metrics collection)
  ├── go-client/
  │   ├── client/
  │   │   └── client.go (Go client library)
  │   └── cmd/
  │       └── main.go (CLI wrapper)
  └── python-client/
      ├── bbmb_client/
      │   └── client.py
      └── cli.py
  ```

#### 1.2 Message & Queue Data Structures ✅
- [x] Define Message struct:
  - GUID (generated on add)
  - Content (arbitrary text)
  - Checksum (for verification)
  - Timeout timestamp (when picked up)
  - State (available/picked_up)
- [x] Implement FIFO Queue:
  - Thread-safe queue operations
  - Lock-based or channel-based synchronization
  - Support for marking messages as "picked up"
  - Timeout-based requeue mechanism
- [x] Implement Queue Manager:
  - Map of queue name -> Queue
  - Ensure queue exists operation
  - Retrieve queue by name

#### 1.3 Protocol Design
Define binary protocol over TCP (length-prefixed frames):

**Message Format:**
```
[4 bytes: total length][1 byte: command type][payload]
```

**Command Types:**
1. **ENSURE_QUEUE (0x01)**
   - Request: `[queue_name_length][queue_name]`
   - Response: `[status: 0x00=success]`

2. **ADD_MESSAGE (0x02)**
   - Request: `[queue_name_length][queue_name][content_length][content][checksum_length][checksum]`
   - Response: `[status][guid_length][guid]`

3. **PICKUP_MESSAGE (0x03)**
   - Request: `[queue_name_length][queue_name][timeout_seconds]`
   - Response: `[status][guid_length][guid][content_length][content][checksum_length][checksum]`
   - Status: 0x00=success, 0x01=empty queue

4. **DELETE_MESSAGE (0x04)**
   - Request: `[queue_name_length][queue_name][guid_length][guid]`
   - Response: `[status: 0x00=success, 0x01=not_found]`

#### 1.4 TCP Server ✅
- [x] Create TCP listener on hardcoded port (9876)
- [x] Handle concurrent connections (goroutine per connection)
- [x] Implement protocol codec (encode/decode)
- [x] Wire up command handlers to queue manager
- [x] Add graceful shutdown handling

#### 1.5 Timeout Requeue Mechanism ✅
- [x] Background goroutine that periodically scans for timed-out messages
- [x] Return timed-out messages to "available" state
- [x] Consider using a priority queue or sorted structure for efficiency

### Phase 2: Observability ✅

#### 2.1 Prometheus Metrics ✅
- [x] Set up HTTP server for metrics endpoint (e.g., port 9877)
- [x] Implement metrics:
  - **Global:**
    - Total messages added
    - Total messages picked up
    - Total messages deleted
    - Total active messages
    - Uptime
    - Memory usage (via runtime.MemStats)
    - Number of active queues
    - Number of active connections
  - **Per-queue:**
    - Messages in queue (gauge)
    - Messages added (counter)
    - Messages picked up (counter)
    - Messages deleted (counter)
    - Messages timed out (counter)

### Phase 3: Go Client ✅

#### 3.1 Go Client Library ✅
- [x] Initialize Go module (`bbmb-client-go`)
- [x] Implement client struct with connection pool
- [x] Implement protocol encoding/decoding (share with server if possible)
- [x] Implement API methods:
  - `EnsureQueue(queueName string) error`
  - `AddMessage(queueName, content, checksum string) (guid string, error)`
  - `PickupMessage(queueName string, timeout int) (guid, content, checksum string, error)`
  - `DeleteMessage(queueName, guid string) error`
- [x] Connection management (connect, reconnect, close)
- [x] Error handling

#### 3.2 Go CLI Wrapper ✅
- [x] Create CLI using standard `flag` package (avoid cobra/etc)
- [x] Commands:
  - `ensure-queue --queue=<name>`
  - `add --queue=<name> --content=<text>`
  - `pickup --queue=<name> --timeout=<seconds>`
  - `delete --queue=<name> --guid=<id>`
- [x] Support server address as flag (default to localhost:9876)

### Phase 4: Python Client ✅

#### 4.1 Python Client Library ✅
- [x] Create package structure with `setup.py` or `pyproject.toml`
- [x] Implement client class
- [x] Implement protocol encoding/decoding
- [x] Implement API methods:
  - `ensure_queue(queue_name: str) -> None`
  - `add_message(queue_name: str, content: str) -> str`
  - `pickup_message(queue_name: str, timeout: int) -> Message`
  - `delete_message(queue_name: str, guid: str) -> None`
- [x] Connection management
- [x] Error handling and exceptions

#### 4.2 Python CLI Wrapper ✅
- [x] Create CLI using `argparse` (avoid click/etc)
- [x] Same commands as Go CLI
- [x] Entry point script

### Phase 5: Testing ✅

#### 5.1 Server Tests ✅
- [x] Unit tests for queue operations
- [x] Unit tests for message lifecycle (add, pickup, delete, timeout)
- [x] Unit tests for protocol encoding/decoding
- [x] Integration tests for TCP server
- [x] Test concurrent access to queues
- [x] Test timeout requeue mechanism
- [ ] Throughput/benchmark tests:
  - Messages per second (single queue)
  - Messages per second (multiple queues)
  - Latency measurements

#### 5.2 Go Client Tests
- [ ] Unit tests for protocol encoding/decoding
- [ ] Integration tests against running server
- [ ] CLI tests

#### 5.3 Python Client Tests
- [ ] Unit tests for protocol encoding/decoding
- [ ] Integration tests against running server
- [ ] CLI tests

### Phase 6: CI/CD ✅

#### 6.1 GitHub Workflows ✅
- [x] **Server workflow:**
  - Run tests on PR
  - Check Go formatting/linting
  - Run benchmarks
- [x] **Go client workflow:**
  - Run tests on PR
  - Check Go formatting/linting
- [x] **Python client workflow:**
  - Run tests on PR
  - Check Python formatting (black/ruff)
  - Type checking (mypy)

### Phase 7: Documentation ✅

- [x] README.md for main project
- [x] Protocol specification document (in README)
- [x] Server documentation with:
  - How to run
  - Metrics documentation
- [x] Go client documentation with:
  - Installation
  - Usage examples (library + CLI)
- [x] Python client documentation with:
  - Installation
  - Usage examples (library + CLI)

## Technical Decisions

### 1. GUID Generation
Use UUID v4 (random) for message IDs. Go: `google/uuid` or `crypto/rand`

### 2. Checksum Verification
Use **SHA256** for all message checksums. The server will:
- Calculate SHA256 hash of message content on ADD
- Validate provided checksum matches calculated hash
- Reject messages with invalid checksums
- Return the SHA256 checksum on PICKUP

### 3. Thread Safety
Use Go's `sync.Mutex` for queue operations or consider using channels for message passing.

### 4. Timeout Scanning
Background goroutine runs every N seconds (e.g., 1 second) to check for timed-out messages.

### 5. Ports
- TCP Server: 9876
- Metrics HTTP: 9877

### 6. Error Handling
Use status codes in protocol. Clients raise exceptions/errors with meaningful messages.

## Design Decisions (Clarified)

1. **Checksum validation:** Server calculates and validates **SHA256** checksums. ADD requests must include correct SHA256 hash or will be rejected.

2. **Persistence:** In-memory only for initial implementation. Architecture should allow for future persistence layer addition.

3. **Message ordering on timeout:** Timed-out messages return to the **front** of the queue to maintain FIFO semantics and original ordering.

4. **Pickup atomicity:** Pickup is **atomic** - once a message is picked up, it becomes unavailable to other clients until either deleted or timed out.

5. **Empty queue behavior:** Pickup returns immediately with "empty" status (non-blocking). No waiting or long-polling.

6. **Message size limit:** **1MB maximum** per message to prevent memory exhaustion while allowing reasonable payloads.

7. **Queue size limit:** No per-queue message limit initially. Rely on monitoring to detect memory issues.

## Development Workflow

1. Commit locally in small increments (as specified)
2. Test each component thoroughly before moving to next phase
3. Start with server core, then observability, then clients
4. Keep it simple - no premature optimization

## Estimated Component Sizes

- **Server:** ~1000-1500 lines of Go
- **Go Client:** ~500-800 lines of Go
- **Python Client:** ~500-800 lines of Python
- **Tests:** 1-2x the code size
