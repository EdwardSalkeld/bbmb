package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edsalkeld/bbmb/server/queue"
)

func TestServeHTTPQueueMetricMetadataEmittedOnce(t *testing.T) {
	manager := queue.NewManager()
	manager.EnsureQueue("q")
	collector := NewCollector(manager)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	collector.ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Count(body, "# HELP bbmb_queue_messages_total") != 1 {
		t.Fatalf("expected one HELP for queue total metric")
	}
	if strings.Count(body, "# TYPE bbmb_queue_messages_total gauge") != 1 {
		t.Fatalf("expected one TYPE for queue total metric")
	}
	if strings.Count(body, "# HELP bbmb_queue_messages_available") != 1 {
		t.Fatalf("expected one HELP for queue available metric")
	}
	if strings.Count(body, "# TYPE bbmb_queue_messages_available gauge") != 1 {
		t.Fatalf("expected one TYPE for queue available metric")
	}
}

func TestServeHTTPEscapesQueueLabelValue(t *testing.T) {
	manager := queue.NewManager()
	queueName := "a\"b\\c\nd"
	manager.EnsureQueue(queueName)
	collector := NewCollector(manager)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	collector.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `bbmb_queue_messages_total{queue="a\"b\\c\nd"}`) {
		t.Fatalf("expected escaped queue label in metrics output, got: %s", body)
	}
}

func TestServeHTTPIncludesLongPollMetrics(t *testing.T) {
	manager := queue.NewManager()
	collector := NewCollector(manager)
	collector.IncrPickupWaits()
	collector.IncrPickupWaits()
	collector.ObservePickupWaitDuration(1500000000)
	collector.IncrEmptyAfterWait()

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	collector.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "bbmb_pickup_waits_total 2") {
		t.Fatalf("expected pickup waits metric, got: %s", body)
	}
	if !strings.Contains(body, "bbmb_pickup_wait_duration_seconds_total 1.500000") {
		t.Fatalf("expected pickup wait duration metric, got: %s", body)
	}
	if !strings.Contains(body, "bbmb_empty_after_wait_total 1") {
		t.Fatalf("expected empty-after-wait metric, got: %s", body)
	}
}
