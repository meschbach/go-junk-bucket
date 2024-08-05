// Package buffernet provides facilities to establish a connection via grpc via internal buffers, allowing for testing
// services and clients without worrying about real networks.
package buffernet

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
)

type BufferTransport struct {
	Listener *bufconn.Listener
}

func New() *BufferTransport {
	//buffer size defaults to 1 megabyte.  this should be sufficient for decently fast tests.
	const bufSize = 1024 * 1024

	var listener *bufconn.Listener
	listener = bufconn.Listen(bufSize)

	return &BufferTransport{
		Listener: listener,
	}
}

// Connect establishes a connection to the listening service via the buffer.  Will block until the connection is
// established.
func (b *BufferTransport) Connect(ctx context.Context, opts ...grpc.DialOption) (conn *grpc.ClientConn, problem error) {
	dialer := func(ctx context.Context, _ string) (net.Conn, error) {
		return b.Listener.DialContext(ctx)
	}
	opts = append(opts, grpc.WithContextDialer(dialer))
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	return grpc.DialContext(ctx, "bufnet", opts...)
}

func (b *BufferTransport) Listen(server *grpc.Server) error {
	return server.Serve(b.Listener)
}

func (b *BufferTransport) ListenAsync(server *grpc.Server) (<-chan error, func()) {
	result := make(chan error, 1)
	serviceContext, done := context.WithCancel(context.Background())
	go func() {
		err := b.Listen(server)
		result <- err
		close(result)
	}()
	go func() {
		<-serviceContext.Done()
		server.GracefulStop()
	}()

	return result, done
}
