package dcsp

type Transport interface {
	SyncSend([]byte) error
	SyncReceive() ([]byte, error)
}
