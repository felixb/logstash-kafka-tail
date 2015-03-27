package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterFilter(t *testing.T) {
	filters := map[string]string{"a": "abc", "b": "1", "c": "true"}
	f := NewFilter(filters, nil)

	m := Message{"a": "abc", "b": 1, "c": true, "foo": "bar"}
	assert.True(t, f.filter(&m))

	m = Message{"abc": "123"}
	assert.False(t, f.filter(&m))
}

func TestFilterFilterMultipleKeys(t *testing.T) {
	filters := map[string]string{"a,b": "abc"}
	f := NewFilter(filters, nil)

	m := Message{"a": "abc"}
	assert.True(t, f.filter(&m))

	m = Message{"b": "abc"}
	assert.True(t, f.filter(&m))

	m = Message{"c": "abc"}
	assert.False(t, f.filter(&m))
}

func TestFilterFilterMultipleValues(t *testing.T) {
	filters := map[string]string{"a": "abc,def"}
	f := NewFilter(filters, nil)

	m := Message{"a": "abc"}
	assert.True(t, f.filter(&m))

	m = Message{"a": "def"}
	assert.True(t, f.filter(&m))

	m = Message{"a": "foo"}
	assert.False(t, f.filter(&m))
}
