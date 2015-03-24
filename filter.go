package main

import (
	"fmt"
)

type Filter struct {
	filters map[string]string
	printer Printer
}

func NewFilter(f map[string]string, p Printer) Filter {
	fltr := Filter{}
	fltr.filters = f
	fltr.printer = p
	return fltr
}

// match message filters
func (f *Filter) filter(m *Message) bool {
	for k, f := range f.filters {
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

// print message with printer
func (f *Filter) Print(m *Message) {
	if f.filter(m) {
		f.printer.Print(m)
	}
}
