package dcsp

import (
	"io"
	"log"
	"net"
	"time"
)

type SendChannel interface {
	Send([]byte) error
}

type ReceiveChannel interface {
	Receive() []byte
}

func NewSendChannel(trans Transport) SendChannel {
	return &sendChannel{
		transport: trans,
	}
}

func NewReceiveChannel(trans Transport) ReceiveChannel {
	return &receiveChannel{
		transport: trans,
	}
}

type sendChannel struct {
	addr      string
	transport Transport
}

type receiveChannel struct {
	addr      string
	transport Transport
}

func (c *sendChannel) Send(msg []byte) error {
	return c.transport.BlockingSend(msg)
}

func (c *receiveChannel) Receive() []byte {
	return c.transport.BlockingReceive()
}

type Transport interface {
	BlockingSend([]byte) error
	BlockingReceive() []byte
}

func NewTCPTransport(addr string) Transport {
	return &tcpTransport{
		addr: addr,
	}
}

type tcpTransport struct {
	addr         string
	recvListener net.Listener
	recvConn     net.Conn
	sendConn     net.Conn
}

func (t *tcpTransport) BlockingSend(msg []byte) error {
	for {
		if t.sendConn == nil {
			log.Println("[SENDER]: attempting to connect with receiver")
			conn, err := net.Dial("tcp", t.addr)
			if err != nil {
				log.Println("[SENDER]:", err)
				log.Println("[SENDER]: retrying")
				time.Sleep(150 * time.Millisecond)
				continue
			}
			log.Println("[SENDER]: conntected")
			t.sendConn = conn
		}

		log.Println("[SENDER]: sending message")
		_, err := t.sendConn.Write(msg)
		if err != nil {
			// TODO: handle client disconnected but conn exists
			log.Println("[SENDER]:", err)
			log.Println("[SENDER]: retrying")
			time.Sleep(150 * time.Millisecond)
			continue
		}

		log.Println("[SENDER]: awating acknowledge signal")
		var buf [512]byte
		n, err := t.sendConn.Read(buf[:])
		if err != nil {
			log.Println("[SENDER]:", err)
			log.Println("[SENDER]: retrying")
			time.Sleep(150 * time.Millisecond)
			continue
		}

		response := string(buf[0:n])
		if response != "OK" {
			log.Println("[SENDER]: did not acknowledge:", response)
			log.Println("[SENDER]: retrying")
			time.Sleep(150 * time.Millisecond)
			continue
		}
		log.Println("[SENDER]: acknowledged")
		return nil
	}
}

func (t *tcpTransport) BlockingReceive() []byte {
	for {
		if t.recvListener == nil {
			log.Println("[RECEIVER] attempting to connect with sender")
			l, err := net.Listen("tcp", t.addr)
			if err != nil {
				log.Println("[RECEIVER]:", err)
				log.Println("[RECEIVER] retrying")
				time.Sleep(150 * time.Millisecond)
				continue
			}
			log.Println("[RECEIVER] connected")
			t.recvListener = l
		}
		break
	}

	for t.recvConn == nil {
		log.Println("[RECEIVER] waiting for connection from sender")
		c, err := t.recvListener.Accept()
		if err != nil && err != io.EOF {
			log.Println("[RECEIVER]:", err)
			log.Println("[RECEIVER] retrying")
			time.Sleep(150 * time.Millisecond)
			continue
		}
		log.Println("[RECEIVER] connection openned")
		t.recvConn = c
	}

	log.Println("[RECEIVER] waiting for message")
	var buf [512]byte
	n, err := t.recvConn.Read(buf[:])
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[RECEIVER] message received")
	_, err = t.recvConn.Write([]byte("OK"))
	if err != nil {
		log.Println(err)
	}

	result := make([]byte, n)
	copy(result[:], buf[:n])

	return result
}
