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
	t.Parallel()

	req, reqErr := http.NewRequest(
		"POST",
		"https://testing.com:3030/test/function",
		bytes.NewReader([]byte("testBody")),
	)
	req.Header.Add("Rando-Header", "Value1")
	req.Header.Add("Rando-Header", "Value2")
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
	assert.Equal(t, "Value1", gwR.Headers["Rando-Header"])
	assert.Equal(
		t,
		[]string{"Value1", "Value2"},
		gwR.MultiValueHeaders["Rando-Header"],
	)
}
