package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	fmt.Println("WHATEVER")
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
