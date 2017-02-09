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
func (t *tcpTransport) EnsureSend() {
	t.sendHandler.EnsureSendConn()
}

func (t *tcpTransport) EnsureReceive() {
	t.sendHandler.EnsureReceiveConn()
}

type tcpHandler struct {
	addr string

	buf [maxMessageSize]byte

	listener net.Listener

	conn net.Conn

	logFields log.FieldLogger

	ready bool
}

func (t *tcpHandler) EnsureSendConn() {
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

		// Send ready signal
		t.logFields.Debug("sending ready message")
		_, err := writeFrame([]byte("RDY?"), t.conn)
		if err != nil {
			t.logFields.Error(err, "retrying in ", retryTime.String())
			t.conn.Close()
			t.conn = nil
			time.Sleep(retryTime)
			continue
		}

		// Block until receiver is ready to be sent a message
		t.logFields.Debug("wating for ready signal")
		n, err := readFrame(t.buf[:], t.conn)
		if err != nil {
			t.logFields.Error("error while waiting for ready signal", err)
			t.conn.Close()
			t.conn = nil
			continue
		}

		response := string(t.buf[0:n])
		if response != "RDY" {
			t.logFields.Error("expected ready message, got:", response, err)
			t.conn.Close()
			t.conn = nil
			continue
		}

		t.ready = true
		return
	}
}

func (t *tcpHandler) EnsureReceiveConn() {
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
				t.logFields.Error(err, "retrying in ", retryTime.String())
				time.Sleep(retryTime)
				continue
			}
			t.logFields.Debug("connection openned")
			t.conn = c
		}

		// Block until sender is ready to send a message
		t.logFields.Debug("wating for ready signal")
		n, err := readFrame(t.buf[:], t.conn)
		if err != nil {
			t.logFields.Error("error while waiting for ready signal", err)
			continue
		}

		response := string(t.buf[0:n])
		if response != "RDY?" {
			t.logFields.Error("expected ready message, got:", response, err)
			continue
		}

		// Send ready signal
		t.logFields.Debug("sending ready message")
		_, err = writeFrame([]byte("RDY"), t.conn)
		if err != nil {
			t.logFields.Error(err, "retrying in ", retryTime.String())
			time.Sleep(retryTime)
			continue
		}
		t.ready = true
		return
	}
}

func (t *tcpTransport) Ready() bool {
	return t.sendHandler.ready || t.receiveHandler.ready
}

// TODO: add shutdown/close signal
func (t *tcpHandler) send(msg []byte) error {
	for {
		t.ready = false
		// Block until conn is available
		t.EnsureSendConn()
		defer func() {
			t.ready = false
		}()

		t.logFields.Debug("sending message")
		_, err := writeFrame(msg, t.conn)
		if err != nil {
			// TODO: handle client disconnected but conn exists
			t.logFields.Error(err, "retrying in ", retryTime.String())
			time.Sleep(retryTime)
			continue
		}

		t.logFields.Debug("wating for acknowledge signal")
		n, err := readFrame(t.buf[:], t.conn)
		if err != nil {
			t.logFields.Error(err, "retrying in ", retryTime.String())
			time.Sleep(retryTime)
			continue
		}

		response := string(t.buf[0:n])
		if response != "ACK" {
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
		t.ready = false
		// Block until conn is available
		t.EnsureReceiveConn()
		defer func() {
			t.ready = false
		}()

		t.logFields.Debug("waiting for message")
		n, err := readFrame(t.buf[:], t.conn)
		if err != nil {
			t.logFields.Error(err)
			continue
		}

		t.logFields.Info("message received. sending acknowledge signal")
		_, err = writeFrame([]byte("ACK"), t.conn)
		if err != nil {
			t.logFields.Error(err)
			continue
		}

		result := make([]byte, n)
		copy(result[:], t.buf[:n])

		return result, nil
	}
}
