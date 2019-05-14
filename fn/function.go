package fn

// Function represents a specific function being run
type Function struct {
	// Name is the name of the function
	Name string

	// Handler is the path to the function binary
	Handler string
}
