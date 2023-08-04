package task

type State uint8

const (
	Incomplete State = iota
	Completed
	Error
)

// Result encapsulates the result of a task
type Result[R any] struct {
	State   State
	Problem error
	Output  R
}

func (r Result[R]) Done() bool {
	switch r.State {
	case Completed, Error:
		return true
	default:
		return false
	}
}
