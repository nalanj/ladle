package gw

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/nalanj/ladle/fn"
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
		event  *fn.Event
		match  bool
		params map[string]string
	}{
		{
			"matches a literal match",
			&fn.Event{
				Source: fn.APISource,
				Meta:   map[string]string{"Route": "/test/function"},
			},
			true,
			map[string]string{},
		},
		{
			"misses a literal miss",
			&fn.Event{
				Source: fn.APISource,
				Meta:   map[string]string{"Route": "/test/not-function"},
			},
			false,
			nil,
		},
		{
			"matches a path param",
			&fn.Event{
				Source: fn.APISource,
				Meta:   map[string]string{"Route": "/{api}/function"},
			},
			true,
			map[string]string{"api": "test"},
		},
		{
			"misses on route longer than path",
			&fn.Event{
				Source: fn.APISource,
				Meta:   map[string]string{"Route": "/{api}/function/testing"},
			},
			false,
			nil,
		},
		{
			"misses on path longer than route",
			&fn.Event{
				Source: fn.APISource,
				Meta:   map[string]string{"Route": "/{api}"},
			},
			false,
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			params, match := routeMatch(req, test.event)
			assert.Equal(t, test.match, match)
			assert.Equal(t, test.params, params)
		})
	}
}
