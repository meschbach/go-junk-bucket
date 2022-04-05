package actors

type FnActor struct {
	perform func(msg any)
}

func (f *FnActor) OnMessage(m any) {
	f.perform(m)
}
