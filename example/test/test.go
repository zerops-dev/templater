package test

import (
	"text/template"
	"time"
)

type SubStruct struct {
	Name string
	Type string
}

//go:generate go run ../../cmd/template2struct/main.go
type Test struct {
	Name string
	Time time.Time
	S    SubStruct
	Ss   []SubStruct
}

func NewTest() Test {
	return Test{
		Name: "aaaa",
	}
}

func (Test) CustomFuncMap() template.FuncMap {
	return map[string]interface{}{
		"Test": func() string {
			return "aaaaaa"
		},
	}
}
