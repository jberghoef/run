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
var debug = false

var requests []string
var envRe *regexp.Regexp
var varRe *regexp.Regexp
var cmdRe *regexp.Regexp
var context map[string]string

func init() {
	flag.StringVar(&filename, "file", filename, "The file to run commands from")
	flag.BoolVar(&verbose, "verbose", verbose, "Whether to show additional information")
	flag.BoolVar(&debug, "debug", debug, "Whether to show debugging information")
	flag.Parse()

	path, err := filepath.Abs(filename)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	filename = path

	requests = flag.Args()
	envRe = regexp.MustCompile(`\${(?P<Key>\S+)}`)
	varRe = regexp.MustCompile(`{{2}\.(?P<Key>\S+)}{2}`)
	cmdRe = regexp.MustCompile(`["'].+?["']|\S+`)
	context = make(map[string]string)
}

func main() {
	files := findRunfiles()

	if len(files) > 0 {
		if len(requests) == 0 {
			runfile := files[0]
			debugPrintf("Active runfile: %s\n", runfile.Path+runfile.Filename)
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
					debugPrintf("Selected runfile: %s\n", runfile.Path+runfile.Filename)
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
