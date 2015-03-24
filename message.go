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
