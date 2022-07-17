package stack

import "container/list"

type Stack[T any] struct {
	l list.List
}

func BuildStack[T any](values ...T) *Stack[T] {
	s := Stack[T]{}

	for i := len(values) - 1; i >= 0; i-- {
		s.Push(values[i])
	}

	return &s
}

func (s *Stack[T]) Push(v T) {
	s.l.PushFront(&v)
}

func (s *Stack[T]) Pop() *T {
	e := s.l.Front()
	if e == nil {
		return nil
	}

	s.l.Remove(e)

	v := e.Value.(*T)
	return v
}

func (s *Stack[T]) Get() *T {
	e := s.l.Front()
	if e == nil {
		return nil
	}

	v := e.Value.(*T)
	return v
}

func (s *Stack[T]) Len() int {
	return s.l.Len()
}

func (s *Stack[T]) Equals(v *Stack[T], cmp func(a, b *T) bool) bool {

	if s.Len() != v.Len() {
		return false
	}

	sV := v.l.Front()

	for sE := s.l.Front(); sE != nil; sE = sE.Next() {

		if sV == nil {
			return false
		}

		tS := sE.Value.(*T)
		tV := sV.Value.(*T)

		if !cmp(tS, tV) {
			return false
		}

		sV = sV.Next()
	}

	return true
}
