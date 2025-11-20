package tcp

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"

	"github.com/edsalkeld/bbmb/server/protocol"
	"github.com/edsalkeld/bbmb/server/queue"
)

type Server struct {
	address string
	manager *queue.Manager
}

func NewServer(address string, manager *queue.Manager) *Server {
	return &Server{
		address: address,
		manager: manager,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %w", err)
	}

	log.Printf("TCP server listening on %s", s.address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		cmdType, payload, err := protocol.ReadFrame(conn)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("error reading frame: %v", err)
			}
			return
		}

		var response []byte

		switch cmdType {
		case protocol.CmdEnsureQueue:
			response = s.handleEnsureQueue(payload)
		case protocol.CmdAddMessage:
			response = s.handleAddMessage(payload)
		case protocol.CmdPickupMessage:
			response = s.handlePickupMessage(payload)
		case protocol.CmdDeleteMessage:
			response = s.handleDeleteMessage(payload)
		default:
			log.Printf("unknown command type: %d", cmdType)
			return
		}

		if err := protocol.WriteFrame(conn, cmdType, response); err != nil {
			log.Printf("error writing response: %v", err)
			return
		}
	}
}

func (s *Server) handleEnsureQueue(payload []byte) []byte {
	req, err := protocol.DecodeEnsureQueue(payload)
	if err != nil {
		return protocol.EncodeEnsureQueueResponse(&protocol.EnsureQueueResponse{
			Status: protocol.StatusInternalError,
		})
	}

	s.manager.EnsureQueue(req.QueueName)

	return protocol.EncodeEnsureQueueResponse(&protocol.EnsureQueueResponse{
		Status: protocol.StatusOK,
	})
}

func (s *Server) handleAddMessage(payload []byte) []byte {
	req, err := protocol.DecodeAddMessage(payload)
	if err != nil {
		return protocol.EncodeAddMessageResponse(&protocol.AddMessageResponse{
			Status: protocol.StatusInternalError,
		})
	}

	if len(req.Content) > protocol.MaxMessageSize {
		return protocol.EncodeAddMessageResponse(&protocol.AddMessageResponse{
			Status: protocol.StatusMessageTooLarge,
		})
	}

	hash := sha256.Sum256([]byte(req.Content))
	expectedChecksum := hex.EncodeToString(hash[:])

	if expectedChecksum != req.Checksum {
		return protocol.EncodeAddMessageResponse(&protocol.AddMessageResponse{
			Status: protocol.StatusInvalidChecksum,
		})
	}

	q, exists := s.manager.GetQueue(req.QueueName)
	if !exists {
		s.manager.EnsureQueue(req.QueueName)
		q, _ = s.manager.GetQueue(req.QueueName)
	}

	msg, err := queue.NewMessage(req.Content, req.Checksum)
	if err != nil {
		return protocol.EncodeAddMessageResponse(&protocol.AddMessageResponse{
			Status: protocol.StatusInternalError,
		})
	}

	q.Add(msg)

	return protocol.EncodeAddMessageResponse(&protocol.AddMessageResponse{
		Status: protocol.StatusOK,
		GUID:   msg.ID,
	})
}

func (s *Server) handlePickupMessage(payload []byte) []byte {
	req, err := protocol.DecodePickupMessage(payload)
	if err != nil {
		return protocol.EncodePickupMessageResponse(&protocol.PickupMessageResponse{
			Status: protocol.StatusInternalError,
		})
	}

	q, exists := s.manager.GetQueue(req.QueueName)
	if !exists {
		return protocol.EncodePickupMessageResponse(&protocol.PickupMessageResponse{
			Status: protocol.StatusEmptyQueue,
		})
	}

	msg, err := q.Pickup(req.TimeoutSeconds)
	if err == queue.ErrQueueEmpty {
		return protocol.EncodePickupMessageResponse(&protocol.PickupMessageResponse{
			Status: protocol.StatusEmptyQueue,
		})
	}
	if err != nil {
		return protocol.EncodePickupMessageResponse(&protocol.PickupMessageResponse{
			Status: protocol.StatusInternalError,
		})
	}

	return protocol.EncodePickupMessageResponse(&protocol.PickupMessageResponse{
		Status:   protocol.StatusOK,
		GUID:     msg.ID,
		Content:  msg.Content,
		Checksum: msg.Checksum,
	})
}

func (s *Server) handleDeleteMessage(payload []byte) []byte {
	req, err := protocol.DecodeDeleteMessage(payload)
	if err != nil {
		return protocol.EncodeDeleteMessageResponse(&protocol.DeleteMessageResponse{
			Status: protocol.StatusInternalError,
		})
	}

	q, exists := s.manager.GetQueue(req.QueueName)
	if !exists {
		return protocol.EncodeDeleteMessageResponse(&protocol.DeleteMessageResponse{
			Status: protocol.StatusNotFound,
		})
	}

	err = q.Delete(req.GUID)
	if err == queue.ErrMessageNotFound {
		return protocol.EncodeDeleteMessageResponse(&protocol.DeleteMessageResponse{
			Status: protocol.StatusNotFound,
		})
	}
	if err != nil {
		return protocol.EncodeDeleteMessageResponse(&protocol.DeleteMessageResponse{
			Status: protocol.StatusInternalError,
		})
	}

	return protocol.EncodeDeleteMessageResponse(&protocol.DeleteMessageResponse{
		Status: protocol.StatusOK,
	})
}
