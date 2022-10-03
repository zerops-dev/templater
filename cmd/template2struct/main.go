package main

import (
	"log"

	"git.vsh-labs.cz/golang/templater/v2/internal"
)

func main() {
	dir, file, line, err := internal.Env()
	if err != nil {
		log.Fatal(err)
	}
	if err := internal.Template(dir, file, line); err != nil {
		log.Fatal(err)
	}

}
