package main

import (
	"github.com/erodrigufer/CTForchestrator/internal/ctfsmd/cli"
)

func main() {

	t := new(tui)
	cli.Execute(t)

}
