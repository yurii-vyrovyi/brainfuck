package stack

import "container/list"

type Stack struct {
	l list.List
}

func (s *Stack) Push(v interface{}) {
	s.l.PushFront(v)
}

func (s *Stack) Pop() interface{} {
	e := s.l.Front()
	if e == nil {
		return nil
	}

	s.l.Remove(e)

	return e.Value
}

func (s *Stack) Get() interface{} {
	e := s.l.Front()
	if e == nil {
		return nil
	}

	return e.Value
}
