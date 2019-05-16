package config

// Function represents a specific function being run
type Function struct {
	// Name is the name of the function
	Name string

	// Package is the go package to be built for the function
	Package string
}
