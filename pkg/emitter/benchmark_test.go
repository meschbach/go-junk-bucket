package emitter

import (
	"context"
	"testing"
)

func benchEmit(b *testing.B, newEmitter func() Emitter[int], listeners int) {
	e := newEmitter()
	for range listeners {
		e.On(func(_ context.Context, _ int) {})
	}
	b.ResetTimer()
	for b.Loop() {
		e.Emit(context.Background(), 1) //nolint
	}
}

func benchEmitParallel(b *testing.B, newEmitter func() Emitter[int], listeners int) {
	e := newEmitter()
	for range listeners {
		e.On(func(_ context.Context, _ int) {})
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			e.Emit(context.Background(), 1) //nolint
		}
	})
}

func BenchmarkDispatcherEmit(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run("Listeners="+itoa(n), func(b *testing.B) {
			benchEmit(b, func() Emitter[int] { return NewDispatcher[int]() }, n)
		})
	}
}

func BenchmarkMutexDispatcherEmit(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run("Listeners="+itoa(n), func(b *testing.B) {
			benchEmit(b, func() Emitter[int] { return NewMutexDispatcher[int]() }, n)
		})
	}
}

func BenchmarkDispatcherEmitParallel(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run("Listeners="+itoa(n), func(b *testing.B) {
			benchEmitParallel(b, func() Emitter[int] { return NewDispatcher[int]() }, n)
		})
	}
}

func BenchmarkMutexDispatcherEmitParallel(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run("Listeners="+itoa(n), func(b *testing.B) {
			benchEmitParallel(b, func() Emitter[int] { return NewMutexDispatcher[int]() }, n)
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
