this project is a go message broker, and reference client implementations in go and python

the heart of it is a collection of fifo queues.
api calls will be limited to:

- ensure queue exists
- add to queue
- pick up from queue
- delete from queue

queues are simply string names
messages are arbitrary text data, the broker is agnostic about the content

pick up from queue will specify a timeout. the message will be marked as handed out
if the message is not deleted within the timeout it returns to being eligible to pick up
messages are given guid ids on add. these are then used when picking up and deleting

there is no authentication of clients

the api is to be tcp buffer based, not http.
add should be content, queue name and a checksum
pick up should just be queue name and timeout and return id, content and checksum
delete should take id and queue and return some success status

don't make anything about the server configurable yet. that can come later
pick a port and hardcode it

## observability

the server should expose prometheus style stats
there should be global stats and per queue
also memory, uptime, etc

## style

don't write comments that say what code does, only why it does it
don't use 3rd party libs for simple functionality

## tests

write tests with coverage
add github workflows to test prs
some throughput tests would be nice

## clients

the client implementations should be standalone
they should work as libs to be imported
they should also have a command line ability as well
this can be a separate build (cli wrapper for lib) if that's easiest
these should also have tests and GH workflows

## git

commit locally in small increments
