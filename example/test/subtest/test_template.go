package subtest

import (
	_ "embed"
)

//go:embed test.tmpl
var _Test string

func (h Test) Template() string {
	return _Test
}
