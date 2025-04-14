package routing

import "github.com/meschbach/go-junk-bucket/pkg/actors"

// Bridge is a strategy for correlating messages with keys and spawning actors for a given key
type Bridge[M any, K any] interface {
	Extract(msg M) K
	SpawnRoutee(bif actors.Runtime, key K, msg M) actors.Pid
}
