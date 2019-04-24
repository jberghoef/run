package main

import (
	"flag"
	"log"
	"path/filepath"
	"regexp"

	"github.com/fatih/color"
)

var filename = "Runfile.yaml"

var requests []string
var eRe *regexp.Regexp
var vRe *regexp.Regexp
var context map[string]string

func init() {
	flag.StringVar(&filename, "file", filename, "The file to run commands from")
	flag.Parse()

	path, err := filepath.Abs(filename)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	filename = path

	requests = flag.Args()
	eRe = regexp.MustCompile(`\${(?P<Key>\S+)}`)
	vRe = regexp.MustCompile(`{{2}\.(?P<Key>\S+)}{2}`)
	context = make(map[string]string)
}

func main() {
	files := findRunfiles()

	if len(files) > 0 {
		if len(requests) == 0 {
			file := files[0]
			for _, command := range file.Commands {
				requests = append(requests, command.Key.(string))
			}
		}

		for _, request := range requests {
			found := false
			for _, runfile := range files {
				c, err := runfile.FindCommand(request)
				if err != nil {
					continue
				}

				found = true
				runfile.ProcessCommand(c)
			}

			if !found {
				color.Red("Command \"%s\" not found.\n", request)
			}
		}
	}
}
