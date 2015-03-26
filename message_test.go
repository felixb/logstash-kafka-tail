package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageGet(t *testing.T) {
	cases := map[string]Message{
		"foo0":    Message{"test": "foo0", "bar": "foo1", "foo": 123},
		"%{null}": Message{"foo": true},
		"foo1":    Message{"bar": "foo1"},
	}

	for want, m := range cases {
		v, _ := m.Get("test,bar")
		assert.Equal(t, want, v)
	}
}
