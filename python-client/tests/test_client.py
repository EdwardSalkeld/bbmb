import unittest
import struct

from bbmb_client.client import (
    BBMBError,
    Client,
    MAX_MESSAGE_SIZE,
    MessageTooLargeError,
)


class ClientTests(unittest.TestCase):
    def test_string_round_trip(self):
        client = Client()
        encoded = client._write_string("hello")
        value, offset = client._read_string(encoded, 0)
        self.assertEqual("hello", value)
        self.assertEqual(len(encoded), offset)

    def test_operations_require_connection(self):
        client = Client()
        with self.assertRaises(BBMBError):
            client._write_frame(1, b"payload")
        with self.assertRaises(BBMBError):
            client._recv_exactly(1)

    def test_add_message_enforces_size_limit(self):
        client = Client()
        with self.assertRaises(MessageTooLargeError):
            client.add_message("queue", "x" * (MAX_MESSAGE_SIZE + 1))

    def test_read_string_rejects_truncated_data(self):
        client = Client()
        with self.assertRaises(BBMBError):
            client._read_string(b"\x00\x00\x00\x05ab", 0)

    def test_pickup_payload_defaults_to_legacy_shape(self):
        client = Client()
        written = {}
        client._write_frame = lambda cmd, payload: written.update(
            {"cmd": cmd, "payload": payload}
        )
        client._read_frame = lambda: (0x03, b"\x01")

        with self.assertRaises(BBMBError):
            client.pickup_message("q", 30)

        expected = client._write_string("q") + struct.pack(">I", 30)
        self.assertEqual(expected, written["payload"])

    def test_pickup_payload_includes_wait_seconds(self):
        client = Client()
        written = {}
        client._write_frame = lambda cmd, payload: written.update(
            {"cmd": cmd, "payload": payload}
        )
        client._read_frame = lambda: (0x03, b"\x01")

        with self.assertRaises(BBMBError):
            client.pickup_message("q", 30, wait_seconds=5)

        expected = (
            client._write_string("q") + struct.pack(">I", 30) + struct.pack(">I", 5)
        )
        self.assertEqual(expected, written["payload"])


if __name__ == "__main__":
    unittest.main()
