package dcsp

import (
	"encoding/binary"
	"io"
)

func writeFrame(b []byte, w io.Writer) (int, error) {
	frameLen := len(b)

	buf := make([]byte, framePrefixSize+frameLen)

	binary.BigEndian.PutUint32(buf[0:framePrefixSize], uint32(frameLen))

	copy(buf[framePrefixSize:], b[:])

	n, err := w.Write(buf)
	if err != nil {
		return -1, err
	}

	return n + framePrefixSize, nil
}

func readFrame(dst []byte, r io.Reader) (int, error) {
	var prefix [framePrefixSize]byte
	_, err := io.ReadFull(r, prefix[:])
	if err != nil {
		if err != io.EOF {
			return -1, err
		}
	}

	frameLen := int(binary.BigEndian.Uint32(prefix[:]))
	n, err := io.LimitReader(r, maxMessageSize).Read(dst[0:frameLen])
	if err != nil {
		if err != io.EOF {
			return -1, err
		}
	}

	return n, nil
}
