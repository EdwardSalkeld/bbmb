package client

import (
	"encoding/binary"
	"io"
)

type CommandType byte

const (
	CmdEnsureQueue   CommandType = 0x01
	CmdAddMessage    CommandType = 0x02
	CmdPickupMessage CommandType = 0x03
	CmdDeleteMessage CommandType = 0x04
)

type StatusCode byte

const (
	StatusOK              StatusCode = 0x00
	StatusEmptyQueue      StatusCode = 0x01
	StatusNotFound        StatusCode = 0x02
	StatusInvalidChecksum StatusCode = 0x03
	StatusMessageTooLarge StatusCode = 0x04
	StatusInternalError   StatusCode = 0x05
)

const MaxMessageSize = 1024 * 1024

func writeFrame(w io.Writer, cmdType CommandType, payload []byte) error {
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

func readFrame(r io.Reader) (CommandType, []byte, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return 0, nil, err
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

func writeString(str string) []byte {
	length := uint32(len(str))
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	copy(buf[4:], str)
	return buf
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

func encodeEnsureQueue(queueName string) []byte {
	return writeString(queueName)
}

func encodeAddMessage(queueName, content, checksum string) []byte {
	buf := writeString(queueName)
	buf = append(buf, writeString(content)...)
	buf = append(buf, writeString(checksum)...)
	return buf
}

func encodePickupMessage(queueName string, timeoutSeconds int) []byte {
	buf := writeString(queueName)
	timeoutBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(timeoutBuf, uint32(timeoutSeconds))
	buf = append(buf, timeoutBuf...)
	return buf
}

func encodeDeleteMessage(queueName, guid string) []byte {
	buf := writeString(queueName)
	buf = append(buf, writeString(guid)...)
	return buf
}

func decodeAddMessageResponse(payload []byte) (StatusCode, string, error) {
	if len(payload) < 1 {
		return 0, "", io.ErrUnexpectedEOF
	}

	status := StatusCode(payload[0])
	if status != StatusOK {
		return status, "", nil
	}

	guid, _, err := readString(payload, 1)
	if err != nil {
		return 0, "", err
	}

	return status, guid, nil
}

func decodePickupMessageResponse(payload []byte) (StatusCode, string, string, string, error) {
	if len(payload) < 1 {
		return 0, "", "", "", io.ErrUnexpectedEOF
	}

	status := StatusCode(payload[0])
	if status != StatusOK {
		return status, "", "", "", nil
	}

	offset := 1
	guid, newOffset, err := readString(payload, offset)
	if err != nil {
		return 0, "", "", "", err
	}
	offset = newOffset

	content, newOffset, err := readString(payload, offset)
	if err != nil {
		return 0, "", "", "", err
	}
	offset = newOffset

	checksum, _, err := readString(payload, offset)
	if err != nil {
		return 0, "", "", "", err
	}

	return status, guid, content, checksum, nil
}

func decodeStatusResponse(payload []byte) (StatusCode, error) {
	if len(payload) < 1 {
		return 0, io.ErrUnexpectedEOF
	}
	return StatusCode(payload[0]), nil
}
