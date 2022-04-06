package supervision

import (
	"github.com/meschbach/go-junk-bucket/pkg/actors/local"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSupervisionTree(t *testing.T) {
	//t.Parallel()

	sys := local.NewLocalActorSystem()
	defer sys.Shutdown()

	mailbox := sys.ExternalMailbox()
	defer mailbox.Close()
	sys.Spawn("supervisor", &supervisor{output: mailbox})
	value, err := mailbox.ReceiveTimeout(10 * time.Second)
	if assert.NoError(t, err) {
		assert.Equal(t, 2, value)
	}
}
