package main

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func findRunfiles() (runfiles []Runfile) {
	tests := []string{filename}
	dir, _ := filepath.Split(filename)
	parts := strings.Split(dir, string(filepath.Separator))
	for i := len(parts) - 1; i > 0; i-- {
		p := append(parts[:i], "Runfile.yaml")
		tests = append(tests, string(filepath.Separator)+filepath.Join(p...))
	}

	for _, test := range tests {
		if _, err := os.Stat(test); err == nil {
			runfiles = append(runfiles, RunfileConstructor(test))
		}
	}

	return
}

func execute(command string) {
	tmpl, err := template.New("test").Parse(command)
	if err != nil {
		log.Fatal(err)
	}

	var result bytes.Buffer
	err = tmpl.Execute(&result, context)
	if err != nil {
		log.Fatal(err)
	}

	color.Green("#! %s\n", result.String())

	parts := cRe.FindAllString(result.String(), -1)
	cmd := exec.Command(parts[0], parts[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}
