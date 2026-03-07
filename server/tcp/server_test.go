package tcp

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/edsalkeld/bbmb/server/metrics"
	"github.com/edsalkeld/bbmb/server/protocol"
	"github.com/edsalkeld/bbmb/server/queue"
)

func TestHandlePickupMessageWaitZeroIsNonBlocking(t *testing.T) {
	s, _ := newTestServer()
	payload := encodePickupPayload("q", 30, 0, true)

	start := time.Now()
	resp := s.handlePickupMessage(payload)
	elapsed := time.Since(start)

	if got := protocol.StatusCode(resp[0]); got != protocol.StatusEmptyQueue {
		t.Fatalf("expected empty queue status, got %d", got)
	}
	if elapsed > 100*time.Millisecond {
		t.Fatalf("expected non-blocking response, took %v", elapsed)
	}
}

func TestHandlePickupMessageWaitBlocksAndWakesOnAdd(t *testing.T) {
	s, manager := newTestServer()
	manager.EnsureQueue("q")
	q, _ := manager.GetQueue("q")

	result := make(chan []byte, 1)
	start := time.Now()
	go func() {
		result <- s.handlePickupMessage(encodePickupPayload("q", 30, 2, true))
	}()

	time.Sleep(100 * time.Millisecond)
	msg, _ := queue.NewMessage("hello", "check")
	q.Add(msg)

	resp := <-result
	elapsed := time.Since(start)

	if got := protocol.StatusCode(resp[0]); got != protocol.StatusOK {
		t.Fatalf("expected success status, got %d", got)
	}
	if elapsed < 100*time.Millisecond {
		t.Fatalf("expected pickup to wait before wakeup, took %v", elapsed)
	}
	if elapsed > 1500*time.Millisecond {
		t.Fatalf("expected pickup to wake before wait timeout, took %v", elapsed)
	}
}

func TestHandlePickupMessageSupportsLegacyPayload(t *testing.T) {
	s, manager := newTestServer()
	manager.EnsureQueue("q")
	q, _ := manager.GetQueue("q")
	msg, _ := queue.NewMessage("hello", "check")
	q.Add(msg)

	resp := s.handlePickupMessage(encodePickupPayload("q", 30, 0, false))
	if got := protocol.StatusCode(resp[0]); got != protocol.StatusOK {
		t.Fatalf("expected success status, got %d", got)
	}
}

func newTestServer() (*Server, *queue.Manager) {
	manager := queue.NewManager()
	collector := metrics.NewCollector(manager)
	return NewServer("127.0.0.1:0", manager, collector), manager
}

func encodePickupPayload(queueName string, timeoutSeconds int, waitSeconds int, includeWait bool) []byte {
	name := []byte(queueName)
	buf := make([]byte, 4+len(name)+4)
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(name)))
	copy(buf[4:4+len(name)], name)
	binary.BigEndian.PutUint32(buf[4+len(name):], uint32(timeoutSeconds))
	if !includeWait {
		return buf
	}

	waitBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(waitBuf, uint32(waitSeconds))
	return append(buf, waitBuf...)
}
