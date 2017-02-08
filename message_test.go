package dcsp

import (
	"testing"

	"github.com/satori/go.uuid"
)

func TestMessageEncodeDecode(t *testing.T) {
	cases := []*message{
		newMessage([]byte("hello\r\n")),
		newMessage([]byte("hello")),
		newMessage([]byte("testing something in here")),
		newMessage([]byte{10, 20, 40, 50, 60, 70, 80, 90, 100}),
	}

	for _, msg := range cases {
		encoded, err := encodeMessage(msg)
		if err != nil {
			t.Error(err)
		}
		if len(encoded) != len(msg.Payload)+minMessageSize {
			t.Errorf("\nexpected encoded message length: %v\nactual: %v", len(msg.Payload)+minMessageSize, len(encoded))
		}
		decoded, err := decodeMessage(encoded)
		if err != nil {
			t.Error(err)
		}
		if decoded == nil {
			t.Fatal("decoded is nil")
		}
		if decoded.ID != msg.ID {
			t.Errorf("\nexpected id: %v\nactual: %v", msg.ID, decoded.ID)
		}
		if decoded.Timestamp != msg.Timestamp {
			t.Errorf("\nexpected timestamp: %v\nactual: %v", msg.Timestamp, decoded.Timestamp)
		}
		if len(decoded.Payload) != len(msg.Payload) {
			t.Errorf("\nexpected payload length: %v\nactual: %v", len(msg.Payload), len(decoded.Payload))
		}
		for i := range decoded.Payload {
			if decoded.Payload[i] != msg.Payload[i] {
				t.Errorf("\nexpected payload: %v\nactual: %v", msg.Payload, decoded.Payload)
				break
			}
		}

		_, err = uuid.FromBytes(decoded.ID[:])
		if err != nil {
			t.Error(err)
		}
	}
}
