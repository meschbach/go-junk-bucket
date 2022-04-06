package actors

type FnActor struct {
	perform func(msg any)
}

func (f *FnActor) OnMessage(m any) {
	f.perform(m)
}

func NewFnActor(fn func(msg any)) *FnActor {
	return &FnActor{perform: fn}
}
