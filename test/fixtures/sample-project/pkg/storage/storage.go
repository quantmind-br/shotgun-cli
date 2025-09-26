package storage

import "errors"

var ErrNotFound = errors.New("record not found")

type InMemory struct {
	items map[string]string
}

func NewInMemory() *InMemory {
	return &InMemory{items: make(map[string]string)}
}

func (s *InMemory) Get(key string) (string, error) {
	value, ok := s.items[key]
	if !ok {
		return "", ErrNotFound
	}
	return value, nil
}
