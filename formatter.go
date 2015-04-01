package main

import (
	"fmt"
	"regexp"
)

type Formatter struct {
	formatString string
	keys         []string
}

// create a new formatter
func NewFormatter(frmtStr string) Formatter {
	re := regexp.MustCompile("%{[^}]+}")
	f := Formatter{}
	f.formatString = re.ReplaceAllStringFunc(frmtStr, func(s string) string {
		key := s[2 : len(s)-1]
		f.keys = append(f.keys, key)
		return "%v"
	})
	return f
}

// format a message
func (f *Formatter) format(m *Message) string {
	if len(f.formatString) > 0 {
		var values []interface{}
		for _, k := range f.keys {
			v, _ := m.Get(k)
			values = append(values, v)
		}

		return fmt.Sprintf(f.formatString, values...)
	} else {
		j, err := m.ToJson()
		if err != nil {
			return fmt.Sprintf("error: %v", err)
		} else {
			return string(j)
		}
	}
}

// print formatted message to stdout
func (f *Formatter) Print(m *Message) {
	fmt.Println(f.format(m))
}
