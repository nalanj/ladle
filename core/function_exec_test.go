package core

import (
	"testing"

	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/nalanj/ladle/config"
	"github.com/stretchr/testify/assert"
)

func TestFunctionExecStart(t *testing.T) {
	t.Parallel()

	echoPkg := "github.com/nalanj/ladle/lambdas/echo"

	t.Run("returns an error if the executable doesn't exist", func(t *testing.T) {
		t.Parallel()

		done := make(chan string, 20)
		f := &config.Function{Name: "Test", Package: "not-here"}
		_, err := StartFunction(&config.Config{}, f, done)
		assert.NotNil(t, err)
	})

	t.Run("starts the function if the handler is valid", func(t *testing.T) {
		t.Parallel()

		done := make(chan string, 20)
		f := &config.Function{Name: "Test", Package: echoPkg}
		fnEx, err := StartFunction(&config.Config{}, f, done)
		assert.Nil(t, err)

		assert.NotNil(t, fnEx.cmd)
		assert.NotNil(t, fnEx.cmd.Process)
		assert.Nil(t, StopFunction(fnEx))
	})
}

func TestFunctionInvoke(t *testing.T) {
	t.Parallel()

	done := make(chan string, 20)
	f := &config.Function{Name: "Test", Package: "../build/echo"}
	fnEx, err := StartFunction(&config.Config{}, f, done)
	assert.Nil(t, err)

	req := &messages.InvokeRequest{Payload: []byte("{}")}
	resp := &messages.InvokeResponse{}
	invokeErr := fnEx.Invoke(req, resp)

	assert.Nil(t, invokeErr)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Payload)
	assert.Nil(t, StopFunction(fnEx))
}
