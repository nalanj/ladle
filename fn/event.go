package fn

const (
	// APISource is the source name of api events
	APISource = "API"
)

// Event represents an event within the system
type Event struct {
	// Source is the source of an event
	Source string

	// Target is the function to be called on the event
	Target string

	// Meta is a map of config options that can be applied to the event source
	Meta map[string]string
}
