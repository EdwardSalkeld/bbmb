package protocol

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeEnsureQueue(t *testing.T) {
	req := &EnsureQueueRequest{
		QueueName: "test-queue",
	}

	buf := writeString(req.QueueName)
	decoded, err := DecodeEnsureQueue(buf)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if decoded.QueueName != req.QueueName {
		t.Errorf("Expected queue name '%s', got '%s'", req.QueueName, decoded.QueueName)
	}
}

func TestEncodeDecodeAddMessage(t *testing.T) {
	req := &AddMessageRequest{
		QueueName: "test-queue",
		Content:   "test content",
		Checksum:  "abc123",
	}

	buf := append(writeString(req.QueueName), writeString(req.Content)...)
	buf = append(buf, writeString(req.Checksum)...)

	decoded, err := DecodeAddMessage(buf)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if decoded.QueueName != req.QueueName {
		t.Errorf("Expected queue name '%s', got '%s'", req.QueueName, decoded.QueueName)
	}
	if decoded.Content != req.Content {
		t.Errorf("Expected content '%s', got '%s'", req.Content, decoded.Content)
	}
	if decoded.Checksum != req.Checksum {
		t.Errorf("Expected checksum '%s', got '%s'", req.Checksum, decoded.Checksum)
	}
}

func TestEncodeDecodePickupMessage(t *testing.T) {
	req := &PickupMessageRequest{
		QueueName:      "test-queue",
		TimeoutSeconds: 30,
	}

	buf := writeString(req.QueueName)
	timeoutBuf := make([]byte, 4)
	timeoutBuf[0] = 0
	timeoutBuf[1] = 0
	timeoutBuf[2] = 0
	timeoutBuf[3] = 30
	buf = append(buf, timeoutBuf...)

	decoded, err := DecodePickupMessage(buf)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if decoded.QueueName != req.QueueName {
		t.Errorf("Expected queue name '%s', got '%s'", req.QueueName, decoded.QueueName)
	}
	if decoded.TimeoutSeconds != req.TimeoutSeconds {
		t.Errorf("Expected timeout %d, got %d", req.TimeoutSeconds, decoded.TimeoutSeconds)
	}
}

func TestEncodeDecodeDeleteMessage(t *testing.T) {
	req := &DeleteMessageRequest{
		QueueName: "test-queue",
		GUID:      "message-id-123",
	}

	buf := append(writeString(req.QueueName), writeString(req.GUID)...)

	decoded, err := DecodeDeleteMessage(buf)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if decoded.QueueName != req.QueueName {
		t.Errorf("Expected queue name '%s', got '%s'", req.QueueName, decoded.QueueName)
	}
	if decoded.GUID != req.GUID {
		t.Errorf("Expected GUID '%s', got '%s'", req.GUID, decoded.GUID)
	}
}

func TestFrameEncodeDecode(t *testing.T) {
	payload := []byte("test payload")
	cmdType := CmdAddMessage

	var buf bytes.Buffer
	if err := WriteFrame(&buf, cmdType, payload); err != nil {
		t.Fatalf("Failed to write frame: %v", err)
	}

	decodedCmd, decodedPayload, err := ReadFrame(&buf)
	if err != nil {
		t.Fatalf("Failed to read frame: %v", err)
	}

	if decodedCmd != cmdType {
		t.Errorf("Expected command type %d, got %d", cmdType, decodedCmd)
	}

	if !bytes.Equal(decodedPayload, payload) {
		t.Errorf("Payload mismatch: expected %v, got %v", payload, decodedPayload)
	}
}

func TestMessageTooLarge(t *testing.T) {
	payload := make([]byte, MaxMessageSize+2000)
	cmdType := CmdAddMessage

	var buf bytes.Buffer
	if err := WriteFrame(&buf, cmdType, payload); err != nil {
		t.Fatalf("Failed to write frame: %v", err)
	}

	_, _, err := ReadFrame(&buf)
	if err != ErrMessageTooLarge {
		t.Errorf("Expected ErrMessageTooLarge, got %v", err)
	}
}

func TestResponseEncoding(t *testing.T) {
	addResp := &AddMessageResponse{
		Status: StatusOK,
		GUID:   "test-guid-123",
	}

	encoded := EncodeAddMessageResponse(addResp)
	if encoded[0] != byte(StatusOK) {
		t.Errorf("Expected status byte %d, got %d", StatusOK, encoded[0])
	}

	pickupResp := &PickupMessageResponse{
		Status:   StatusOK,
		GUID:     "test-guid",
		Content:  "test content",
		Checksum: "checksum",
	}

	encoded = EncodePickupMessageResponse(pickupResp)
	if encoded[0] != byte(StatusOK) {
		t.Errorf("Expected status byte %d, got %d", StatusOK, encoded[0])
	}
}
