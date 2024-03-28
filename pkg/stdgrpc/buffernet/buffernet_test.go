package buffernet

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"testing"
)

func TestBufferNet(t *testing.T) {
	t.Run("Able to spawn a new service and client", func(t *testing.T) {
		ctx, done := context.WithCancel(context.Background())
		t.Cleanup(done)

		healthService := health.NewServer()
		healthService.SetServingStatus("test", healthgrpc.HealthCheckResponse_SERVING)

		net := New()
		server := grpc.NewServer()
		healthgrpc.RegisterHealthServer(server, healthService)
		_, shutdownServer := net.ListenAsync(server)
		t.Cleanup(shutdownServer)

		client, err := net.Connect(ctx)
		require.NoError(t, err)
		healthClient := healthgrpc.NewHealthClient(client)
		out, err := healthClient.Check(ctx, &healthgrpc.HealthCheckRequest{Service: "test"})
		require.NoError(t, err)
		assert.Equal(t, healthgrpc.HealthCheckResponse_SERVING, out.Status)
	})
}
