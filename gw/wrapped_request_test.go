package gw

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestPrepareRequest(t *testing.T) {
	req, reqErr := http.NewRequest(
		"POST",
		"https://testing.com:3030/test/function",
		bytes.NewReader([]byte("testBody")),
	)
	assert.Nil(t, reqErr)

	wr := newRequest(req)
	pathParams := map[string]string{"param": "payload"}
	ri, prepErr := wr.prepareRequest(pathParams)
	assert.Nil(t, prepErr)

	gwR := &events.APIGatewayProxyRequest{}
	unmarshalErr := json.Unmarshal(ri.Payload, gwR)
	assert.Nil(t, unmarshalErr)

	assert.Equal(t, "/test/function", gwR.Path)
	assert.Equal(t, pathParams, gwR.PathParameters)
	assert.Equal(t, "POST", gwR.HTTPMethod)
	assert.Equal(t, "testBody", gwR.Body)
	assert.Equal(t, wr.id, gwR.RequestContext.RequestID)
}
