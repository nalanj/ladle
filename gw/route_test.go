package gw

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/nalanj/ladle/config"
	"github.com/stretchr/testify/assert"
)

func TestRouteMatch(t *testing.T) {
	t.Parallel()

	req, reqErr := http.NewRequest(
		"POST",
		"https://testing.com:3030/test/function",
		bytes.NewReader([]byte("testBody")),
	)
	assert.Nil(t, reqErr)

	tests := []struct {
		name   string
		event  *config.Event
		match  bool
		params map[string]string
	}{
		{
			"matches a literal match",
			&config.Event{
				Source: config.APISource,
				Meta:   map[string]string{"Route": "/test/function"},
			},
			true,
			map[string]string{},
		},
		{
			"misses a literal miss",
			&config.Event{
				Source: config.APISource,
				Meta:   map[string]string{"Route": "/test/not-function"},
			},
			false,
			nil,
		},
		{
			"matches a path param",
			&config.Event{
				Source: config.APISource,
				Meta:   map[string]string{"Route": "/{api}/function"},
			},
			true,
			map[string]string{"api": "test"},
		},
		{
			"misses on route longer than path",
			&config.Event{
				Source: config.APISource,
				Meta:   map[string]string{"Route": "/{api}/function/testing"},
			},
			false,
			nil,
		},
		{
			"misses on path longer than route",
			&config.Event{
				Source: config.APISource,
				Meta:   map[string]string{"Route": "/{api}"},
			},
			false,
			nil,
		},
		{
			"misses on non API source",
			&config.Event{
				Source: "whatever",
				Meta:   map[string]string{"Route": "/test/function"},
			},
			false,
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			params, match := routeMatch(req, test.event)
			assert.Equal(t, test.match, match)
			assert.Equal(t, test.params, params)
		})
	}
}
