#!/usr/bin/env python3

import sys
import argparse
from bbmb_client import Client, QueueEmptyError, NotFoundError, BBMBError


def main():
    parser = argparse.ArgumentParser(
        description="BBMB Client CLI",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Commands:
  ensure-queue  Ensure a queue exists
  add           Add a message to a queue
  pickup        Pick up a message from a queue
  consume       Pick up and delete a message (pickup + delete)
  delete        Delete a message from a queue

Examples:
  %(prog)s ensure-queue --queue myqueue
  %(prog)s add --queue myqueue --content "hello world"
  %(prog)s pickup --queue myqueue --timeout 30
  %(prog)s pickup --queue myqueue --timeout 30 --wait 5
  %(prog)s consume --queue myqueue --timeout 30
  %(prog)s delete --queue myqueue --guid <message-id>
        """,
    )

    parser.add_argument(
        "command",
        choices=["ensure-queue", "add", "pickup", "consume", "delete"],
        help="Command to execute",
    )
    parser.add_argument(
        "--server",
        default="localhost:9876",
        help="Server address (default: localhost:9876)",
    )
    parser.add_argument("--queue", help="Queue name (required)")
    parser.add_argument("--content", help="Message content (for add command)")
    parser.add_argument("--guid", help="Message GUID (for delete command)")
    parser.add_argument(
        "--timeout",
        type=int,
        default=30,
        help="Timeout in seconds (for pickup command, default: 30)",
    )
    parser.add_argument(
        "--wait",
        type=int,
        default=0,
        help="Long-poll wait in seconds (for pickup/consume, default: 0)",
    )

    args = parser.parse_args()

    if not args.queue:
        print("Error: --queue is required", file=sys.stderr)
        sys.exit(1)

    try:
        with Client(args.server) as client:
            if args.command == "ensure-queue":
                client.ensure_queue(args.queue)
                print(f"Queue '{args.queue}' ensured")

            elif args.command == "add":
                if not args.content:
                    print(
                        "Error: --content is required for add command", file=sys.stderr
                    )
                    sys.exit(1)

                guid = client.add_message(args.queue, args.content)
                print(f"Message added with GUID: {guid}")

            elif args.command == "pickup":
                try:
                    msg = client.pickup_message(args.queue, args.timeout, args.wait)
                    print(f"GUID: {msg.guid}")
                    print(f"Content: {msg.content}")
                    print(f"Checksum: {msg.checksum}")
                except QueueEmptyError:
                    print("Queue is empty")

            elif args.command == "consume":
                try:
                    msg = client.pickup_message(args.queue, args.timeout, args.wait)
                    print(f"GUID: {msg.guid}")
                    print(f"Content: {msg.content}")
                    print(f"Checksum: {msg.checksum}")
                    client.delete_message(args.queue, msg.guid)
                    print("Message consumed (deleted)")
                except QueueEmptyError:
                    print("Queue is empty")

            elif args.command == "delete":
                if not args.guid:
                    print(
                        "Error: --guid is required for delete command", file=sys.stderr
                    )
                    sys.exit(1)

                client.delete_message(args.queue, args.guid)
                print(f"Message {args.guid} deleted from queue '{args.queue}'")

    except NotFoundError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    except BBMBError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    except KeyboardInterrupt:
        print("\nInterrupted", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Unexpected error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
