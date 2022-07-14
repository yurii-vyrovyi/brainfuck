package stack

import "container/list"

type Stack[T any] struct {
	l list.List
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
