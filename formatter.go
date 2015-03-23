package main

import (
	"fmt"
	"regexp"
	"strings"
)

type Formatter struct {
	FormatString string
	Keys         []string
}

// create a new formatter
func NewFormatter(formatString string) Formatter {
	re := regexp.MustCompile("%{[^}]+}")
	f := Formatter{}
	f.FormatString = re.ReplaceAllStringFunc(formatString, func(s string) string {
		key := s[2 : len(s)-1]
		f.Keys = append(f.Keys, key)
		return "%v"
	})
	return f
}

// format a message
func (f *Formatter) Format(m *Message) string {
	var values []interface{}
	for _, k := range f.Keys {
		v, _ := m.Get(k)
		values = append(values, v)
	}

	return fmt.Sprintf(f.FormatString, values...)
}

// print formatted message to stdout
func (f *Formatter) Print(m *Message) {
	fmt.Println(f.Format(m))
}

func (f *Formatter) String() string {
	return f.FormatString + " (" + strings.Join(f.Keys, ", ") + ")"
}
