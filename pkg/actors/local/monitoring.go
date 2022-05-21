package local

import (
	"context"
	"github.com/meschbach/go-junk-bucket/pkg/actors"
	"github.com/meschbach/go-junk-bucket/pkg/fx"
)

type startMonitoring struct {
	listener actors.Pid
	what     any
}

func (s *startMonitoring) execute(ctx context.Context, r *runtime) {
	r.monitoring = append(r.monitoring, *s)
}

func (s *startMonitoring) name() string {
	return "startMonitoring"
}

type stopMonitoring struct {
	listener actors.Pid
}

func (s *stopMonitoring) execute(ctx context.Context, r *runtime) {
	r.monitoring = fx.Filter[startMonitoring](r.monitoring, func(m startMonitoring) bool {
		return m.listener != s.listener
	})
}

func (s *stopMonitoring) name() string {
	return "stopMonitoring"
}
