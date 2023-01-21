package main

import (
	"github.com/erodrigufer/pongo/internal/pongo/cli"
)

func main() {

	t := new(tui)
	cli.Execute(t)

}
