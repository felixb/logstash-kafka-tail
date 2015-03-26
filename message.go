package main

import (
	"fmt"
	"strings"
)

type Message map[string]interface{}

func (m *Message) Get(key string) (string, bool) {
	keys := strings.Split(key, ",")
	for _, k := range keys {
		v, ok := (*m)[k]
		if ok {
			return fmt.Sprint(v), true
		}
	}
	return "%{null}", false
}
