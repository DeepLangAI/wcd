package utils

import (
	"errors"
)

// Stack 是一个基于切片的栈实现
type Stack[T any] struct {
	items []T
}

// NewStack 创建一个新的栈
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{items: make([]T, 0)}
}

// Push 将一个元素压入栈
func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

// Pop 从栈中弹出一个元素
func (s *Stack[T]) Pop() (T, error) {
	if len(s.items) == 0 {
		var zero T
		return zero, errors.New("stack underflow")
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, nil
}

// Peek 查看栈顶元素，但不弹出
func (s *Stack[T]) Peek() (T, error) {
	if len(s.items) == 0 {
		var zero T
		return zero, errors.New("stack is empty")
	}
	return s.items[len(s.items)-1], nil
}

// IsEmpty 检查栈是否为空
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}
