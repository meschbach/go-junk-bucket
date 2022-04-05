package dispatcher

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestDispatcherNothing(t *testing.T) {
	t.Parallel()

	d := NewDispatcher[string]()
	d.Broadcast("test")
	d.Close()
}

func TestNotifiesListeners(t *testing.T) {
	t.Parallel()

	gate := sync.WaitGroup{}
	gate.Add(1)
	received := false
	d := NewDispatcher[string]()
	d.Consume(func(m string) {
		received = true
		gate.Done()
	})
	d.Broadcast("test")
	d.Close()

	gate.Wait()

	assert.Equal(t, true, received)
}
