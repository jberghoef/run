package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

// RunfileConstructor ...
func RunfileConstructor(path string) Runfile {
	r := Runfile{}
	r.Path, r.Filename = filepath.Split(path)

	file, err := ioutil.ReadFile(r.FilePath())
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = yaml.Unmarshal(file, &r.Commands)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return r
}

// Runfile ...
type Runfile struct {
	Commands yaml.MapSlice
	Path     string
	Filename string
	Offset   int
}

// FilePath ...
func (r *Runfile) FilePath() string {
	return filepath.Join(r.Path, r.Filename)
}

// FindCommand ...
func (r *Runfile) FindCommand(name interface{}) (m yaml.MapItem, err error) {
	for _, command := range r.Commands {
		if command.Key.(string) == name.(string) {
			m = command
			return
		}
	}
	message := fmt.Sprintf("command %s not found", name)
	err = errors.New(message)
	return
}

// ProcessCommand ...
func (r *Runfile) ProcessCommand(command interface{}) {
	t := reflect.TypeOf(command)

	if m, ok := command.(yaml.MapSlice); ok {
		for _, command := range m {
			r.ProcessCommand(command)
		}
		return
	}

	switch t.Kind() {
	case reflect.String:
		if strings.HasPrefix(command.(string), ":") {
			cmd, err := r.FindCommand(strings.Replace(command.(string), ":", "", 1))
			if err != nil {
				color.Red("#! Command named '%s' does not exist.\n", command)
			} else {
				r.ProcessCommand(cmd)
			}
		} else {
			r.handleString(command)
		}
	case reflect.Slice:
		r.handleArray(command.([]interface{}))
	case reflect.Map:
		r.handleMap(command.(map[interface{}]interface{}))
	case reflect.TypeOf(yaml.MapItem{}).Kind():
		r.ProcessCommand(command.(yaml.MapItem).Value)
	default:
		color.Red("Can't handle command of kind %s.\n", t.Kind())
	}
}

// ProcessEnv ...
func (r *Runfile) ProcessEnv(env interface{}) {
	t := reflect.TypeOf(env)
	if t.Kind() == reflect.Map {
		for key, value := range env.(map[interface{}]interface{}) {
			os.Setenv(key.(string), value.(string))
		}
	} else {
		color.Red("env has to be a map", env)
	}
}

func (r *Runfile) handleString(input interface{}) {
	matches := vRe.FindAllStringSubmatch(input.(string), -1)
	for i := range matches {
		name := matches[i][1]

		if _, ok := context[name]; !ok {
			reader := bufio.NewReader(os.Stdin)
			color.Yellow(fmt.Sprintf("Enter value for variable '%s': ", name))
			text, _ := reader.ReadString('\n')
			context[name] = strings.TrimSuffix(text, "\n")
		}
	}

	input = eRe.ReplaceAllStringFunc(input.(string), func(s string) string {
		match := eRe.FindAllStringSubmatch(s, -1)
		value, set := os.LookupEnv(match[0][1])
		if set {
			return value
		}
		return s
	})

	execute(input.(string))
}

func (r *Runfile) handleArray(input []interface{}) {
	for _, line := range input {
		r.ProcessCommand(line)
	}
}

func (r *Runfile) handleMap(input map[interface{}]interface{}) {
	options := input

	if input["optional"] != nil {
		if input["optional"].(bool) {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Execute command (Y/n): ")
			text, _ := reader.ReadString('\n')

			text = strings.ToLower(text)
			if text == "n\n" || text == "no\n" {
				return
			}
		}

		delete(options, "optional")
	}

	if input["env"] != nil {
		r.ProcessEnv(input["env"])
		delete(options, "env")
	}

	if input["command"] != nil {
		r.ProcessCommand(input["command"])
		delete(options, "command")
	} else if input["commands"] != nil {
		r.ProcessCommand(input["commands"])
		delete(options, "commands")
	}

	for key, value := range options {
		color.Magenta("#! %s\n", key)
		r.ProcessCommand(value)
	}
}
