package dispatcher

type dispatcherOperation[T any] interface {
	perform(d *Dispatcher[T])
}

type Dispatcher[T any] struct {
	out []chan T
	in  chan dispatcherOperation[T]
}

func NewDispatcher[T any]() *Dispatcher[T] {
	in := make(chan dispatcherOperation[T])
	d := &Dispatcher[T]{
		out: nil,
		in:  in,
	}
	go func() {
		for op := range in {
			op.perform(d)
		}
	}()
	return d
}

func (d *Dispatcher[T]) Close() {
	close(d.in)
}

func (d *Dispatcher[T]) Broadcast(message T) {
	d.in <- &broadcastOperation[T]{value: message}
}

type broadcastOperation[T any] struct {
	value T
}

func (b *broadcastOperation[T]) perform(d *Dispatcher[T]) {
	for _, t := range d.out {
		t <- b.value
	}
}

func (d *Dispatcher[T]) Listen() (<-chan T, func()) {
	out := make(chan T)
	d.in <- &listenOperation[T]{target: out}
	return out, func() {
		d.in <- &closeOperation[T]{
			target: out,
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

type listenOperation[T any] struct {
	target chan T
}

func (l *listenOperation[T]) perform(d *Dispatcher[T]) {
	d.out = append(d.out, l.target)
}

type closeOperation[T any] struct {
	target chan T
}

func (b *closeOperation[T]) perform(d *Dispatcher[T]) {
	close(b.target)
	newListeners := make([]chan T, 0, len(d.out)-1)
	for _, o := range d.out {
		if o != b.target {
			newListeners = append(newListeners, o)
		}
	}
	d.out = newListeners
}
