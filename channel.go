package dcsp

// TODO: add message frames (uuid, timestamp, context?)

type Middleware func([]byte) []byte

type SendChannel interface {
	Send([]byte) error
	Ready() bool
}

type ReceiveChannel interface {
	Receive() ([]byte, error)
	Ready() bool
}

func NewSendChannel(trans Transport, middleware ...Middleware) SendChannel {
	go trans.EnsureSend()
	return &sendChannel{
		transport:  trans,
		middleware: middleware,
	}
}

func NewReceiveChannel(trans Transport, middleware ...Middleware) ReceiveChannel {
	go trans.EnsureReceive()
	return &receiveChannel{
		transport:  trans,
		middleware: middleware,
	}
}

// TODO: combine sendChannel and receiveChannel into one struct?
type sendChannel struct {
	transport  Transport
	middleware []Middleware
}

type receiveChannel struct {
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

func (c *sendChannel) Ready() bool {
	return c.transport.Ready()
}

func (c *receiveChannel) Ready() bool {
	return c.transport.Ready()
}
