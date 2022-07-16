package stack

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Stack(t *testing.T) {
	t.Parallel()

	s := Stack[int32]{}

	src := []int32{1, 2, 3, 4, 5}

	for _, v := range src {
		s.Push(v)
	}

	lenSrc := s.Len()
	require.Equal(t, len(src), lenSrc)

	for i := lenSrc - 1; i >= 0; i-- {
		top := s.Get()
		require.Equal(t, src[i], *top)

		popped := s.Pop()
		require.Equal(t, src[i], *popped)
	}

	require.Equal(t, 0, s.Len())

	nilGet := s.Get()
	require.Nil(t, nilGet)

	nilPop := s.Pop()
	require.Nil(t, nilPop)
}
