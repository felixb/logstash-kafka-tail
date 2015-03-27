package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockedPrinter struct {
	mock.Mock
}

func (mck *MockedPrinter) Print(m *Message) {
	mck.Called()
}

func TestFilterPrint(t *testing.T) {
	p := new(MockedPrinter)
	p.On("Print").Return()

	filters := map[string]string{"a": "1"}
	f := NewFilter(filters, p)

	m := Message{"b": "2"}
	f.Print(&m)
	p.AssertNotCalled(t, "Print")

	m = Message{"a": "1"}
	f.Print(&m)
	p.AssertCalled(t, "Print")
}

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
