package actors

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPanickedActor(t *testing.T) {
	sys := NewLocalActorSystem()
	defer sys.Shutdown()

	ref := sys.SpawnFn("panic", func(msg any) {
		panic("example")
	})
	mailbox := sys.ExternalMailbox()
	unlink, err := sys.Monitor(ref, mailbox)
	defer unlink()
	if assert.NoError(t, err) {
		ref.Tell("pleases")
		_, err := mailbox.ReceiveTimeout(100 * time.Millisecond)
		if assert.NoError(t, err) {
		}
	}
}
