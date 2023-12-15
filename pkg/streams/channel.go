package streams

type channelPortFeedback int
type ChannelPort[T any] struct {
	Input    *ChannelSink[T]
	Feedback chan channelPortFeedback
	Output   *ChannelSource[T]
}

func NewChannelPort[T any](size int) *ChannelPort[T] {
	pipe := make(chan T, size)
	feedback := make(chan channelPortFeedback, size)

	input := NewChannelSink(pipe)

	output := NewChannelSource(pipe)
	output.feedback = feedback

	port := &ChannelPort[T]{
		Input:    input,
		Output:   output,
		Feedback: feedback,
	}
	return port
}
