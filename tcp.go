package dcsp

import (
	"io"
	"log"
	"net"
	"time"
)

func NewTCPTransport(addr string) Transport {
	return &tcpTransport{
		addr:           addr,
		loggingEnabled: false,
	}
}

type tcpTransport struct {
	loggingEnabled bool
	sendBuf        [maxMessageSize]byte
	recvBuf        [maxMessageSize]byte
	addr           string
	recvListener   net.Listener
	recvConn       net.Conn
	sendConn       net.Conn
}

// TODO: handle disconnected cases
// TODO: make retries and timers configurable
func (t *tcpTransport) BlockingSend(msg []byte) error {
	for {
		if t.sendConn == nil {
			t.log("[SENDER]: attempting to connect with receiver")
			conn, err := net.Dial("tcp", t.addr)
			if err != nil {
				t.log("[SENDER]:", err)
				t.log("[SENDER]: retrying")
				time.Sleep(150 * time.Millisecond)
				continue
			}
			t.log("[SENDER]: conntected")
			t.sendConn = conn
		}

		t.log("[SENDER]: sending message")

		_, err := writeFrame(msg, t.sendConn)
		if err != nil {
			// TODO: handle client disconnected but conn exists
			t.log("[SENDER]:", err)
			t.log("[SENDER]: retrying")
			time.Sleep(150 * time.Millisecond)
			continue
		}

		t.log("[SENDER]: awating acknowledge signal")
		n, err := readFrame(t.sendBuf[:], t.sendConn)
		if err != nil {
			t.log("[SENDER]:", err)
			t.log("[SENDER]: retrying")
			time.Sleep(150 * time.Millisecond)
			continue
		}

		response := string(t.sendBuf[0:n])
		if response != "OK" {
			t.log("[SENDER]: did not acknowledge:", response)
			t.log("[SENDER]: retrying")
			time.Sleep(150 * time.Millisecond)
			continue
		}
		t.log("[SENDER]: acknowledged")
		return nil
	}
}

func (t *tcpTransport) BlockingReceive() []byte {
	for {
		if t.recvListener == nil {
			t.log("[RECEIVER] attempting to connect with sender")
			l, err := net.Listen("tcp", t.addr)
			if err != nil {
				t.log("[RECEIVER]:", err)
				t.log("[RECEIVER] retrying")
				time.Sleep(150 * time.Millisecond)
				continue
			}
			t.log("[RECEIVER] connected")
			t.recvListener = l
		}
		break
	}

	for t.recvConn == nil {
		t.log("[RECEIVER] waiting for connection from sender")
		c, err := t.recvListener.Accept()
		if err != nil && err != io.EOF {
			t.log("[RECEIVER]:", err)
			t.log("[RECEIVER] retrying")
			time.Sleep(150 * time.Millisecond)
			continue
		}
		t.log("[RECEIVER] connection openned")
		t.recvConn = c
	}

	t.log("[RECEIVER] waiting for message")
	n, err := readFrame(t.recvBuf[:], t.recvConn)
	if err != nil {
		log.Fatal(err)
	}

	t.log("[RECEIVER] message received")
	_, err = writeFrame([]byte("OK"), t.recvConn)
	if err != nil {
		t.log(err)
	}

	result := make([]byte, n)
	copy(result[:], t.recvBuf[:n])

	return result
}

func (t *tcpTransport) log(v ...interface{}) {
	if t.loggingEnabled {
		log.Println(v...)
	}
}

func (t *tcpTransport) logf(format string, v ...interface{}) {
	if t.loggingEnabled {
		log.Printf(format, v...)
	}
}
