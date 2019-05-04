package fn

import (
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/stretchr/testify/assert"
)

func TestFunctionStart(t *testing.T) {
	t.Run("returns an error if the executable doesn't exist", func(t *testing.T) {
		f := &Function{Name: "Test", Handler: "not-here"}
		err := f.Start()
		assert.NotNil(t, err)
	})

	t.Run("starts the function if the handler is valid", func(t *testing.T) {
		f := &Function{Name: "Test", Handler: "../build/echo"}
		err := f.Start()
		assert.Nil(t, err)

		assert.NotNil(t, f.cmd)
		assert.NotNil(t, f.cmd.Process)
		assert.Nil(t, f.cmd.Process.Kill())
	})
}

func TestFunctionInvoke(t *testing.T) {
	f := &Function{Name: "Test", Handler: "../build/echo"}
	err := f.Start()
	assert.Nil(t, err)

	time.Sleep(100 * time.Millisecond)

	req := &messages.InvokeRequest{Payload: []byte("{}")}
	resp := &messages.InvokeResponse{}
	invokeErr := f.Invoke(req, resp)

	assert.Nil(t, invokeErr)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Payload)
	assert.Nil(t, f.cmd.Process.Kill())
}
