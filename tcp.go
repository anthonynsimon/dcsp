package dcsp

import (
	"io"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	retryTime = 300 * time.Millisecond
)

func NewTCPTransport(addr string) Transport {
	t := tcpTransport{
		logFields: log.WithFields(log.Fields{
			"transport": "tcp",
		}),
	}

	t.sendingHandler.addr = addr
	t.receiveHandler.addr = addr

	t.sendingHandler.logFields = t.logFields.WithFields(log.Fields{
		"channelType": "sender",
		"sendAddr":    addr,
	})

	t.receiveHandler.logFields = log.WithFields(log.Fields{
		"channelType": "receive",
		"recieveAddr": addr,
	})

	return &t
}

type tcpTransport struct {
	sendingHandler tcpHandler
	receiveHandler tcpHandler

	logFields log.FieldLogger
}

func (t *tcpTransport) BlockingSend(b []byte) error {
	return t.sendingHandler.send(b)
}

func (t *tcpTransport) BlockingReceive() []byte {
	return t.receiveHandler.receive()
}

type tcpHandler struct {
	addr string

	buf [maxMessageSize]byte

	listener net.Listener

	conn net.Conn

	logFields log.FieldLogger
}

func (t *tcpHandler) send(msg []byte) error {
	for {
		if t.conn == nil {
			t.logFields.Info("attempting to connect with receiver")
			conn, err := net.Dial("tcp", t.addr)
			if err != nil {
				t.logFields.Debug(err)
				t.logFields.Debug("retrying in ", retryTime.String())
				time.Sleep(retryTime)
				continue
			}
			t.logFields.Info("conntected")
			t.conn = conn
		}

		t.logFields.Debug("sending message")

		_, err := writeFrame(msg, t.conn)
		if err != nil {
			// TODO: handle client disconnected but conn exists
			t.logFields.Debug(err, "retrying in ", retryTime.String())
			time.Sleep(retryTime)
			continue
		}

		t.logFields.Debug("awating acknowledge signal")
		n, err := readFrame(t.buf[:], t.conn)
		if err != nil {
			t.logFields.Debug(err, "retrying in ", retryTime.String())
			time.Sleep(retryTime)
			continue
		}

		response := string(t.buf[0:n])
		if response != "OK" {
			t.logFields.Debug("did not acknowledge:", response)
			t.logFields.Debug("retrying in ", retryTime.String())
			time.Sleep(retryTime)
			continue
		}
		t.logFields.Info("message sent")
		return nil
	}
}

func (t *tcpHandler) receive() []byte {
	for {
		if t.listener == nil {
			t.logFields.Info("attempting to connect with sender")
			l, err := net.Listen("tcp", t.addr)
			if err != nil {
				t.logFields.Debug(err, "retrying in ", retryTime.String())
				time.Sleep(retryTime)
				continue
			}
			t.logFields.Info("connected")
			t.listener = l
		}
		break
	}

	for t.conn == nil {
		t.logFields.Debug("waiting for connection from sender")
		c, err := t.listener.Accept()
		if err != nil && err != io.EOF {
			t.logFields.Debug(err, "retrying in ", retryTime.String())
			time.Sleep(retryTime)
			continue
		}
		t.logFields.Debug("connection openned")
		t.conn = c
	}

	t.logFields.Debug("waiting for message")
	n, err := readFrame(t.buf[:], t.conn)
	if err != nil {
		log.Fatal(err)
	}

	t.logFields.Info("message received")
	_, err = writeFrame([]byte("OK"), t.conn)
	if err != nil {
		t.logFields.Debug(err)
		// TODO: return error!
		return []byte{}
	}

	result := make([]byte, n)
	copy(result[:], t.buf[:n])

	return result
}
