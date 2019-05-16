package rpc

import (
	"errors"

	"github.com/aws/aws-lambda-go/lambda/messages"
)

var globalInvoker Invoker

// InvokeWrapper exposes Invoke for the given function
type InvokeWrapper struct {
	Name string
}

// Invoke invokes the given function
func (r *InvokeWrapper) Invoke(
	req *messages.InvokeRequest,
	resp *messages.InvokeResponse,
) error {
	if globalInvoker == nil {
		return errors.New("No global invoker")
	}

	return globalInvoker(r.Name, req, resp)
}
