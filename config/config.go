package config

import (
	"errors"
	"os"

	"github.com/nalanj/confl"
	"github.com/nalanj/ladle/fn"
)

// Config is a struct representing the configuration of the service
type Config struct {

	// Functions is a map of the named functions
	Functions map[string]*fn.Function
}

// Parse parses the config file and returns the resulting config
func Parse(path string) (*Config, error) {
	conf := &Config{
		Functions: make(map[string]*fn.Function),
	}

	file, openErr := os.Open(path)
	if openErr != nil {
		return nil, openErr
	}

	doc, parseErr := confl.Parse(file)
	if parseErr != nil {
		return nil, parseErr
	}

	var key confl.Node
	for _, node := range doc.Children() {
		if key == nil {
			key = node
		} else {
			if key.Type() != confl.WordType && key.Type() != confl.StringType {
				return nil, errors.New("Invalid function name")
			}

			fnDef, fnErr := readFunction(key.Value(), node)
			if fnErr != nil {
				return nil, fnErr
			}

			conf.Functions[fnDef.Name] = fnDef
			key = nil
		}
	}

	return conf, nil
}

// readFunction reads a function def from the config
func readFunction(name string, fnNode confl.Node) (*fn.Function, error) {
	if fnNode.Type() != confl.MapType {
		return nil, errors.New("Invalid function definition")
	}

	out := &fn.Function{Name: name}

	var key confl.Node
	for _, node := range fnNode.Children() {
		if key == nil {
			key = node
		} else {
			switch key.Value() {
			case "Handler":
				if key.Type() != confl.WordType && key.Type() != confl.StringType {
					return nil, errors.New("Invalid key")
				}

				if node.Type() != confl.WordType && node.Type() != confl.StringType {
					return nil, errors.New("Invalid handler path")
				}

				out.Handler = node.Value()
			default:
				return nil, errors.New("Invalid key")
			}
			key = nil
		}
	}

	return out, nil
}
