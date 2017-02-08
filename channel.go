package dcsp

// TODO: add message frames (uuid, timestamp, context?)

type Middleware func([]byte) []byte

type SendChannel interface {
	Send([]byte) error
}

type ReceiveChannel interface {
	Receive() ([]byte, error)
}

func NewSendChannel(trans Transport, middleware ...Middleware) SendChannel {
	return &sendChannel{
		transport:  trans,
		middleware: middleware,
	}
}

func NewReceiveChannel(trans Transport, middleware ...Middleware) ReceiveChannel {
	return &receiveChannel{
		transport:  trans,
		middleware: middleware,
	}
}

// TODO: combine sendChannel and receiveChannel into one struct?
type sendChannel struct {
	addr       string
	transport  Transport
	middleware []Middleware
}

type receiveChannel struct {
	addr       string
	transport  Transport
	middleware []Middleware
}

func (c *sendChannel) Send(msg []byte) error {
	for _, mid := range c.middleware {
		msg = mid(msg)
	}
	return c.transport.SyncSend(msg)
}

func (c *receiveChannel) Receive() ([]byte, error) {
	msg, err := c.transport.SyncReceive()
	if err != nil {
		return nil, err
	}
	for _, mid := range c.middleware {
		msg = mid(msg)
	}
	return msg, nil
}
