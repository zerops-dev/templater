package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type TemplateContext struct {
	Fs                 *token.FileSet
	AstFile            *ast.File
	Success            bool
	WorkingDirectory   string
	Filename           string
	TemplateFilename   string
	TemplateGoFilename string
	File               string
	Package            string
	StructName         string
}

func Env() (dir, file string, line int, err error) {

	dir, err = os.Getwd()
	if err != nil {
		return
	}

	dir = strings.TrimSuffix(dir, string(filepath.Separator)) + string(filepath.Separator)

	line, err = strconv.Atoi(os.Getenv("GOLINE"))
	if err != nil {
		return
	}
	file = os.Getenv("GOFILE")
	return
}

func TemplateInfoLine(dir, filename string, structName string) (int, string, error) {
	if filename == "" {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return 0, filename, err
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			line, filename, err := TemplateInfoLine(dir, file.Name(), structName)
			if err != nil {
				continue
			}
			return line, filename, nil
		}
		return 0, filename, fmt.Errorf("not found")
	}

	fileData, err := os.Open(path.Join(dir, filename))
	if err != nil {
		return 0, filename, err
	}
	defer fileData.Close()

	fs := token.NewFileSet()
	astFile, err := parser.ParseFile(fs, filename, fileData, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return 0, filename, err
	}

	line := 0
	var found bool
	ast.Inspect(astFile, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		switch c := n.(type) {
		case *ast.TypeSpec:
			if c.Name.Name == structName {
				line = fs.Position(n.Pos()).Line - 1
				found = true
				return false
			}
		}
		return true
	})
	if !found {
		return line, filename, errors.New("not found")
	}
	return line, filename, nil
}

func TemplateInfo(dir, filename string, line int) (TemplateContext, error) {
	rand.Seed(time.Now().UnixNano())

	ctx := TemplateContext{
		Fs:                 token.NewFileSet(),
		WorkingDirectory:   dir,
		Filename:           filename,
		TemplateFilename:   strings.TrimSuffix(filename, ".go") + ".tmpl",
		TemplateGoFilename: path.Join(dir, strings.TrimSuffix(filename, ".go")+"_template.go"),
		File:               path.Join(dir, filename),
	}

	fileData, err := os.Open(ctx.File)
	if err != nil {
		panic(err)
	}
	defer fileData.Close()

	ctx.AstFile, err = parser.ParseFile(ctx.Fs, ctx.Filename, fileData, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return ctx, err
	}

	ctx.Package = ctx.AstFile.Name.Name

	ast.Inspect(ctx.AstFile, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		p := ctx.Fs.Position(n.Pos())
		if p.Line != line+1 {
			return true
		}

		switch c := n.(type) {
		case *ast.TypeSpec:
			ctx.StructName = c.Name.Name
			ctx.Success = true
			return false
		}
		return true
	})

	if !ctx.Success {
		return ctx, fmt.Errorf("not found")
	}
	return ctx, nil
}
