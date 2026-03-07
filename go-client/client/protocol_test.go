package client

import (
	"bytes"
	"encoding/binary"
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

func TestEncodePickupMessageLegacyShape(t *testing.T) {
	payload := encodePickupMessage("q", 30)
	if len(payload) != 9 {
		t.Fatalf("expected payload length 9, got %d", len(payload))
	}

	timeout := binary.BigEndian.Uint32(payload[5:9])
	if timeout != 30 {
		t.Fatalf("expected timeout 30, got %d", timeout)
	}
}

func TestEncodePickupMessageWithWaitSeconds(t *testing.T) {
	payload := encodePickupMessage("q", 30, 5)
	if len(payload) != 13 {
		t.Fatalf("expected payload length 13, got %d", len(payload))
	}

	timeout := binary.BigEndian.Uint32(payload[5:9])
	if timeout != 30 {
		t.Fatalf("expected timeout 30, got %d", timeout)
	}

	wait := binary.BigEndian.Uint32(payload[9:13])
	if wait != 5 {
		t.Fatalf("expected wait 5, got %d", wait)
	}
}
