package gw

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/gofrs/uuid"
	"github.com/nalanj/ladle/config"
	"github.com/nalanj/ladle/fn"
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
func (r *wrappedRequest) prepareRequest(pathParams map[string]string) (*messages.InvokeRequest, error) {
	body, bodyErr := ioutil.ReadAll(r.r.Body)
	if bodyErr != nil {
		return nil, bodyErr
	}

	gwR := events.APIGatewayProxyRequest{
		Resource:        "",
		Path:            r.r.URL.Path,
		PathParameters:  pathParams,
		HTTPMethod:      r.r.Method,
		Headers:         make(map[string]string),
		Body:            string(body),
		IsBase64Encoded: false,
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

// invoke wraps http invocation and makes it easier to deal with logging out
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
		r.log(fmt.Sprintf("Invocation Error: %s", resp.Error))
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

// route converts a request to its corresponding function
func route(conf *config.Config, r *http.Request) (*fn.Function, map[string]string) {
	for _, event := range conf.Events {
		pathParams, ok := routeMatch(r, event)
		if ok {
			return conf.Functions[event.Target], pathParams
		}
	}

	return nil, nil
}

// routeMatch tests if a route matches and returns path parts if it does
func routeMatch(r *http.Request, event *fn.Event) (map[string]string, bool) {
	if event.Source != fn.APISource {
		return nil, false
	}

	reqParts := strings.Split(r.URL.Path, "/")
	if reqParts[0] == "" {
		reqParts = reqParts[1:]
	}

	routeParts := strings.Split(event.Meta["Route"], "/")
	if routeParts[0] == "" {
		routeParts = routeParts[1:]
	}

	pathParams := make(map[string]string)
	for i := 0; i < len(routeParts); i++ {
		if len(reqParts) <= i {
			return nil, false
		}

		reqPart := reqParts[i]
		routePart := routeParts[i]

		if strings.HasPrefix(routePart, "{") && strings.HasSuffix(routePart, "}") {
			pathParams[routePart[1:len(routePart)-1]] = reqPart
		} else {
			if routePart != reqPart {
				return nil, false
			}
		}
	}

	return pathParams, true
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
