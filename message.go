package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Message map[string]interface{}

func NewMessageFromJson(v []byte) (*Message, error) {
	var m Message
	err := json.Unmarshal(v, &m)
	return &m, err
}

func (m *Message) ToJson() ([]byte, error) {
	return json.Marshal(m)
}

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
