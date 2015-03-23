package main

import (
	"fmt"
)

type Message map[string]interface{}

func (m *Message) Get(key string) (string, bool) {
	if (*m)[key] == nil {
		return "%{null}", false
	} else {
		return fmt.Sprint((*m)[key]), true
	}
}

// match message filters
func (m *Message) Filter() bool {
	for k, f := range filters {
		v, ok := m.Get(k)
		if !ok {
			return false
		}
		if fmt.Sprint(v) != f {
			return false
		}
	}
	return true
}
