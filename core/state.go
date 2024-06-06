package core

import (
	"fmt"
)

type State struct {
	data map[string][]byte
}

func NewState() *State {
	return &State{
		data: make(map[string][]byte),
	}
}

func (s *State) Add(k, v []byte) {
	s.data[string(k)] = v
}

func (s *State) Get(k []byte) ([]byte, error) {
	value, ok := s.data[string(k)]
	if !ok {
		return nil, fmt.Errorf("State.Get: given key (%s) not found", k)
	}
	return value, nil
}

func (s *State) Delete(k []byte) {
	delete(s.data, string(k))
}
