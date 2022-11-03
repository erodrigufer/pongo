package main

import (
	"github.com/erodrigufer/CTForchestrator/cmd/ctfsmd/cli"
)

func main() {

	t := new(tui)
	cli.Execute(t)

}
