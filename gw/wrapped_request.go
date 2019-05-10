package gw

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/gofrs/uuid"
)

// wrappedRequest wraps an http request with a struct
type wrappedRequest struct {
	id string
	r  *http.Request
}

// newRequest initializes a new wrapped request
func newRequest(r *http.Request) *wrappedRequest {
	return &wrappedRequest{
		id: uuid.Must(uuid.NewV4()).String(),
		r:  r,
	}
}

// log logs a message relating to this request
func (r *wrappedRequest) log(msg string) {
	log.Printf("HTTP %s: %s", r.id, msg)
}

// errorLog writes an error log message for this request
func (r *wrappedRequest) errorLog(err error) {
	r.log(fmt.Sprintf("Error: %s", err))
}

// prepareRequest converts an http.Request into an InvokeRequest
func (r *wrappedRequest) prepareRequest(
	pathParams map[string]string,
) (*messages.InvokeRequest, error) {
	body, bodyErr := ioutil.ReadAll(r.r.Body)
	if bodyErr != nil {
		return nil, bodyErr
	}

	headers := make(map[string]string)
	for k, v := range r.r.Header {
		headers[k] = v[0]
	}

	gwR := events.APIGatewayProxyRequest{
		Path:              r.r.URL.Path,
		PathParameters:    pathParams,
		HTTPMethod:        r.r.Method,
		Headers:           headers,
		MultiValueHeaders: r.r.Header,
		Body:              string(body),
		IsBase64Encoded:   false,
		RequestContext: events.APIGatewayProxyRequestContext{
			RequestID: r.id,
		},
	}

	payload, marshalErr := json.Marshal(gwR)
	if marshalErr != nil {
		return nil, marshalErr
	}

	return &messages.InvokeRequest{RequestId: r.id, Payload: payload}, nil
}
