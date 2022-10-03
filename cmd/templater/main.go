package main

import (
	"flag"
	"log"

	"git.vsh-labs.cz/golang/templater/v2/internal"
)

func main() {

	config := internal.NewConfig()

	flag.BoolVar(&config.ForceDelete, "forceDelete", config.ForceDelete, "force delete generated files")
	flag.StringVar(&config.Tags, "tags", config.Tags, "set build tags")
	flag.StringVar(&config.TemplateTags, "templateTags", config.TemplateTags, "set output file build tags")
	flag.StringVar(&config.CommentPrefix, "commentPrefix", config.CommentPrefix, "set line comment prefix, empty means no header")
	flag.BoolVar(&config.Verbose, "verbose", config.Verbose, "verbose output")
	flag.BoolVar(&config.Override, "override", config.Override, "override already generated file")
	flag.Parse()

	dir, file, line, err := internal.Env()
	if err != nil {
		log.Fatal(err)
	}
	if err := internal.Templater(dir, file, line, internal.WithConfig(config)); err != nil {
		log.Fatal(err)
		return
	}
}
