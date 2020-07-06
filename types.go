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
	debugPrintf("FindCommand: %+v\n", name)

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
	debugPrintf("└ ProcessCommand: %v\n", command)

	t := reflect.TypeOf(command)

	if m, ok := command.(yaml.MapItem); ok {
		debugPrintf("  └ Found MapItem: %+v\n", m)

		if s, ok := m.Value.(yaml.MapSlice); ok {
			p := make(map[interface{}]interface{})
			for _, item := range s {
				p[item.Key] = item.Value
			}
			r.handleMap(p)
		} else {
			r.ProcessCommand(m.Value)
		}

		return
	}

	if m, ok := command.(yaml.MapSlice); ok {
		debugPrintf("  └ Found MapSlice: %+v\n", m)
		for _, command := range m {
			r.ProcessCommand(command)
		}
		return
	}

	switch t.Kind() {
	case reflect.String:
		debugPrintf("  └ Found String: %+v\n", command)
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
		debugPrintf("  └ Found Slice: %+v\n", command)
		r.handleArray(command.([]interface{}))
	case reflect.Map:
		debugPrintf("  └ Found Map: %+v\n", command)
		r.handleMap(command.(map[interface{}]interface{}))
	default:
		color.Red("Can't handle command of kind %s.\n", t.Kind())
	}
}

// ProcessEnv ...
func (r *Runfile) ProcessEnv(env interface{}) {
	debugPrintf("└ ProcessEnv: %v\n", env)

	t := reflect.TypeOf(env)
	if t.Kind() == reflect.Map {
		for key, value := range env.(map[interface{}]interface{}) {
			os.Setenv(key.(string), value.(string))
			if verbose {
				color.Cyan("$%s %s", key, value)
			}
		}
	} else if m, ok := env.(yaml.MapSlice); ok {
		for _, e := range m {
			os.Setenv(e.Key.(string), e.Value.(string))
			if verbose {
				color.Cyan("$%s %s", e.Key, e.Value)
			}
		}
	} else {
		color.Red("env has to be a map", env)
	}
}

func (r *Runfile) handleString(input interface{}) {
	debugPrintf("    └ handleString: %+v\n", input)

	matches := varRe.FindAllStringSubmatch(input.(string), -1)
	for i := range matches {
		name := matches[i][1]

		if _, ok := context[name]; !ok {
			debugPrintf("    └ requestVariable: %+v\n", name)

			reader := bufio.NewReader(os.Stdin)
			color.Yellow(fmt.Sprintf("Enter value for variable '%s': ", name))
			text, _ := reader.ReadString('\n')
			context[name] = strings.TrimSuffix(text, "\n")
		}
	}

	input = envRe.ReplaceAllStringFunc(input.(string), func(s string) string {
		match := envRe.FindAllStringSubmatch(s, -1)
		value, set := os.LookupEnv(match[0][1])
		if set {
			return value
		}
		return s
	})

	execute(input.(string))
}

func (r *Runfile) handleArray(input []interface{}) {
	debugPrintf("    └ handleArray: %+v\n", input)

	for _, line := range input {
		r.ProcessCommand(line)
	}
}

func (r *Runfile) handleMap(input map[interface{}]interface{}) {
	debugPrintf("    └ handleMap: %+v\n", input)

	options := input
	if input["optional"] != nil {
		if input["optional"].(string) != "" {
			color.Yellow(input["optional"].(string))
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
