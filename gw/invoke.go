package gw

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/nalanj/ladle/config"
	"github.com/nalanj/ladle/fn"
)

// InvokeHandler returns a handler that can invoke called functions via http
func InvokeHandler(conf *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, pathParams := route(conf, r)
		if f == nil {
			log.Printf("HTTP: No function match for %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		invokeReq, prepareErr := prepareRequest(r, pathParams)
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

// prepareRequest converts an http.Request into an InvokeRequest
func prepareRequest(r *http.Request, pathParams map[string]string) (*messages.InvokeRequest, error) {
	body, bodyErr := ioutil.ReadAll(r.Body)
	if bodyErr != nil {
		return nil, bodyErr
	}

	gwR := events.APIGatewayProxyRequest{
		Resource:        "",
		Path:            r.URL.Path,
		PathParameters:  pathParams,
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
