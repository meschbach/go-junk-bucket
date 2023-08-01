// Package dispatcher provides a multi-goproc safe event dispatching system
package dispatcher

type dispatcherOperation[T any] interface {
	perform(d *Dispatcher[T])
}

type Dispatcher[T any] struct {
	out []*listener[T]
	in  chan dispatcherOperation[T]
}

type listener[T any] struct {
	active bool
	queue  chan T
}

func NewDispatcher[T any]() *Dispatcher[T] {
	in := make(chan dispatcherOperation[T])
	d := &Dispatcher[T]{
		out: nil,
		in:  in,
	}
	go d.run()
	return d
}
