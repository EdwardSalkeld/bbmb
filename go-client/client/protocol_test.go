package client

import (
	"bytes"
	"io"
	"testing"
)

func TestReadFrameRejectsZeroLength(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte{0, 0, 0, 0})
	buf.WriteByte(byte(CmdAddMessage))

	_, _, err := readFrame(&buf)
	if err != io.ErrUnexpectedEOF {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

func TestWriteReadFrameRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	payload := []byte("payload")

	if err := writeFrame(&buf, CmdPickupMessage, payload); err != nil {
		t.Fatalf("writeFrame failed: %v", err)
	}

	cmdType, decodedPayload, err := readFrame(&buf)
	if err != nil {
		t.Fatalf("readFrame failed: %v", err)
	}
	if cmdType != CmdPickupMessage {
		t.Fatalf("expected cmd %d, got %d", CmdPickupMessage, cmdType)
	}
	if !bytes.Equal(decodedPayload, payload) {
		t.Fatalf("payload mismatch: expected %q, got %q", payload, decodedPayload)
	}
}
