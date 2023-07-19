package reactive

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type FromType int
type ToType int

func TestTransform(t *testing.T) {
	t.Run("Interface Compliance", func(t *testing.T) {
		transformer := NewTransform[FromType, ToType](func(ctx context.Context, input FromType) (output ToType, err error) {
			return ToType(input * 2), nil
		})
		assert.Implements(t, (*Sink[FromType])(nil), transformer)
		assert.Implements(t, (*Source[ToType])(nil), transformer)
	})

	t.Run("Given a fixed source and accumulator", func(t *testing.T) {
		source := FromSlice([]FromType{0, 1, 1, 2, 3, 5, 8})
		out := NewSliceAccumulator[ToType]()

		t.Run("When attached with a doubling transform", func(t *testing.T) {
			transformer := NewTransform[FromType, ToType](func(ctx context.Context, input FromType) (output ToType, err error) {
				return ToType(input * 2), nil
			})
			require.ErrorIs(t, transformer.Pump(context.Background(), source, out), End)

			t.Run("Then the accumulator has correct results", func(t *testing.T) {
				assert.Equal(t, []ToType{0, 2, 2, 4, 6, 10, 16}, out.Output)
			})
		})
	})
}
