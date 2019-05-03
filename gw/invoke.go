package gw

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/nalanj/ladle/config"
	"github.com/nalanj/ladle/fn"
)

// InvokeHandler returns a handler that can invoke called functions via http
func InvokeHandler(conf *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := route(conf, r)
		if f == nil {
			log.Printf("HTTP: No function match for %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		invokeReq, prepareErr := prepareRequest(r)
		if prepareErr != nil {
			log.Printf("HTTP: Error: %s\n", prepareErr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := &messages.InvokeResponse{}
		invokeErr := f.Invoke(invokeReq, resp)
		if invokeErr != nil {
			log.Printf("HTTP: Error: %s\n", invokeErr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if resp.Error != nil {
			log.Printf("HTTP: Error: %s\n", resp.Error.Message)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		writeErr := writeInvokeResponse(w, resp)
		if writeErr != nil {
			log.Printf("HTTP: Error: %s\n", writeErr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

// route converts a request to its corresponding function
func route(conf *config.Config, r *http.Request) *fn.Function {
	fnName := r.URL.Path[1:]

	for _, event := range conf.Events {
		if event.Source == fn.APISource && event.Target == fnName {
			f, ok := conf.Functions[event.Target]
			if ok {
				return f
			}
		}
	}

	return nil
}

// prepareRequest converts an http.Request into an InvokeRequest
func prepareRequest(r *http.Request) (*messages.InvokeRequest, error) {
	body, bodyErr := ioutil.ReadAll(r.Body)
	if bodyErr != nil {
		return nil, bodyErr
	}

	gwR := events.APIGatewayProxyRequest{
		Resource:        "",
		Path:            r.URL.Path,
		HTTPMethod:      r.Method,
		Headers:         make(map[string]string),
		Body:            string(body),
		IsBase64Encoded: false,
		RequestContext:  events.APIGatewayProxyRequestContext{},
	}

	payload, marshalErr := json.Marshal(gwR)
	if marshalErr != nil {
		return nil, marshalErr
	}

	return &messages.InvokeRequest{Payload: payload}, nil
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
