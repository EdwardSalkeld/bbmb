import socket
import struct
import hashlib
from typing import Optional
from dataclasses import dataclass


class BBMBError(Exception):
    pass


class QueueEmptyError(BBMBError):
    pass


class NotFoundError(BBMBError):
    pass


class InvalidChecksumError(BBMBError):
    pass


class MessageTooLargeError(BBMBError):
    pass


class ServerError(BBMBError):
    pass


class CommandType:
    ENSURE_QUEUE = 0x01
    ADD_MESSAGE = 0x02
    PICKUP_MESSAGE = 0x03
    DELETE_MESSAGE = 0x04


class StatusCode:
    OK = 0x00
    EMPTY_QUEUE = 0x01
    NOT_FOUND = 0x02
    INVALID_CHECKSUM = 0x03
    MESSAGE_TOO_LARGE = 0x04
    INTERNAL_ERROR = 0x05


MAX_MESSAGE_SIZE = 1024 * 1024  # 1MB


@dataclass
class Message:
    guid: str
    content: str
    checksum: str


class Client:
    def __init__(self, address: str = "localhost:9876"):
        parts = address.split(":")
        self.host = parts[0]
        self.port = int(parts[1]) if len(parts) > 1 else 9876
        self.sock: Optional[socket.socket] = None

    def _require_socket(self) -> socket.socket:
        if self.sock is None:
            raise BBMBError("Client is not connected")
        return self.sock

    def connect(self):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.connect((self.host, self.port))

    def close(self):
        if self.sock:
            self.sock.close()
            self.sock = None

    def __enter__(self):
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()

    def _write_frame(self, cmd_type: int, payload: bytes):
        sock = self._require_socket()
        length = len(payload) + 1
        frame = struct.pack(">I", length)
        frame += struct.pack("B", cmd_type)
        frame += payload
        sock.sendall(frame)

    def _read_frame(self) -> tuple[int, bytes]:
        length_data = self._recv_exactly(4)
        length = struct.unpack(">I", length_data)[0]

        cmd_type_data = self._recv_exactly(1)
        cmd_type = struct.unpack("B", cmd_type_data)[0]

        payload = self._recv_exactly(length - 1)
        return cmd_type, payload

    def _recv_exactly(self, n: int) -> bytes:
        sock = self._require_socket()
        data = b""
        while len(data) < n:
            chunk = sock.recv(n - len(data))
            if not chunk:
                raise BBMBError("Connection closed by server")
            data += chunk
        return data

    def _write_string(self, s: str) -> bytes:
        encoded = s.encode("utf-8")
        return struct.pack(">I", len(encoded)) + encoded

    def _read_string(self, data: bytes, offset: int) -> tuple[str, int]:
        if offset + 4 > len(data):
            raise BBMBError("Unexpected end of data")

        length = struct.unpack(">I", data[offset : offset + 4])[0]
        offset += 4

        if offset + length > len(data):
            raise BBMBError("Unexpected end of data")

        string = data[offset : offset + length].decode("utf-8")
        offset += length

        return string, offset

    def ensure_queue(self, queue_name: str):
        payload = self._write_string(queue_name)
        self._write_frame(CommandType.ENSURE_QUEUE, payload)

        _, resp_payload = self._read_frame()
        status = struct.unpack("B", resp_payload[:1])[0]

        if status != StatusCode.OK:
            raise BBMBError(f"Ensure queue failed with status: {status}")

    def add_message(self, queue_name: str, content: str) -> str:
        if len(content) > MAX_MESSAGE_SIZE:
            raise MessageTooLargeError("Message exceeds 1MB limit")

        checksum = hashlib.sha256(content.encode("utf-8")).hexdigest()

        payload = self._write_string(queue_name)
        payload += self._write_string(content)
        payload += self._write_string(checksum)

        self._write_frame(CommandType.ADD_MESSAGE, payload)

        _, resp_payload = self._read_frame()
        status = struct.unpack("B", resp_payload[:1])[0]

        if status == StatusCode.OK:
            guid, _ = self._read_string(resp_payload, 1)
            return guid
        elif status == StatusCode.INVALID_CHECKSUM:
            raise InvalidChecksumError("Server rejected checksum")
        elif status == StatusCode.MESSAGE_TOO_LARGE:
            raise MessageTooLargeError("Message too large")
        elif status == StatusCode.INTERNAL_ERROR:
            raise ServerError("Server internal error")
        else:
            raise BBMBError(f"Unexpected status: {status}")

    def pickup_message(
        self, queue_name: str, timeout_seconds: int = 30, wait_seconds: int = 0
    ) -> Message:
        payload = self._write_string(queue_name)
        payload += struct.pack(">I", timeout_seconds)
        if wait_seconds > 0:
            payload += struct.pack(">I", wait_seconds)

        self._write_frame(CommandType.PICKUP_MESSAGE, payload)

        _, resp_payload = self._read_frame()
        status = struct.unpack("B", resp_payload[:1])[0]

        if status == StatusCode.OK:
            offset = 1
            guid, offset = self._read_string(resp_payload, offset)
            content, offset = self._read_string(resp_payload, offset)
            checksum, _ = self._read_string(resp_payload, offset)
            return Message(guid=guid, content=content, checksum=checksum)
        elif status == StatusCode.EMPTY_QUEUE:
            raise QueueEmptyError("Queue is empty")
        elif status == StatusCode.INTERNAL_ERROR:
            raise ServerError("Server internal error")
        else:
            raise BBMBError(f"Unexpected status: {status}")

    def delete_message(self, queue_name: str, guid: str):
        payload = self._write_string(queue_name)
        payload += self._write_string(guid)

        self._write_frame(CommandType.DELETE_MESSAGE, payload)

        _, resp_payload = self._read_frame()
        status = struct.unpack("B", resp_payload[:1])[0]

        if status == StatusCode.OK:
            return
        elif status == StatusCode.NOT_FOUND:
            raise NotFoundError("Message not found")
        elif status == StatusCode.INTERNAL_ERROR:
            raise ServerError("Server internal error")
        else:
            raise BBMBError(f"Unexpected status: {status}")
