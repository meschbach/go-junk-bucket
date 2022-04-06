package dispatcher

import "fmt"

func (d *Dispatcher[T]) Broadcast(message T) {
	d.in <- &broadcastOperation[T]{value: message}
}

type broadcastOperation[T any] struct {
	value T
}

func (b *broadcastOperation[T]) perform(d *Dispatcher[T]) {
	fmt.Printf("Dispatching %v\n", b.value)
	for _, t := range d.out {
		if t.active {
			t.queue <- b.value
		}
	}
}
