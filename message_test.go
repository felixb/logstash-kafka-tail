package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageFilter(t *testing.T) {
	filters["a"] = "abc"
	filters["b"] = "1"
	filters["c"] = "true"

	m := Message{"a": "abc", "b": 1, "c": true, "foo": "bar"}
	assert.True(t, m.Filter())

	m = Message{"abc": "123"}
	assert.False(t, m.Filter())
}
