package main

import "testing"

func TestFilter(t *testing.T) {
	filters["a"] = "abc"
	filters["b"] = "1"
	filters["c"] = "true"

	m := Message{}
	m["a"] = "abc"
	m["b"] = 1
	m["c"] = true
	m["foo"] = "bar"

	if !(filter(m)) {
		t.Errorf("filter(%q) == false with filters=%q, want true", m, filters)
	}

	m = Message{}
	m["abc"] = "123"

	if filter(m) {
		t.Errorf("filter(%q) == true with filters=%q, want false", m, filters)
	}
}

func TestFormat(t *testing.T) {
	formatString = "abc %{test} %{foo} %{glibberish} %{c} bar"

	m := Message{}
	m["test"] = "foo"
	m["foo"] = 1
	m["c"] = true

	want := "abc foo 1 %{null} true bar"
	got := format(m)

	if got != want {
		t.Errorf("format(%q) == %q, want %q", m, got, want)
	}
}
