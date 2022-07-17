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

		require.Equal(t, i, s.Len())
	}

	require.Equal(t, 0, s.Len())

	nilGet := s.Get()
	require.Nil(t, nilGet)

	nilPop := s.Pop()
	require.Nil(t, nilPop)
}

func TestStack_Build(t *testing.T) {
	t.Parallel()

	src := []int{1, 2, 3, 4, 5}

	stck := BuildStack(src...)

	require.Equal(t, stck.l.Len(), len(src))

	i := 0
	v := stck.l.Front()
	for v != nil && i < len(src) {

		nV := v.Value.(*int)
		require.NotNil(t, nV)

		require.Equal(t, src[i], *nV)

		v = v.Next()
		i++
	}

}

func TestStack_Equals(t *testing.T) {
	t.Parallel()

	type Test struct {
		s1     *Stack[int16]
		s2     *Stack[int16]
		expRes bool
	}

	tests := map[string]Test{
		"Equal": {
			s1:     BuildStack[int16](1, 2, 3, 4, 5),
			s2:     BuildStack[int16](1, 2, 3, 4, 5),
			expRes: true,
		},

		"Not equal. Different length": {
			s1:     BuildStack[int16](1, 2, 3, 4, 5),
			s2:     BuildStack[int16](1, 2, 3, 4),
			expRes: false,
		},

		"Not equal. Different values": {
			s1:     BuildStack[int16](1, 2, 3, 4, 5),
			s2:     BuildStack[int16](2, 3, 4, 5, 6),
			expRes: false,
		},
	}

	//nolint:paralleltest
	for description, test := range tests {
		test := test

		t.Run(description, func(t *testing.T) {
			t.Parallel()

			res := test.s1.Equals(test.s2, func(a, b *int16) bool { return *a == *b })

			require.Equal(t, test.expRes, res)
		})
	}
}
