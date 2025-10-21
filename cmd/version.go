package cmd

import (
	"fmt"

	"github.com/racg0092/rhombifer"
)

var VerCmd = &rhombifer.Command{
	Name:      "version",
	ShortDesc: "App version",
	Run: func(args ...string) error {
		fmt.Println("v0.1.0")
		return nil
	},
}

func init() {
	root := rhombifer.Root()
	root.AddSub(VerCmd)
}
