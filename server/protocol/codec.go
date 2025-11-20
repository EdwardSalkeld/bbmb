package protocol

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrMessageTooLarge = errors.New("message too large")
	ErrInvalidCommand  = errors.New("invalid command type")
)

func ReadFrame(r io.Reader) (CommandType, []byte, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return 0, nil, err
	}

	if length > MaxMessageSize+1024 {
		return 0, nil, ErrMessageTooLarge
	}

	var cmdType CommandType
	if err := binary.Read(r, binary.BigEndian, &cmdType); err != nil {
		return 0, nil, err
	}

	payload := make([]byte, length-1)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, err
	}

	return cmdType, payload, nil
}

func WriteFrame(w io.Writer, cmdType CommandType, payload []byte) error {
	length := uint32(len(payload) + 1)

	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, cmdType); err != nil {
		return err
	}

	if _, err := w.Write(payload); err != nil {
		return err
	}

	return nil
}

func readString(data []byte, offset int) (string, int, error) {
	if offset+4 > len(data) {
		return "", 0, io.ErrUnexpectedEOF
	}

	length := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	if offset+int(length) > len(data) {
		return "", 0, io.ErrUnexpectedEOF
	}

	str := string(data[offset : offset+int(length)])
	offset += int(length)

	return str, offset, nil
}

func writeString(str string) []byte {
	length := uint32(len(str))
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	copy(buf[4:], str)
	return buf
}

func DecodeEnsureQueue(payload []byte) (*EnsureQueueRequest, error) {
	queueName, _, err := readString(payload, 0)
	if err != nil {
		return nil, err
	}

	return &EnsureQueueRequest{
		QueueName: queueName,
	}, nil
}

func EncodeEnsureQueueResponse(resp *EnsureQueueResponse) []byte {
	return []byte{byte(resp.Status)}
}

func DecodeAddMessage(payload []byte) (*AddMessageRequest, error) {
	offset := 0

	queueName, newOffset, err := readString(payload, offset)
	if err != nil {
		return nil, err
	}
	offset = newOffset

	content, newOffset, err := readString(payload, offset)
	if err != nil {
		return nil, err
	}
	offset = newOffset

	checksum, _, err := readString(payload, offset)
	if err != nil {
		return nil, err
	}

	return &AddMessageRequest{
		QueueName: queueName,
		Content:   content,
		Checksum:  checksum,
	}, nil
}

func EncodeAddMessageResponse(resp *AddMessageResponse) []byte {
	buf := []byte{byte(resp.Status)}
	if resp.Status == StatusOK {
		buf = append(buf, writeString(resp.GUID)...)
	}
	return buf
}

func DecodePickupMessage(payload []byte) (*PickupMessageRequest, error) {
	queueName, offset, err := readString(payload, 0)
	if err != nil {
		return nil, err
	}

	if offset+4 > len(payload) {
		return nil, io.ErrUnexpectedEOF
	}

	timeoutSeconds := int(binary.BigEndian.Uint32(payload[offset : offset+4]))

	return &PickupMessageRequest{
		QueueName:      queueName,
		TimeoutSeconds: timeoutSeconds,
	}, nil
}

func EncodePickupMessageResponse(resp *PickupMessageResponse) []byte {
	buf := []byte{byte(resp.Status)}
	if resp.Status == StatusOK {
		buf = append(buf, writeString(resp.GUID)...)
		buf = append(buf, writeString(resp.Content)...)
		buf = append(buf, writeString(resp.Checksum)...)
	}
	return buf
}

func DecodeDeleteMessage(payload []byte) (*DeleteMessageRequest, error) {
	offset := 0

	queueName, newOffset, err := readString(payload, offset)
	if err != nil {
		return nil, err
	}
	offset = newOffset

	guid, _, err := readString(payload, offset)
	if err != nil {
		return nil, err
	}

	return &DeleteMessageRequest{
		QueueName: queueName,
		GUID:      guid,
	}, nil
}

func EncodeDeleteMessageResponse(resp *DeleteMessageResponse) []byte {
	return []byte{byte(resp.Status)}
}
