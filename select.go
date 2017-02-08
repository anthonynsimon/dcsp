package dcsp

import "sync"

type selector struct {
	fn func([]byte, error)
	ch ReceiveChannel
}

func (s *selector) Func() func([]byte, error) {
	return s.fn
}

func (s *selector) Chan() ReceiveChannel {
	return s.ch
}

type Selector interface {
	Func() func([]byte, error)
	Chan() ReceiveChannel
}

// TODO: poll channels randomly for pulling/pushing
// TODO: fix this, needs to poll the chan before pulling/pushing
// maybe acknowledge on return, via defer
func Select(selectors ...Selector) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	for _, s := range selectors {
		go func() {
			msg, err := s.Chan().Receive()
			wg.Done()
			s.Func()(msg, err)
		}()
	}
	wg.Wait()
}

func NewSelector(ch ReceiveChannel, fn func([]byte, error)) Selector {
	return &selector{ch: ch, fn: fn}
}
