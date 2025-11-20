package protocol

type CommandType byte

const (
	CmdEnsureQueue  CommandType = 0x01
	CmdAddMessage   CommandType = 0x02
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

const MaxMessageSize = 1024 * 1024 // 1MB

type EnsureQueueRequest struct {
	QueueName string
}

type EnsureQueueResponse struct {
	Status StatusCode
}

type AddMessageRequest struct {
	QueueName string
	Content   string
	Checksum  string
}

type AddMessageResponse struct {
	Status StatusCode
	GUID   string
}

type PickupMessageRequest struct {
	QueueName      string
	TimeoutSeconds int
}

type PickupMessageResponse struct {
	Status   StatusCode
	GUID     string
	Content  string
	Checksum string
}

type DeleteMessageRequest struct {
	QueueName string
	GUID      string
}

type DeleteMessageResponse struct {
	Status StatusCode
}
