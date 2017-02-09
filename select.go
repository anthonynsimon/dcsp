package dcsp

import "math/rand"

type selector struct {
	fn func()
	ch ReceiveChannel
}

func (s *selector) Func() func() {
	return s.fn
}

func (s *selector) Chan() ReceiveChannel {
	return s.ch
}

type Selector interface {
	Func() func()
	Chan() ReceiveChannel
}

// TODO: poll channels randomly for pulling/pushing
// TODO: fix this, needs to poll the chan before pulling/pushing
// maybe acknowledge on return, via defer
func Select(selectors ...Selector) {
	for {
		l := len(selectors)
		s := selectors[rand.Intn(l)]
		if s.Chan().Ready() {
			s.Func()()
			return
		}
	}
}

func NewSelector(ch ReceiveChannel, fn func()) Selector {
	return &selector{ch: ch, fn: fn}
}
