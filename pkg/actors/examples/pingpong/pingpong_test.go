package pingpong

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/actors/local"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPingPongActors(t *testing.T) {
	t.Parallel()
	ctx, done := context.WithCancel(context.Background())
	t.Cleanup(done)

	sys := local.NewSystem()
	ping := sys.Spawn(ctx, &appender{suffix: "ping"})
	pong := sys.Spawn(ctx, &appender{suffix: "pong"})
	out := sys.NewPort()
	director := sys.Spawn(ctx, &stringDirector{
		apply:  pong,
		inform: out.Pid(),
	})

	sys.Tell(ctx, ping, &appendString{
		to:   "a",
		next: director,
	})
	v, err := out.ReceiveWith(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, "apingpong", v)
	}
}
