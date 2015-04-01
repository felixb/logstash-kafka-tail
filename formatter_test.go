package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFormatter(t *testing.T) {
	formatString := "abc %{test} %{foo} bar"
	got := NewFormatter(formatString)
	want := Formatter{"abc %v %v bar", []string{"test", "foo"}}

	assert.Equal(t, want, got)
}

func TestFormatterFormat(t *testing.T) {
	formatString := "abc %{test} %{foo} bar"
	f := NewFormatter(formatString)

	cases := map[string]Message{
		"abc foo 123 bar":      Message{"test": "foo", "foo": 123},
		"abc %{null} true bar": Message{"foo": true},
	}

	for want, m := range cases {
		assert.Equal(t, want, f.format(&m))
	}
}

func TestFormatterFormatWithNoFormatString(t *testing.T) {
	f := NewFormatter("")

	cases := map[string]Message{
		"{\"foo\":123,\"test\":\"foo\"}": Message{"test": "foo", "foo": 123},
		"{\"foo\":true}":                 Message{"foo": true},
	}

	for want, m := range cases {
		assert.Equal(t, want, f.format(&m))
	}
}
