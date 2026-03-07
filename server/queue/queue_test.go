package queue

import (
	"sync"
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	msg, err := NewMessage("test content", "checksum123")
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	if msg.ID == "" {
		t.Error("Message ID should not be empty")
	}

	if msg.Content != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", msg.Content)
	}

	if msg.Checksum != "checksum123" {
		t.Errorf("Expected checksum 'checksum123', got '%s'", msg.Checksum)
	}

	if msg.State != StateAvailable {
		t.Errorf("New message should be in StateAvailable, got %v", msg.State)
	}
}

func TestQueueAddAndPickup(t *testing.T) {
	q := NewQueue("test-queue")

	msg1, _ := NewMessage("message 1", "check1")
	msg2, _ := NewMessage("message 2", "check2")

	q.Add(msg1)
	q.Add(msg2)

	if q.Size() != 2 {
		t.Errorf("Expected queue size 2, got %d", q.Size())
	}

	pickedMsg, err := q.Pickup(10)
	if err != nil {
		t.Fatalf("Failed to pickup message: %v", err)
	}

	if pickedMsg.ID != msg1.ID {
		t.Error("Should pickup first message (FIFO)")
	}

	if pickedMsg.State != StatePickedUp {
		t.Error("Picked up message should be in StatePickedUp")
	}

	if q.AvailableCount() != 1 {
		t.Errorf("Expected 1 available message, got %d", q.AvailableCount())
	}
}

func TestQueueEmpty(t *testing.T) {
	q := NewQueue("test-queue")

	_, err := q.Pickup(10)
	if err != ErrQueueEmpty {
		t.Errorf("Expected ErrQueueEmpty, got %v", err)
	}
}

func TestQueueDelete(t *testing.T) {
	q := NewQueue("test-queue")

	msg, _ := NewMessage("message", "check")
	q.Add(msg)

	if err := q.Delete(msg.ID); err != nil {
		t.Fatalf("Failed to delete message: %v", err)
	}

	if q.Size() != 0 {
		t.Errorf("Expected queue size 0 after delete, got %d", q.Size())
	}
}

func TestQueueDeleteNotFound(t *testing.T) {
	q := NewQueue("test-queue")

	err := q.Delete("non-existent-id")
	if err != ErrMessageNotFound {
		t.Errorf("Expected ErrMessageNotFound, got %v", err)
	}
}

func TestQueueRequeueTimedOut(t *testing.T) {
	q := NewQueue("test-queue")

	msg, _ := NewMessage("message", "check")
	q.Add(msg)

	pickedMsg, _ := q.Pickup(1)
	pickedMsg.TimeoutAt = time.Now().Add(-1 * time.Second)

	count := q.RequeueTimedOut()
	if count != 1 {
		t.Errorf("Expected 1 message to be requeued, got %d", count)
	}

	if pickedMsg.State != StateAvailable {
		t.Error("Timed out message should be back to StateAvailable")
	}

	if q.AvailableCount() != 1 {
		t.Errorf("Expected 1 available message after requeue, got %d", q.AvailableCount())
	}
}

func TestQueueConcurrency(t *testing.T) {
	q := NewQueue("test-queue")

	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			msg, _ := NewMessage("message", "check")
			q.Add(msg)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			q.Pickup(10)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	<-done
	<-done

	if q.Size() < 50 {
		t.Error("Queue should have at least 50 messages after concurrent operations")
	}
}

func TestQueuePickupWithWaitWakesOnAdd(t *testing.T) {
	q := NewQueue("test-queue")

	done := make(chan *Message, 1)
	go func() {
		msg, err := q.PickupWithWait(10, 1)
		if err != nil {
			done <- nil
			return
		}
		done <- msg
	}()

	time.Sleep(50 * time.Millisecond)

	msg, _ := NewMessage("hello", "check")
	q.Add(msg)

	select {
	case picked := <-done:
		if picked == nil {
			t.Fatal("expected pickup to succeed after add")
		}
		if picked.ID != msg.ID {
			t.Fatalf("expected message %s, got %s", msg.ID, picked.ID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for pickup")
	}
}

func TestQueuePickupWithWaitTimesOut(t *testing.T) {
	q := NewQueue("test-queue")

	start := time.Now()
	_, err := q.PickupWithWait(10, 1)
	elapsed := time.Since(start)
	if err != ErrQueueEmpty {
		t.Fatalf("expected ErrQueueEmpty, got %v", err)
	}
	if elapsed < 900*time.Millisecond {
		t.Fatalf("pickup returned too quickly: %v", elapsed)
	}
}

func TestQueuePickupWithWaitMultipleWaiters(t *testing.T) {
	q := NewQueue("test-queue")

	const waiters = 3
	var wg sync.WaitGroup
	wg.Add(waiters)

	results := make(chan error, waiters)
	for i := 0; i < waiters; i++ {
		go func() {
			defer wg.Done()
			_, err := q.PickupWithWait(10, 2)
			results <- err
		}()
	}

	time.Sleep(100 * time.Millisecond)
	for i := 0; i < waiters; i++ {
		msg, _ := NewMessage("message", "check")
		q.Add(msg)
	}

	wg.Wait()
	close(results)

	for err := range results {
		if err != nil {
			t.Fatalf("expected waiter to receive message, got err %v", err)
		}
	}
}
