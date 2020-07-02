package main

import (
	"flag"
	"log"
	"path/filepath"
	"regexp"

	"github.com/fatih/color"
)

var filename = "Runfile.yaml"
var verbose = false

var requests []string
var eRe *regexp.Regexp
var vRe *regexp.Regexp
var cRe *regexp.Regexp
var context map[string]string

func init() {
	flag.StringVar(&filename, "file", filename, "The file to run commands from")
	flag.BoolVar(&verbose, "verbose", verbose, "Whether to show additional information")
	flag.Parse()

	path, err := filepath.Abs(filename)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	filename = path

	requests = flag.Args()
	eRe = regexp.MustCompile(`\${(?P<Key>\S+)}`)
	vRe = regexp.MustCompile(`{{2}\.(?P<Key>\S+)}{2}`)
	cRe = regexp.MustCompile(`["'].+?["']|\S+`)
	context = make(map[string]string)
}

func main() {
	files := findRunfiles()

	if len(files) > 0 {
		if len(requests) == 0 {
			runfile := files[0]
			for _, command := range runfile.Commands {
				c, err := runfile.FindCommand(command.Key.(string))
				if err != nil {
					continue
				}

				runfile.ProcessCommand(c)
			}
		} else {
			for _, request := range requests {
				found := false
				for _, runfile := range files {
					c, err := runfile.FindCommand(request)
					if err != nil {
						continue
					}

					found = true
					runfile.ProcessCommand(c)
					break
				}

				if !found {
					color.Red("command \"%s\" not found.\n", request)
				}
			}
		}
	}
}
