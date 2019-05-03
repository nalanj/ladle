package main

// Echo is a lambda for simply echoing the request payload back to the response
// via http. It's helpful for debugging.

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	out, outErr := json.Marshal(req)
	if outErr != nil {
		return nil, outErr
	}

	return &events.APIGatewayProxyResponse{
		Body:       string(out),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
