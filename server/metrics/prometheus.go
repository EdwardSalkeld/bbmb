package metrics

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/edsalkeld/bbmb/server/queue"
)

type Collector struct {
	manager           *queue.Manager
	startTime         time.Time
	messagesAdded     atomic.Uint64
	messagesPickedUp  atomic.Uint64
	messagesDeleted   atomic.Uint64
	messagesTimedOut  atomic.Uint64
	pickupWaits       atomic.Uint64
	pickupWaitNanos   atomic.Uint64
	emptyAfterWait    atomic.Uint64
	activeConnections atomic.Int32
}

func NewCollector(manager *queue.Manager) *Collector {
	return &Collector{
		manager:   manager,
		startTime: time.Now(),
	}
}

func (c *Collector) IncrMessagesAdded() {
	c.messagesAdded.Add(1)
}

func (c *Collector) IncrMessagesPickedUp() {
	c.messagesPickedUp.Add(1)
}

func (c *Collector) IncrMessagesDeleted() {
	c.messagesDeleted.Add(1)
}

func (c *Collector) IncrMessagesTimedOut() {
	c.messagesTimedOut.Add(1)
}

func (c *Collector) IncrPickupWaits() {
	c.pickupWaits.Add(1)
}

func (c *Collector) ObservePickupWaitDuration(d time.Duration) {
	c.pickupWaitNanos.Add(uint64(d.Nanoseconds()))
}

func (c *Collector) IncrEmptyAfterWait() {
	c.emptyAfterWait.Add(1)
}

func (c *Collector) IncrActiveConnections() {
	c.activeConnections.Add(1)
}

func (c *Collector) DecrActiveConnections() {
	c.activeConnections.Add(-1)
}

func (c *Collector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	uptime := time.Since(c.startTime).Seconds()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Fprintf(w, "# HELP bbmb_uptime_seconds Server uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE bbmb_uptime_seconds gauge\n")
	fmt.Fprintf(w, "bbmb_uptime_seconds %.2f\n", uptime)

	fmt.Fprintf(w, "# HELP bbmb_memory_alloc_bytes Current memory allocated in bytes\n")
	fmt.Fprintf(w, "# TYPE bbmb_memory_alloc_bytes gauge\n")
	fmt.Fprintf(w, "bbmb_memory_alloc_bytes %d\n", m.Alloc)

	fmt.Fprintf(w, "# HELP bbmb_memory_total_alloc_bytes Total memory allocated in bytes\n")
	fmt.Fprintf(w, "# TYPE bbmb_memory_total_alloc_bytes counter\n")
	fmt.Fprintf(w, "bbmb_memory_total_alloc_bytes %d\n", m.TotalAlloc)

	fmt.Fprintf(w, "# HELP bbmb_memory_sys_bytes Memory obtained from system in bytes\n")
	fmt.Fprintf(w, "# TYPE bbmb_memory_sys_bytes gauge\n")
	fmt.Fprintf(w, "bbmb_memory_sys_bytes %d\n", m.Sys)

	fmt.Fprintf(w, "# HELP bbmb_queues_total Total number of queues\n")
	fmt.Fprintf(w, "# TYPE bbmb_queues_total gauge\n")
	fmt.Fprintf(w, "bbmb_queues_total %d\n", c.manager.QueueCount())

	fmt.Fprintf(w, "# HELP bbmb_messages_added_total Total messages added\n")
	fmt.Fprintf(w, "# TYPE bbmb_messages_added_total counter\n")
	fmt.Fprintf(w, "bbmb_messages_added_total %d\n", c.messagesAdded.Load())

	fmt.Fprintf(w, "# HELP bbmb_messages_picked_up_total Total messages picked up\n")
	fmt.Fprintf(w, "# TYPE bbmb_messages_picked_up_total counter\n")
	fmt.Fprintf(w, "bbmb_messages_picked_up_total %d\n", c.messagesPickedUp.Load())

	fmt.Fprintf(w, "# HELP bbmb_messages_deleted_total Total messages deleted\n")
	fmt.Fprintf(w, "# TYPE bbmb_messages_deleted_total counter\n")
	fmt.Fprintf(w, "bbmb_messages_deleted_total %d\n", c.messagesDeleted.Load())

	fmt.Fprintf(w, "# HELP bbmb_messages_timed_out_total Total messages timed out\n")
	fmt.Fprintf(w, "# TYPE bbmb_messages_timed_out_total counter\n")
	fmt.Fprintf(w, "bbmb_messages_timed_out_total %d\n", c.messagesTimedOut.Load())

	fmt.Fprintf(w, "# HELP bbmb_pickup_waits_total Total pickup requests with long-poll wait enabled\n")
	fmt.Fprintf(w, "# TYPE bbmb_pickup_waits_total counter\n")
	fmt.Fprintf(w, "bbmb_pickup_waits_total %d\n", c.pickupWaits.Load())

	fmt.Fprintf(w, "# HELP bbmb_pickup_wait_duration_seconds_total Total wall time spent in long-poll waits\n")
	fmt.Fprintf(w, "# TYPE bbmb_pickup_wait_duration_seconds_total counter\n")
	fmt.Fprintf(w, "bbmb_pickup_wait_duration_seconds_total %.6f\n", float64(c.pickupWaitNanos.Load())/1e9)

	fmt.Fprintf(w, "# HELP bbmb_empty_after_wait_total Total long-poll pickup requests that returned empty\n")
	fmt.Fprintf(w, "# TYPE bbmb_empty_after_wait_total counter\n")
	fmt.Fprintf(w, "bbmb_empty_after_wait_total %d\n", c.emptyAfterWait.Load())

	fmt.Fprintf(w, "# HELP bbmb_active_connections Current number of active connections\n")
	fmt.Fprintf(w, "# TYPE bbmb_active_connections gauge\n")
	fmt.Fprintf(w, "bbmb_active_connections %d\n", c.activeConnections.Load())

	queues := c.manager.GetAllQueues()
	fmt.Fprintf(w, "# HELP bbmb_queue_messages_total Total messages in queue\n")
	fmt.Fprintf(w, "# TYPE bbmb_queue_messages_total gauge\n")
	fmt.Fprintf(w, "# HELP bbmb_queue_messages_available Available messages in queue\n")
	fmt.Fprintf(w, "# TYPE bbmb_queue_messages_available gauge\n")
	for _, q := range queues {
		queueName := escapeLabelValue(q.Name())
		size := q.Size()
		available := q.AvailableCount()

		fmt.Fprintf(w, "bbmb_queue_messages_total{queue=\"%s\"} %d\n", queueName, size)

		fmt.Fprintf(w, "bbmb_queue_messages_available{queue=\"%s\"} %d\n", queueName, available)
	}
}

func escapeLabelValue(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\n", "\\n",
		"\"", "\\\"",
	)
	return replacer.Replace(value)
}

func (c *Collector) StartServer(address string) {
	http.Handle("/metrics", c)
	log.Printf("Metrics server listening on %s", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Failed to start metrics server: %v", err)
	}
}
