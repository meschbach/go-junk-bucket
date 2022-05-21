package supervisor

import "github.com/meschbach/go-junk-bucket/pkg/actors"

type WatchState struct {
	Observer actors.Pid
}

type StateReady struct {
}
