package dcsp

type Transport interface {
	EnsureSend()
	EnsureReceive()
	SyncSend([]byte) error
	SyncReceive() ([]byte, error)
	Ready() bool
}
