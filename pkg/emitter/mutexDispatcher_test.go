package emitter

import "testing"

func TestMutexDispatcherIsEmitter(t *testing.T) {
	applyTestEventEmitter(t, func() Emitter[int] {
		return NewMutexDispatcher[int]()
	})
}
