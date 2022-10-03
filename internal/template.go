package internal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"text/template"

	"git.vsh-labs.cz/golang/templater/v2/templates"
	"golang.org/x/tools/imports"
)

func Template(dir, filename string, line int) error {
	ctx, err := TemplateInfo(dir, filename, line)
	if err != nil {
		return err
	}

	file, err := os.Open(path.Join(ctx.WorkingDirectory, ctx.TemplateFilename))
	if err != nil {
		return err
	}
	defer file.Close()

	output := &strings.Builder{}
	if err := func() error {
		base64Encoder := base64.NewEncoder(base64.StdEncoding, output)
		defer base64Encoder.Close()
		if _, err := io.Copy(base64Encoder, file); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}
	tmpls, err := template.New("template").Parse(GetTemplate(&templates.Template{}))
	if err != nil {
		return err
	}
	data := struct {
		Package          string
		Template         string
		StructName       string
		TemplateFilename string
	}{
		Package:          ctx.Package,
		StructName:       ctx.StructName,
		Template:         output.String(),
		TemplateFilename: ctx.TemplateFilename,
	}

	buf := &bytes.Buffer{}
	if err := tmpls.Execute(buf, data); err != nil {
		return err
	}

	formatted, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		os.Stdout.Write(buf.Bytes())
		return err
	}

	templateFile, err := os.Create(ctx.TemplateGoFilename)
	if err != nil {
		return err
	}
	defer templateFile.Close()

	if size, err := templateFile.Write(formatted); err != nil || size != len(formatted) {
		if err == nil {
			err = fmt.Errorf("unfinnished buffer write %d != %d", size, len(formatted))
		}
		return err
	}

	return nil
}
