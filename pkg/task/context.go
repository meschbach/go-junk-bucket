package task

import "context"

func EnjoinContext(parentContext context.Context, enjoining context.Context) (context.Context, func()) {
	subcontext, done := context.WithCancel(parentContext)
	go func() {
		select {
		case <-enjoining.Done():
			done()
		case <-subcontext.Done():
		}
	}()

	return subcontext, done
}
