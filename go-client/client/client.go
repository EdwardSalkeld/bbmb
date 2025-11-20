package client

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
)

var (
	ErrQueueEmpty      = errors.New("queue is empty")
	ErrNotFound        = errors.New("message not found")
	ErrInvalidChecksum = errors.New("invalid checksum")
	ErrMessageTooLarge = errors.New("message too large")
	ErrServerError     = errors.New("server internal error")
)

type Client struct {
	address string
	conn    net.Conn
}

func NewClient(address string) *Client {
	return &Client{
		address: address,
	}
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) EnsureQueue(queueName string) error {
	payload := encodeEnsureQueue(queueName)

	if err := writeFrame(c.conn, CmdEnsureQueue, payload); err != nil {
		return err
	}

	_, respPayload, err := readFrame(c.conn)
	if err != nil {
		return err
	}

	status, err := decodeStatusResponse(respPayload)
	if err != nil {
		return err
	}

	if status != StatusOK {
		return fmt.Errorf("ensure queue failed with status: %d", status)
	}

	return nil
}

func (c *Client) AddMessage(queueName, content string) (string, error) {
	if len(content) > MaxMessageSize {
		return "", ErrMessageTooLarge
	}

	hash := sha256.Sum256([]byte(content))
	checksum := hex.EncodeToString(hash[:])

	payload := encodeAddMessage(queueName, content, checksum)

	if err := writeFrame(c.conn, CmdAddMessage, payload); err != nil {
		return "", err
	}

	_, respPayload, err := readFrame(c.conn)
	if err != nil {
		return "", err
	}

	status, guid, err := decodeAddMessageResponse(respPayload)
	if err != nil {
		return "", err
	}

	switch status {
	case StatusOK:
		return guid, nil
	case StatusInvalidChecksum:
		return "", ErrInvalidChecksum
	case StatusMessageTooLarge:
		return "", ErrMessageTooLarge
	case StatusInternalError:
		return "", ErrServerError
	default:
		return "", fmt.Errorf("unexpected status: %d", status)
	}
}

type Message struct {
	GUID     string
	Content  string
	Checksum string
}

func (c *Client) PickupMessage(queueName string, timeoutSeconds int) (*Message, error) {
	payload := encodePickupMessage(queueName, timeoutSeconds)

	if err := writeFrame(c.conn, CmdPickupMessage, payload); err != nil {
		return nil, err
	}

	_, respPayload, err := readFrame(c.conn)
	if err != nil {
		return nil, err
	}

	status, guid, content, checksum, err := decodePickupMessageResponse(respPayload)
	if err != nil {
		return nil, err
	}

	switch status {
	case StatusOK:
		return &Message{
			GUID:     guid,
			Content:  content,
			Checksum: checksum,
		}, nil
	case StatusEmptyQueue:
		return nil, ErrQueueEmpty
	case StatusInternalError:
		return nil, ErrServerError
	default:
		return nil, fmt.Errorf("unexpected status: %d", status)
	}
}

func (c *Client) DeleteMessage(queueName, guid string) error {
	payload := encodeDeleteMessage(queueName, guid)

	if err := writeFrame(c.conn, CmdDeleteMessage, payload); err != nil {
		return err
	}

	_, respPayload, err := readFrame(c.conn)
	if err != nil {
		return err
	}

	status, err := decodeStatusResponse(respPayload)
	if err != nil {
		return err
	}

	switch status {
	case StatusOK:
		return nil
	case StatusNotFound:
		return ErrNotFound
	case StatusInternalError:
		return ErrServerError
	default:
		return fmt.Errorf("unexpected status: %d", status)
	}
}
