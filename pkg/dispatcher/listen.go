package dispatcher

type listenOperation[T any] struct {
	target *listener[T]
}

func (l *listenOperation[T]) perform(d *Dispatcher[T]) {
	d.out = append(d.out, l.target)
}

func (d *Dispatcher[T]) Listen() (<-chan T, func()) {
	out := make(chan T)
	l := &listener[T]{
		active: true,
		queue:  out,
	}
	d.in <- &listenOperation[T]{target: l}
	return out, func() {
		d.in <- &closeOperation[T]{
			target: l,
		}
	}
}

func (d *Dispatcher[T]) Consume(fn func(msg T)) {
	c, done := d.Listen()
	go func() {
		defer done()
		for m := range c {
			fn(m)
		}
	}()
}
