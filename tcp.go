package dcsp

import (
	"io"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	retryTime = 1000 * time.Millisecond
)

func NewTCPTransport(addr string) Transport {
	t := tcpTransport{
		logFields: log.WithFields(log.Fields{
			"transport": "tcp",
		}),
	}

	t.sendHandler.addr = addr
	t.receiveHandler.addr = addr

	t.sendHandler.logFields = t.logFields.WithFields(log.Fields{
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
	sendHandler    tcpHandler
	receiveHandler tcpHandler

	logFields log.FieldLogger
}

func (t *tcpTransport) SyncSend(b []byte) error {
	return t.sendHandler.send(b)
}

func (t *tcpTransport) SyncReceive() ([]byte, error) {
	return t.receiveHandler.receive()
}

type tcpHandler struct {
	addr string

	buf [maxMessageSize]byte

	listener net.Listener

	conn net.Conn

	logFields log.FieldLogger
}

// TODO: add shutdown/close signal
func (t *tcpHandler) send(msg []byte) error {
	for {
		for t.conn == nil {
			t.logFields.Info("attempting to connect with receiver")
			conn, err := net.Dial("tcp", t.addr)
			if err != nil {
				t.logFields.Error(err, "retrying in ", retryTime.String())
				time.Sleep(retryTime)
				continue
			}
			t.logFields.Info("conntected")
			t.conn = conn
		}

		t.logFields.Debug("sending message")

		_, err := writeFrame(msg, t.conn)
		if err != nil {
			t.conn.Close()
			t.conn = nil
			// TODO: handle client disconnected but conn exists
			t.logFields.Error(err, "retrying in ", retryTime.String())
			time.Sleep(retryTime)
			continue
		}

		t.logFields.Debug("wating for acknowledge signal")
		n, err := readFrame(t.buf[:], t.conn)
		if err != nil {
			t.conn.Close()
			t.conn = nil
			t.logFields.Error(err, "retrying in ", retryTime.String())
			time.Sleep(retryTime)
			continue
		}

		response := string(t.buf[0:n])
		if response != "OK" {
			t.conn.Close()
			t.conn = nil
			t.logFields.Error("did not acknowledge. responded:", response, "retrying in ", retryTime.String())
			time.Sleep(retryTime)
			continue
		}
		t.logFields.Info("message sent")
		return nil
	}
}

func (t *tcpHandler) receive() ([]byte, error) {
	for {
		for t.listener == nil {
			t.logFields.Info("attempting to connect with sender")
			l, err := net.Listen("tcp", t.addr)
			if err != nil {
				t.logFields.Error(err, "retrying in ", retryTime.String())
				time.Sleep(retryTime)
				continue
			}
			t.logFields.Info("connected")
			t.listener = l
		}

		for t.conn == nil {
			t.logFields.Debug("waiting for connection from sender")
			c, err := t.listener.Accept()
			if err != nil && err != io.EOF {
				t.listener.Close()
				t.listener = nil
				t.logFields.Error(err, "retrying in ", retryTime.String())
				time.Sleep(retryTime)
				continue
			}
			t.logFields.Debug("connection openned")
			t.conn = c
		}

		t.logFields.Debug("waiting for message")
		n, err := readFrame(t.buf[:], t.conn)
		if err != nil {
			t.listener.Close()
			t.listener = nil
			t.conn.Close()
			t.conn = nil
			t.logFields.Error(err)
			continue
		}

		t.logFields.Info("message received")
		_, err = writeFrame([]byte("OK"), t.conn)
		if err != nil {
			t.listener.Close()
			t.listener = nil
			t.conn.Close()
			t.conn = nil
			t.logFields.Error(err)
			continue
		}

		result := make([]byte, n)
		copy(result[:], t.buf[:n])

		return result, nil
	}
}
