package config

import (
	"path"
	"testing"

	"github.com/nalanj/ladle/fn"
	"github.com/stretchr/testify/assert"
)

// TestParsePath tests the ParsePath function and everything under it as a
// series of tests against the whole configuration file format.
func TestParsePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		want    *Config
		wantErr bool
	}{
		{"invalid path", "non-existent.confl", nil, true},
		{"invalid confl", "invalid.confl", nil, true},
		{"unknown doc key", "unknown_doc_key.confl", nil, true},
		{"invalid Functions type", "invalid_functions.confl", nil, true},
		{"invalid Function type", "invalid_function_type.confl", nil, true},
		{"invalid Function key", "invalid_function_key.confl", nil, true},
		{
			"invalid Function handler type",
			"invalid_function_handler_type.confl",
			nil,
			true,
		},
		{"unknown function key", "unknown_function_key.confl", nil, true},
		{"invalid events", "invalid_events.confl", nil, true},
		{"invalid event type", "invalid_event_type.confl", nil, true},
		{"invalid event source", "invalid_event_source.confl", nil, true},
		{"invalid event target", "invalid_event_target.confl", nil, true},
		{"invalid event meta", "invalid_event_meta.confl", nil, true},
		{"invalid event key", "invalid_event_key.confl", nil, true},
		{
			"valid config",
			"valid.confl",
			&Config{
				Functions: map[string]*fn.Function{
					"Testing": &fn.Function{
						Name:    "Testing",
						Handler: "function",
					},
				},
				Events: []*fn.Event{
					&fn.Event{
						Source: fn.APISource,
						Target: "Testing",
						Meta:   map[string]string{"Route": "/Testing"},
					},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParsePath(path.Join("fixtures", tt.path))

			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
