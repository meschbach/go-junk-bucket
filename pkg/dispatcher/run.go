package dispatcher

func (d *Dispatcher[T]) run() {
	for op := range d.in {
		op.perform(d)
	}
}
