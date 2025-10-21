package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"g0.codeh.io/chapgx/liv/pkg/wserv"
	"github.com/racg0092/rhombifer"
	"github.com/racg0092/rhombifer/pkg/models"
)

// Starts the live server
var ServeCmd = &rhombifer.Command{
	Name:      "serve",
	ShortDesc: "Serves the files in the specified location",
	Run: func(args ...string) error {
		if len(args) == 0 {
			return fmt.Errorf("expected filepath found none")
		}

		src, e := rhombifer.FindFlag("src")
		if e != nil {
			return e
		}

		source := src.Values[0]
		info, e := os.Stat(source)
		if e != nil {
			return e
		}

		var basedir string
		var indexfile string
		if info.IsDir() {
			basedir = source
			idxf, e := rhombifer.FindFlag("index")
			if e != rhombifer.ErroFlagNotFound && e != nil {
				return e
			}

			if e != nil {
				fmt.Println("index flag was not set will attemp to look up index file in dir")
				//TODO: need to finish logic
				return nil
			} else {
				indexfile = filepath.Base(idxf.Values[0])
			}

		} else {
			basedir = filepath.Dir(source)
			indexfile = filepath.Base(source)
		}

		fmt.Println("made it here")
		go wserv.RunServer(basedir, indexfile)

		_, e = rhombifer.FindFlag("open")
		if e == nil {
			switch runtime.GOOS {
			case "linux":
				// e := exec.Command("xdg-open", path).Start()
				e := exec.Command("zen", "--new-window", "http://"+wserv.Host+":"+wserv.PORT).Start()
				if e != nil {
					fmt.Println("failed to open browser", e)
				}
			}
		}

		r := <-wserv.Done
		if r != 0 {
			return fmt.Errorf("error server close with code %d", r)
		}

		return nil
	},
}

var servecmd_open = &models.Flag{
	Name:  "open",
	Short: "Opens the browser for you",
}

var servercmd_src = &models.Flag{
	Name:     "src",
	Short:    "Specified the source of the web server can be a file or a directory",
	Required: true,
}

var servecmd_index = &models.Flag{
	Name:  "index",
	Short: "Specifies index file to server",
}

func init() {
	ServeCmd.AddFlags(
		servecmd_open,
		servercmd_src,
		servecmd_index,
	)

	root := rhombifer.Root()
	root.AddSub(ServeCmd)
}
