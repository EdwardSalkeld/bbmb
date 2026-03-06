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
