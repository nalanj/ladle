package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/nalanj/confl"
	"github.com/nalanj/ladle/fn"
)

// Config is a struct representing the configuration of the service
type Config struct {

	// Functions is a map of the named functions
	Functions map[string]*fn.Function

	// Events is a slice of defined events
	Events []*fn.Event
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
			switch key.Value() {
			case "Functions":
				functions, fnErr := readFunctions(node)
				if fnErr != nil {
					return nil, fnErr
				}
				conf.Functions = functions
			case "Events":
				events, eventsErr := readEvents(node)
				if eventsErr != nil {
					return nil, eventsErr
				}
				conf.Events = events
			default:
				return nil, fmt.Errorf("Unknown key")
			}

			key = nil
		}
	}

	return conf, nil
}

// readFunctions reads the functions for the document
func readFunctions(fnsNode confl.Node) (map[string]*fn.Function, error) {
	if fnsNode.Type() != confl.MapType {
		return nil, errors.New("Expected map for Functions section")
	}

	functions := make(map[string]*fn.Function)

	var key confl.Node
	for _, node := range fnsNode.Children() {
		if key == nil {
			key = node
		} else {
			fnDef, fnErr := readFunction(key, node)
			if fnErr != nil {
				return nil, fnErr
			}

			functions[fnDef.Name] = fnDef
			key = nil
		}
	}

	return functions, nil
}

// readFunction reads a function def from the config
func readFunction(fnKey confl.Node, fnNode confl.Node) (*fn.Function, error) {
	if fnKey.Type() != confl.StringType && fnKey.Type() != confl.WordType {
		return nil, errors.New("Invalid function name")
	}
	name := fnKey.Value()

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

// readEvents reads confl nodes and converts them to events
func readEvents(eventsNode confl.Node) ([]*fn.Event, error) {
	if eventsNode.Type() != confl.ListType {
		return nil, errors.New("Invalid Events section")
	}

	events := []*fn.Event{}

	for _, node := range eventsNode.Children() {
		event, eventErr := readEvent(node)
		if eventErr != nil {
			return nil, eventErr
		}

		events = append(events, event)
	}

	return events, nil
}

// readEvent reads a single event from confl
func readEvent(eventNode confl.Node) (*fn.Event, error) {
	if eventNode.Type() != confl.MapType {
		return nil, errors.New("Invalid event")
	}

	event := &fn.Event{}

	var key confl.Node
	for _, node := range eventNode.Children() {
		if key == nil {
			key = node
		} else {
			switch key.Value() {
			case "Source":
				if node.Type() != confl.WordType && node.Type() != confl.StringType {
					return nil, errors.New("Invalid event source")
				}

				event.Source = node.Value()
			case "Target":
				if node.Type() != confl.WordType && node.Type() != confl.StringType {
					return nil, errors.New("Invalid event source")
				}

				event.Target = node.Value()
			}
			key = nil
		}
	}

	return event, nil
}
