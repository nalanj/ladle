package gw

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nalanj/ladle/config"
	"github.com/nalanj/ladle/fn"
	"github.com/stretchr/testify/assert"
)

func TestInvoke(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		path   string
		status int
	}{
		{
			"returns not found when no route matches",
			"/not-found",
			http.StatusNotFound,
		},
		{
			"returns internal error on invoke error",
			"/invoke-error",
			http.StatusInternalServerError,
		},
		{
			"returns success on success",
			"/echo",
			http.StatusOK,
		},
	}

	// Test against two functions. Echo is will be running and InvokeError
	// will not be running
	functions := map[string]*fn.Function{
		"Echo":        &fn.Function{Name: "Echo", Handler: "../build/echo"},
		"InvokeError": &fn.Function{Name: "InvokeError", Handler: "n/a"},
	}
	startErr := fn.Start(functions["Echo"])
	assert.Nil(t, startErr)

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{
				Functions: functions,
				Events: []*fn.Event{
					&fn.Event{
						Source: fn.APISource,
						Target: "Echo",
						Meta:   map[string]string{"Route": "/echo"},
					},
					&fn.Event{
						Source: fn.APISource,
						Target: "InvokeError",
						Meta:   map[string]string{"Route": "/invoke-error"},
					},
				},
			}

			w := httptest.NewRecorder()

			req, reqErr := http.NewRequest(
				"POST",
				fmt.Sprintf("https://testing.com:3030%s", test.path),
				bytes.NewReader([]byte("testBody")),
			)
			assert.Nil(t, reqErr)
			wr := newRequest(req)

			invoke(cfg, w, wr)

			assert.Equal(t, test.status, w.Code)
		})
	}

	stopErr := fn.Stop(functions["Echo"])
	assert.Nil(t, stopErr)
}
