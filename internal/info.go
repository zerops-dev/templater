package internal

import (
	"crypto/md5"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type ValueType int

const (
	ValueTypeValue = ValueType(iota + 1)
	ValueTypeFunc
)

type InfoContext struct {
	Tags             string
	CommentPrefix    string
	TemplateTags     string
	Fs               *token.FileSet
	AstFile          *ast.File
	Line             int
	Success          bool
	WorkingDirectory string

	Filename string
	File     string

	TemplatePkg      string
	TemplatePath     string
	TemplateStruct   string
	TemplateFilename string
	TemplateLine     int

	TemplateRandName string

	DataPkg           string
	ProjectPkg        string
	PackagePath       string
	ProjectPath       string
	Package           string
	IdentName         string
	ValueType         ValueType
	IsPtr             bool
	GeneratedFilename string
	ValueGetter       string
}

func (i InfoContext) PtrStar() string {
	if i.IsPtr {
		return "*"
	}
	return ""
}

func Info(dir, filename string, line int, options ...Option) (InfoContext, error) {
	config := NewConfig()
	for _, option := range options {
		config = option(config)
	}
	info := func(v ...interface{}) {
		if config.Verbose {
			log.Print(v...)
		}
	}

	infof := func(format string, v ...interface{}) {
		if config.Verbose {
			log.Printf(format, v...)
		}
	}

	ctx := InfoContext{
		Line:             line,
		Tags:             config.Tags,
		CommentPrefix:    config.CommentPrefix,
		TemplateTags:     config.TemplateTags,
		Fs:               token.NewFileSet(),
		WorkingDirectory: dir,
		Filename:         filename,
		File:             path.Join(dir, filename),
		TemplateRandName: fmt.Sprintf("TemplateName_%x", md5.Sum([]byte(filename))),
	}

	fileData, err := os.Open(ctx.File)
	if err != nil {
		return ctx, errors.Wrapf(err, "open file %s ", ctx.File)
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
		case *ast.GenDecl:
			if i, isI := c.Specs[0].(*ast.ValueSpec); isI {
				i.Names[0].Name = ctx.TemplateRandName
				if len(i.Values) == 0 {
					return true
				}
				switch value := (i.Values[0]).(type) {
				case *ast.CompositeLit:

					if exprSel, isExprSel := (value.Type).(*ast.SelectorExpr); isExprSel {
						parseStructFromSelectorExpr(&ctx, exprSel)
						ctx.ValueType = ValueTypeValue
						return false
					}

				case *ast.FuncLit:
					if len(value.Type.Results.List) == 0 {
						return true
					}
					selType := value.Type.Results.List[0].Type

					if starExp, isStarExp := selType.(*ast.StarExpr); isStarExp {
						ctx.IsPtr = true
						selType = starExp.X
					}

					if exprSel, isExprSel := selType.(*ast.SelectorExpr); isExprSel {
						parseStructFromSelectorExpr(&ctx, exprSel)
						ctx.ValueType = ValueTypeFunc
						return false
					}
				}
			}
		}
		return true
	})

	if !ctx.Success {
		return ctx, fmt.Errorf("not found")
	}

	ctx.GeneratedFilename = strings.TrimSuffix(ctx.Filename, ".go") + "_" + FirstLower(ctx.TemplateStruct) + ".go"
	switch ctx.ValueType {
	case ValueTypeValue:
		ctx.ValueGetter = ctx.TemplateRandName
	case ValueTypeFunc:
		ctx.ValueGetter = ctx.TemplateRandName + "()"
	}
	gopath := os.Getenv("GOPATH")
	info("gopath: ", gopath)
	info("wd: ", ctx.WorkingDirectory)
	info("prefix: ", gopath+string(filepath.Separator))
	if strings.HasPrefix(ctx.WorkingDirectory, gopath+string(filepath.Separator)) && gopath != "" {
		// gomod disabled
		info("gomod disabled")
		ctx.DataPkg = strings.TrimPrefix(ctx.WorkingDirectory, path.Join(gopath, "src"))
		ctx.ProjectPkg = ctx.DataPkg
		ctx.ProjectPath = path.Join(gopath, "src", ctx.ProjectPkg)
		ctx.PackagePath = path.Join(gopath, "src", ctx.DataPkg)
	} else {
		info("gomod enabled")
		ctx.DataPkg, ctx.ProjectPkg, ctx.PackagePath, ctx.ProjectPath, err = getDirectoryPkgByModules(ctx.WorkingDirectory)
		info("gomod enabled", ctx.ProjectPkg)
		if err != nil {
			return ctx, err
		}
	}

	infof("DataPkg: %s", ctx.DataPkg)
	infof("ProjectPkg: %s", ctx.ProjectPkg)
	infof("TemplatePkg: %s", ctx.TemplatePkg)
	infof("PackagePath: %s", ctx.PackagePath)
	infof("ProjectPath: %s", ctx.ProjectPath)

	if strings.HasPrefix(ctx.TemplatePkg, ctx.ProjectPkg) {
		infof("generate template")
		ctx.TemplatePath = path.Join(ctx.ProjectPath, strings.TrimPrefix(ctx.TemplatePkg, ctx.ProjectPkg))

		infof("TemplateStruct: %s", ctx.TemplateStruct)
		infof("TemplatePath: %s", ctx.TemplatePath)

		ctx.TemplateLine, ctx.TemplateFilename, err = TemplateInfoLine(ctx.TemplatePath, "", ctx.TemplateStruct)

		if err != nil {
			return ctx, err
		}
		infof("TemplateLine: %d", ctx.TemplateLine)
		infof("TemplateFilename: %s", ctx.TemplateFilename)

		if err := Template(ctx.TemplatePath, ctx.TemplateFilename, ctx.TemplateLine); err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

func parseStructFromSelectorExpr(ctx *InfoContext, exprSel *ast.SelectorExpr) {
	ctx.TemplateStruct = exprSel.Sel.Name
	if ident, isIdent := exprSel.X.(*ast.Ident); isIdent {
		ctx.IdentName = ident.Name
	}

	for _, i := range ctx.AstFile.Imports {
		if i.Name != nil && i.Name.String() == ctx.IdentName {
			ctx.TemplatePkg, _ = strconv.Unquote(i.Path.Value)
			break
		}
		unqotePath, _ := strconv.Unquote(i.Path.Value)
		if path.Base(unqotePath) == ctx.IdentName {
			ctx.TemplatePkg = unqotePath
			break
		}
	}
	ctx.Success = true

}
