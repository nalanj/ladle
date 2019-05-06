package gw

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/nalanj/ladle/config"
)

// InvokeHandler returns a handler that can invoke called functions via http
func InvokeHandler(conf *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		wr := newRequest(r)
		wr.log(fmt.Sprintf("Start %s", wr.r.URL.Path))
		invoke(conf, w, wr)

		wr.log(
			fmt.Sprintf(
				"Invoke (%.3fms)",
				float64(time.Now().Sub(startTime).Nanoseconds())/1000000,
			),
		)
	})
}

// invoke wraps http invocation and makes it easier to deal with logging
// of requests
func invoke(conf *config.Config, w http.ResponseWriter, r *wrappedRequest) {
	f, pathParams := route(conf, r.r)
	if f == nil {
		r.log("No matching route")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	invokeReq, prepareErr := r.prepareRequest(pathParams)
	if prepareErr != nil {
		r.errorLog(prepareErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := &messages.InvokeResponse{}
	invokeErr := f.Invoke(invokeReq, resp)
	if invokeErr != nil {
		r.errorLog(invokeErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if resp.Error != nil {
		r.log(fmt.Sprintf("Invocation Error: %s", resp.Error.Message))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeErr := writeInvokeResponse(w, resp)
	if writeErr != nil {
		r.errorLog(writeErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// writes an http response based on the given InvokeResponse
func writeInvokeResponse(w http.ResponseWriter, resp *messages.InvokeResponse) error {
	var gwResp events.APIGatewayProxyResponse
	if unmarshalErr := json.Unmarshal(resp.Payload, &gwResp); unmarshalErr != nil {
		return unmarshalErr
	}

	for key, val := range gwResp.Headers {
		w.Header().Add(key, val)
	}
	w.WriteHeader(gwResp.StatusCode)
	w.Write([]byte(gwResp.Body))

	return nil
}
