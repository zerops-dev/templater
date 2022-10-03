package internal

import (
	"fmt"
	"go/printer"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"

	"git.vsh-labs.cz/golang/templater/v2/templates"
)

type Config struct {
	Tags          string
	TemplateTags  string
	ForceDelete   bool
	Verbose       bool
	Override      bool
	CommentPrefix string
}

func NewConfig() Config {
	return Config{
		Tags:          "templater",
		TemplateTags:  "!templater",
		ForceDelete:   false,
		Verbose:       false,
		Override:      true,
		CommentPrefix: "//",
	}
}

type Option func(Config) Config

func WithConfig(config Config) Option {
	return func(_ Config) Config {
		return config
	}
}

func Templater(dir, file string, line int, options ...Option) error {
	config := NewConfig()
	for _, option := range options {
		config = option(config)
	}
	ctx, err := Info(dir, file, line, options...)
	if err != nil {
		return err
	}

	filename := path.Join(ctx.WorkingDirectory, ctx.Filename)
	renameFilename := filename + ".tmp"
	if err := os.Rename(filename, renameFilename); err != nil {
		return err
	}
	defer os.Rename(renameFilename, filename)

	tmpFilename := path.Join(ctx.WorkingDirectory, ctx.TemplateRandName+".go")
	tmpFile, err := os.Create(tmpFilename)
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	if err := printer.Fprint(tmpFile, ctx.Fs, ctx.AstFile); err != nil {
		return err
	}
	if !config.ForceDelete {
		defer os.Remove(tmpFilename)
	}

	if err := os.Mkdir(ctx.TemplateRandName, 0755); err != nil {
		return err
	}
	if !config.ForceDelete {
		defer os.RemoveAll(ctx.TemplateRandName)
	}

	generateFile := path.Join(ctx.TemplateRandName, "generate.go")
	if err := func() error {
		generate, err := os.Create(generateFile)
		if err != nil {
			return err
		}
		defer generate.Close()

		funcMap := template.FuncMap{
			"GoValue": func(value interface{}) string {
				return fmt.Sprintf("%#v", value)
			},
			"quote":   strconv.Quote,
			"ToUpper": strings.ToUpper,
			"ToLower": strings.ToLower,
		}

		tmpls, err := template.New("generate").Funcs(funcMap).Parse(GetTemplate(&templates.Generate{}))
		if err != nil {
			return err
		}
		if err := tmpls.Execute(generate, ctx); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}
	if !config.ForceDelete {
		defer os.Remove(generateFile)
	}

	args := []string{
		"run",
		"-gcflags", "all=-l -N", // -l disable inlining & -N disable optimizations
		"-tags", config.Tags, path.Join(ctx.TemplateRandName, "generate.go"),
	}
	if config.Override {
		args = append(args, "-override=true")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func GetTemplate(in interface{}) string {
	if t, isT := in.(interface {
		Template() string
	}); isT {
		return t.Template()
	}
	return ""
}
