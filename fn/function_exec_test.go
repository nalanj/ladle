package fn

import (
	"testing"

	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/stretchr/testify/assert"
)

func TestFunctionExecStart(t *testing.T) {
	t.Parallel()

	t.Run("returns an error if the executable doesn't exist", func(t *testing.T) {
		t.Parallel()

		done := make(chan string, 20)
		f := &Function{Name: "Test", Handler: "not-here"}
		_, err := Start(f, done)
		assert.NotNil(t, err)
	})

	t.Run("starts the function if the handler is valid", func(t *testing.T) {
		t.Parallel()

		done := make(chan string, 20)
		f := &Function{Name: "Test", Handler: "../build/echo"}
		fnEx, err := Start(f, done)
		assert.Nil(t, err)

		assert.NotNil(t, fnEx.cmd)
		assert.NotNil(t, fnEx.cmd.Process)
		assert.Nil(t, Stop(fnEx))
	})
}

func TestFunctionInvoke(t *testing.T) {
	t.Parallel()

	done := make(chan string, 20)
	f := &Function{Name: "Test", Handler: "../build/echo"}
	fnEx, err := Start(f, done)
	assert.Nil(t, err)

	req := &messages.InvokeRequest{Payload: []byte("{}")}
	resp := &messages.InvokeResponse{}
	invokeErr := fnEx.Invoke(req, resp)

	assert.Nil(t, invokeErr)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Payload)
	assert.Nil(t, Stop(fnEx))
}
