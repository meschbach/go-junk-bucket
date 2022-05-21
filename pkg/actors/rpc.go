package actors

import "time"

func CallService[S any, R any](bif Runtime, target Pid, action RpcAction[S, R]) R {
	mailbox := bif.SpawnMailbox()
	p := mailbox.Pid()
	bif.Monitor2(target, p)
	bif.Tell(target, RpcCall[S, R]{
		tell:   p,
		action: action,
	})
	result, problem := mailbox.ReceiveTimeout(100 * time.Millisecond)
	if problem != nil {
		bif.Log().Fatal("RPC failure %s", problem.Error())
	}
	if out, ok := result.(R); ok {
		return out
	} else {
		panic(result)
	}
}

type RpcCall[S any, R any] struct {
	tell   Pid
	action RpcAction[S, R]
}

func (r *RpcCall[S, R]) Perform(bif Runtime, state S) {
	result := r.action.Invoke(bif, state)
	bif.Tell(r.tell, result)
}

type RpcAction[S any, R any] interface {
	Invoke(bif Runtime, state S) R
}

type Action[S any] interface {
	Perform(bif Runtime, state *S)
}

func PerformOn[S any](ctx Ingestor, target Pid, action Action[S]) {
	ctx.Tell(target, action)
}
