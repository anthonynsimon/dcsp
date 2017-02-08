package dcsp

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
