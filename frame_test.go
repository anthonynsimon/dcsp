package dcsp

import (
	"bytes"
	"testing"
)

func TestReadWriteFrame(t *testing.T) {
	inputstr := "hello\n"
	frame := bytes.NewBuffer(nil)
	n, err := writeFrame([]byte(inputstr), frame)
	if err != nil {
		t.Error(err)
	}
	frameBytes := frame.Bytes()
	expectedBytes := []byte{0, 0, 0, 6, 104, 101, 108, 108, 111, 10}

	if n != len(expectedBytes)+framePrefixSize {
		t.Errorf("unexpected written frame length. expected %d, got %d", len(expectedBytes)+framePrefixSize, n)
	}

	if len(frameBytes) != len(expectedBytes) {
		t.Error("unexpected frame bytes length")
	}

	for i := range frameBytes {
		if frameBytes[i] != expectedBytes[i] {
			t.Errorf("unexpected frame byte at index %d. expected %d, got %d", i, expectedBytes[i], frameBytes[i])
		}
	}

	var readBuf [512]byte
	n, err = readFrame(readBuf[:], frame)
	if err != nil {
		t.Error(err)
	}

	if n != len(inputstr) {
		t.Error("unexpected frame read n")
	}

	result := string(readBuf[:n])

	if result != inputstr {
		t.Errorf("expected %s, got %s", inputstr, result)
	}
}
