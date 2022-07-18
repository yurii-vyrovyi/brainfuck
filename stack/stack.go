package stack

import "container/list"

// Stack implements FILO stack.
type Stack[T any] struct {
	l list.List
}

// BuildStack creates Stack instance and optionally may initialise it with values.
// Values are in reverse order – values[0] item will be popped first.
func BuildStack[T any](values ...T) *Stack[T] {
	s := Stack[T]{}

	for i := len(values) - 1; i >= 0; i-- {
		s.Push(values[i])
	}

	return &s
}

// Push puts the value on the top of the stack
func (s *Stack[T]) Push(v T) {
	s.l.PushFront(&v)
}

// Pop returns the top value and removes it from the stack.
// If stack is empty Pop() returns nil.
func (s *Stack[T]) Pop() *T {
	e := s.l.Front()
	if e == nil {
		return nil
	}

	s.l.Remove(e)

	v := e.Value.(*T)
	return v
}

// Get returns the top value keeping it in stack
// If stack is empty Pop() returns nil.
func (s *Stack[T]) Get() *T {
	e := s.l.Front()
	if e == nil {
		return nil
	}

	v := e.Value.(*T)
	return v
}

// Len returns the number of items in stack
func (s *Stack[T]) Len() int {
	return s.l.Len()
}

// Equals compares two stacks and returns true if the values in stack are identical
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
