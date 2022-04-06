package dispatcher

func (d *Dispatcher[T]) Close() {
	close(d.in)
}

type closeOperation[T any] struct {
	target *listener[T]
}

func (b *closeOperation[T]) perform(d *Dispatcher[T]) {
	if b.target.active {
		b.target.active = false
		close(b.target.queue)
	}
	//TODO: Compaction heuristics
	//newListeners := make([]*listener[T])
	//for _, o := range d.out {
	//	if o != b.target {
	//		newListeners = append(newListeners, o)
	//	}
	//}
	//d.out = newListeners
}
