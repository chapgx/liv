package main

import (
	_ "github.com/chapgx/liv/cmd"
	"github.com/racg0092/rhombifer"
	"github.com/racg0092/rhombifer/pkg/builtin"
)

func main() {
	if e := rhombifer.Start(); e != nil {
		panic(e)
	}
}

func init() {
	cfg := rhombifer.GetConfig()
	cfg.RunHelpIfNoInput = true

	root := rhombifer.Root()
	root.Name = "Liv "
	root.ShortDesc = "Live web server"
	root.LongDesc = `
Liv is a lightweigth simple web server. Intended for hot reload web development
`

	help := builtin.HelpCommand(nil, nil)
	root.AddSub(&help)
}
