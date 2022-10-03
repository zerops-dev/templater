package templates

import (
	_ "embed"
)

//go:embed template.tmpl
var _Template string

func (h Template) Template() string {
	return _Template
}
