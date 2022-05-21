package local

import "context"

type terminateSignal struct {
}

func (t *terminateSignal) execute(ctx context.Context, r *runtime) {
	r.done()
}

func (t *terminateSignal) name() string {
	return "terminateSignal"
}
