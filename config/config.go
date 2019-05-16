package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	"github.com/nalanj/confl"
)

// Config is a struct representing the configuration of the service
type Config struct {
	// Path is the path to the config file
	Path string

	// RPCAddress is the address for listening for RPC
	RPCAddress string

	// HTTPAddress is the address for listening for HTTP
	HTTPAddress string

	// Functions is a map of the named functions for access to their
	// configurations
	Functions map[string]*Function

	// Events is a slice of defined events
	Events []*Event
}

// ParsePath parses the config file at the given path and returns the resulting
// config
func ParsePath(path string) (*Config, error) {
	file, openErr := os.Open(path)
	if openErr != nil {
		return nil, openErr
	}
	defer file.Close()

	conf, parseErr := parse(file)
	if parseErr != nil {
		return nil, parseErr
	}

	conf.Path = path
	return conf, nil
}

// RuntimeDir returns the .ladle dir where ladle caches builds
func (conf *Config) RuntimeDir() string {
	return path.Join(path.Dir(conf.Path), ".ladle")
}

// EnsureRuntimeDir creates the runtime directory, .ladle
// if it's not present
func (conf *Config) EnsureRuntimeDir() error {
	_, statErr := os.Stat(conf.RuntimeDir())
	if os.IsNotExist(statErr) {
		return os.Mkdir(conf.RuntimeDir(), os.ModePerm)
	}
	return statErr
}

// FunctionExecutable returns the executable path for  the given function
func (conf *Config) FunctionExecutable(f *Function) string {
	ext := ""
	if runtime.GOOS == "windows" {
		// add .exe on windows
		ext = ".exe"
	}

	return path.Join(conf.RuntimeDir(), f.Name) + ext
}

// parse parses the config file and returns the resulting config
func parse(reader io.Reader) (*Config, error) {
	conf := &Config{
		Functions: make(map[string]*Function),
	}

	doc, parseErr := confl.Parse(reader)
	if parseErr != nil {
		return nil, parseErr
	}

	for _, pair := range confl.KVPairs(doc) {
		switch pair.Key.Value() {
		case "Functions":
			functions, fnErr := readFunctions(pair.Value)
			if fnErr != nil {
				return nil, fnErr
			}
			conf.Functions = functions
		case "Events":
			events, eventsErr := readEvents(pair.Value)
			if eventsErr != nil {
				return nil, eventsErr
			}
			conf.Events = events
		default:
			return nil, fmt.Errorf("Unknown key")
		}
	}

	return conf, nil
}

// readFunctions reads the functions for the document
func readFunctions(fnsNode confl.Node) (map[string]*Function, error) {
	if fnsNode.Type() != confl.MapType {
		return nil, errors.New("Expected map for Functions section")
	}

	functions := make(map[string]*Function)

	for _, pair := range confl.KVPairs(fnsNode) {
		fnDef, fnErr := readFunction(pair.Key, pair.Value)
		if fnErr != nil {
			return nil, fnErr
		}

		functions[fnDef.Name] = fnDef
	}

	return functions, nil
}

// readFunction reads a function def from the config
func readFunction(fnKey confl.Node, fnNode confl.Node) (*Function, error) {
	name := fnKey.Value()

	if fnNode.Type() != confl.MapType {
		return nil, errors.New("Invalid function definition")
	}

	out := &Function{Name: name}

	for _, pair := range confl.KVPairs(fnNode) {
		switch pair.Key.Value() {
		case "Package":
			if !confl.IsText(pair.Value) {
				return nil, errors.New("Invalid package")
			}

			out.Package = pair.Value.Value()
		default:
			return nil, errors.New("Invalid key")
		}
	}

	return out, nil
}

// readEvents reads confl nodes and converts them to events
func readEvents(eventsNode confl.Node) ([]*Event, error) {
	if eventsNode.Type() != confl.ListType {
		return nil, errors.New("Invalid Events section")
	}

	events := []*Event{}

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
func readEvent(eventNode confl.Node) (*Event, error) {
	if eventNode.Type() != confl.MapType {
		return nil, errors.New("Invalid event")
	}

	event := &Event{}

	for _, pair := range confl.KVPairs(eventNode) {
		switch pair.Key.Value() {
		case "Source":
			if !confl.IsText(pair.Value) {
				return nil, errors.New("Invalid event source")
			}

			event.Source = pair.Value.Value()
		case "Target":
			if !confl.IsText(pair.Value) {
				return nil, errors.New("Invalid event source")
			}

			event.Target = pair.Value.Value()

		case "Meta":
			meta, metaErr := readEventMeta(pair.Value)
			if metaErr != nil {
				return nil, metaErr
			}

			event.Meta = meta
		default:
			return nil, errors.New("Invalid key")
		}
	}

	return event, nil
}

// readEventMeta reads metadata
func readEventMeta(metaNode confl.Node) (map[string]string, error) {
	if metaNode.Type() != confl.MapType {
		return nil, errors.New("Invalid meta section")
	}

	meta := make(map[string]string)

	var key confl.Node
	for _, node := range metaNode.Children() {
		if key == nil {
			key = node
		} else {
			meta[key.Value()] = node.Value()
			key = nil
		}
	}

	return meta, nil
}
