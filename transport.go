package dcsp

type Transport interface {
	BlockingSend([]byte) error
	BlockingReceive() []byte
}
