package streams

type ChannelSink[T any] struct {
	target chan<- T
}

func NewChannelSink[T any](target chan<- T) *ChannelSink[T] {
	return &ChannelSink[T]{
		target: target,
	}
}
