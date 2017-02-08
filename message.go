package dcsp

import (
	"encoding/binary"
	"errors"
	"time"

	"github.com/satori/go.uuid"
)

const (
	framePrefixSize = 4
	maxPayloadSize  = 1000000 // num of bytes, max 1 MB
	minMessageSize  = 8 + 16  // timestamp + id
	maxMessageSize  = minMessageSize + maxPayloadSize
)

// message will be encoded as a byte array to be transmitted over the wire.
// Once encoded, it is length-prefixed to facilitate buffered reading.
// The first field, the frame size, is an int32 denoting the size of the frame without
// counting itself (size includes timestamp, id and payload).
//
// Message as frame:
//
// |      4-bytes     |       8-bytes      |      16-bytes    |      n-bytes    |
// ------------------------------------------------------------------------------
// |    frame size    |       timestamp    |        UUID      |      payload    |
// |      int32       |        int64       |     hex string   |       binary    |
//
type message struct {
	Timestamp int64    // [8]byte in unix nanoseconds
	ID        [16]byte // [16]byte hex string
	Payload   []byte
}

func newMessage(payload []byte) *message {
	return &message{
		ID:        uuid.NewV4(),
		Timestamp: time.Now().UnixNano(),
		Payload:   payload,
	}
}

func decodeMessage(data []byte) (*message, error) {
	dataLen := len(data)
	if dataLen < minMessageSize {
		return nil, errors.New("message length is too small")
	}
	if dataLen-minMessageSize > maxPayloadSize {
		return nil, errors.New("message payload length is too large")
	}

	msg := message{
		Payload: make([]byte, dataLen-minMessageSize),
	}
	msg.Timestamp = int64(binary.BigEndian.Uint64(data[0:8]))
	copy(msg.ID[:], data[8:24])
	copy(msg.Payload[:], data[24:])

	return &msg, nil
}

func encodeMessage(msg *message) ([]byte, error) {
	dataLen := len(msg.Payload)
	if dataLen > maxPayloadSize {
		return nil, errors.New("message payload length is too large")
	}

	buf := make([]byte, minMessageSize+dataLen)

	binary.BigEndian.PutUint64(buf[:8], uint64(msg.Timestamp))

	copy(buf[8:24], msg.ID[:])

	copy(buf[24:], msg.Payload)

	return buf, nil
}
