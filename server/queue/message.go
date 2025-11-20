package queue

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type MessageState int

const (
	StateAvailable MessageState = iota
	StatePickedUp
)

type Message struct {
	ID             string
	Content        string
	Checksum       string
	State          MessageState
	TimeoutAt      time.Time
	OriginalIndex  int
}

func NewMessage(content, checksum string) (*Message, error) {
	id, err := generateGUID()
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:       id,
		Content:  content,
		Checksum: checksum,
		State:    StateAvailable,
	}, nil
}

func generateGUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
