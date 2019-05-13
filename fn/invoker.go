package fn

import "github.com/aws/aws-lambda-go/lambda/messages"

// An Invoker looks up a function and invokes it
type Invoker func(
	name string,
	req *messages.InvokeRequest,
	resp *messages.InvokeResponse,
) error
