package queue

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrQueueEmpty      = errors.New("queue is empty")
	ErrMessageNotFound = errors.New("message not found")
)

type Queue struct {
	name     string
	messages []*Message
	mu       sync.Mutex
}

func NewQueue(name string) *Queue {
	return &Queue{
		name:     name,
		messages: make([]*Message, 0),
	}
}

func (q *Queue) Add(msg *Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.messages = append(q.messages, msg)
}

func (q *Queue) Pickup(timeoutSeconds int) (*Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, msg := range q.messages {
		if msg.State == StateAvailable {
			msg.State = StatePickedUp
			msg.TimeoutAt = time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
			return msg, nil
		}
	}

	return nil, ErrQueueEmpty
}

func (q *Queue) Delete(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, msg := range q.messages {
		if msg.ID == id {
			q.messages = append(q.messages[:i], q.messages[i+1:]...)
			return nil
		}
	}

	return ErrMessageNotFound
}

func (q *Queue) RequeueTimedOut() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := 0
	now := time.Now()

	for _, msg := range q.messages {
		if msg.State == StatePickedUp && now.After(msg.TimeoutAt) {
			msg.State = StateAvailable
			count++
		}
	}

	return count
}

func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.messages)
}

func (q *Queue) AvailableCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := 0
	for _, msg := range q.messages {
		if msg.State == StateAvailable {
			count++
		}
	}
	return count
}

func (q *Queue) Name() string {
	return q.name
}
