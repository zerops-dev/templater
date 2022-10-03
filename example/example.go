package example

import (
	"fmt"
	"time"

	"git.vsh-labs.cz/golang/templater/v2/example/test"
	"git.vsh-labs.cz/golang/templater/v2/example/test/subtest"
)

//go:generate go run ../cmd/templater/main.go
var _ = test.Test{
	Name: "aaa.go",
}

//go:generate go run ../cmd/templater/main.go
var _ = subtest.Test{
	Name: "subtest.go",
}

//go:generate go run ../cmd/templater/main.go
var _ = func() test.Test {
	t := test.NewTest()
	t.Name = fmt.Sprintf("xxxxx %d %d %d %d", 1, 2, 3, 4)
	t.Time = time.Now()
	for i := 1; i < 10; i++ {
		t.Ss = append(t.Ss, test.SubStruct{
			Name: fmt.Sprintf("%05d", i),
			Type: fmt.Sprintf("Type%5d", i),
		})
	}
	return t
}

//go:generate go run ../cmd/templater/main.go
var _ = func() test.Test {
	return test.Test{
		Name: fmt.Sprintf("xxxxx %d %d %d %d %d", 1, 2, 3, 4, 5),
		S: test.SubStruct{
			Name: "aaaaa",
			Type: "tyepAAAA",
		},
		Ss: []test.SubStruct{
			{
				Name: "aaa.go",
				Type: "Bbbbb",
			},
			{
				Name: "bbb",
				Type: "Cccc",
			},
		},
		Time: time.Now(),
	}
}
