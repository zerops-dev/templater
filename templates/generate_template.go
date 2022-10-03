package templates

import (
	_ "embed"
)

//go:embed generate.tmpl
var _Generate string

func (h Generate) Template() string {
	return _Generate
}
