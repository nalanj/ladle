package gw

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda/messages"

	"github.com/nalanj/ladle/config"
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
	functions := map[string]*config.Function{
		"Echo":        &config.Function{Name: "Echo", Package: "../build/echo"},
		"InvokeError": &config.Function{Name: "InvokeError", Package: "n/a"},
	}

	invoker := func(
		name string,
		req *messages.InvokeRequest,
		resp *messages.InvokeResponse,
	) error {
		if name == "Echo" {
			data, marshalErr := json.Marshal(
				&events.APIGatewayProxyResponse{
					Body:       "OK",
					StatusCode: http.StatusOK,
				},
			)
			assert.Nil(t, marshalErr)
			resp.Payload = data

			return nil
		}
		return errors.New("Invoke error")
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{
				Functions: functions,
				Events: []*config.Event{
					&config.Event{
						Source: config.APISource,
						Target: "Echo",
						Meta:   map[string]string{"Route": "/echo"},
					},
					&config.Event{
						Source: config.APISource,
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

			invoke(cfg, invoker, w, wr)

			assert.Equal(t, test.status, w.Code)
		})
	}
}

func TestWriteInvokeResponse(t *testing.T) {
	w := httptest.NewRecorder()
	gwResp := &events.APIGatewayProxyResponse{
		Body:       "test body",
		Headers:    map[string]string{"Cool-Header": "yes"},
		StatusCode: http.StatusCreated,
	}
	gwRespBytes, marshalErr := json.Marshal(gwResp)
	assert.Nil(t, marshalErr)

	writeErr := writeInvokeResponse(
		w,
		&messages.InvokeResponse{Payload: gwRespBytes},
	)
	assert.Nil(t, writeErr)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "test body", w.Body.String())
	assert.Equal(t, "yes", w.Header().Get("Cool-Header"))
}
